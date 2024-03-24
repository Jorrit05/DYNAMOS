package main

import (
	"context"
	"fmt"
	"slices"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
)

type UnauthorizedProviderError struct {
	ProviderName string
}

func (e *UnauthorizedProviderError) Error() string {
	return fmt.Sprintf("third party '%s' is not online", e.ProviderName)
}

func startCompositionRequest(ctx context.Context, validationResponse *pb.ValidationResponse, authorizedProviders map[string]lib.AgentDetails, compositionRequest *pb.CompositionRequest) (map[string]string, context.Context, error) {
	logger.Debug("Entering startCompositionRequest")

	ctx, span := trace.StartSpan(ctx, "startCompositionRequest")
	defer span.End()

	archetype, err := chooseArchetype(validationResponse, authorizedProviders)
	if err != nil {
		return nil, ctx, err
	}
	logger.Sugar().Infof("Chosen archetype: %s", archetype)

	var archetypeConfig api.Archetype
	_, err = etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/archetypes/%s", archetype), &archetypeConfig)
	if err != nil {
		return nil, ctx, err
	}

	compositionRequest.User = validationResponse.User
	compositionRequest.DataProviders = []string{}
	compositionRequest.ArchetypeId = archetype
	compositionRequest.RequestType = validationResponse.RequestType
	compositionRequest.JobName = lib.GenerateJobName(validationResponse.User.UserName, 8)

	// Save the ActiveJob to etcd
	// var activeJob = &pb.ActiveJob{
	// 	JobId:               compositionRequest.JobName,
	// 	Type:                validationResponse.RequestType,
	// 	User:                validationResponse.User,
	// 	Archetype:           archetype,
	// 	AuthorizedProviders: make(map[string]string),
	// }
	// for name, agent := range authorizedProviders {
	// 	activeJob.AuthorizedProviders[name] = agent.Dns
	// }

	// etcd.SaveStructToEtcd(etcdClient, fmt.Sprintf("/activeJobs/%s/%s", validationResponse.User.Id, compositionRequest.JobName), activeJob)

	// Use to return the proper endpoints to the user
	userTargets := make(map[string]string)

	if archetypeConfig.ComputeProvider != "other" {
		// Compute to data
		compositionRequest.Role = "all"
		for key := range authorizedProviders {
			compositionRequest.DestinationQueue = authorizedProviders[key].RoutingKey
			c.SendCompositionRequest(ctx, compositionRequest)
			userTargets[key] = authorizedProviders[key].Dns
		}
	} else {
		// TODO: Here I am assuming that the initial archetype choice is correct
		// and that this is the only possible archetype.
		// I should probably build in that if there is no TTP online or available
		// Or Universities have different TTPs. That these scenarios are handled as well.
		ttp, err := chooseThirdParty(validationResponse)
		if err != nil {
			return nil, ctx, err
		}

		// Send to each validData provider the role data provider
		// Send to the thirdParty the role Compute provider
		compositionRequest.Role = "dataProvider"
		tmpDataProvider := []string{}
		for key := range authorizedProviders {
			tmpDataProvider = append(tmpDataProvider, key)
			compositionRequest.DestinationQueue = authorizedProviders[key].RoutingKey
			c.SendCompositionRequest(ctx, compositionRequest)
		}

		compositionRequest.DataProviders = tmpDataProvider
		compositionRequest.Role = "computeProvider"
		compositionRequest.DestinationQueue = ttp.RoutingKey
		userTargets[ttp.Name] = ttp.Dns

		c.SendCompositionRequest(ctx, compositionRequest)
	}
	return userTargets, ctx, nil
}

// Simple sorting to return the archetype with the least weight
func pickArchetypeBasedOnWeight() (*api.Archetype, error) {
	logger.Sugar().Info("Start pickArchetypeBasedOnWeight")

	target := &api.Archetype{}

	// Assuming GetPrefixListEtcd returns a slice of Archetype and an error
	archeTypes, err := etcd.GetPrefixListEtcd(etcdClient, "/archetypes", target)
	if err != nil {
		return nil, err
	}

	if len(archeTypes) == 0 {
		return nil, fmt.Errorf("no archetypes available")
	}

	lightest := archeTypes[0]

	// Iterate to find the one with the lowest weight
	for _, archeType := range archeTypes {
		if archeType.Weight < lightest.Weight {
			lightest = archeType
		}
	}

	return lightest, nil
}

