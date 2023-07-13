package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/api"
	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

func sqlDataRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Entering sqlDataRequestHandler")

		body, err := api.GetRequestBody(w, r, serviceName)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		sqlDataRequest := &pb.SqlDataRequest{}
		sqlDataRequest.RequestMetada = &pb.RequestMetada{}

		err = protojson.Unmarshal(body, sqlDataRequest)
		if err != nil {
			logger.Sugar().Errorf("Error unmarshalling sqlDataRequest: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if sqlDataRequest.RequestMetada == nil {
			sqlDataRequest.RequestMetada = &pb.RequestMetada{}

		}

		// Get the jobname of this user
		// /agents/jobs/UVA/jorrit.stutterheim@cloudnation.nl -> jorrit-3141334
		jobName, err := getJobName(sqlDataRequest.User.UserName)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Get the matching composition request and determine our role
		// /agents/jobs/UVA/jorrit-3141334
		compositionRequest, err := getCompositionRequest(jobName)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Generate correlationID for this request
		correlationId := uuid.New().String()

		// Switch on the role we have in this data request
		if strings.EqualFold(compositionRequest.Role, "computeProvider") {
			err = handleSqlComputeProvider(jobName, compositionRequest, sqlDataRequest, correlationId)
			if err != nil {
				logger.Sugar().Errorf("Error in computeProvider role: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

		} else if strings.EqualFold(compositionRequest.Role, "all") {
			err = handleSqlAll(jobName, compositionRequest, sqlDataRequest, correlationId)
			if err != nil {
				logger.Sugar().Errorf("Error in all role: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		} else {
			logger.Sugar().Warnf("Unknown role or unexpected HTTP request: %s", compositionRequest.Role)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Create a channel to receive the response
		responseChan := make(chan *pb.MicroserviceCommunication)

		// Store the request information in the map
		mutex.Lock()
		responseMap[sqlDataRequest.RequestMetada.CorrelationId] = &dataResponse{response: responseChan}
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

func handleSqlAll(jobName string, compositionRequest *pb.CompositionRequest, sqlDataRequest *pb.SqlDataRequest, correlationId string) error {
	// Create msChain and deploy job.

	actualJobName, err := generateChainAndDeploy(compositionRequest, sqlDataRequest)
	if err != nil {
		logger.Sugar().Errorf("error deploying job: %v", err)
		return err
	}

	msComm := &pb.MicroserviceCommunication{}
	msComm.DestinationQueue = actualJobName
	msComm.ReturnAddress = agentConfig.RoutingKey

	any, err := anypb.New(sqlDataRequest)
	if err != nil {
		logger.Sugar().Error(err)
		return err
	}

	msComm.UserRequest = any
	msComm.CorrelationId = correlationId

	logger.Sugar().Debugf("Sending SendMicroserviceInput to: %s", actualJobName)

	go c.SendMicroserviceComm(context.Background(), msComm)
	return nil
}

func handleSqlComputeProvider(jobName string, compositionRequest *pb.CompositionRequest, sqlDataRequest *pb.SqlDataRequest, correlationId string) error {

	// pack and send request to all data providers, add own routing key as return address
	// check request and spin up own job (generate mschain, deployjob)
	if len(compositionRequest.DataProviders) == 0 {
		return fmt.Errorf("expected to know dataproviders")
	}

	for _, dataProvider := range compositionRequest.DataProviders {
		dataProviderRoutingKey := fmt.Sprintf("/agents/%s", dataProvider)
		var agentData lib.AgentDetails
		_, err := etcd.GetAndUnmarshalJSON(etcdClient, dataProviderRoutingKey, &agentData)
		if err != nil {
			return fmt.Errorf("error getting dataProvider dns")
		}

		sqlDataRequest.RequestMetada.DestinationQueue = agentData.RoutingKey

		// This is a bit confusing, but it tells the other agent to go back here.
		// The other agent, will reset the address to get the message from the job.
		sqlDataRequest.RequestMetada.ReturnAddress = agentConfig.RoutingKey

		sqlDataRequest.RequestMetada.CorrelationId = correlationId
		sqlDataRequest.RequestMetada.JobName = compositionRequest.JobName
		logger.Sugar().Debugf("Sending sqlDataRequest to: %s", sqlDataRequest.RequestMetada.DestinationQueue)

		c.SendSqlDataRequest(context.Background(), sqlDataRequest)
	}

	// TODO: Parse SQL request for extra compute services
	actualJobName, err := generateChainAndDeploy(compositionRequest, sqlDataRequest)
	if err != nil {
		logger.Sugar().Errorf("error deploying job: %v", err)
	}

	waitingJobMutex.Lock()
	waitingJobMap[sqlDataRequest.RequestMetada.CorrelationId] = actualJobName
	waitingJobMutex.Unlock()

	return nil
}
