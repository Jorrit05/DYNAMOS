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
	"go.opencensus.io/trace"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

func sqlDataRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Entering sqlDataRequestHandler")
		// Start a new span with the context that has a timeout

		ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		ctx, span := trace.StartSpan(ctxWithTimeout, serviceName+"/func: sqlDataRequestHandler")
		defer span.End()

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

		actualJobMutex.Lock()
		localJobname, ok := actualJobMap[compositionRequest.JobName]
		actualJobMutex.Unlock()

		if ok {
			actualJobMutex.Lock()
			delete(actualJobMap, compositionRequest.JobName)
			actualJobMutex.Unlock()
		} else {
			logger.Sugar().Error("Unknown actualjob")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Switch on the role we have in this data request
		if strings.EqualFold(compositionRequest.Role, "computeProvider") {
			ctx, err = handleSqlComputeProvider(ctx, localJobname, compositionRequest, sqlDataRequest, correlationId)
			if err != nil {
				logger.Sugar().Errorf("Error in computeProvider role: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

		} else if strings.EqualFold(compositionRequest.Role, "all") {
			ctx, err = handleSqlAll(ctx, localJobname, compositionRequest, sqlDataRequest, correlationId)
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
		responseChan := make(chan dataResponse)

		// Store the request information in the map
		mutex.Lock()
		responseMap[correlationId] = responseChan
		mutex.Unlock()

		select {
		case dataResponseStruct := <-responseChan:
			msg := dataResponseStruct.response

			logger.Sugar().Infof("Received response, %s", msg.RequestMetada.CorrelationId)

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

func handleSqlAll(ctx context.Context, jobName string, compositionRequest *pb.CompositionRequest, sqlDataRequest *pb.SqlDataRequest, correlationId string) (context.Context, error) {
	// Create msChain and deploy job.

	ctx, span := trace.StartSpan(ctx, serviceName+"/func: handleSqlAll")
	defer span.End()

	err := generateChainAndDeploy(ctx, compositionRequest, jobName, sqlDataRequest)
	if err != nil {
		logger.Sugar().Errorf("error deploying job: %v", err)
		return ctx, err
	}

	msComm := &pb.MicroserviceCommunication{}
	msComm.RequestMetada = &pb.RequestMetada{}
	msComm.Type = "microserviceCommunication"
	msComm.RequestMetada.DestinationQueue = jobName
	msComm.RequestMetada.ReturnAddress = agentConfig.RoutingKey
	msComm.RequestType = compositionRequest.RequestType

	any, err := anypb.New(sqlDataRequest)
	if err != nil {
		logger.Sugar().Error(err)
		return ctx, err
	}

	msComm.OriginalRequest = any
	msComm.RequestMetada.CorrelationId = correlationId

	logger.Sugar().Debugf("Sending SendMicroserviceInput to: %s", jobName)

	go c.SendMicroserviceComm(ctx, msComm)
	return ctx, nil
}

func handleSqlComputeProvider(ctx context.Context, jobName string, compositionRequest *pb.CompositionRequest, sqlDataRequest *pb.SqlDataRequest, correlationId string) (context.Context, error) {
	ctx, span := trace.StartSpan(ctx, serviceName+"/func: handleSqlComputeProvider")
	defer span.End()

	// pack and send request to all data providers, add own routing key as return address
	// check request and spin up own job (generate mschain, deployjob)
	if len(compositionRequest.DataProviders) == 0 {
		return ctx, fmt.Errorf("expected to know dataproviders")
	}

	for _, dataProvider := range compositionRequest.DataProviders {
		dataProviderRoutingKey := fmt.Sprintf("/agents/%s", dataProvider)
		var agentData lib.AgentDetails
		_, err := etcd.GetAndUnmarshalJSON(etcdClient, dataProviderRoutingKey, &agentData)
		if err != nil {
			return ctx, fmt.Errorf("error getting dataProvider dns")
		}

		sqlDataRequest.RequestMetada.DestinationQueue = agentData.RoutingKey

		// This is a bit confusing, but it tells the other agent to go back here.
		// The other agent, will reset the address to get the message from the job.
		sqlDataRequest.RequestMetada.ReturnAddress = agentConfig.RoutingKey

		sqlDataRequest.RequestMetada.CorrelationId = correlationId
		sqlDataRequest.RequestMetada.JobName = compositionRequest.JobName
		logger.Sugar().Debugf("Sending sqlDataRequest to: %s", sqlDataRequest.RequestMetada.DestinationQueue)

		c.SendSqlDataRequest(ctx, sqlDataRequest)
	}

	// TODO: Parse SQL request for extra compute services
	err := generateChainAndDeploy(ctx, compositionRequest, jobName, sqlDataRequest)
	if err != nil {
		logger.Sugar().Errorf("error deploying job: %v", err)
	}

	waitingJobMutex.Lock()
	waitingJobMap[sqlDataRequest.RequestMetada.CorrelationId] = jobName
	waitingJobMutex.Unlock()

	return ctx, nil
}
