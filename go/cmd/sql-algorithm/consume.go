package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// This is called if this is the first microservice, coming in from rabbitMQ
func sideCarMessageHandler() func(ctx context.Context, grpcMsg *pb.SideCarMessage) error {
	return func(ctx context.Context, grpcMsg *pb.SideCarMessage) error {

		ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: messageHandler, process MS", grpcMsg.Traces)
		if err != nil {
			logger.Sugar().Warnf("Error starting span: %v", err)
		}
		defer span.End()

		switch grpcMsg.Type {
		case "microserviceCommunication":
			msComm := &pb.MicroserviceCommunication{}
			msComm.RequestMetadata = &pb.RequestMetadata{}

			if err := grpcMsg.Body.UnmarshalTo(msComm); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal msComm message: %v", err)
			}

			incomingMessageWrapper(ctx, msComm)

		default:
			logger.Sugar().Errorf("Unknown message type: %v", grpcMsg.Type)
			return fmt.Errorf("unknown message type: %s", grpcMsg.Type)
		}

		return nil
	}
}

// This is the function being called by the last microservice
func sendDataHandler(ctx context.Context, msComm *pb.MicroserviceCommunication) (*emptypb.Empty, error) {
	logger.Debug("Start sendDataHandler")
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		logger.Sugar().Infof("OK: %v", md)
	}

	incomingMessageWrapper(ctx, msComm)

	return &emptypb.Empty{}, nil
}

// Wrapper function to handle incoming messages either from rabbitMQ or a previous microservice
// TODO: Move to lib.
func incomingMessageWrapper(ctx context.Context, msComm *pb.MicroserviceCommunication) {
	ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: incomingMessageWrapper, process grpc MS", msComm.Traces)
	if err != nil {
		logger.Sugar().Warnf("Error starting span: %v", err)
	}
	defer span.End()

	// Wait till all services and connections have started
	<-COORDINATOR

	c := pb.NewMicroserviceClient(config.NextConnection)

	switch msComm.RequestType {
	case "sqlDataRequest":
		handleSqlDataRequest(ctx, msComm)
	default:
		logger.Sugar().Errorf("Unknown RequestType type: %v", msComm.RequestType)
	}

	c.SendData(ctx, msComm)

	close(config.StopMicroservice)
}
