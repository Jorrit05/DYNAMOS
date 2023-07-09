package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
)

func sqlDataRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Entering sqlDataRequestHandler")

		body, err := api.GetRequestBody(w, r, serviceName)
		if err != nil {
			return
		}

		sqlDataRequest := &pb.SqlDataRequest{}

		err = protojson.Unmarshal(body, sqlDataRequest)
		if err != nil {
			logger.Sugar().Errorf("Error unmarshalling sqlDataRequest: %v", err)
			return
		}

		jobName, err := etcd.GetValueFromEtcd(etcdClient, fmt.Sprintf("/activeJobs/%s", sqlDataRequest.User.UserName))
		if err != nil {
			logger.Sugar().Errorf("Error getting target: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		sqlDataRequest.Target = jobName
		sqlDataRequest.ReturnAddress = agentConfig.RoutingKey
		sqlDataRequest.CorrelationId = uuid.New().String()
		logger.Debug("Sending sqlDataReqeuest ")
		go c.SendSqlDataRequest(context.Background(), sqlDataRequest)

		// Create a channel to receive the response
		responseChan := make(chan *pb.SqlDataRequestResponse)

		// Store the request information in the map
		mutex.Lock()
		responseMap[sqlDataRequest.CorrelationId] = &dataResponse{response: responseChan}
		mutex.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		select {
		case msg := <-responseChan:
			logger.Sugar().Infof("Received response, %s", msg.CorrelationId)

			// Marshaling google.protobuf.Struct to JSON
			m := &jsonpb.Marshaler{}
			jsonString, err := m.MarshalToString(msg.Data)
			if err != nil {
				logger.Sugar().Errorf("Error in unmarshalling data: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error in returning result"))
			}

			//Handle response information
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(jsonString))
			return

		case <-ctx.Done():
			http.Error(w, "Request timed out", http.StatusRequestTimeout)
			return
		}
	}
}
