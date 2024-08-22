package msinit

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	logger = lib.InitLogger(zap.DebugLevel)
)

func (s *Configuration) CloseConnection() {
	if s.NextClientConnection != nil {
		s.NextClientConnection.Close()
	}

	if s.RabbitMsgClientConnection != nil {
		s.RabbitMsgClientConnection.Close()
	}
}

type Configuration struct {
	Port         uint32
	FirstService bool
	LastService  bool
	ServiceName  string

	MessageHandler   func(conf *Configuration) func(ctx context.Context, data *pb.MicroserviceCommunication) error
	StopMicroservice chan struct{} // channel to continue the main routine to kill the MS

	GrpcServer                *grpc.Server
	RabbitMsgClientConnection *grpc.ClientConn
	NextClientConnection      *grpc.ClientConn
	RabbitMsgClient           pb.RabbitMQClient
	NextClient                pb.MicroserviceClient
}

func NewConfiguration(
	ctx context.Context,
	serviceName string,
	grpcAddr string,
	COORDINATOR chan struct{},
	messageHandler func(conf *Configuration) func(ctx context.Context, data *pb.MicroserviceCommunication) error,
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

	jobName := os.Getenv("JOB_NAME")
	if jobName == "" {
		logger.Sugar().Fatalf("Jobname not defined.")
	}

	logger.Sugar().Debugf("NewConfiguration %s, firstServer: %s, port: %s. lastservice: %s", serviceName, firstService, port, lastService)

	conf := &Configuration{
		Port:                      uint32(port),
		FirstService:              firstService > 0,
		LastService:               lastService > 0,
		ServiceName:               serviceName,
		RabbitMsgClient:           nil,
		RabbitMsgClientConnection: nil,
		NextClientConnection:      nil,
		NextClient:                nil,
		MessageHandler:            messageHandler,
		StopMicroservice:          make(chan struct{}), // Continue the main routine to kill the MS
		GrpcServer:                nil,
	}

	if conf.FirstService {
		conf.GrpcServer = grpc.NewServer()
		conf.StartGrpcServer()
		conf.RabbitMsgClientConnection = lib.GetGrpcConnection(grpcAddr + os.Getenv("SIDECAR_PORT"))
		conf.RabbitMsgClient = pb.NewRabbitMQClient(conf.RabbitMsgClientConnection)

		if conf.LastService {
			conf.NextClientConnection = lib.GetGrpcConnection(grpcAddr + os.Getenv("SIDECAR_PORT"))
			conf.NextClient = pb.NewMicroserviceClient(conf.NextClientConnection)
		} else {
			conf.NextClientConnection = lib.GetGrpcConnection(grpcAddr + strconv.Itoa(int(conf.Port)+1))
			conf.NextClient = pb.NewMicroserviceClient(conf.NextClientConnection)
		}

		// When being called from an MS, QueueAutoDelete should be false, queues
		// are managed in the compositionRequest handler.
		chainRequest := &pb.ChainRequest{
			ServiceName:     conf.ServiceName,
			RoutingKey:      jobName,
			QueueAutoDelete: false,
			Port:            conf.Port,
		}

		conf.RabbitMsgClient.InitRabbitForChain(ctx, chainRequest)

	} else if conf.LastService {
		conf.GrpcServer = grpc.NewServer()
		conf.StartGrpcServer()
		conf.NextClientConnection = lib.GetGrpcConnection(grpcAddr + os.Getenv("SIDECAR_PORT"))
		conf.NextClient = pb.NewMicroserviceClient(conf.NextClientConnection)
		conf.RabbitMsgClientConnection = conf.NextClientConnection
		conf.RabbitMsgClient = pb.NewRabbitMQClient(conf.RabbitMsgClientConnection)
	} else {
		conf.GrpcServer = grpc.NewServer()
		conf.StartGrpcServer()
		conf.NextClientConnection = lib.GetGrpcConnection(grpcAddr + strconv.Itoa(int(conf.Port)+1))
		conf.NextClient = pb.NewMicroserviceClient(conf.NextClientConnection)
	}

	close(COORDINATOR)
	return conf, nil
}

// Register a gRPC server on our designated port
// StartGrpcServer starts the gRPC server for the Configuration instance.
// It listens on the specified port and registers the MicroserviceServer and HealthServer
// with the gRPC server. It also sets up this server with a callback from the initiating service.
//
// The server is started in a separate goroutine.
// parameters:
// - none
//
// returns:
// - none
func (s *Configuration) StartGrpcServer() {

	go func() {
		logger.Sugar().Infof("Start listening on port: %v", s.Port)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%v", s.Port))
		if err != nil {
			logger.Sugar().Fatalw("failed to listen: %v", err)
		}
		serverInstance := &lib.SharedServer{ServiceName: s.ServiceName, Callback: s.MessageHandler(s)}

		pb.RegisterMicroserviceServer(s.GrpcServer, serverInstance)
		pb.RegisterHealthServer(s.GrpcServer, serverInstance)

		if err := s.GrpcServer.Serve(lis); err != nil {
			logger.Sugar().Fatalw("failed to serve: %v", err)
		}
	}()
}

func (s *Configuration) SafeExit(oce *ocagent.Exporter, serviceName string) {
	logger.Debug("Start SafeExit")

	if s.LastService {
		if s.RabbitMsgClient == nil {
			logger.Sugar().Error("RabbitMsgClient is nil while we should send a StopReceivingRabbit signal")
		} else {
			logger.Sugar().Debugw("Send StopReceivingRabbit", "service", serviceName)
			_, err := s.RabbitMsgClient.StopReceivingRabbit(context.Background(), &pb.StopRequest{})
			if err != nil {
				logger.Sugar().Errorf("Error stopping receiving rabbit: %v", err)
			}
		}
	}

	logger.Sugar().Infof("Wait 2 seconds before ending %s", serviceName)
	oce.Flush()
	time.Sleep(2 * time.Second)
	oce.Stop()
	logger.Sugar().Debug("Start closing gRPC connections NextClientConnection and RabbitConnection")

	s.CloseConnection()

	if s.GrpcServer != nil {
		logger.Sugar().Debug("Close own gRPC server")
		s.StopGrpcServer()
	}
}

func (s *Configuration) StopGrpcServer() {
	logger.Info("Stopping StopGrpcServer")
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
