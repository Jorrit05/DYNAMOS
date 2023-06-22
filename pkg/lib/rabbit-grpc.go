package lib

import (
	"context"
	"time"

	"github.com/avast/retry-go"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitializeRabbit(grpcAddr string, in *pb.ServiceRequest) (pb.SideCarClient, *grpc.ClientConn) {
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Sugar().Fatalf("could not connect to grpc server: %v", err)
	}

	c := pb.NewSideCarClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// _, err = c.StartService(ctx, in)
	// if err != nil {
	// 	logger.Sugar().Fatalw("could not establish connection with RabbitMQ: %v", err)
	// }

	retryOptions := []retry.Option{
		retry.Attempts(5),            // Retries up to 5 times
		retry.Delay(2 * time.Second), // Initial delay is 1 second
		retry.MaxDelay(time.Minute),  // Maximum delay is 1 minute
		retry.OnRetry(func(n uint, err error) {
			// This function is called each time a retry is made
			logger.Sugar().Errorf("Attempt %d: could not establish connection with RabbitMQ: %v", n, err)
		}),
	}

	err = retry.Do(func() error {
		_, err := c.StartService(ctx, in)
		return err
	}, retryOptions...)
	if err != nil {
		logger.Sugar().Fatalw("could not establish connection with RabbitMQ after retries: %v", err)
	}

	return c, conn

}
