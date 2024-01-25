package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func handleIncomingMessages(ctx context.Context, grpcMsg *pb.SideCarMessage) error {
	logger.Debug("start orchestrator handleIncomingMessages")
	ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: handleIncomingMessages", grpcMsg.Traces)
	if err != nil {
		logger.Sugar().Warnf("Error starting span: %v", err)
	}
	defer span.End()

	logger.Sugar().Debugw("Type:", "MessageType", grpcMsg.Type)

	switch grpcMsg.Type {
	case "validationResponse":
		validationResponse := &pb.ValidationResponse{}
		if err := grpcMsg.Body.UnmarshalTo(validationResponse); err != nil {
			logger.Sugar().Fatalf("Failed to unmarshal message: %v", err)
		}
		mutex.Lock()
		// Look up the corresponding channel in the request map
		validationChannel, ok := validationMap[validationResponse.User.Id]

		if ok {
			logger.Sugar().Info("Sending validation to channel")
			// Send a signal on the channel to indicate that the response is ready
			validationChannel <- validation{response: validationResponse, localContext: ctx}
			delete(validationMap, validationResponse.User.Id)
		} else {
			logger.Sugar().Errorw("unknown validation response", "GUID", validationResponse.User.Id)
		}

		mutex.Unlock()
	case "policyUpdate":
		policyUpdate := &pb.PolicyUpdate{}
		if err := grpcMsg.Body.UnmarshalTo(policyUpdate); err != nil {
			logger.Sugar().Fatalf("Failed to unmarshal message: %v", err)
		}
		policyUpdateMutex.Lock()
		// Look up the corresponding channel in the request map
		jobCompositionRequest, ok := policyUpdateMap[policyUpdate.RequestMetadata.CorrelationId]
		if ok {
			delete(policyUpdateMap, policyUpdate.RequestMetadata.CorrelationId)
			processPolicyUpdate(ctx, jobCompositionRequest, policyUpdate)
		} else {
			logger.Sugar().Error("no job information available for this policy update")
		}
		policyUpdateMutex.Unlock()

	case "requestApprovalRequest":
		requestApproval := &pb.RequestApproval{}
		if err := grpcMsg.Body.UnmarshalTo(requestApproval); err != nil {
			logger.Sugar().Fatalf("Failed to unmarshal message: %v", err)
		}
		requestApprovalMutex.Lock()
		// Look up the corresponding channel in the request map
		approvalRequest, ok := requestApprovalMap[requestApproval.User.Id]
		if ok {
			handleRequestApproval(ctx, approvalRequest, requestApproval)
			delete(requestApprovalMap, requestApproval.User.Id)
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
