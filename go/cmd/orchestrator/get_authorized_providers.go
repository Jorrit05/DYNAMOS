package main

import (
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func getAuthorizedProviders(validationResponse *pb.ValidationResponse) (map[string]lib.AgentDetails, error) {
	authorizedProviders := make(map[string]lib.AgentDetails)
	// logger.Sugar().Debugf("ValidDataproviders: %v", validationResponse.ValidDataproviders)
	// logger.Sugar().Debugf("InvalidDataproviders: %v", validationResponse.InvalidDataproviders)

	for key := range validationResponse.ValidDataproviders {
		logger.Sugar().Debugf("key: %s", key)
		var agentData lib.AgentDetails
		json, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/agents/online/%s", key), &agentData)
		if err != nil {
			logger.Sugar().Warnf("error getAuthorizedProviders: %v", err)
			return nil, err
		} else if json == nil {
			logger.Sugar().Warnf("no JSON in getAuthorizedProviders, key: %v", key)
			// invalidProviders = append(invalidProviders, key)
			continue
		}
		authorizedProviders[key] = agentData
		logger.Sugar().Debugf("Added agent: %s", agentData.Name)
	}

	return authorizedProviders, nil
}
