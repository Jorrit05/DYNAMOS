package main

import (
	"context"
	"fmt"

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

	archetype, err := chooseArchetype(validationResponse.ValidDataproviders)
	if err != nil {
		return nil, ctx, err
	}

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

// Just returns one of the entries that match. No logic behind it.
// TODO: Make smarter
func chooseArchetype(validDataproviders map[string]*pb.DataProvider) (string, error) {
	intersection := make(map[string]bool)

	first := true
	for _, dataProvider := range validDataproviders {
		if first {
			for _, archType := range dataProvider.Archetypes {
				intersection[archType] = true
			}
			first = false
		} else {
			newIntersection := make(map[string]bool)
			for _, archType := range dataProvider.Archetypes {
				if intersection[archType] {
					newIntersection[archType] = true
				}
			}
			intersection = newIntersection
		}
	}

	if len(intersection) == 0 {
		return "", fmt.Errorf("no common archetypes found")
	}

	// return the first common archetype
	for key := range intersection {
		return key, nil
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
