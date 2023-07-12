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
	stop   = make(chan struct{}) // channel to tell the server to stop
)

type server struct {
	pb.UnimplementedSideCarServer
	pb.UnimplementedEtcdServer
}

func main() {
	flag.Parse()
	finished := make(chan struct{}) // channel to tell us the server has finished

	go func() {
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

		go func() {
			<-stop
			logger.Info("Stopping sidecar")
			s.Stop()
			// s.GracefulStop() // or s.Stop() if you need it to stop immediately
			logger.Info("1")
			if channel != nil {
				logger.Info("2")
				close_channel(channel)
			}
			logger.Info("3")
			close(finished)
			logger.Info("4")
		}()

		if err := s.Serve(lis); err != nil {
			logger.Sugar().Fatalw("failed to serve: %v", err)
		}
	}()

	<-finished

	logger.Sugar().Infof("Exiting sidecar server")
	os.Exit(0)
}
