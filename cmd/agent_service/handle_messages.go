package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	"github.com/docker/docker/client"

	amqp "github.com/rabbitmq/amqp091-go"
)

func createDockerServices(cli *client.Client, msData lib.MicroServiceData) {
	log.Info("createDockerServices length msData: ", fmt.Sprint(len(msData.Services)))
	// Get some values from etcd
	for _, microservice := range msData.Services {
		serviceSpec := lib.CreateServiceSpec(
			microservice.Image,
			microservice.Tag,
			microservice.EnvVars,
			microservice.Networks,
			microservice.NetworkList,
			microservice.Secrets,
			microservice.Volumes,
			microservice.Ports,
			cli,
		)
		lib.CreateDockerService(cli, serviceSpec)

		jsonData, err := json.Marshal(microservice)
		if err != nil {
			log.Warn("Error marshaling payload to JSON:", err)
		}

		_, err = etcdClient.Put(context.Background(), fmt.Sprintf("%s/%s", msEtcdPath, microservice.Image), string(jsonData))
		if err != nil {
			log.Fatalf("Failed creating an item in etcd: %s", err)
		}
	}
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

func getTargetNetwork() map[string]lib.Network {
	// TODO: How to handle this if there are more networks?
	// for now just return the one that is not 'core'
	// Also this assumes the agent networks are always a map (which makes sense)
	for k, v := range agentConfig.AgentDetails.Networks {

		if !strings.Contains(k, "core") {
			return map[string]lib.Network{k: v}
		}
	}

	log.Error("No viable network config found for this agent.")
	return nil
}

func createServiceFromArchetype(archetype lib.ArcheType) error {
	var microServiceData lib.MicroServiceData
	microServiceData.Services = make(map[string]lib.MicroService) // Initialize the map

	startQueue = hostname + "_start_queue"
	queue, err := lib.DeclareQueue(startQueue, channel, true)
	if err != nil {
		log.Errorf("Unable to declare queue, err: %v", err)
		return err
	}

	if err := channel.QueueBind(
		queue.Name,         // name
		routingKey+"start", // key
		"topic_exchange",   // exchange
		false,              // noWait
		nil,                // args
	); err != nil {
		log.Fatalf("Queue Bind: %s", err)
		return err
	}

	for msName, inputService := range archetype.IoConfig.ServiceIO {

		var microservice lib.MicroService
		_, err := lib.GetAndUnmarshalJSON(etcdClient, "/microservices/"+msName, &microservice)
		if err != nil {
			log.Errorf("Error unmarshalling requested Microservice %s, error: %v", msName, err)
			return err
		}
		log.Infof("current MS name: %s", microservice.Image)
		if inputService == "start" {
			inputService = startQueue
		}
		microservice.EnvVars["INPUT_QUEUE"] = inputService
		microservice.Networks = getTargetNetwork()
		microservice.NetworkList = nil

		microServiceData.Services[msName] = microservice
	}

	createDockerServices(dockerClient, microServiceData)
	return nil
}

func startMessageLoop(
	messages <-chan amqp.Delivery,
	exchangeName string,
) {

	for msg := range messages {
		if exchangeName == "" {
			exchangeName = "topic_exchange"
		}

		msg.Type = strings.ToLower(msg.Type)

		switch msg.Type {
		case "datarequest":
		case "createarchitecture":
			log.Info("Create architecture message received")
			var requestor lib.Requestor
			var archetype lib.ArcheType

			err := json.Unmarshal(msg.Body, &requestor)
			if err != nil {
				log.Errorf("Error unmarshaling JSON:", err)
			}

			log.Info("requestor name: %s", requestor.Name)
			log.Info("requestor CurrentArchetype: %s", requestor.CurrentArchetype)

			archJson, err := lib.GetAndUnmarshalJSON(etcdClient, "/reasoner/archetype_config/"+requestor.CurrentArchetype, &archetype)
			if err != nil {
				log.Errorf("Error unmarshaling archetype:", err)
			}
			log.Println("Json dump: %s", string(archJson))

			createServiceFromArchetype(archetype)

		case "CreateService":
			var payload lib.MicroServiceData
			err := json.Unmarshal(msg.Body, &payload)
			if err != nil {
				log.Printf("Error decoding CreateServicePayload: %v", err)
				return
			}
			createDockerServices(dockerClient, payload)
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
			log.Printf("Unknown message type: %s", msg.Type)
		}

		// ... (Acknowledge the message)
	}
}
