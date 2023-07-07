package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/docker/docker/client"
	"github.com/rabbitmq/amqp091-go"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	logger                      = lib.InitLogger()
	dockerClient *client.Client = lib.GetDockerClient()
	routingKey   string
	channel      *amqp091.Channel
	conn         *amqp091.Connection
	startQueue   string
	// externalRoutingKey  string
	// externalServiceName string
	etcdClient  *clientv3.Client = lib.GetEtcdClient(etcdEndpoints)
	agentConfig lib.AgentDetails
	msEtcdPath  string = fmt.Sprintf("/%s/services", hostname)
)

func main() {
	defer logger.Sync() // flushes buffer, if any
	defer etcdClient.Close()

	// Because there will be several agents running in this test setup add (and register) a guid for uniqueness
	routingKey = lib.GetDefaultRoutingKey(serviceName)
	// externalRoutingKey = fmt.Sprintf("%s-input", routingKey)
	// externalServiceName = fmt.Sprintf("%s-input", serviceName)

	// Define a WaitGroup
	var wg sync.WaitGroup
	wg.Add(2)

	var err error
	var messages <-chan amqp091.Delivery
	// Connect to AMQ queue, declare own routingKey as queue
	messages, conn, channel, err = lib.SetupConnection(serviceName, routingKey, true, true)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// The agent will need an external queue to receive requests from the orchestrator
	// messages := registerExternalQueue(channel)

	// Register all basic info of this agents environment
	registerAgent()

	// Start listening for messages, this method keeps this method 'alive'
	go func() {

		startMessageLoop(messages, "")
		wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()

	// Wait for both goroutines to finish
	wg.Wait()
}

func registerAgent() {
	// Prepare agent configuration data
	var service lib.MicroServiceData = lib.UnmarshalStackFile("/var/log/stack-files/agents.yaml")

	agentConfig = lib.AgentDetails{
		Name: hostname,
		// RoutingKeyOutput: routingKey,
		ServiceName:  serviceName,
		RoutingKey:   routingKey,
		AgentDetails: service.Services[hostname],
	}

	// Serialize agent configuration data as JSON
	configData, err := json.Marshal(agentConfig)
	if err != nil {
		log.Fatal(err)
	}

	go lib.CreateEtcdLeaseObject(etcdClient, fmt.Sprintf("/agents/%s", agentConfig.Name), string(configData))
}

func updateAgent() {

	// Update the ActiveSince field
	now := time.Now()
	agentConfig.ConfigUpdated = &now

	// Serialize agent configuration data as JSON
	configData, err := json.Marshal(agentConfig)
	if err != nil {
		log.Fatal(err)
	}

	go lib.CreateEtcdLeaseObject(etcdClient, fmt.Sprintf("/agents/%s", agentConfig.Name), string(configData))

}

// func registerExternalQueue(queueName, string, channel *amqp.Channel) <-chan amqp.Delivery {

// 	queue, err := lib.DeclareQueue(externalServiceName, channel, true)
// 	if err != nil {
// 		log.Fatalf("Failed to declare queue: %v", err)
// 	}

// 	// Bind queue to "topic_exchange"
// 	// TODO: Make "topic_exchange" flexible?
// 	if err := channel.QueueBind(
// 		queue.Name,         // name
// 		externalRoutingKey, // key
// 		"topic_exchange",   // exchange
// 		false,              // noWait
// 		nil,                // args
// 	); err != nil {
// 		log.Fatalf("Queue Bind: %s", err)
// 	}

// 	// Start listening to queue
// 	messages, err := lib.Consume(externalServiceName, channel)
// 	if err != nil {
// 		log.Fatalf("Failed to register consumer: %v", err)

// 	} else {
// 		log.Printf("Registered consumer: %s", externalServiceName)
// 	}

// 	return messages
// }
