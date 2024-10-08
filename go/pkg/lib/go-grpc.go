package lib

import (
	"context"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GetGrpcConnection establishes a gRPC connection to the specified address.
// It returns a *grpc.ClientConn if the connection is successfully established,
// otherwise it logs a fatal error and exits the program.
// The function makes up to 7 retries with a 1-second delay between each retry
// until the gRPC server is serving or until it reaches the maximum number of retries.
// The function uses the pb.HealthClient to check the health status of the gRPC server.
// It uses an insecure transport credentials and an ocgrpc.ClientHandler for stats handling.
// The function takes the gRPC address as a parameter.
func GetGrpcConnection(grpcAddr string) *grpc.ClientConn {
	var conn *grpc.ClientConn
	var err error
	conn, err = grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(new(ocgrpc.ClientHandler)))

	if err != nil {
		logger.Sugar().Fatalw("could not establish connection with grpc: %v", err)
	}
	h := pb.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	for i := 1; i <= 7; i++ { // maximum of 7 retries
		r, err := h.Check(ctx, &pb.HealthCheckRequest{})
		if err != nil {
			logger.Sugar().Warnf("could not check: %v", err)
		}

		if r.GetStatus() == pb.HealthCheckResponse_SERVING {
			break // The sidecar is ready, so break the loop.
		}
		logger.Debug("Sleep 1 second")
		time.Sleep(time.Second) // Wait a second before checking again.

		if i == 8 {
			logger.Sugar().Fatalf("could not connect with gRPC after %s tries: %v", "8", err)
		}
	}
	logger.Debug("returning conn GetGrpcConnection")
	return conn
}

func InitializeSidecarMessaging(conn *grpc.ClientConn, in *pb.InitRequest) pb.RabbitMQClient {
	c := pb.NewRabbitMQClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logger.Sugar().Debugf("Initializing sidecar messaging with: %v", in)
	_, err := c.InitRabbitMq(ctx, in)
	if err != nil {
		logger.Sugar().Errorf("could not establish connection with: %v", err)
	}
	return c
}
