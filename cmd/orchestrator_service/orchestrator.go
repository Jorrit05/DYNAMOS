package main

import (
	"encoding/json"
	"fmt"

	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Jorrit05/micro-recomposer/pkg/lib"
	"github.com/docker/docker/client"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	serviceName string = "orchestrator_service"
	routingKey  string

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
	processedServices, err := lib.SetMicroservicesEtcd(&lib.EtcdClientWrapper{Client: etcdClient}, "/var/log/stack-files/config/microservices_config.yml", "")
	if err != nil {
		log.Fatalf("Error setting microservices in etcd: %v", err)
	}

	for serviceName, _ := range processedServices {
		log.Infof("serviceName added to etcd, %s", serviceName)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatalf("Error starting HTTP server: %s", err)
		}
	}()

	select {}
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

		if len(agentData) == 0 {
			w.Write([]byte("No providers of that name are currently available"))
		} else if len(agentData) != len(orchestratorRequest.Providers) {
			var agentList []string
			for k := range agentData {
				agentList = append(agentList, k)
			}

			// Filter out which providers aren't online currently
			nonExistingProviders := strings.Join(lib.SliceDifferenceString(orchestratorRequest.Providers, agentList), ",")
			w.Write([]byte(fmt.Sprintf("Providers %s, currently not available. Other requests, if any, are accepted.", nonExistingProviders)))

			return
		} else {
			w.Write([]byte("Request accepted, check output queue"))
			return
		}

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
