package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func handleIncomingMessages(ctx context.Context, grpcMsg *pb.RabbitMQMessage) error {

	ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: handleSidecarMessages", grpcMsg.Trace)
	if err != nil {
		logger.Sugar().Errorf("Error starting span: %v", err)
		return err
	}
	defer span.End()

	logger.Sugar().Debugw("Type:", "MessageType", grpcMsg.Type)

	switch grpcMsg.Type {
	case "requestApproval":
		logger.Debug("sidecar/Consume: Received a requestApproval")
		var requestApproval pb.RequestApproval
		if err := grpcMsg.Body.UnmarshalTo(&requestApproval); err != nil {
			logger.Sugar().Fatalf("Failed to unmarshal message: %v", err)
		}

		logger.Sugar().Infof("User name: %s", requestApproval.User.UserName)
		checkRequestApproval(ctx, &requestApproval)
	default:
		logger.Sugar().Errorf("Unknown message type: %s", grpcMsg.Type)
		return fmt.Errorf("unknown message type: %s", grpcMsg.Type)
	}

	return nil
}
