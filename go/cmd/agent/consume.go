package main

import (
	"context"
	"io"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func startConsumingWithRetry(c pb.SideCarClient, name string, maxRetries int, waitTime time.Duration) {
	for i := 0; i < maxRetries; i++ {
		err := startConsuming(c, name)
		if err == nil {
			return
		}

		logger.Sugar().Errorf("Failed to start consuming (attempt %d/%d): %v", i+1, maxRetries, err)

		// Wait for some time before retrying
		time.Sleep(waitTime)
	}
}

func startConsuming(c pb.SideCarClient, from string) error {
	stream, err := c.Consume(context.Background(), &pb.ConsumeRequest{QueueName: from, AutoAck: true})
	if err != nil {
		logger.Sugar().Errorf("Error on consume: %v", err)
	}

	for {
		grpcMsg, err := stream.Recv()
		if err == io.EOF {
			// The stream has ended.
			logger.Sugar().Warnw("Stream has ended", "error:", err)
			break
		}

		if err != nil {
			logger.Sugar().Fatalf("Failed to receive: %v", err)
		}

		logger.Sugar().Debugw("Type:", "MessageType", grpcMsg.Type)

		switch grpcMsg.Type {
		case "compositionRequest":
			logger.Debug("Received compositionRequest")

			compositionRequest := &pb.CompositionRequest{}

			if err := grpcMsg.Body.UnmarshalTo(compositionRequest); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal compositionRequest message: %v", err)
			}
			go compositionRequestHandler(compositionRequest)
		case "sqlDataRequestResponse":
			logger.Debug("Received sqlDataRequestResponse")
			sqlResult := &pb.SqlDataRequestResponse{}

			if err := grpcMsg.Body.UnmarshalTo(sqlResult); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal sqlResult message: %v", err)
			}

			mutex.Lock()
			// Look up the corresponding channel in the request map
			requestData, ok := responseMap[sqlResult.CorrelationId]
			mutex.Unlock()

			if ok {
				logger.Sugar().Info("Sending requestData to channel")
				// Send a signal on the channel to indicate that the response is ready
				requestData.response <- sqlResult
			} else {
				logger.Sugar().Errorw("unknown requestData response", "CorrelationId", sqlResult.CorrelationId)
			}
		default:
			logger.Sugar().Fatalf("Unknown message type: %s", grpcMsg.Type)
		}
	}
	return err
}
