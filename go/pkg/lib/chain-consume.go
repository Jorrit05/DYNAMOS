package lib

import (
	"context"
	"io"
	"sync"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func ChainConsumeWithRetry(serviceName string, c pb.SideCarClient, queueName string, handler MessageHandlerFunc, maxRetries int, waitTime time.Duration, receiveMutex *sync.Mutex) {
	for i := 0; i < maxRetries; i++ {
		err := chainConsume(serviceName, c, queueName, handler, receiveMutex)
		if err == nil {
			return
		}

		logger.Sugar().Errorf("Failed to start consuming (attempt %d/%d): %v", i+1, maxRetries, err)

		// Wait for some time before retrying
		time.Sleep(waitTime)
	}
}

// Specific handler for microservices in the microservice chain
func chainConsume(serviceName string, c pb.SideCarClient, from string, handler MessageHandlerFunc, receiveMutex *sync.Mutex) error {
	logger.Sugar().Debugf("Starting %s/chainConsume", serviceName)
	ctx := context.Background()
	stream, err := c.ChainConsume(ctx, &pb.ConsumeRequest{QueueName: from, AutoAck: true})
	if err != nil {
		logger.Sugar().Fatalf("Error on consume: %v", err)
	}

	for {
		receiveMutex.Lock()
		grpcMsg, err := stream.Recv()
		receiveMutex.Unlock()

		if err == io.EOF {
			// The stream has ended.
			logger.Sugar().Warnw("Stream has ended", "error:", err)
			return err
		}

		if err != nil {
			logger.Sugar().Fatalf("Failed to receive: %v", err)
		}

		err = handler(ctx, grpcMsg)
		if err != nil {
			logger.Sugar().Fatalf("Failed to handle message: %v", err)
		}

		if err := stream.CloseSend(); err != nil {
			logger.Sugar().Fatalf("Failed to close stream: %v", err)
		}
	}
	logger.Sugar().Debugf("At the end %s/chainConsume", serviceName)
	return err
}
