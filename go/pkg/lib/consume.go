package lib

import (
	"context"
	"io"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

type MessageHandlerFunc func(ctx context.Context, grpcMsg *pb.SideCarMessage) error

func StartConsumingWithRetry(serviceName string, c pb.SideCarClient, queueName string, handler MessageHandlerFunc, maxRetries int, waitTime time.Duration) {
	for i := 0; i < maxRetries; i++ {
		err := startConsuming(serviceName, c, queueName, handler)
		if err == nil {
			return
		}

		logger.Sugar().Errorf("Failed to start consuming (attempt %d/%d): %v", i+1, maxRetries, err)

		// Wait for some time before retrying
		time.Sleep(waitTime)
	}
}

func startConsuming(serviceName string, c pb.SideCarClient, from string, handler MessageHandlerFunc) error {
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
		// spanContext, ok := propagation.FromBinary(grpcMsg.Trace)
		// if !ok {
		// 	return errors.New("invalid span context")
		// }
		// _, span := trace.StartSpanWithRemoteParent(ctx, serviceName+"/consume/"+grpcMsg.Type, spanContext)
		// logger.Sugar().Debugw("Type:", "MessageType", grpcMsg.Type)

		err = handler(ctx, grpcMsg)
		if err != nil {
			logger.Sugar().Fatalf("Failed to handle message: %v", err)
		}

		// span.End()
	}
	return err
}
