package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"net/http"
	"strings"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
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
	defer logger.Sync() // flushes buffer, if any

	routingKey = lib.GetDefaultRoutingKey(serviceName)

	// Register a yaml file of available microservices in etcd.
	_, err := lib.SetMicroservicesEtcd(&lib.EtcdClientWrapper{Client: etcdClient}, "/var/log/stack-files/config/microservices_config.yaml", "")
	if err != nil {
		logger.Sugar().Fatalw("Error setting microservices in etcd: %v", err)
	}

	// Define a WaitGroup
	var wg sync.WaitGroup
	wg.Add(1)

	// Connect to AMQ queue, declare own routingKey as queue, start listening for messages
	_, conn, channel, err = lib.SetupConnection(serviceName, routingKey, false)
	if err != nil {
		logger.Sugar().Fatalw("Failed to setup proper connection to RabbitMQ: %v", err)
	}
	defer conn.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	go func() {
		if err := http.ListenAndServe(":8080", mux); err != nil {
			logger.Sugar().Fatalw("Error starting HTTP server: %s", err)
		}
		wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()

	wg.Wait()
}

func handler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Sugar().Infof("handler: Error reading body: %v", err)
		http.Error(w, "handler: Error reading request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var orchestratorRequest lib.OrchestratorRequest
	err = json.Unmarshal([]byte(body), &orchestratorRequest)
	if err != nil {
		logger.Sugar().Infof("Error unmarshalling: %v", err)
		http.Error(w, "Error parsing request", http.StatusBadRequest)
		return
	}
	orchestratorRequest.Type = strings.ToLower(orchestratorRequest.Type)

	switch orchestratorRequest.Type {
	case "datarequest":
		handleDataRequest(w, orchestratorRequest, etcdClient, channel)
		return

	case "architecture":

	default:
		logger.Sugar().Infof("Unknown message type: %s", orchestratorRequest.Type)
		http.Error(w, "Unknown request", http.StatusNotFound)
		return
	}
}

func handleDataRequest(w http.ResponseWriter, orchestratorRequest lib.OrchestratorRequest, etcdClient *clientv3.Client, channel *amqp.Channel) {

	// Get available agents
	agentData, err := lib.GetAndUnmarshalJSONMap[lib.AgentDetails](etcdClient, "/agents/")
	if err != nil {
		logger.Sugar().Errorw("Getting available agents: %v", err)
		http.Error(w, "Internal error getting available agents", http.StatusInternalServerError)
		return
	}

	// Check if requested organizationts are online
	matched, err := getMatchedProviders(orchestratorRequest.Providers, agentData)
	if err != nil {
		logger.Sugar().Errorw("Error getting matched providers: %v", err)
		http.Error(w, "Internal error getting matched providers", http.StatusInternalServerError)
		return
	}

	if len(matched) == 0 {
		w.Write([]byte("No providers of that name are currently available"))
		return
	}

	// Check if the requesting user has access to all these organizations
	var requestorConfig lib.Requestor
	matches, requestorJSON, err := getAllowedProviders(orchestratorRequest.Name, matched, &requestorConfig)
	if err != nil {
		logger.Sugar().Errorw("Error getting requestor configuration", err)
		http.Error(w, "Internal error getting requestor data", http.StatusInternalServerError)
		return
	}

	// _, err = lib.GetAndUnmarshalJSON(etcdClient, "/reasoner/requestor_config/"+orchestratorRequest.Name, &requestorConfig)
	// if err != nil {
	// 	logger.Sugar().Errorw("Error getting requestor configuration", err)
	// 	http.Error(w, "Internal error getting requestor data", http.StatusInternalServerError)
	// 	return
	// }

	// matches, _ := lib.SliceIntersectAndDifference(matched, requestorConfig.AllowedPartners)
	// log.Infof(" matches: %s", strings.Join(matches, ","))
	requestID := uuid.New().String()

	if len(orchestratorRequest.Architecture.ServiceIO) != 0 {
		// TODO: Handle new architecture
	} else {

		// var requestor lib.Requestor
		// requestorJSON, err := lib.GetAndUnmarshalJSON(etcdClient, "/reasoner/requestor_config/"+orchestratorRequest.Name, &requestor)
		// if err != nil {
		// 	logger.Sugar().Errorw("Error getting requestor configuration", err)
		// 	http.Error(w, "Internal error getting requestor data", http.StatusInternalServerError)
		// 	return
		// }

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
	resp := []byte(fmt.Sprintf("Requests for %s accepted, check output queue.", strings.Join(matched, ",")))

	w.Write(resp)
}

func getMatchedProviders(requestProviders []string, agentData map[string]lib.AgentDetails) ([]string, error) {
	var agentList []string

	for k := range agentData {
		agentList = append(agentList, k)
	}

	matched, _ := lib.SliceIntersectAndDifference(requestProviders, agentList)
	return matched, nil
}

func getAllowedProviders(name string, requestedOrganizations []string, reqConfig *lib.Requestor) ([]string, []byte, error) {

	// Check if the requesting user has access to all these organizations
	requesterJSON, err := lib.GetAndUnmarshalJSON(etcdClient, "/reasoner/requestor_config/"+name, reqConfig)
	if err != nil {
		return nil, nil, err
	}

	matches, _ := lib.SliceIntersectAndDifference(requestedOrganizations, reqConfig.AllowedPartners)
	log.Infof(" matches: %s", strings.Join(matches, ","))
	return matches, requesterJSON, nil
}

// func handler(w http.ResponseWriter, req *http.Request) {
// 	body, err := ioutil.ReadAll(req.Body)
// 	if err != nil {
// 		logger.Sugar().Infof("handler: Error reading body: %v", err)
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
// 	logger.Sugar().Infof("handler: 3, %s", routingKey)

// 	if err := lib.Publish(outputChannel, routingKey, convertedAmqMessage, ""); err != nil {
// 		logger.Sugar().Infof("Handler 4: Error publishing: %s", err)
// 	}

// 	// Wait for the response from the response channel
// 	select {
// 	case msg := <-responseChan:
// 		logger.Sugar().Infof("handler: 5, msg received: %s", msg.Body)
// 		w.Write(msg.Body)
// 	case <-ctx.Done():
// 		logger.Info("handler: 6, context timed out")
// 		http.Error(w, "handler: Request timed out", http.StatusRequestTimeout)
// 	}
// }
