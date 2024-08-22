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
	port             = flag.Int("port", grpcPort, "The server port")
	logger           = lib.InitLogger(logLevel)
	stop             = make(chan struct{}) // channel to tell the server to stop
	sendMutex        = &sync.Mutex{}
	running_messages = 0
)

func setupTracing() {
	_, err := lib.InitTracer("sidecar")
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}
}

func setupListener() net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Sugar().Fatalw("Failed to listen: %v", err)
	}
	logger.Sugar().Infof("Serving on %v", *port)

	return lis
}

func setupGRPCServer() (*grpc.Server, *serverInstance) {
	grpcServer := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))

	// register RabbitMQ, Etcd, and Microservice services on the gRPC server
	sideCarServer := &serverInstance{}
	pb.RegisterRabbitMQServer(grpcServer, sideCarServer)
	pb.RegisterEtcdServer(grpcServer, sideCarServer)
	pb.RegisterMicroserviceServer(grpcServer, sideCarServer)

	// register Health and Generic services on the gRPC server
	sharedServer := &lib.SharedServer{}
	pb.RegisterHealthServer(grpcServer, sharedServer)
	pb.RegisterGenericServer(grpcServer, sharedServer)

	return grpcServer, sideCarServer
}

func stopSidecar(s *grpc.Server, grpcServerInstance *serverInstance, finished chan struct{}) {
	<-stop
	logger.Info("Stopping sidecar wait for a few 4 seconds before initating stop")
	time.Sleep(4 * time.Second)

	if grpcServerInstance.channel != nil {
		close_channel(grpcServerInstance.channel, grpcServerInstance.conn)
	} else {
		logger.Sugar().Debug("Channel is nil")
	}

	s.Stop()

	close(finished)
}

func main() {
	flag.Parse()
	finished := make(chan struct{}) // channel to tell us the server has finished

	setupTracing()

	lis := setupListener()

	grpcServer, grpcServerInstance := setupGRPCServer()

	go stopSidecar(grpcServer, grpcServerInstance, finished)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Sugar().Fatalw("failed to serve: %v", err)
	}

	<-finished

	logger.Sugar().Infof("Exiting sidecar server")
	os.Exit(0)
}
