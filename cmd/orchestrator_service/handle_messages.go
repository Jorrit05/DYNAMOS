package main

import (
	"encoding/json"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/docker/docker/client"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handleCreateService(cli *client.Client, payload lib.MicroServiceData) {
	fmt.Println("Handling Create Service")

	// for _, microservice := range payload {
	// 	serviceSpec := lib.CreateServiceSpec(
	// 		microservice.ImageName,
	// 		microservice.Tag,
	// 		microservice.EnvVars,
	// 		microservice.Networks,
	// 		microservice.Secrets,
	// 		microservice.Volumes,
	// 		microservice.Ports,
	// 		cli,
	// 	)
	// 	lib.CreateDockerService(cli, serviceSpec)
	// }
}

func handleDetachService(payload lib.DetachAttachServicePayload) {
	fmt.Println("Handling Detach Service")
	// Detach the service from the queue
}

func handleAttachService(payload lib.DetachAttachServicePayload) {
	fmt.Println("Handling Attach Service")
	// Attach the service to the queue
}

func handleKillService(payload lib.KillServicePayload) {
	fmt.Println("Handling Kill Service")
	// Kill the service
}

func startMessageLoop(
	messages <-chan amqp.Delivery,
	exchangeName string,
) {

	for msg := range messages {
		if exchangeName == "" {
			exchangeName = "topic_exchange"
		}

		switch msg.Type {
		case "CreateService":
			var payload lib.MicroServiceData
			err := json.Unmarshal(msg.Body, &payload)
			if err != nil {
				logger.Sugar().Infof("Error decoding CreateServicePayload: %v", err)
				return
			}
			handleCreateService(dockerClient, payload)
		case "DetachService":
			// Handle DetachService
			// ...

		case "AttachService":
			// Handle AttachService
			// ...

		case "KillService":
			// Handle KillService
			// ...

		default:
			logger.Sugar().Infof("Unknown message type: %s", msg.Type)
		}

		// ... (Acknowledge the message)
	}
}
