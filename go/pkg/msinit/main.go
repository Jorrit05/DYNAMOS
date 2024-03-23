package msinit

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	logger = lib.InitLogger(zap.DebugLevel)
)

func (s *Configuration) CloseConnection() {
	if s.NextConnection != nil {
		s.NextConnection.Close()
	}

	if s.SidecarConnection != nil {
		s.SidecarConnection.Close()
	}
}

type Configuration struct {
	Port              int
	FirstService      bool
	LastService       bool
	ServiceName       string
	SideCarClient     pb.SideCarClient
	SidecarConnection *grpc.ClientConn
	NextConnection    *grpc.ClientConn
	SideCarCallback   func() func(ctx context.Context, grpcMsg *pb.SideCarMessage) error
	GrpcCallback      func(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error)
	StopMicroservice  chan struct{} // channel to continue the main routine to kill the MS
	// Exit              chan struct{} // Final exit
	GrpcServer *grpc.Server
}

func NewConfiguration(serviceName string,
	grpcAddr string,
	COORDINATOR chan struct{},
	sidecarCallback func() func(ctx context.Context, grpcMsg *pb.SideCarMessage) error,
	grpcCallback func(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error),
	receiveMutex *sync.Mutex,
) (*Configuration, error) {

	port, err := strconv.Atoi(os.Getenv("DESIGNATED_GRPC_PORT"))
	if err != nil {
		return nil, fmt.Errorf("error determining port number: %w", err)
	}
	firstService, err := strconv.Atoi(os.Getenv("FIRST"))
	if err != nil {
		return nil, fmt.Errorf("error determining first service: %w", err)
	}

	lastService, err := strconv.Atoi(os.Getenv("LAST"))
	if err != nil {
		return nil, fmt.Errorf("error determining last service: %w", err)
	}

	logger.Sugar().Debugf("NewConfiguration %s, firstServer: %s, port: %s. lastservice: %s", serviceName, firstService, port, lastService)

	conf := &Configuration{
		Port:              port,
		FirstService:      firstService > 0,
		LastService:       lastService > 0,
		ServiceName:       serviceName,
		SidecarConnection: nil,
		NextConnection:    nil,
		SideCarCallback:   sidecarCallback,
		GrpcCallback:      grpcCallback,
		StopMicroservice:  make(chan struct{}), // Continue the main routine to kill the MS
		// StopGrpcServer:    make(chan struct{}), // Tell the self-hosted GRPC listener to stop
		GrpcServer: nil,
		// Exit:              make(chan struct{}), // Final exit
	}

	if conf.FirstService && conf.LastService {
		conf.SidecarConnection = lib.GetGrpcConnection(grpcAddr + os.Getenv("SIDECAR_PORT"))
		conf.InitSidecarMessaging(receiveMutex)
		conf.NextConnection = lib.GetGrpcConnection(grpcAddr + os.Getenv("SIDECAR_PORT"))
	} else if conf.FirstService {
		conf.SidecarConnection = lib.GetGrpcConnection(grpcAddr + os.Getenv("SIDECAR_PORT"))
		conf.InitSidecarMessaging(receiveMutex)
		conf.NextConnection = lib.GetGrpcConnection(grpcAddr + strconv.Itoa(conf.Port+1))
	} else if conf.LastService {
		conf.GrpcServer = grpc.NewServer()
		conf.StartGrpcServer()
		conf.NextConnection = lib.GetGrpcConnection(grpcAddr + os.Getenv("SIDECAR_PORT"))
	} else {
		conf.GrpcServer = grpc.NewServer()
		conf.StartGrpcServer()
		conf.NextConnection = lib.GetGrpcConnection(grpcAddr + strconv.Itoa(conf.Port+1))
	}
	close(COORDINATOR)
	return conf, nil
}

func (s *Configuration) InitSidecarMessaging(receiveMutex *sync.Mutex) {
	logger.Debug("InitSidecarMessaging")

	jobName := os.Getenv("JOB_NAME")
	if jobName == "" {
		logger.Sugar().Fatalf("Jobname not defined.")
	}

	// When being called from an MS, QueueAutoDelete should be false, it is no managed in the compositionrequest handler.
	// This might be up for refactoring....
	s.SideCarClient = lib.InitializeSidecarMessaging(s.SidecarConnection, &pb.InitRequest{
		ServiceName:     jobName,
		RoutingKey:      jobName,
		QueueAutoDelete: false,
	})

	go func() {
		lib.ChainConsumeWithRetry(s.ServiceName, s.SideCarClient, jobName, s.SideCarCallback(), 5, 5*time.Second, receiveMutex)
	}()
}

// Register a gRPC server on our designated port
func (s *Configuration) StartGrpcServer() {

	go func() {
		logger.Sugar().Infof("Start listening on port: %v", s.Port)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%v", s.Port))
		if err != nil {
			logger.Sugar().Fatalw("failed to listen: %v", err)
		}
		serverInstance := &lib.SharedServer{}

		pb.RegisterMicroserviceServer(s.GrpcServer, serverInstance)
		pb.RegisterHealthServer(s.GrpcServer, serverInstance)
		serverInstance.RegisterCallback("microserviceCommunication", s.GrpcCallback)

		if err := s.GrpcServer.Serve(lis); err != nil {
			logger.Sugar().Fatalw("failed to serve: %v", err)
		}
	}()

}

func (s *Configuration) SafeExit(oce *ocagent.Exporter, serviceName string) {
	logger.Sugar().Infof("Wait 2 seconds before ending %s", serviceName)

	oce.Flush()
	time.Sleep(2 * time.Second)
	oce.Stop()
	s.CloseConnection()

	if s.GrpcServer != nil {
		s.StopGrpcServer()
	}
}

func (s *Configuration) StopGrpcServer() {
	logger.Info("Stopping StartGrpcServer")
	timeout := time.After(5 * time.Second)
	done := make(chan bool)

	go func() {
		s.GrpcServer.GracefulStop()
		done <- true
	}()

	select {
	case <-timeout:
		logger.Info("Hard stop")
		s.GrpcServer.Stop() // forcefully stop if graceful stop did not complete within timeout
	case <-done:
		logger.Info("Finished graceful stop")
	}
}
