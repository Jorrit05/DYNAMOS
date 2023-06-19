package main

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 3005, "The server port")
)

type server struct {
	pb.UnimplementedSideCarServer
}

func (s *server) StartService(ctx context.Context, in *pb.ServiceRequest) (*pb.ServiceReply, error) {
	logger.Sugar().Infow("Received: %v", in.ServiceName)

	// Call the SetupConnection function and handle the message consumption inside this function
	_, _, _, err := lib.SetupConnection(in.ServiceName, in.RoutingKey, in.StartConsuming, in.QueueAutoDelete)

	if err != nil {
		return &pb.ServiceReply{Message: "Failed to setup connection"}, nil
	}

	return &pb.ServiceReply{Message: "Successfully setup connection"}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Sugar().Fatalw("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSideCarServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		logger.Sugar().Fatalw("failed to serve: %v", err)
	}
}
