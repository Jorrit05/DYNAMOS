package lib

import (
	"context"
	"io"
	"sync"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

type MessageHandlerFunc func(ctx context.Context, grpcMsg *pb.SideCarMessage) error

func StartConsumingWithRetry(serviceName string, c pb.SideCarClient, queueName string, handler MessageHandlerFunc, maxRetries int, waitTime time.Duration, receiveMutex *sync.Mutex) {
	for i := 0; i < maxRetries; i++ {
		err := startConsuming(serviceName, c, queueName, handler, receiveMutex)
		if err == nil {
			return
		}

		logger.Sugar().Errorf("Failed to start consuming (attempt %d/%d): %v", i+1, maxRetries, err)

		// Wait for some time before retrying
		time.Sleep(waitTime)
	}
}

func startConsuming(serviceName string, c pb.SideCarClient, from string, handler MessageHandlerFunc, receiveMutex *sync.Mutex) error {
	ctx := context.Background()
	stream, err := c.Consume(ctx, &pb.ConsumeRequest{QueueName: from, AutoAck: true})
	if err != nil {
		logger.Sugar().Fatalf("Error on consume: %v", err)
	}

	for {
		receiveMutex.Lock()
		grpcMsg, err := stream.Recv()
		receiveMutex.Unlock()

		logger.Sugar().Debugw("startConsuming receiving", "serviceName:", serviceName)
		if err == io.EOF {
			// The stream has ended.
			logger.Sugar().Warnw("Stream has ended", "error:", err)
			break
		}

		if err != nil {
			logger.Sugar().Fatalf("Failed to receive: %v", err)
		}

		err = handler(ctx, grpcMsg)
		if err != nil {
			logger.Sugar().Fatalf("Failed to handle message: %v", err)
		}
	}
	return err
}
