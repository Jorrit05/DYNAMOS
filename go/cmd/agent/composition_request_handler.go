package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/mschain"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
)

func generateJobName(jobName string) (string, error) {
	if serviceName == "" {
		return "", fmt.Errorf("env variable DATA_STEWARD_NAME not defined")
	}

	dataStewardName := strings.ToLower(serviceName)
	logger.Sugar().Debugw("Pod info:", "dataStewardName: ", dataStewardName, "jobName: ", jobName)

	// Get the jobname of this user
	jobMutex.Lock()
	jobCounter[jobName]++
	newValue := jobCounter[jobName]
	jobMutex.Unlock()

	return jobName + dataStewardName + strconv.Itoa(newValue), nil
}

func compositionRequestHandler(ctx context.Context, compositionRequest *pb.CompositionRequest) {
	// get local requiredServices
	// Generate microservice chain
	// Spin up pod
	// Save session information in etcd

	ctx, span := trace.StartSpan(ctx, serviceName+"/func: startCompositionRequest")
	defer span.End()

	err := registerUserWithJob(compositionRequest)
	if err != nil {
		logger.Sugar().Errorf("Error in registering Job %v", err)
		return
	}

	localJobname, err := generateJobName(compositionRequest.JobName)
	if err != nil {
		logger.Sugar().Errorf("generateJobName err: %v", err)
	}

	actualJobMutex.Lock()
	actualJobMap[compositionRequest.JobName] = localJobname
	actualJobMutex.Unlock()
	// Create queue for this job
	// I think the actualJobname should be stored in etcd as well on a timer.
	// think on this in combination with the actualJobMap.. The user should be able to send multiple
	// sqlDAtaReqeuest, so the initial queue should not be autoDeleted, but deleted when the timer on
	// this active job expires
	queueInfo := &pb.QueueInfo{}
	queueInfo.AutoDelete = true
	queueInfo.QueueName = localJobname

	c.CreateQueue(ctx, queueInfo)

	if strings.EqualFold(compositionRequest.Role, "dataProvider") {
		err := generateChainAndDeploy(ctx, compositionRequest, localJobname, &pb.SqlDataRequest{})
		if err != nil {
			logger.Sugar().Errorf("Error in deploying job: %v", err)
			return
		}
		logger.Sugar().Warnf("jobName: %v", compositionRequest.JobName)
		logger.Sugar().Warnf("actualJobName: %v", localJobname)
		waitingJobMutex.Lock()
		waitingJobMap[compositionRequest.JobName] = localJobname
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
