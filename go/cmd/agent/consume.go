package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/protobuf/types/known/anypb"
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
	case "compositionRequest":
		logger.Debug("Received compositionRequest")

		compositionRequest := &pb.CompositionRequest{}

		if err := grpcMsg.Body.UnmarshalTo(compositionRequest); err != nil {
			logger.Sugar().Errorf("Failed to unmarshal compositionRequest message: %v", err)
		}
		go compositionRequestHandler(ctx, compositionRequest)
	case "microserviceCommunication":
		logger.Debug("Received microserviceCommunication")
		msComm := &pb.MicroserviceCommunication{}
		msComm.RequestMetada = &pb.RequestMetada{}

		if err := grpcMsg.Body.UnmarshalTo(msComm); err != nil {
			logger.Sugar().Errorf("Failed to unmarshal msComm message: %v", err)
		}

		correlationId := msComm.RequestMetada.CorrelationId
		// Check if there is a job waiting for this result
		waitingJobMutex.Lock()
		waitingJobName, ok := waitingJobMap[correlationId]
		waitingJobMutex.Unlock()

		if ok {

			// There was still a job waiting for this response
			handleFurtherProcessing(ctx, waitingJobName, msComm)
			waitingJobMutex.Lock()
			delete(waitingJobMap, correlationId)
			waitingJobMutex.Unlock()
			return nil
		}

		// Check if there is a http result waiting for this
		mutex.Lock()
		// Look up the corresponding channel in the request map
		dataResponseChan, ok := responseMap[correlationId]
		mutex.Unlock()
		logger.Sugar().Info("Passed waitingJobName ")

		if ok {
			logger.Sugar().Info("Sending requestData to channel")

			// Send a signal on the channel to indicate that the response is ready
			dataResponseChan <- dataResponse{response: msComm, localContext: ctx}

			mutex.Lock()
			delete(responseMap, correlationId)
			mutex.Unlock()

			logger.Warn("returning from responding......")
			return nil
		}

		// Check if there is a third party where this goes back to
		ttpMutex.Lock()
		returnAddress, ok := thirdPartyMap[correlationId]
		ttpMutex.Unlock()

		if ok {
			logger.Sugar().Infof("Sending sql response to returnAddress: %s", returnAddress)
			// Send a signal on the channel to indicate that the response is ready
			msComm.RequestMetada.DestinationQueue = returnAddress

			c.SendMicroserviceComm(context.Background(), msComm)

			logger.Warn("returning from forwarding to 3rd party......")
			return nil
		}
		logger.Sugar().Errorw("unknown requestData response", "CorrelationId", correlationId)

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
			msComm.RequestMetada.ReturnAddress = sqlDataRequest.RequestMetada.ReturnAddress
			msComm.RequestMetada.CorrelationId = sqlDataRequest.RequestMetada.CorrelationId
			// Initialize the rest?
			any, err := anypb.New(sqlDataRequest)
			if err != nil {
				logger.Sugar().Error(err)
				return err
			}

			msComm.OriginalRequest = any

			logger.Sugar().Debugf("Sending SendMicroserviceInput to: %s", actualJobName)

			go c.SendMicroserviceComm(context.Background(), msComm)

		} else {
			logger.Sugar().Warnf("No job found for: %v", sqlDataRequest.RequestMetada.JobName)
		}
	default:
		logger.Sugar().Errorf("Unknown message type: %s", grpcMsg.Type)
		return fmt.Errorf("unknown message type: %s", grpcMsg.Type)
	}

	return nil
}
