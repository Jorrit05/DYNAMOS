package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// handleIncomingMessages processes incoming gRPC messages based on their type.
// It uses OpenTelemetry tracing to start a span for the operation.
func handleIncomingMessages(ctx context.Context, grpcMsg *pb.SideCarMessage) error {
	// Start a tracing span using the incoming trace context from grpcMsg
	ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: handleIncomingMessages/"+grpcMsg.Type, grpcMsg.Traces)
	if err != nil {
		logger.Sugar().Warnf("Error starting span: %v", err)
	}
	defer span.End() // Always end the span at the end of the function

	// Decide how to handle the message based on its type
	switch grpcMsg.Type {

	case "compositionRequest":
		// Handle a compositionRequest message (used to initiate a service composition workflow)
		logger.Debug("Received compositionRequest")

		// Unmarshal the protobuf message body into a CompositionRequest struct
		compositionRequest := &pb.CompositionRequest{}
		if err := grpcMsg.Body.UnmarshalTo(compositionRequest); err != nil {
			logger.Sugar().Errorf("Failed to unmarshal compositionRequest message: %v", err)
		}

		// Call the handler in a separate goroutine to process asynchronously
		go compositionRequestHandler(ctx, compositionRequest)

	case "microserviceCommunication":
		// Handle internal microservice-to-microservice communication message
		handleMicroserviceCommunication(ctx, grpcMsg)

	case "sqlDataRequest":
		// Handle SQL data requests coming from the computeProvider through RabbitMQ
		// handleSqlRequestDataProvider
		// Receive sqlDataRequest through RabbitMQ, means we received the request from the computeProvider
		// Implicitly this means I am only a dataProvider
		logger.Debug("Received sqlDataRequest from Rabbit (third party)")

		// Unmarshal the message into a SqlDataRequest struct
		sqlDataRequest := &pb.SqlDataRequest{}
		if err := grpcMsg.Body.UnmarshalTo(sqlDataRequest); err != nil {
			logger.Sugar().Errorf("Failed to unmarshal sqlResult message: %v", err)
		}

		// Store the original sender's return address in a map using the CorrelationId
		ttpMutex.Lock()
		thirdPartyMap[sqlDataRequest.RequestMetadata.CorrelationId] = sqlDataRequest.RequestMetadata.ReturnAddress
		ttpMutex.Unlock()

		// Prepare a MicroserviceCommunication message to forward the request internally
		msComm := &pb.MicroserviceCommunication{}
		msComm.RequestMetadata = &pb.RequestMetadata{}
		msComm.Type = "microserviceCommunication"
		msComm.RequestType = sqlDataRequest.Type
		// Set own routing key as return address to ensure the response comes back to me and then returned to where it needs
		msComm.RequestMetadata.ReturnAddress = agentConfig.RoutingKey
		msComm.RequestMetadata.CorrelationId = sqlDataRequest.RequestMetadata.CorrelationId

		// Pack the SqlDataRequest into a protobuf Any type and attach to the microservice message
		any, err := anypb.New(sqlDataRequest)
		if err != nil {
			logger.Sugar().Error(err)
			return err
		}
		msComm.OriginalRequest = any

		// Retrieve the original composition request associated with this job/user
		compositionRequest, err := getCompositionRequest(sqlDataRequest.User.UserName, sqlDataRequest.RequestMetadata.JobId)
		if err != nil {
			logger.Sugar().Errorf("Error getting matching composition request: %v", err)
			return err
		}

		// Set the internal destination queue based on the compositionRequest
		msComm.RequestMetadata.DestinationQueue = compositionRequest.LocalJobName

		// Construct etcd key/value for queue coordination
		key := fmt.Sprintf("/agents/jobs/%s/queueInfo/%s", serviceName, compositionRequest.LocalJobName)
		value := compositionRequest.LocalJobName

		// Build and deploy service chain dynamically
		generateChainAndDeploy(ctx, compositionRequest, compositionRequest.LocalJobName, sqlDataRequest.Options)

		// Send the prepared message through RabbitMQ
		c.SendMicroserviceComm(ctx, msComm)

		// Log key/value that will be stored in etcd
		logger.Sugar().Warnf("key: %v", key)
		logger.Sugar().Warnf("value: %v", value)

		// Store key in etcd with expiration (used for tracking and auto-cleanup)
		err = etcd.PutEtcdWithGrant(ctx, etcdClient, key, value, queueDeleteAfter)
		if err != nil {
			logger.Sugar().Errorf("Error PutEtcdWithGrant: %v", err)
		}

	default:
		// Handle unknown message types gracefully with logging
		logger.Sugar().Errorf("Unknown message type: %s", grpcMsg.Type)
		return fmt.Errorf("unknown message type: %s", grpcMsg.Type)
	}

	return nil
}
