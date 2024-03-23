package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/plugin/ocgrpc"

	"google.golang.org/grpc"
)

var (
	port      = flag.Int("port", grpcPort, "The server port")
	logger    = lib.InitLogger(logLevel)
	stop      = make(chan struct{}) // channel to tell the server to stop
	sendMutex = &sync.Mutex{}
)

type server struct {
	pb.UnimplementedSideCarServer
	pb.UnimplementedEtcdServer
	pb.UnimplementedMicroserviceServer
}

func main() {
	flag.Parse()
	finished := make(chan struct{}) // channel to tell us the server has finished

	_, err := lib.InitTracer("sidecar")
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Sugar().Fatalw("failed to listen: %v", err)
	}
	logger.Sugar().Infof("Serving on %v", *port)

	s := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))

	serverInstance := &server{}
	sharedServer := &lib.SharedServer{}
	pb.RegisterSideCarServer(s, serverInstance)
	pb.RegisterEtcdServer(s, serverInstance)
	pb.RegisterHealthServer(s, sharedServer)
	pb.RegisterGenericServer(s, sharedServer)

	// This env variable is only defined if this job is deployed
	// by a distributed agent as a datasharing pod
	if os.Getenv("TEMPORARY_JOB") != "" {
		pb.RegisterMicroserviceServer(s, sharedServer)
		sharedServer.RegisterCallback("microserviceCommunication", SendDataThroughAMQ)
	}

	go func(s *grpc.Server, finished chan struct{}) {
		<-stop
		logger.Info("Stopping sidecar wait for a few 4 seconds before initating stop")
		time.Sleep(4 * time.Second)

		s.Stop()

		if channel != nil {
			close_channel(channel)
		}

		close(finished)
	}(s, finished)

	if err := s.Serve(lis); err != nil {
		logger.Sugar().Fatalw("failed to serve: %v", err)
	}

	<-finished

	logger.Sugar().Infof("Exiting sidecar server")
	os.Exit(0)
}
