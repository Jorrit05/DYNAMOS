package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func handleIncomingMessages(ctx context.Context, grpcMsg *pb.SideCarMessage) error {
	switch grpcMsg.Type {
	case "requestApproval":
		ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: checkRequestApproval", grpcMsg.Traces)
		if err != nil {
			logger.Sugar().Errorf("error starting trace: %v", err)
		}
		defer span.End()

		var requestApproval pb.RequestApproval
		if err := grpcMsg.Body.UnmarshalTo(&requestApproval); err != nil {
			logger.Sugar().Fatalf("Failed to unmarshal message: %v", err)
		}

		logger.Sugar().Infof("User name: %s", requestApproval.User.UserName)
		checkRequestApproval(ctx, &requestApproval)
	case "policyUpdate":
		ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: policyUpdate", grpcMsg.Traces)
		if err != nil {
			logger.Sugar().Errorf("error starting trace: %v", err)
		}
		defer span.End()

		policyUpdate := &pb.PolicyUpdate{}
		if err := grpcMsg.Body.UnmarshalTo(policyUpdate); err != nil {
			logger.Sugar().Fatalf("Failed to unmarshal message: %v", err)
		}
		checkPolicyUpdate(ctx, policyUpdate)

	default:
		logger.Sugar().Errorf("Unknown message type: %s", grpcMsg.Type)
		return fmt.Errorf("unknown message type: %s", grpcMsg.Type)
	}

	return nil
}
