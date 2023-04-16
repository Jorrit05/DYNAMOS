package main

import "github.com/Jorrit05/micro-recomposer/pkg/lib"

func createRabbitMq() {
	// Example service specification
	image := "rabbitmq"
	version := "3-management"

	envVars := map[string]string{
		"RABBITMQ_ERLANG_COOKIE": "mysecretcookie",
		"RABBITMQ_DEFAULT_USER":  "guest",
		"RABBITMQ_DEFAULT_PASS":  "guest",
		"RABBITMQ_LOGS":          "-",
	}

	networks := []string{
		"core_network",
	}

	secrets := []string{}

	volumes := map[string]string{
		"~/.docker-conf/rabbitmq/data/": "/var/lib/rabbitmq/",
		"~/.docker-conf/rabbitmq/log/":  "/var/log/rabbitmq",
		"./rabbitmq.conf":               "/etc/rabbitmq/conf.d/11-custom.conf",
		"./definitions.json":            "/opt/definitions.json",
	}

	ports := map[string]string{
		"5672":  "5672",
		"15672": "15672",
	}

	spec := lib.CreateServiceSpec(image, version, envVars, networks, secrets, volumes, ports, dockerClient)
	lib.CreateDockerService(dockerClient, spec)
}
