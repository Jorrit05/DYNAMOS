package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func handleIncomingMessages(ctx context.Context, grpcMsg *pb.SideCarMessage) error {
	ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: handleIncomingMessages/"+grpcMsg.Type, grpcMsg.Traces)
	if err != nil {
		logger.Sugar().Warnf("Error starting span: %v", err)
	}
	defer span.End()

	switch grpcMsg.Type {
	case "compositionRequest":
		logger.Debug("Received compositionRequest")

		compositionRequest := &pb.CompositionRequest{}

		if err := grpcMsg.Body.UnmarshalTo(compositionRequest); err != nil {
			logger.Sugar().Errorf("Failed to unmarshal compositionRequest message: %v", err)
		}
		go compositionRequestHandler(ctx, compositionRequest)
	case "microserviceCommunication":

		handleMicroserviceCommunication(ctx, grpcMsg)
	case "sqlDataRequest":
		// handleSqlRequestDataProvider
		// Receive sqlDataRequest through RabbitMQ, means we received the request from the computeProvider
		// Implicitly this means I am only a dataProvider
		logger.Debug("Received sqlDataRequest from Rabbit (third party)")

		sqlDataRequest := &pb.SqlDataRequest{}

		if err := grpcMsg.Body.UnmarshalTo(sqlDataRequest); err != nil {
			logger.Sugar().Errorf("Failed to unmarshal sqlResult message: %v", err)
		}

		ttpMutex.Lock()
		thirdPartyMap[sqlDataRequest.RequestMetadata.CorrelationId] = sqlDataRequest.RequestMetadata.ReturnAddress
		ttpMutex.Unlock()

		msComm := &pb.MicroserviceCommunication{}
		msComm.RequestMetadata = &pb.RequestMetadata{}

		msComm.Type = "microserviceCommunication"
		msComm.RequestType = sqlDataRequest.Type
		// Set own routing key as return address to ensure the response comes back to me and then returned to where it needs
		msComm.RequestMetadata.ReturnAddress = agentConfig.RoutingKey
		msComm.RequestMetadata.CorrelationId = sqlDataRequest.RequestMetadata.CorrelationId

		any, err := anypb.New(sqlDataRequest)
		if err != nil {
			logger.Sugar().Error(err)
			return err
		}

		msComm.OriginalRequest = any
		compositionRequest, err := getCompositionRequest(sqlDataRequest.User.UserName, sqlDataRequest.RequestMetadata.JobId)
		if err != nil {

			logger.Sugar().Errorf("Error getting matching composition request: %v", err)
			return err
		}
		msComm.RequestMetadata.DestinationQueue = compositionRequest.LocalJobName
		key := fmt.Sprintf("/agents/jobs/%s/queueInfo/%s", serviceName, compositionRequest.LocalJobName)
		value := compositionRequest.LocalJobName
		generateChainAndDeploy(ctx, compositionRequest, compositionRequest.LocalJobName, sqlDataRequest.Options)
		c.SendMicroserviceComm(ctx, msComm)

		logger.Sugar().Warnf("key: %v", key)
		logger.Sugar().Warnf("value: %v", value)
		err = etcd.PutEtcdWithGrant(ctx, etcdClient, key, value, queueDeleteAfter)
		if err != nil {
			logger.Sugar().Errorf("Error PutEtcdWithGrant: %v", err)
		}
	default:
		logger.Sugar().Errorf("Unknown message type: %s", grpcMsg.Type)
		return fmt.Errorf("unknown message type: %s", grpcMsg.Type)
	}

	return nil
}
