package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	outputChannel *amqp.Channel
	requestMap    = make(map[string]*requestInfo)
	mutex         = &sync.Mutex{}
)

type requestInfo struct {
	id       string
	response chan amqp.Delivery
}

func main() {

	var initialRequest lib.OrchestratorRequest
	jsonData, err := lib.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal([]byte(jsonData), &initialRequest)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	go func() {
		if err := http.ListenAndServe(":3001", mux); err != nil {
			log.Fatalf("Error starting HTTP server: %s", err)
		}
	}()

	// Start a separate go routine to handle reply messages
	go startReplyLoop()

	select {}
}

// Asynchoronous handler function.
// Create a channel for a response, publishes message to the gateway_queue
func handler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("handler: Error reading body: %v", err)
		http.Error(w, "handler: Error reading request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	// Generate a unique identifier for the request
	requestID := uuid.New().String()

	// Create a channel to receive the response
	responseChan := make(chan amqp.Delivery)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Store the request information in the map
	mutex.Lock()
	requestMap[requestID] = &requestInfo{id: requestID, response: responseChan}
	mutex.Unlock()

	// Send the message to the start queue
	convertedAmqMessage := amqp.Publishing{
		// DeliveryMode: amqp.Persistent,
		Timestamp:     time.Now(),
		ContentType:   "application/json",
		CorrelationId: requestID,
		Body:          body,
		// Headers:       amqp.Table{"context": json.Marshal()},
	}

	if err := lib.Publish(outputChannel, "routingKey", convertedAmqMessage, ""); err != nil {
		log.Printf("Handler 4: Error publishing: %s", err)
	}

	// Wait for the response from the response channel
	select {
	case msg := <-responseChan:
		log.Printf("handler: 5, msg received: %s", msg.Body)
		w.Write(msg.Body)
	case <-ctx.Done():
		log.Println("handler: 6, context timed out")
		http.Error(w, "handler: Request timed out", http.StatusRequestTimeout)
	}
}

// Consume messages from the last microservice (environment var INPUT_QUEUE). Send the output back to
// the http handler to return to the requestor.
func startReplyLoop() {
	// Start consuming from environment var INPUT_QUEUE
	messages, conn := lib.StartNewConsumer()
	defer conn.Close()

	for msg := range messages {
		fmt.Println(string(msg.Body))
	}
}

func Publish(ctx context.Context, chann *amqp.Channel, routingKey string, message amqp.Publishing, exchangeName string) error {
	if exchangeName == "" {
		exchangeName = "topic_exchange"
	}
	log.Printf("Publish: 1 %s", routingKey)

	err := chann.PublishWithContext(ctx, exchangeName, routingKey, false, false, message)
	if err != nil {
		log.Printf("Publish: 2 %s", err)
		return err
	}
	log.Println("Publish: 3")

	return nil
}
