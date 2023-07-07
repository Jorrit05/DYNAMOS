package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"

	"google.golang.org/grpc"
)

var (
	port   = flag.Int("port", grpcPort, "The server port")
	logger = lib.InitLogger(logLevel)
)

type server struct {
	pb.UnimplementedSideCarServer
	pb.UnimplementedEtcdServer
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Sugar().Fatalw("failed to listen: %v", err)
	}
	logger.Sugar().Infof("Serving on %v", *port)
	s := grpc.NewServer()
	serverInstance := &server{}
	sharedServer := &lib.SharedServer{}
	pb.RegisterSideCarServer(s, serverInstance)
	pb.RegisterEtcdServer(s, serverInstance)
	pb.RegisterHealthServer(s, sharedServer)

	// This env variable is only defined if this job is deployed
	// by a distributed agent
	if os.Getenv("TEMPORARY_JOB") != "" {
		pb.RegisterMicroserviceServer(s, sharedServer)
		sharedServer.RegisterCallback("sqlDataRequest", handleSqlDataRequest)

	}

	if err := s.Serve(lis); err != nil {
		logger.Sugar().Fatalw("failed to serve: %v", err)
	}
}
