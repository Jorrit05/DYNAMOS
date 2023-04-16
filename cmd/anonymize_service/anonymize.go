package main

import (
	"encoding/json"
	"sync"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	log, logFile        = lib.InitLogger(serviceName)
	serviceName  string = "anonymize_service"
	routingKey   string = lib.GetDefaultRoutingKey(serviceName)
)

func main() {
	defer logFile.Close()
	defer lib.HandlePanicAndFlushLogs(log, logFile)

	// Define a WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	// Connect to AMQ queue, declare own routingKey as queue, start listening for messages
	messages, conn, channel, err := lib.SetupConnection(serviceName, routingKey, true)
	if err != nil {
		log.Fatalf("Failed to setup proper connection to RabbitMQ: %v", err)
	}
	defer conn.Close()
	log.Printf("Anonymize:  %s", routingKey)

	go func() {
		lib.StartMessageLoop(anonymize, messages, channel, routingKey, "")
		wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()

	// Wait for both goroutines to finish
	wg.Wait()
}

type SkillQuery struct {
	PersonId  int    `json:"person_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Sex       string `json:"sex"`
	// DriversLicense string `json:"drivers_license"`
	// Programming    string `json:"programming"`
}

func anonymize(message amqp.Delivery) (amqp.Publishing, error) {
	var skillQueries []SkillQuery

	err := json.Unmarshal(message.Body, &skillQueries)
	if err != nil {
		log.Printf("Error unmarshaling JSON:", err)
		return amqp.Publishing{}, err
	}

	// Anonymise last name
	for i := range skillQueries {
		skillQueries[i].LastName = "anonymized"
	}

	jsonMessage, err := json.Marshal(skillQueries)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return amqp.Publishing{}, err
	}

	return amqp.Publishing{
		Body:          jsonMessage,
		Type:          "application/json",
		CorrelationId: message.CorrelationId,
	}, nil
}
