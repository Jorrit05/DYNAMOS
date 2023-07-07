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
)

func requestApprovalHandler(c pb.SideCarClient) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Starting requestApprovalHandler")
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

		logger.Debug("Sending request approval")
		go c.SendRequestApproval(context.Background(), protoRequest)

		// Create a channel to receive the response
		responseChan := make(chan *pb.ValidationResponse)

		// Store the request information in the map
		mutex.Lock()
		validationMap[protoRequest.User.Id] = &validation{response: responseChan}
		mutex.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		select {
		case msg := <-responseChan:
			logger.Sugar().Infof("Received response, %s", msg.Type)
			if msg.Type != "validationResponse" {
				logger.Sugar().Errorf("Unexpected message received, type: %s", msg.Type)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// On succesful requestApproval
			//   - Reply with AcceptedDataRequest
			//   - Start a composition request

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

			go startCompositionRequest(msg, &authorizedProviders, c)

			createAcceptedDataRequest(msg, w, &authorizedProviders)
			return

		case <-ctx.Done():
			http.Error(w, "Request timed out", http.StatusRequestTimeout)
			return
		}
	}

}

func getAuthorizedProviders(validationResponse *pb.ValidationResponse) (map[string]string, error) {
	authorizedProviders := make(map[string]string)

	for key, _ := range validationResponse.ValidDataproviders {
		var agentData lib.AgentDetails
		json, err := etcd.GetAndUnmarshalJSON(etcdClient, fmt.Sprintf("/agents/%s", key), &agentData)
		if err != nil {
			return nil, err
		} else if json == nil {
			// invalidProviders = append(invalidProviders, key)
			continue
		}
		authorizedProviders[key] = agentData.Dns
	}
	return authorizedProviders, nil
}
