package msinit

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

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
	if s.GrpcConnection != nil {
		s.GrpcConnection.Close()
	}
}
func (s *Configuration) GetConnection() *grpc.ClientConn {
	if s.GrpcConnection != nil {
		return s.GrpcConnection
	}
	logger.Sugar().Errorf("GetConnecton, s.GrpcConnection is nil")
	return nil
}

type Configuration struct {
	Port            int
	FirstService    bool
	LastService     bool
	ServiceName     string
	SideCarClient   pb.SideCarClient
	GrpcConnection  *grpc.ClientConn
	SideCarCallback func(config *Configuration) func(ctx context.Context, grpcMsg *pb.RabbitMQMessage) error
	GrpcCallback    func(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error)
	Stopped         chan struct{} // channel to tell us the server has stopped
	StopServer      chan struct{} // Tell the server to stop
}

func NewConfiguration(serviceName string,
	grpcAddr string,
	sidecarCallback func(config *Configuration) func(ctx context.Context, grpcMsg *pb.RabbitMQMessage) error,
	grpcCallback func(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error)) (*Configuration, error) {

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

	conf := &Configuration{
		Port:            port,
		FirstService:    firstService > 0,
		LastService:     lastService > 0,
		ServiceName:     serviceName,
		SideCarCallback: sidecarCallback,
		GrpcCallback:    grpcCallback,
		Stopped:         make(chan struct{}), // channel to tell us the server has stopped
		StopServer:      make(chan struct{}), // Tell the server to stop
	}

	conf.ConnectNextService(grpcAddr)

	if conf.FirstService {
		conf.InitSidecarMessaging()
	} else {
		conf.StartGrpcServer()
	}

	return conf, nil
}

func (s *Configuration) ConnectNextService(grpcAddr string) {
	if s.LastService {
		// We are the last service, connect to the sidecar
		s.GrpcConnection = lib.GetGrpcConnection(grpcAddr + os.Getenv("SIDECAR_PORT"))
	} else {
		// Connect to following service
		s.GrpcConnection = lib.GetGrpcConnection(grpcAddr + strconv.Itoa(s.Port+1))
	}
}

func (s *Configuration) InitSidecarMessaging() {
	jobName := os.Getenv("JOB_NAME")
	if jobName == "" {
		logger.Sugar().Fatalf("Jobname not defined.")
	}

	s.SideCarClient = lib.InitializeSidecarMessaging(s.GrpcConnection, &pb.InitRequest{
		ServiceName:     jobName,
		RoutingKey:      jobName,
		QueueAutoDelete: true,
	})

	go func() {
		lib.StartConsumingWithRetry(s.ServiceName, s.SideCarClient, jobName, s.SideCarCallback(s), 5, 5*time.Second)
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
		server := grpc.NewServer()
		serverInstance := &lib.SharedServer{}

		pb.RegisterMicroserviceServer(server, serverInstance)
		pb.RegisterHealthServer(server, serverInstance)
		serverInstance.RegisterCallback("microserviceCommunication", s.GrpcCallback)
		go func() {
			<-s.StopServer
			logger.Info("Stopping StartGrpcServer")
			timeout := time.After(5 * time.Second)
			done := make(chan bool)

			go func() {
				server.GracefulStop()
				done <- true
			}()

			select {
			case <-timeout:
				logger.Info("Hard stop")
				server.Stop() // forcefully stop if graceful stop did not complete within timeout
			case <-done:
				logger.Info("Finished graceful stop")
			}

			close(s.Stopped)
		}()

		if err := server.Serve(lis); err != nil {
			logger.Sugar().Fatalw("failed to serve: %v", err)
		}
	}()
}
