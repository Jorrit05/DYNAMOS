package main

import (
	"context"
	"log"
	"net"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedSideCarServer
}

func (s *server) StartService(ctx context.Context, in *pb.ServiceRequest) (*pb.ServiceReply, error) {
	log.Printf("Received: %v", in.GetService())

	// Call the SetupConnection function and handle the message consumption inside this function
	_, _, _, err := SetupConnection(in.GetService(), in.GetRoutingKey(), true)

	if err != nil {
		return &pb.ServiceReply{Message: "Failed to setup connection"}, nil
	}

	return &pb.ServiceReply{Message: "Successfully setup connection"}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSideCarServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
