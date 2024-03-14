package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func sideCarMessageHandler() func(ctx context.Context, grpcMsg *pb.SideCarMessage) error {
	return func(ctx context.Context, grpcMsg *pb.SideCarMessage) error {

		ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: messageHandler, process MS", grpcMsg.Traces)
		if err != nil {
			logger.Sugar().Warnf("Error starting span: %v", err)
		}
		defer span.End()

		switch grpcMsg.Type {
		case "microserviceCommunication":

			logger.Sugar().Info("switching on microserviceCommunication")
			msComm := &pb.MicroserviceCommunication{}
			msComm.RequestMetadata = &pb.RequestMetadata{}

			if err := grpcMsg.Body.UnmarshalTo(msComm); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal msComm message: %v", err)
			}

			switch msComm.RequestType {
			case "sqlDataRequest":
				handleSqlDataRequest(ctx, msComm)
			default:
				logger.Sugar().Errorf("Unknown RequestType type: %v", msComm.RequestType)
				return fmt.Errorf("unknown RequestType type: %s", msComm.RequestType)
			}

		default:
			logger.Sugar().Errorf("Unknown message type: %v", grpcMsg.Type)
			return fmt.Errorf("unknown message type: %s", grpcMsg.Type)
		}

		return nil
	}
}

func sendDataHandler(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error) {
	logger.Debug("Start sendDataHandler")
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		logger.Sugar().Infof("OK: %v", md)

	}

	ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: sendDataHandler, process grpc MS", data.Traces)
	if err != nil {
		logger.Sugar().Warnf("Error starting span: %v", err)
	}
	defer span.End()

	switch data.RequestType {
	case "sqlDataRequest":
		handleSqlDataRequest(ctx, data)
	default:
		logger.Sugar().Errorf("Unknown RequestType type: %v", data.RequestType)

		return nil, fmt.Errorf("unknown RequestType type: %s", data.RequestType)
	}

	return &emptypb.Empty{}, nil
}
