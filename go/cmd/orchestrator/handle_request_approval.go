package main

import (
	"context"
	"encoding/json"
	"fmt"

	"net/http"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
)

func requestApprovalHandler(c pb.SideCarClient) http.HandlerFunc {

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
			DataProviders: reqApproval.DataProviders,
			SyncServices:  reqApproval.SyncServices,
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
		mutex.Lock()
		validationMap[protoRequest.User.Id] = responseChan
		mutex.Unlock()

		select {
		case validationStruct := <-responseChan:
			msg := validationStruct.response

			logger.Sugar().Infof("Received response, %s", msg.Type)
			if msg.Type != "validationResponse" {
				logger.Sugar().Errorf("Unexpected message received, type: %s", msg.Type)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			//TODO: This has a bug I think. When the third party is not online but the agent is.
			// there will be a valid provider, so the code will go through, but crashes on 'startCompositionRequest'
			authorizedProviders, err := getAuthorizedProviders(msg)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if len(authorizedProviders) == 0 {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Request was processed, but no agreements or available dataproviders have been found"))
				return
			}

			// TODO: Might be able to improve processing by converting functions to go routines
			// Seems a bit tricky though due to the response writer.

			compositionRequest := &pb.CompositionRequest{}
			compositionRequest.User = &pb.User{}
			userTargets, ctx, err := startCompositionRequest(validationStruct.localContext, msg, authorizedProviders, c, compositionRequest)
			if err != nil {
				logger.Sugar().Errorf("Error starting composition request: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			createAcceptedDataRequest(ctx, msg, w, userTargets, compositionRequest.JobName)
			return

		case <-ctx.Done():
			http.Error(w, "Request timed out", http.StatusRequestTimeout)
			return
		}
	}
}

func getAuthorizedProviders(validationResponse *pb.ValidationResponse) (map[string]lib.AgentDetails, error) {
	authorizedProviders := make(map[string]lib.AgentDetails)

	for key := range validationResponse.ValidDataproviders {
		var agentData lib.AgentDetails
		json, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/agents/%s", key), &agentData)
		if err != nil {
			return nil, err
		} else if json == nil {
			// invalidProviders = append(invalidProviders, key)
			continue
		}
		authorizedProviders[key] = agentData
	}
	return authorizedProviders, nil
}
