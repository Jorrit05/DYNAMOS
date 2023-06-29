package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
)

func deployJob(compositionRequest *pb.CompositionRequest) error {
	logger.Debug("Starting deployJob")
	// Create context
	ctx := context.TODO()

	// Use the current context in kubeconfig
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Sugar().Errorf("failed to build config: %v", err)
		return err
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Sugar().Errorf("failed to create clientset: %v", err)
		return err
	}

	dataStewardName := strings.ToLower(os.Getenv("DATA_STEWARD_NAME"))

	// Define the job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      compositionRequest.User.UserName,
			Namespace: dataStewardName,
			Labels:    map[string]string{"app": dataStewardName},
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
	for i, name := range compositionRequest.Microservices {
		order := i + 1
		port = port + order
		lastService := "0"
		if i == nrOfServices-1 {
			// This will make the port nr > 60000
			// The last service both realizes it's last and know
			// which port number to set up for the service before it (-100000)
			lastService = "1"
		}

		container := v1.Container{
			Name:            name,
			Image:           fmt.Sprintf("%s:latest", name),
			ImagePullPolicy: v1.PullIfNotPresent,
			Env:             []v1.EnvVar{{Name: "ORDER", Value: strconv.Itoa(port)}, {Name: "LAST", Value: lastService}},
			// Add additional container configuration here as needed
		}
		job.Spec.Template.Spec.Containers = append(job.Spec.Template.Spec.Containers, container)
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
		Env:             []v1.EnvVar{{Name: "ORDER", Value: strconv.Itoa(50052)}, {Name: "AMQ_PASSWORD", Value: os.Getenv("AMQ_PASSWORD")}, {Name: "AMQ_USER", Value: os.Getenv("AMQ_USER")}},
		// Add additional container configuration here as needed
	}
}
