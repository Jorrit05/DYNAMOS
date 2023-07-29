package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/Jorrit05/DYNAMOS/pkg/mschain"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	clientv3 "go.etcd.io/etcd/client/v3"
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

func watchQueue(ctx context.Context, key string) {
	watchChan := etcdClient.Watch(ctx, key)
	go func() {
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				// event.Type will be either PUT or DELETE
				logger.Sugar().Infof("Event received! Type: %s Key:%s Value:%s\n", event.Type, event.Kv.Key, event.Kv.Value)

				// Take action if key is deleted
				if event.Type == clientv3.EventTypeDelete {
					logger.Sugar().Infof("Key has been deleted! Taking action...")
					//TODO probably should figure out a way to also delete the jobs compositionrequest here (/agents/jobs/UVA/jorrit.stutterheim@cloudnation.nl/jorrit-stutterheim-7e2d9c4c)
					// userKey := fmt.Sprintf("%s/%s/%s/%s", etcdJobRootKey, agentConfig.Name, compositionRequest.User.UserName, compositionRequest.JobName)

					c.DeleteQueue(ctx, &pb.QueueInfo{QueueName: lib.LastPartAfterSlash(string(event.Kv.Key)), AutoDelete: false})
				}
			}
		}
	}()
}

func compositionRequestHandler(ctx context.Context, compositionRequest *pb.CompositionRequest) context.Context {
	// get local requiredServices
	// Generate microservice chain
	// Spin up pod
	// Save session information in etcd

	ctx, span := trace.StartSpan(ctx, serviceName+"/func: startCompositionRequest")
	defer span.End()

	localJobname, err := generateJobName(compositionRequest.JobName)
	if err != nil {
		logger.Sugar().Errorf("generateJobName err: %v", err)
	}

	compositionRequest.LocalJobName = localJobname
	err = registerUserWithJob(ctx, compositionRequest)
	if err != nil {
		logger.Sugar().Errorf("Error in registering Job %v", err)
		return ctx
	}

	queueInfo := &pb.QueueInfo{}
	queueInfo.AutoDelete = false
	queueInfo.QueueName = localJobname

	key := fmt.Sprintf("/agents/jobs/%s/queueInfo/%s", serviceName, localJobname)
	err = etcd.PutEtcdWithGrant(ctx, etcdClient, key, localJobname, queueDeleteAfter)
	if err != nil {
		logger.Sugar().Errorf("Error PutEtcdWithGrant: %v", err)
	}
	watchQueue(ctx, key)
	c.CreateQueue(ctx, queueInfo)

	if strings.EqualFold(compositionRequest.Role, "dataProvider") {
		ctx, err = generateChainAndDeploy(ctx, compositionRequest, localJobname, &pb.SqlDataRequest{})
		if err != nil {
			logger.Sugar().Errorf("Error in deploying job: %v", err)
			return ctx
		}
		logger.Sugar().Warnf("jobName: %v", compositionRequest.JobName)
		logger.Sugar().Warnf("actualJobName: %v", localJobname)
		waitingJobMutex.Lock()
		waitingJobMap[compositionRequest.JobName] = localJobname
		waitingJobMutex.Unlock()
	}
	return ctx
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
