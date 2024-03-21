// This file contains the handlers for the requests that the API Gateway receives from the client
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"github.com/golang/protobuf/jsonpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.opencensus.io/trace"
)

func requestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Starting requestApprovalHandler")
		ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Start a new span with the context that has a timeout
		ctx, span := trace.StartSpan(ctxWithTimeout, "requestApprovalHandler")
		defer span.End()

		body, err := api.GetRequestBody(w, r, serviceName)
		if err != nil {
			return
		}

		var apiReqApproval api.RequestApproval
		if err := json.Unmarshal(body, &apiReqApproval); err != nil {
			logger.Sugar().Errorf("Error unmMarshalling get apiReqApproval: %v", err)
			return
		}

		userPb := &pb.User{
			Id:       reqApproval.User.Id,
			UserName: reqApproval.User.UserName,
		}

		// Convert the JSON request to a protobuf message
		protoRequest := &pb.RequestApproval{
			Type:             reqApproval.Type,
			User:             userPb,
			DataProviders:    reqApproval.DataProviders,
			DestinationQueue: "policyEnforcer-in",
		}

		// Unmarshal JSON into a regular Go struct
		var bodyJsonObj map[string]interface{}
		if err := json.Unmarshal(body, &bodyJsonObj); err != nil {
			logger.Sugar().Errorf("Error unmarhsalling get request: %v", err)
			return
		}

		dataRequest, err := prepareDataRequestStruct(protoRequest.Type, bodyJsonObj, userPb)
		if err != nil {
			logger.Sugar().Errorf("Error preparing data request: %v", err)
			return
		}
		protoRequest.Options = dataRequest.Options

		logger.Sugar().Debugf("Data Request of type %s prepared", dataRequest.Type)

		go func() {
			_, err := c.SendRequestApproval(ctx, protoRequest)
			if err != nil {
				logger.Sugar().Errorf("error in sending requestapproval: %v", err)
			}
		}()

		// Create a channel to receive the response
		responseChan := make(chan validation)

		requestApprovalMutex.Lock()
		requestApprovalMap[protoRequest.User.Id] = responseChan
		requestApprovalMutex.Unlock()

		_, err = c.SendRequestApproval(ctx, protoRequest)
		if err != nil {
			logger.Sugar().Errorf("error in sending requestapproval: %v", err)
		}

		select {
		case validationStruct := <-responseChan:
			msg := validationStruct.response

			logger.Sugar().Infof("Received response, %s", msg.Type)
			if msg.Type != "requestApprovalResponse" {
				logger.Sugar().Errorf("Unexpected message received, type: %s", msg.Type)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			logger.Sugar().Infof("Data Prepared Response: %s", dataRequest)

			// Add the job ID from the request approval to the data request
			dataRequest.RequestMetadata.JobId = msg.JobId

			responses := sendDataToAuthProviders(dataRequest, msg.AuthorizedProviders)
			w.WriteHeader(http.StatusOK)
			w.Write(responses)
			return

		case <-ctx.Done():
			http.Error(w, "Request timed out", http.StatusRequestTimeout)
			return
		}
	}
}

func prepareDataRequestStruct(dataRequestType string, bodyJsonObj map[string]interface{}, userPb *pb.User) (*pb.SqlDataRequest, error) {
	// Marshal the JSON object into JSON string
	dataRequestJsonString, err := json.Marshal(bodyJsonObj["data_request"])
	if err != nil {
		return nil, err
	}

	// If DataRequest is sqlDataRequest
	switch dataRequestType {
	case "sqlDataRequest":
		dataRequest := &pb.SqlDataRequest{}
		if err := jsonpb.UnmarshalString(string(dataRequestJsonString), dataRequest); err != nil {
			return nil, err
		}
		// This is part of the request approval but it is also required for the SQL data request
		dataRequest.User = userPb

		return dataRequest, nil

	}

	return nil, nil
}

// Use the data request that was previously built and send it to the authorised providers
// acquired from the request approval
func sendDataToAuthProviders(dataRequest *pb.SqlDataRequest, authorizedProviders map[string]string) []byte {
	// Prepare the data to send
	jsonData, err := json.Marshal(dataRequest)
	if err != nil {
		logger.Sugar().Fatalf("error marshalling sqldatarequest: %v", err)
	}

	// Setup the wait group for async data requests
	var wg sync.WaitGroup
	// Prepare the variables to be used for the requests
	var responses []string

	// This will be replaced with RMQ in the future
	agentPort := "8080"
	// Iterate over each auth provider
	for auth, url := range authorizedProviders {
		wg.Add(1)
		target := strings.ToLower(auth)
		// Construct the end point
		endpoint := fmt.Sprintf("http://%s:%s/agent/v1/%s/%s", url, agentPort, dataRequest.Type, target)

		logger.Sugar().Infof("Sending request to %s.\nEndpoint: %s\nJSON:%v", target, endpoint, string(jsonData))

		// Async call send the data
		go func() {
			respData, err := sendData(endpoint, jsonData)
			if err != nil {
				logger.Sugar().Errorf("Error sending data, %v", err)
			}
			responses = append(responses, respData)
			// Signal that the data request has been sent to all auth providers
			wg.Done()
		}()
	}
	// Wait until all the requests are complete
	wg.Wait()
	logger.Sugar().Debug("Returning responses")

	responseMap := map[string]interface{}{
		"jobId":     dataRequest.RequestMetadata.JobId,
		"responses": responses,
	}

	jsonResponse, _ := json.Marshal(responseMap)
	return jsonResponse

}

func sendData(endpoint string, jsonData []byte) (string, error) {

	// FIXME: Change to an actual token in the future?
	headers := map[string]string{
		"Authorization": "bearer 1234",
	}
	body, err := api.PostRequest(endpoint, string(jsonData), headers)
	if err != nil {
		return "", err
	}

	// Here we should send the request over the socket
	// For now we should append it to a list so that we gather all responses and send them in bulk
	logger.Sugar().Infof("Body: %v", body)
	return string(body), nil
}

func availableProvidersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Starting requestApprovalHandler")
		var availableProviders = make(map[string]lib.AgentDetails)
		resp, err := getAvailableProviders()
		if err != nil {
			logger.Sugar().Errorf("Error getting available providers: %v", err)
			return
		}

		// Bind resp to availableProviders
		availableProviders = resp

		jsonResponse, err := json.Marshal(availableProviders)
		if err != nil {
			logger.Sugar().Errorf("Error marshalling result, %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

// Maybe this should be moved into the orchestrarot
func getAvailableProviders() (map[string]lib.AgentDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the value from etcd.
	resp, err := etcdClient.Get(ctx, "/agents/online", clientv3.WithPrefix())
	if err != nil {
		logger.Sugar().Errorf("failed to get value from etcd: %v", err)
		return nil, err
	}

	// Initialize an empty map to store the unmarshaled structs.
	result := make(map[string]lib.AgentDetails)
	// Iterate through the key-value pairs and unmarshal the values into structs.
	for _, kv := range resp.Kvs {
		var target lib.AgentDetails
		err = json.Unmarshal(kv.Value, &target)
		if err != nil {
			// return nil, fmt.Errorf("failed to unmarshal JSON for key %s: %v", key, err)
		}
		result[string(target.Name)] = target
	}

	return result, nil

}
