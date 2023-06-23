package main

import (
	"context"
	"encoding/json"

	"net/http"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func requestApprovalHandler() http.HandlerFunc {

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
				ID:       reqApproval.User.ID,
				UserName: reqApproval.User.UserName,
			},
			DataProviders: reqApproval.DataProviders,
			SyncServices:  reqApproval.SyncServices,
		}

		logger.Info("Sending request approval")
		go c.SendRequestApproval(context.Background(), protoRequest)

		// Create a channel to receive the response
		responseChan := make(chan *pb.ValidationResponse)

		// Store the request information in the map
		mutex.Lock()
		validationMap[protoRequest.User.ID] = &validation{response: responseChan}
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

			go startCompositionRequest(msg)
			createAcceptedDataRequest()
			w.WriteHeader(http.StatusOK)
			// w.Write(jsonData)

			return

		case <-ctx.Done():
			http.Error(w, "Request timed out", http.StatusRequestTimeout)
			return
		}
	}

}
