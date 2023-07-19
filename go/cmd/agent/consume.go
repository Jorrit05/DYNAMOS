package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/protobuf/types/known/anypb"
)

func handleIncomingMessages(ctx context.Context, grpcMsg *pb.RabbitMQMessage) error {

	// logger.Sugar().Infof("Jorrit check: servicename: %s ", serviceName)
	// spanContext, _ := propagation.FromBinary(grpcMsg.Trace)
	// lib.PrettyPrintSpanContext(spanContext)

	ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: handleIncomingMessages/"+grpcMsg.Type, grpcMsg.Trace)
	if err != nil {
		logger.Sugar().Errorf("Error starting span: %v", err)
	}
	defer span.End()

	// logger.Sugar().Debugw("Type:", "MessageType", grpcMsg.Type)
	// lib.PrettyPrintSpanContext(span.SpanContext())
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
		// Implicitly this means I am only a dataProvider
		logger.Debug("Received sqlDataRequest from Rabbit (third party)")

		sqlDataRequest := &pb.SqlDataRequest{}

		if err := grpcMsg.Body.UnmarshalTo(sqlDataRequest); err != nil {
			logger.Sugar().Errorf("Failed to unmarshal sqlResult message: %v", err)
		}

		waitingJobMutex.Lock()
		actualJobName, ok := waitingJobMap[sqlDataRequest.RequestMetada.JobName]
		waitingJobMutex.Unlock()

		ttpMutex.Lock()
		thirdPartyMap[sqlDataRequest.RequestMetada.CorrelationId] = sqlDataRequest.RequestMetada.ReturnAddress
		ttpMutex.Unlock()

		logger.Sugar().Warnf("jobName: %v", sqlDataRequest.RequestMetada.JobName)
		logger.Sugar().Warnf("actualJobName: %v", actualJobName)
		if ok {
			waitingJobMutex.Lock()
			delete(waitingJobMap, sqlDataRequest.RequestMetada.JobName)
			waitingJobMutex.Unlock()

			msComm := &pb.MicroserviceCommunication{}
			msComm.RequestMetada = &pb.RequestMetada{}

			msComm.Type = "microserviceCommunication"
			msComm.RequestType = sqlDataRequest.Type
			msComm.RequestMetada.DestinationQueue = actualJobName
			msComm.RequestMetada.ReturnAddress = agentConfig.RoutingKey
			msComm.RequestMetada.CorrelationId = sqlDataRequest.RequestMetada.CorrelationId

			sc := trace.FromContext(ctx).SpanContext()
			binarySc := propagation.Binary(sc)
			msComm.Trace = binarySc

			// Initialize the rest?
			any, err := anypb.New(sqlDataRequest)
			if err != nil {
				logger.Sugar().Error(err)
				return err
			}

			msComm.OriginalRequest = any

			logger.Sugar().Debugf("Sending SendMicroserviceInput to: %s", actualJobName)

			go c.SendMicroserviceComm(ctx, msComm)

		} else {
			logger.Sugar().Warnf("No job found for: %v", sqlDataRequest.RequestMetada.JobName)
		}
	default:
		logger.Sugar().Errorf("Unknown message type: %s", grpcMsg.Type)
		return fmt.Errorf("unknown message type: %s", grpcMsg.Type)
	}

	return nil
}
