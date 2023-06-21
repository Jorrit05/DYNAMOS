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
		logger.Sugar().Fatalf("Error on consume: %v", err)
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
		case "validationResponse":

			var validationResponse pb.ValidationResponse
			if err := grpcMsg.Body.UnmarshalTo(&validationResponse); err != nil {
				logger.Sugar().Fatalf("Failed to unmarshal message: %v", err)
			}

			logger.Sugar().Infof("Allowed Archetypes: %s", validationResponse.GetAllowedArcheTypes)

		default:
			logger.Sugar().Fatalf("Unknown message type: %s", grpcMsg.Type)
		}
	}
	return err
}
