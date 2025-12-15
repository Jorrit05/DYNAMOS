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
			Id:       apiReqApproval.User.Id,
			UserName: apiReqApproval.User.UserName,
		}

		var dataRequestInterface map[string]interface{}
		if err := json.Unmarshal(apiReqApproval.DataRequest, &dataRequestInterface); err != nil {
			logger.Sugar().Errorf("Error unmarhsalling get request: %v", err)
			return
		}

		dataRequestOptions := &api.DataRequestOptions{}
		dataRequestOptions.Options = make(map[string]bool)
		if err := json.Unmarshal(apiReqApproval.DataRequest, &dataRequestOptions); err != nil {
			logger.Sugar().Errorf("Error unmMarshalling get apiReqApproval: %v", err)
			return
		}

		dataRequestInterface["user"] = userPb

		// Create protobuf struct for the req approval flow
		protoRequest := &pb.RequestApproval{
			Type:             apiReqApproval.Type,
			User:             userPb,
			DataProviders:    apiReqApproval.DataProviders,
			DestinationQueue: "policyEnforcer-in",
			Options:          dataRequestOptions.Options,
		}

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

			requestMetadata := &pb.RequestMetadata{
				JobId: msg.JobId,
			}
			dataRequestInterface["requestMetadata"] = requestMetadata

			// Marshal the combined data back into JSON for forwarding
			dataRequestJson, err := json.Marshal(dataRequestInterface)
			if err != nil {
				logger.Sugar().Errorf("Error marshalling combined data: %v", err)
				return
			}
			// note: maybe here I need to parse the MScomm and return it to the request response
			//responseData, err := sendData(endpoint, dataRequestJson)
			//
			//if err != nil {
			//	logger.Sugar().Errorf("Error sending data, %v", err)
			//} else {
			//	responseJson := &pb.MicroserviceCommunication{}
			//	err = json.Unmarshal([]byte(responseData), responseJson)
			//
			//	if err != nil {
			//		logger.Sugar().Error("Unmarshalling response did not go well: ", err)
			//	}
			//
			//	dataJson := responseJson.Data.AsMap()
			//	trace, ok := dataJson["trace"].(string)
			//
			//	if !ok {
			//		logger.Sugar().Error("No embeddings found in the return data.")
			//		trace = ""
			//	}
			//	responses[target] = trace
			//}

			logger.Sugar().Infof("Data Prepared jsonData: %s", dataRequestJson)
			logger.Sugar().Debugf("msg.AuthorizedProviders: %v", msg.AuthorizedProviders)
			responses := sendDataToAuthProviders(dataRequestJson, msg.AuthorizedProviders, apiReqApproval.Type, msg.JobId)
			w.WriteHeader(http.StatusOK)
			w.Write(responses)
			return

		case <-ctx.Done():
			http.Error(w, "Request timed out", http.StatusRequestTimeout)
			return
		}
	}
}

// Use the data request that was previously built and send it to the authorised providers
// acquired from the request approval
func sendDataToAuthProviders(dataRequest []byte, authorizedProviders map[string]string, msgType string, jobId string) []byte {
	var wg sync.WaitGroup
	responsesMu := sync.Mutex{}
	responses := make(map[string]string)

	agentPort := "8080"

	for auth, url := range authorizedProviders {
		wg.Add(1)
		target := strings.ToLower(auth)
		endpoint := fmt.Sprintf("http://%s:%s/agent/v1/%s/%s", url, agentPort, msgType, target)

		logger.Sugar().Infof("Sending request to %s.\nEndpoint: %s\nJSON:%v", target, endpoint, string(dataRequest))

		go func(target, endpoint string) {
			defer wg.Done()
			respData, err := sendData(endpoint, dataRequest)
			if err != nil {
				logger.Sugar().Errorf("Error sending data, %v", err)
				responsesMu.Lock()
				responses[target] = ""
				responsesMu.Unlock()
				return
			}

			responseJson := &pb.MicroserviceCommunication{}
			err = json.Unmarshal([]byte(respData), responseJson)
			if err != nil {
				logger.Sugar().Error("Unmarshalling response did not go well: ", err)
				responsesMu.Lock()
				responses[target] = ""
				responsesMu.Unlock()
				return
			}

			dataJson := responseJson.Data.AsMap()
			embeddings, ok := dataJson["embeddings"].(string)
			if !ok {
				logger.Sugar().Error("No embeddings found in the return data.")
				embeddings = ""
			}

			responsesMu.Lock()
			responses[target] = embeddings
			responsesMu.Unlock()
		}(target, endpoint)
	}

	wg.Wait()
	logger.Sugar().Debug("Returning responses")

	responseMap := map[string]interface{}{
		"jobId":     jobId,
		"responses": responses,
	}

	return cleanupAndMarshalResponse(responseMap)
}

// Now assumes input is map[string]interface{} and directly marshals it to prettified JSON.
func cleanupAndMarshalResponse(responseMap map[string]interface{}) []byte {
	prettifiedJSON, err := json.MarshalIndent(responseMap, "", "    ")
	if err != nil {
		logger.Sugar().Errorf("Error marshalling cleaned response: %v", err)
	}
	return prettifiedJSON
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

	// TODO: unmarshal and then return to the original request
	//     check jake's code
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
