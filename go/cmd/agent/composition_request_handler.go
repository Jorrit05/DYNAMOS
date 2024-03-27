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

					// delete the jobs compositionrequest here
					queueName := lib.LastPartAfterSlash(string(event.Kv.Key))
					queueInfoMutex.Lock()
					queueInfo, ok := queueInfoMap[queueName]
					if ok {
						key := fmt.Sprintf("%s/%s/%s/%s", etcdJobRootKey, agentConfig.Name, queueInfo.UserName, queueInfo.JobName)

						_, err := etcdClient.Delete(ctx, key)
						if err != nil {
							logger.Sugar().Warnf("error deleting key from etcd: %v", err)
						}
					} else {
						logger.Sugar().Warnf("Can't find queueInfo for this expired key")
					}
					delete(queueInfoMap, queueName)
					queueInfoMutex.Unlock()

					c.DeleteQueue(ctx, &pb.QueueInfo{QueueName: queueName, AutoDelete: false})
				}
			}
		}
	}()
}

func compositionRequestHandler(ctx context.Context, compositionRequest *pb.CompositionRequest) context.Context {
	// Register user and job
	// Create queue (this might be up for refactoring)

	ctx, span := trace.StartSpan(ctx, serviceName+"/func: startCompositionRequest")
	defer span.End()

	var err error
	localJobname := ""
	value, ok := jobCounter[compositionRequest.JobName]
	if !ok {
		logger.Sugar().Debug("NOK")

		localJobname, err = generateJobName(compositionRequest.JobName)
		if err != nil {
			logger.Sugar().Errorf("generateJobName err: %v", err)
		}
	} else {
		logger.Sugar().Debugf("value: %v", value)
		localJobname = compositionRequest.JobName + strings.ToLower(serviceName) + strconv.Itoa(value)
	}

	compositionRequest.LocalJobName = localJobname
	ctx, err = registerUserWithJob(ctx, compositionRequest)
	if err != nil {
		logger.Sugar().Errorf("Error in registering Job %v", err)
		return ctx
	}

	ctx = handleQueue(ctx, compositionRequest.JobName, localJobname, compositionRequest.User.UserName)

	return ctx
}

func handleQueue(ctx context.Context, jobName string, localJobname string, userName string) context.Context {
	queueInfo := &pb.QueueInfo{}
	queueInfo.AutoDelete = false
	queueInfo.QueueName = localJobname
	queueInfo.JobName = jobName
	queueInfo.UserName = userName

	key := fmt.Sprintf("/agents/jobs/%s/queueInfo/%s", serviceName, localJobname)
	err := etcd.PutEtcdWithGrant(ctx, etcdClient, key, localJobname, queueDeleteAfter)
	if err != nil {
		logger.Sugar().Errorf("Error PutEtcdWithGrant: %v", err)
	}
	queueInfoMutex.Lock()
	queueInfoMap[localJobname] = queueInfo
	queueInfoMutex.Unlock()
	watchQueue(ctx, key)
	c.CreateQueue(ctx, queueInfo)
	return ctx
}

func generateMicroserviceChain(compositionRequest *pb.CompositionRequest, options map[string]bool) ([]mschain.MicroserviceMetadata, error) {
	logger.Sugar().Debugf("Starting generateMicroserviceChain")

	var requestType mschain.RequestType
	_, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/requestTypes/%s", compositionRequest.RequestType), &requestType)
	if err != nil {
		return nil, err
	}
	logger.Sugar().Debugf("After GetAndUnmarshalJSON  compositionRequest.RequestType)")

	var msMetadata []mschain.MicroserviceMetadata

	// Returns required Microservices
	err = getRequiredMicroservices(&msMetadata, &requestType, compositionRequest.Role)
	if err != nil {
		return nil, err
	}

	err = getOptionalMicroservices(&msMetadata, &requestType, compositionRequest.Role, compositionRequest.RequestType, options)
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
