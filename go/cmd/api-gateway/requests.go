// This file contains the handlers for the requests that the API Gateway receives from the client
package main

import (
	"context"
	"encoding/json"
	"net/http"
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
		ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Start a new span with the context that has a timeout
		ctx, span := trace.StartSpan(ctxWithTimeout, "requestApprovalHandler")
		defer span.End()

		body, err := api.GetRequestBody(w, r, serviceName)
		if err != nil {
			return
		}

		var reqApproval api.RequestApproval
		err = json.Unmarshal(body, &reqApproval)
		if err != nil {
			logger.Sugar().Errorf("Error unmarshalling reqApproval: %v", err)
			return
		}

		// Convert the JSON request to a protobuf message
		protoRequest := &pb.RequestApproval{
			Type: reqApproval.Type,
			User: &pb.User{
				Id:       reqApproval.User.Id,
				UserName: reqApproval.User.UserName,
			},
			DataProviders:    reqApproval.DataProviders,
			SyncServices:     reqApproval.SyncServices,
			DestinationQueue: "policyEnforcer-in",
		}

		go func() {
			_, err := c.SendRequestApproval(ctx, protoRequest)
			if err != nil {
				logger.Sugar().Errorf("error in sending requestapproval: %v", err)
			}
		}()

		// Create a channel to receive the response
		responseChan := make(chan validation)

		// Store the request information in the map
		requestApprovalMutex.Lock()
		requestApprovalMap[protoRequest.User.Id] = responseChan
		requestApprovalMutex.Unlock()

		select {
		case validationStruct := <-responseChan:
			msg := validationStruct.response

			logger.Sugar().Infof("Received response, %s", msg.Type)
			if msg.Type != "requestApprovalResponse" {
				logger.Sugar().Errorf("Unexpected message received, type: %s", msg.Type)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			jsonResponse, err := json.Marshal(msg)
			if err != nil {
				logger.Sugar().Errorf("Error marshalling result, %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(jsonResponse)
			return

		case <-ctx.Done():
			http.Error(w, "Request timed out", http.StatusRequestTimeout)
			return
		}
	}
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
