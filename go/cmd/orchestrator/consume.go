package main

import (
	"context"
	"errors"
	"io"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
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
	ctx := context.Background()

	stream, err := c.Consume(ctx, &pb.ConsumeRequest{QueueName: from, AutoAck: true})
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

		// Deserialize the span context
		spanContext, ok := propagation.FromBinary(grpcMsg.Trace)
		if !ok {
			return errors.New("invalid span context")
		}
		_, span := trace.StartSpanWithRemoteParent(ctx, serviceName+"/consume/"+grpcMsg.Type, spanContext)
		logger.Sugar().Debugw("Type:", "MessageType", grpcMsg.Type)

		switch grpcMsg.Type {
		case "validationResponse":
			validationResponse := &pb.ValidationResponse{}
			if err := grpcMsg.Body.UnmarshalTo(validationResponse); err != nil {
				logger.Sugar().Fatalf("Failed to unmarshal message: %v", err)
			}
			mutex.Lock()
			// Look up the corresponding channel in the request map
			requestData, ok := validationMap[validationResponse.User.Id]
			mutex.Unlock()

			if ok {
				logger.Sugar().Info("Sending validation to channel")
				// Send a signal on the channel to indicate that the response is ready
				requestData.response <- validationResponse
			} else {
				logger.Sugar().Errorw("unknown validation response", "GUID", validationResponse.User.Id)
			}

		default:
			logger.Sugar().Fatalf("Unknown message type: %s", grpcMsg.Type)
		}
		span.End()
	}
	return err
}
