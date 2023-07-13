package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func startCompositionRequest(validationResponse *pb.ValidationResponse, authorizedProviders map[string]lib.AgentDetails, c pb.SideCarClient) (map[string]string, error) {
	logger.Debug("Entering startCompositionRequest")
	archetype, err := chooseArchetype(validationResponse)
	if err != nil {
		return nil, err
	}
	logger.Debug("ARCHETYPE: " + archetype)

	var archetypeConfig api.Archetype
	_, err = etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/archetypes/%s", archetype), &archetypeConfig)
	if err != nil {
		return nil, err
	}
	logger.Debug("1: ")
	compositionRequest := &pb.CompositionRequest{}
	compositionRequest.User = &pb.User{}
	compositionRequest.User = validationResponse.User
	compositionRequest.DataProviders = []string{}
	compositionRequest.ArchetypeId = archetype
	compositionRequest.RequestType = validationResponse.RequestType
	compositionRequest.JobName = lib.GeneratePodNameWithGUID(validationResponse.User.UserName, 8)

	// Use to return the proper endpoints to the user
	userTargets := make(map[string]string)

	if archetypeConfig.ComputeProvider != "other" {
		// Compute to data
		compositionRequest.Role = "all"
		for key := range authorizedProviders {
			compositionRequest.DestinationQueue = authorizedProviders[key].RoutingKey
			c.SendCompositionRequest(context.Background(), compositionRequest)
			userTargets[key] = authorizedProviders[key].Dns
		}
	} else {
		// TODO: Here I am assuming that the initial archetype choice is correct
		// and that this is the only possible archetype.
		// I should probably build in that if there is no TTP online or available
		// Or Universities have different TTPs. That these scenarios are handled as well.
		ttp, err := chooseThirdParty(validationResponse)
		if err != nil {
			return nil, err
		}
		logger.Debug("2: ")
		// Send to each validData provider the role data provider
		// Send to the thirdParty the role Compute provider
		compositionRequest.Role = "dataProvider"
		tmpDataProvider := []string{}
		for key := range authorizedProviders {
			tmpDataProvider = append(tmpDataProvider, key)
			compositionRequest.DestinationQueue = authorizedProviders[key].RoutingKey
			c.SendCompositionRequest(context.Background(), compositionRequest)
		}
		logger.Debug("3: ")
		compositionRequest.DataProviders = tmpDataProvider
		compositionRequest.Role = "computeProvider"
		compositionRequest.DestinationQueue = ttp.RoutingKey
		userTargets[ttp.Name] = ttp.Dns
		c.SendCompositionRequest(context.Background(), compositionRequest)
	}
	// var request RequestType
	// _, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/requestTypes/%s", validationResponse.RequestType), &request)
	// if err != nil {
	// 	return err
	// }

	// var msMetadata []MicroserviceMetada
	// // Returns required Microservices
	// err = getRequiredMicroservices(&msMetadata, &request)
	// if err != nil {
	// 	return err
	// }

	// msChain, err := generateChain([]string{}, msMetadata)
	// if err != nil {
	// 	return err
	// }
	// for _, ms := range msChain {
	// 	logger.Info(ms.Name)
	// }

	// //TODO: aanpassen
	// compositionRequest := &pb.CompositionRequest{}
	// compositionRequest.User = &pb.User{}
	// compositionRequest.User = validationResponse.User
	// compositionRequest.DataProvider = ""
	// compositionRequest.ArchetypeId = chooseArchetype(validationResponse)
	// compositionRequest.RequestType = "sqlDataRequest"
	// compositionRequest.Target = "UVA-in"
	// for _, chain := range msChain {
	// 	compositionRequest.Microservices = append(compositionRequest.Microservices, chain.Name)
	// }

	// c.SendCompositionRequest(context.Background(), compositionRequest)
	logger.Debug("4: ")
	return userTargets, nil
}

// Just returns one of the entries that match. No logic behind it.
// TODO: Make smarter
func chooseArchetype(validationResponse *pb.ValidationResponse) (string, error) {
	intersection := make(map[string]bool)

	first := true
	for _, dataProvider := range validationResponse.ValidDataproviders {
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
	json, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/agents/%s", intersection[0]), &agentData)
	if err != nil {
		return lib.AgentDetails{}, err
	} else if json == nil {
		return lib.AgentDetails{}, fmt.Errorf("compute provider not online")
	}

	return agentData, nil
}
