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
	archetype := chooseArchetype(validationResponse)

	var archetypeConfig api.Archetype
	_, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/archetypes/%s", archetype), &archetypeConfig)
	if err != nil {
		return nil, err
	}

	compositionRequest := &pb.CompositionRequest{}
	compositionRequest.User = &pb.User{}
	compositionRequest.User = validationResponse.User
	compositionRequest.DataProviders = []string{}
	compositionRequest.ArchetypeId = archetype
	compositionRequest.RequestType = validationResponse.RequestType

	userTargets := make(map[string]string)

	// TODO: Here I am assuming that the initial archetype choice is correct
	// and that this is the only possible archetype.
	// I should probably build in that if there is no TTP online or available
	// Or Universities have different TTPs. That these scenarios are handled as well.
	if archetypeConfig.ComputeProvider != "other" {
		compositionRequest.Role = "computeProvider"
		for key := range authorizedProviders {
			compositionRequest.Target = authorizedProviders[key].RoutingKey
			c.SendCompositionRequest(context.Background(), compositionRequest)
			userTargets[key] = authorizedProviders[key].Dns
		}
	} else {
		ttp, err := chooseThirdParty(validationResponse)
		if err != nil {
			return nil, err
		}
		// Send to each validData provider the role data provider
		// Send to the thirdParty the role Compute provider
		compositionRequest.Role = "dataProvider"

		for key := range authorizedProviders {
			compositionRequest.DataProviders = append(compositionRequest.DataProviders, key)
			compositionRequest.Target = authorizedProviders[key].RoutingKey
			c.SendCompositionRequest(context.Background(), compositionRequest)
		}

		compositionRequest.Role = "computeProvider"
		compositionRequest.Target = ttp.RoutingKey
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

	return userTargets, nil
}

func chooseArchetype(validationResponse *pb.ValidationResponse) string {
	return "computeToData"
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
