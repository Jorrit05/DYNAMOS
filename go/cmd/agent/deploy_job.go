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

func generateChainAndDeploy(ctx context.Context, compositionRequest *pb.CompositionRequest, localJobName string, sqlDataRequest *pb.SqlDataRequest) error {
	logger.Debug("Starting generateChainAndDeploy")

	ctx, span := trace.StartSpan(ctx, serviceName+"/func: generateChainAndDeploy")
	defer span.End()

	// TODO: Parse SQL request for extra compute services

	msChain, err := generateMicroserviceChain(compositionRequest)
	if err != nil {
		logger.Sugar().Errorf("Error generating microservice chain %v", err)
		return err
	}

	err = deployJob(ctx, msChain, localJobName)
	if err != nil {
		logger.Sugar().Errorf("Error generating microservice chain %v", err)
		return err
	}

	logger.Sugar().Infow("Deployed job.", "actualJobName", localJobName, "msChain", msChain)
	return nil
}

func deployJob(ctx context.Context, msChain []mschain.MicroserviceMetadata, jobName string) error {
	logger.Debug("Starting deployJob")

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

	dataStewardName := strings.ToLower(serviceName)
	if dataStewardName == "" {
		return fmt.Errorf("env variable DATA_STEWARD_NAME not defined")
	}

	// Define the job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: dataStewardName,
			Labels:    map[string]string{"app": dataStewardName, "jobName": jobName},
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds:   &activeDeadlineSeconds,
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
	nrOfServices := len(msChain)
	firstService := "1"
	lastService := "0"
	for i, microservice := range msChain {
		order := i
		port = port + order

		if i == nrOfServices-1 {
			lastService = "1"
		}

		logger.Sugar().Debugw("job info:", "name: ", microservice.Name, "Port: ", port)

		container := v1.Container{
			Name:            microservice.Name,
			Image:           fmt.Sprintf("%s:latest", microservice.Name),
			ImagePullPolicy: v1.PullIfNotPresent,
			Env: []v1.EnvVar{
				{Name: "DESIGNATED_GRPC_PORT", Value: strconv.Itoa(port)},
				{Name: "FIRST", Value: firstService},
				{Name: "LAST", Value: lastService},
				{Name: "JOB_NAME", Value: jobName},
				{Name: "SIDECAR_PORT", Value: strconv.Itoa(firstPortMicroservice - 1)},
				{Name: "OC_AGENT_HOST", Value: tracingHost},
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

func addSidecar() v1.Container {
	name := "sidecar"

	return v1.Container{
		Name:            name,
		Image:           fmt.Sprintf("%s:latest", name),
		ImagePullPolicy: v1.PullIfNotPresent,
		Env: []v1.EnvVar{
			{Name: "DESIGNATED_GRPC_PORT", Value: strconv.Itoa(firstPortMicroservice - 1)},
			{Name: "TEMPORARY_JOB", Value: "true"},
			{Name: "AMQ_USER", Value: rabbitMqUser},
			{Name: "OC_AGENT_HOST", Value: tracingHost},
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

func getRequiredMicroservices(microserviceMetada *[]mschain.MicroserviceMetadata, request *mschain.RequestType, role string) error {

	for _, ms := range request.RequiredServices {
		var metadataObject mschain.MicroserviceMetadata

		_, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/microservices/%s/chainMetadata", ms), &metadataObject)
		if err != nil {
			return err
		}

		if strings.EqualFold(metadataObject.Label, role) {
			*microserviceMetada = append(*microserviceMetada, metadataObject)
		} else if strings.EqualFold("all", role) {
			// Only append dataProvider microservices
			*microserviceMetada = append(*microserviceMetada, metadataObject)
		}
	}

	return nil
}

func getOptionalMicroservices(microserviceMetada *[]mschain.MicroserviceMetadata, request *mschain.RequestType, role string) error {
	//TODO: figure out a way to include enforced microservices

	return nil
}
func RequestTypeMicroservices(requestType string) (mschain.RequestType, error) {
	var request mschain.RequestType
	_, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/requestTypes/%s", requestType), &request)
	if err != nil {
		return mschain.RequestType{}, err
	}

	return request, nil
}
