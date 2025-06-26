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
	// Custom message variables to handle the expected messages, such as when multiple messages are expected
	ExpectedMessages int32
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

	// TODO: add custom logic here for ExpectedMessages. Use AGENT_ROLE to determine if this is a data provider or compute provider.
	// For SURF it is compute provider and it should then change the expected messages to nr of data providers. 
	// All other cases keep it as the default of 1 to avoid breaking anything.
	// TODO: then further add logic in all other services like sql-algorithm, etc. 
	// Sql-query should be fine and not affected, as this is not part of the compute provider, add that as a comment
	// TODO: likely sql-anonymize as well as this is only used by the data provider, not compute provider.

	// Set default expected messages to 1 for other cases
	expectedMessages := int32(1)

	// Check if the agent is acting as a compute provider. This logic is specific to the trusted third party use case, 
	// where the compute provider might receive messages from multiple data providers. In that case, the deployed microservice(s)
	// need to receive multiple messages (one for each data provider) before processing can continue.
	if os.Getenv("AGENT_ROLE") == "computeProvider" {
		// Retrieve the number of expected data providers from the environment
		if val := os.Getenv("NR_OF_DATA_PROVIDERS"); val != "" {
			// Try to convert the value to an integer
			if num, err := strconv.Atoi(val); err == nil && num > 0 {
				// Set expectedMessages based on the number of data providers
				expectedMessages = int32(num)
				logger.Sugar().Infof("computeProvider detected, expecting %d messages", expectedMessages)
			} else {
				// Log a warning if the value is invalid (e.g., non-integer or <= 0)
				logger.Sugar().Warnf("Invalid NR_OF_DATA_PROVIDERS value: %s", val)
			}

		} else {
			// Log a warning if the variable is not set
			logger.Sugar().Warn("NR_OF_DATA_PROVIDERS not set for computeProvider")
		}
	} else {
		logger.Sugar().Infof("Default message expectation: %d", expectedMessages)
	}

	// Create a new configuration instance with the provided parameters
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
		// Number of expected messages for this service, used to determine when to stop receiving messages.
		ExpectedMessages:          expectedMessages,
	}

	// Debug for printing message variables from the configuration
	logger.Sugar().Debugf("ExpectedMessages: %d", conf.ExpectedMessages)

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