func getArchetypeBasedOnOptions(validationResponse *pb.ValidationResponse, authorizedDataProviders map[string]lib.AgentDetails) string {
	logger.Sugar().Debugf("Start getArchetypeBasedOnOptions, options: %v", validationResponse.Options)

	// This ranges over the options. And selects an archetype based on the options.
	for option, value := range validationResponse.Options {
		switch option {
		case "aggregate":
			// If aggregate is enabled, it will select the 'dataThroughTtp' archetype, if this is allowed on all the authorizedDataProviders
			if value {
				allowed := true
				for provider, _ := range authorizedDataProviders {
					if !slices.Contains(validationResponse.ValidArchetypes.Archetypes[provider].Archetypes, "dataThroughTtp") {
						logger.Sugar().Debugf("allowed false, slice: %v", validationResponse.ValidArchetypes.Archetypes[provider].Archetypes)

						allowed = false
					}
				}

				if allowed {
					return "dataThroughTtp"
				}
			}
		}
	}
	return ""
}

// TODO: Make smarter
func chooseArchetype(validationResponse *pb.ValidationResponse, authorizedDataProviders map[string]lib.AgentDetails) (string, error) {
	logger.Sugar().Debug("starting chooseArchetype")
	logger.Sugar().Debugf("length options: %v", len(validationResponse.Options))

	for k, _ := range validationResponse.ValidDataproviders {
		logger.Sugar().Debug("validDataprovider: %s ", k)
	}

	if validationResponse.Options != nil && len(validationResponse.Options) > 0 {
		archetype := getArchetypeBasedOnOptions(validationResponse, authorizedDataProviders)
		if archetype != "" {
			return archetype, nil
		}
	}

	//Below is messy, up for refactoring and making it smarter
	archeType, err := pickArchetypeBasedOnWeight()
	if err != nil {
		return "", err
	}
	allowed := true
	for provider := range authorizedDataProviders {
		if !slices.Contains(validationResponse.ValidArchetypes.Archetypes[provider].Archetypes, archeType.Name) {
			allowed = false
		}
	}
	if allowed {
		return archeType.Name, nil
	}

	for provider := range authorizedDataProviders {
		someArchetype := validationResponse.ValidArchetypes.Archetypes[provider].Archetypes[0]
		if someArchetype != "" {
			return someArchetype, nil
		}
	}

	return "", fmt.Errorf("unexpected error: could not retrieve an archetype from the intersection")
}

func chooseThirdParty(validationResponse *pb.ValidationResponse) (lib.AgentDetails, error) {

	intersectionMap := make(map[string]int)
	totalProviders := len(validationResponse.ValidDataproviders)

	// Iterate over all valid dataproviders
	for _, dataProvider := range validationResponse.ValidDataproviders {
		// For each compute provider in a valid dataprovider
		for _, computeProvider := range dataProvider.ComputeProviders {
			intersectionMap[computeProvider]++
		}
	}

	// Extract the intersection from the map
	var intersection []string
	for provider, count := range intersectionMap {
		if count == totalProviders {
			intersection = append(intersection, provider)
		}
	}

	// If the intersection is empty, return an error
	if len(intersection) == 0 {
		return lib.AgentDetails{}, fmt.Errorf("no common compute providers found")
	}

	// If the intersection is not empty, return the first item
	var agentData lib.AgentDetails

	json, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/agents/online/%s", intersection[0]), &agentData)
	if err != nil {
		return lib.AgentDetails{}, err
	} else if json == nil {
		return lib.AgentDetails{}, &UnauthorizedProviderError{ProviderName: intersection[0]}
	}

	return agentData, nil
}
