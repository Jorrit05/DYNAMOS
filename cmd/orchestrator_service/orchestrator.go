package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	serviceName  string = "orchestrator_service"
	routingKey   string
	channel      *amqp.Channel
	conn         *amqp.Connection
	log, logFile                = lib.InitLogger(serviceName)
	dockerClient *client.Client = lib.GetDockerClient()

	externalRoutingKey  string
	externalServiceName string
	etcdClient          *clientv3.Client = lib.GetEtcdClient()
	agentConfig         lib.AgentData
)

func main() {
	defer logFile.Close()
	defer lib.HandlePanicAndFlushLogs(log, logFile)

	routingKey = lib.GetDefaultRoutingKey(serviceName)

	// Register a yaml file of available microservices in etcd.
	_, err := lib.SetMicroservicesEtcd(&lib.EtcdClientWrapper{Client: etcdClient}, "/var/log/stack-files/config/microservices_config.yaml", "")
	if err != nil {
		log.Fatalf("Error setting microservices in etcd: %v", err)
	}

	// Define a WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	// Connect to AMQ queue, declare own routingKey as queue, start listening for messages
	_, conn, channel, err = lib.SetupConnection(serviceName, routingKey, false)
	if err != nil {
		log.Fatalf("Failed to setup proper connection to RabbitMQ: %v", err)
	}
	defer conn.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatalf("Error starting HTTP server: %s", err)
		}
		wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()

	wg.Wait()
}

func handler(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("handler: Error reading body: %v", err)
		http.Error(w, "handler: Error reading request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var orchestratorRequest lib.OrchestratorRequest
	err = json.Unmarshal([]byte(body), &orchestratorRequest)
	if err != nil {
		log.Printf("Error unmarshalling: %v", err)
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}
	orchestratorRequest.Type = strings.ToLower(orchestratorRequest.Type)

	switch orchestratorRequest.Type {
	case "datarequest":
		agentData, err := lib.GetAndUnmarshalJSONMap[lib.AgentDetails](etcdClient, "/agents/")
		if err != nil {
			log.Errorf("Getting available agents: %v", err)
			http.Error(w, "Internal error getting available agents", http.StatusInternalServerError)
			return
		}

		var resp []byte
		var agentList []string

		for k := range agentData {
			agentList = append(agentList, k)
		}

		// Filter out which providers aren't online currently
		matched, _ := lib.SliceIntersectAndDifference(orchestratorRequest.Providers, agentList)

		if len(matched) == 0 {
			resp = []byte("No providers of that name are currently available")
			// } else if len(notMatched) > 0 {

		} else {

			var requestorConfig lib.Requestor
			_, err = lib.GetAndUnmarshalJSON(etcdClient, "/reasoner/requestor_config/"+orchestratorRequest.Name, &requestorConfig)
			if err != nil {
				log.Errorf("Error getting requestor configuration", err)
				http.Error(w, "Internal error getting requestor data", http.StatusInternalServerError)
				return
			}

			matches, _ := lib.SliceIntersectAndDifference(matched, requestorConfig.AllowedPartners)
			log.Infof(" matches: %s", strings.Join(matches, ","))
			requestID := uuid.New().String()

			if len(orchestratorRequest.Architecture.ServiceIO) != 0 {
				// TODO: Handle new architecture
			} else {

				var requestor lib.Requestor
				requestorJSON, err := lib.GetAndUnmarshalJSON(etcdClient, "/reasoner/requestor_config/"+orchestratorRequest.Name, &requestor)
				if err != nil {
					log.Errorf("Error getting requestor configuration", err)
					http.Error(w, "Internal error getting requestor data", http.StatusInternalServerError)
					return
				}

				// Send architecture to agents. Who decide whether it is already created or being created or not.
				for _, v := range matches {
					mes := amqp.Publishing{
						Body:          requestorJSON,
						Type:          "createArchitecture",
						CorrelationId: requestID,
					}

					lib.Publish(channel, agentData[v].RoutingKey, mes, "")
				}

			}
			resp = []byte(fmt.Sprintf("Requests for %s accepted, check output queue.", strings.Join(matched, ",")))
			// resp = []byte("Request accepted, check output queue")
		}

		w.Write(resp)
		return

	case "architecture":

	default:
		log.Printf("Unknown message type: %s", orchestratorRequest.Type)
		http.Error(w, "Unknown request", http.StatusNotFound)
		return
	}
}

// func handler(w http.ResponseWriter, req *http.Request) {
// 	body, err := ioutil.ReadAll(req.Body)
// 	if err != nil {
// 		log.Printf("handler: Error reading body: %v", err)
// 		http.Error(w, "handler: Error reading request body", http.StatusBadRequest)
// 		return
// 	}
// 	defer req.Body.Close()

// 	// Generate a unique identifier for the request
// 	requestID := uuid.New().String()

// 	// Create a channel to receive the response
// 	responseChan := make(chan amqp.Delivery)

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
// 	defer cancel()

// 	// Store the request information in the map
// 	mutex.Lock()
// 	requestMap[requestID] = &requestInfo{id: requestID, response: responseChan}
// 	mutex.Unlock()

// 	// Send the message to the start queue
// 	convertedAmqMessage := amqp.Publishing{
// 		// DeliveryMode: amqp.Persistent,
// 		Timestamp:     time.Now(),
// 		ContentType:   "application/json",
// 		CorrelationId: requestID,
// 		Body:          body,
// 		// Headers:       amqp.Table{"context": json.Marshal()},
// 	}
// 	log.Printf("handler: 3, %s", routingKey)

// 	if err := lib.Publish(outputChannel, routingKey, convertedAmqMessage, ""); err != nil {
// 		log.Printf("Handler 4: Error publishing: %s", err)
// 	}

// 	// Wait for the response from the response channel
// 	select {
// 	case msg := <-responseChan:
// 		log.Printf("handler: 5, msg received: %s", msg.Body)
// 		w.Write(msg.Body)
// 	case <-ctx.Done():
// 		log.Println("handler: 6, context timed out")
// 		http.Error(w, "handler: Request timed out", http.StatusRequestTimeout)
// 	}
// }
