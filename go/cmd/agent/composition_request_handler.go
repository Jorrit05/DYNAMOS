package main

import (
	"fmt"
	"strings"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/mschain"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func compositionRequestHandler(compositionRequest *pb.CompositionRequest) {
	// get local requiredServices
	// Generate microservice chain
	// Spin up pod
	// Save session information in etcd
	//
	logger.Debug("-----")
	logger.Sugar().Debugf("%v", compositionRequest)
	logger.Debug("-----")

	err := registerUserWithJob(compositionRequest)
	if err != nil {
		logger.Sugar().Errorf("Error in registering Job %v", err)
		return
	}

	if strings.EqualFold(compositionRequest.Role, "dataProvider") {
		actualJobName, err := generateChainAndDeploy(compositionRequest, &pb.SqlDataRequest{})
		if err != nil {
			logger.Sugar().Errorf("Error in deploying job: %v", err)
			return
		}
		logger.Sugar().Warnf("jobName: %v", compositionRequest.JobName)
		logger.Sugar().Warnf("actualJobName: %v", actualJobName)
		waitingJobMutex.Lock()
		waitingJobMap[compositionRequest.JobName] = actualJobName
		waitingJobMutex.Unlock()
	}
}

func generateMicroserviceChain(compositionRequest *pb.CompositionRequest) ([]mschain.MicroserviceMetadata, error) {
	var requestType mschain.RequestType
	_, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/requestTypes/%s", compositionRequest.RequestType), &requestType)
	if err != nil {
		return nil, err
	}

	var msMetadata []mschain.MicroserviceMetadata

	// Returns required Microservices
	err = getRequiredMicroservices(&msMetadata, &requestType, compositionRequest.Role)
	if err != nil {
		return nil, err
	}

	err = getOptionalMicroservices(&msMetadata, &requestType, compositionRequest.Role)
	if err != nil {
		return nil, err
	}

	msChain, err := mschain.GenerateChain(msMetadata)
	if err != nil {
		return nil, err
	}
	for _, ms := range msChain {
		logger.Info(ms.Name)
	}

	return msChain, nil
}
