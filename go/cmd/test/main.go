package main

import (
	"context"
	"fmt"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	logger      = lib.InitLogger(zap.DebugLevel)
	conn        *grpc.ClientConn
	serviceName string = "test"
	grpcAddr           = "localhost:50052"
)

func main() {
	defer logger.Sync() // flushes buffer, if any

	conn = lib.GetGrpcConnection(grpcAddr, serviceName)
	defer conn.Close()
	c := lib.InitializeSidecarMessaging(conn, &pb.InitRequest{ServiceName: fmt.Sprintf("%s-in", serviceName), RoutingKey: fmt.Sprintf("%s-in", serviceName), QueueAutoDelete: false})

	c.SendTest(context.TODO(), &pb.SqlDataRequest{})
	select {}
}
