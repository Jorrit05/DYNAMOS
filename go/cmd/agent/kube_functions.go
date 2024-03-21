package main

import (
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubeConfig() *rest.Config {
	var config *rest.Config
	var err error

	if local {
		// Use out-of-cluster configuration
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			logger.Sugar().Fatalf("failed to build config from Flags: %v", err)
		}
	} else {
		// Use in-cluster configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			logger.Sugar().Fatalf("failed to build config inClluster: %v", err)
		}
	}

	return config
}

func getKubeClient() *kubernetes.Clientset {
	config := getKubeConfig()

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create kube client: %v", err)
	}

	return clientSet
}
