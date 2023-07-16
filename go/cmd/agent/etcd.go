package main

import (
	"fmt"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func getJobName(user string) (string, error) {
	// /agents/jobs/UVA/jorrit.stutterheim@cloudnation.nl -> jorrit-3141334
	jobNameKey := fmt.Sprintf("%s/%s/%s", etcdJobRootKey, agentConfig.Name, user)
	jobName, err := etcd.GetValueFromEtcd(etcdClient, jobNameKey)
	if err != nil {
		logger.Sugar().Errorf("etcd error: %v", err.Error())
		return "", err
	}
	return jobName, nil
}

func getCompositionRequest(jobName string) (*pb.CompositionRequest, error) {
	var compositionRequest *pb.CompositionRequest
	_, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("%s/%s/%s", etcdJobRootKey, agentConfig.Name, jobName), &compositionRequest)
	if err != nil {
		logger.Sugar().Errorf("Error getting composition request: %v", err)
		return nil, err
	}
	return compositionRequest, nil
}

func registerUserWithJob(compositionRequest *pb.CompositionRequest) error {
	logger.Debug("Entering registerUserWithJob")

	// /agents/jobs/UVA/jorrit-3141334 ->  pb.CompositionRequest
	jobNameKey := fmt.Sprintf("%s/%s/%s", etcdJobRootKey, agentConfig.Name, compositionRequest.JobName)
	// /agents/jobs/UVA/jorrit.stutterheim@cloudnation.nl -> jorrit-3141334
	userKey := fmt.Sprintf("%s/%s/%s", etcdJobRootKey, agentConfig.Name, compositionRequest.User.UserName)

	// One entry with all related info with the jobName as key
	err := etcd.SaveStructToEtcd[*pb.CompositionRequest](etcdClient, jobNameKey, compositionRequest)
	if err != nil {
		logger.Sugar().Warnf("Error saving struct to etcd: %v", err)
		return err
	}

	// One entry with the jobName with the userName as key
	err = etcd.PutValueToEtcd(etcdClient, userKey, compositionRequest.JobName, etcd.WithMaxElapsedTime(time.Second*10))
	if err != nil {
		logger.Sugar().Warnf("Error saving jobname to etcd: %v", err)
		return err
	}
	return nil
}
