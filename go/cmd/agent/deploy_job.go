package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeConfig() (*rest.Config, error) {
	var config *rest.Config
	var err error

	if local {
		// Use out-of-cluster configuration
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logger.Sugar().Errorf("failed to build config: %v", err)
			return nil, err
		}
	} else {
		// Use in-cluster configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			logger.Sugar().Errorf("failed to build config: %v", err)
			return nil, err
		}
	}

	return config, nil
}

func deployJob(compositionRequest *pb.CompositionRequest) error {
	logger.Debug("Starting deployJob")
	// Create context
	ctx := context.TODO()

	config, err := getKubeConfig()
	if err != nil {
		return err
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Sugar().Errorf("failed to create clientset: %v", err)
		return err
	}

	dataStewardName := strings.ToLower(os.Getenv("DATA_STEWARD_NAME"))
	jobName := lib.GeneratePodNameWithGUID(compositionRequest.User.UserName, 8)
	logger.Sugar().Debugw("Pod info:", "dataStewardName: ", dataStewardName, "jobName: ", jobName)

	go registerUserWithJob(compositionRequest, jobName)

	// Define the job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: dataStewardName,
			Labels:    map[string]string{"app": dataStewardName, "jobName": jobName},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl, // Clean up job TTL after it finishes
			BackoffLimit:            &backoffLimit,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": dataStewardName},
				},
				Spec: v1.PodSpec{
					Containers:    []v1.Container{},
					RestartPolicy: v1.RestartPolicyOnFailure,
				},
			},
		},
	}

	// Add the containers to the job
	port := firstPortMicroservice
	nrOfServices := len(compositionRequest.Microservices)
	firstService := "1"
	lastService := "0"
	for i, name := range compositionRequest.Microservices {
		order := i
		port = port + order

		if i == nrOfServices-1 {
			lastService = "1"
		}

		logger.Sugar().Debugw("job info:", "name: ", name, "Port: ", port)

		container := v1.Container{
			Name:            name,
			Image:           fmt.Sprintf("%s:latest", name),
			ImagePullPolicy: v1.PullIfNotPresent,
			Env: []v1.EnvVar{
				{Name: "DESIGNATED_GRPC_PORT", Value: strconv.Itoa(port)},
				{Name: "FIRST", Value: firstService},
				{Name: "LAST", Value: lastService},
				{Name: "JOB_NAME", Value: jobName},
				{Name: "SIDECAR_PORT", Value: strconv.Itoa(firstPortMicroservice - 1)},
			},
			// Add additional container configuration here as needed
		}
		job.Spec.Template.Spec.Containers = append(job.Spec.Template.Spec.Containers, container)
		firstService = "0"
	}

	job.Spec.Template.Spec.Containers = append(job.Spec.Template.Spec.Containers, addSidecar())
	// Create the job
	_, err = clientset.BatchV1().Jobs(dataStewardName).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		logger.Sugar().Errorf("failed to create job: %v", err)
		return err
	}

	return nil
}

func registerUserWithJob(compositionRequest *pb.CompositionRequest, jobName string) {
	logger.Debug("Entering registerUserWithJob")
	var userData lib.JobUserInfo
	userData.ArcheType = compositionRequest.ArchetypeId
	userData.JobName = jobName
	userData.Name = compositionRequest.User.UserName
	userData.RequestType = compositionRequest.RequestType

	jobNameKey := fmt.Sprintf("/activeJobs/%s", jobName)
	userKey := fmt.Sprintf("/activeJobs/%s", compositionRequest.User.UserName)

	// One entry with all related info with the jobName as key
	err := etcd.SaveStructToEtcd[lib.JobUserInfo](etcdClient, jobNameKey, userData)
	if err != nil {
		logger.Sugar().Warnf("Error saving struct to etcd: %v", err)
	}

	// One entry with the jobName with the userName as key
	err = etcd.PutValueToEtcd(etcdClient, userKey, jobName, etcd.WithMaxElapsedTime(time.Second*10))
	if err != nil {
		logger.Sugar().Warnf("Error saving jobname to etcd: %v", err)
	}
}

func addSidecar() v1.Container {
	name := "sidecar"

	return v1.Container{
		Name:            name,
		Image:           fmt.Sprintf("%s:latest", name),
		ImagePullPolicy: v1.PullIfNotPresent,
		Env: []v1.EnvVar{
			{Name: "DESIGNATED_GRPC_PORT", Value: strconv.Itoa(firstPortMicroservice - 1)},
			{Name: "AMQ_USER", Value: rabbitMqUser},
			{Name: "AMQ_PASSWORD",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "rabbit",
						},
						Key: "password",
					},
				},
			}},
		// Add additional container configuration here as needed
	}
}
