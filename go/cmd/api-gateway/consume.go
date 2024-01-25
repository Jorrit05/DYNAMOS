package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func handleIncomingMessages(ctx context.Context, grpcMsg *pb.SideCarMessage) error {
	logger.Debug("start api gateway handleIncomingMessages")
	ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: handleIncomingMessages", grpcMsg.Traces)
	if err != nil {
		logger.Sugar().Warnf("Error starting span: %v", err)
	}
	defer span.End()

	logger.Sugar().Debugw("Type:", "MessageType", grpcMsg.Type)

	switch grpcMsg.Type {
	case "requestApprovalResponse":
		// Return the approval response to the client
		acceptedDataRequest := &pb.AcceptedDataRequest{}
		if err := grpcMsg.Body.UnmarshalTo(acceptedDataRequest); err != nil {
			logger.Sugar().Fatalf("Failed to unmarshal message: %v", err)
		}
		requestApprovalMutex.Lock()
		// Look up the corresponding channel in the request map
		approvalRequest, ok := requestApprovalMap[acceptedDataRequest.User.Id]
		if ok {
			approvalRequest <- validation{response: acceptedDataRequest, localContext: ctx}
			delete(requestApprovalMap, acceptedDataRequest.User.Id)
		} else {
			logger.Sugar().Error("no job information available for this policy update")
		}
		requestApprovalMutex.Unlock()
	default:
		logger.Sugar().Errorf("Unknown message type: %s", grpcMsg.Type)
		return fmt.Errorf("unknown message type: %s", grpcMsg.Type)
	}
	logger.Debug("end orchestrator handleIncomingMessages")

	return nil
}
