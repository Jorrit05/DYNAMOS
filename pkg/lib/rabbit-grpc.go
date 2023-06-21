package lib

import (
	"context"
	"time"

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

	_, err = c.StartService(ctx, in)
	if err != nil {
		logger.Sugar().Fatalw("could not establish connection with RabbitMQ: %v", err)
	}

	return c, conn

}
