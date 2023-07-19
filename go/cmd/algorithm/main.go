package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/grpc"
)

var (
	logger = lib.InitLogger(logLevel)
	conn   *grpc.ClientConn
	config *Configuration
	stop   = make(chan struct{}) // channel to tell the server to stop

)

type Configuration struct {
	Port           int
	FirstService   bool
	LastService    bool
	ServiceName    string
	SideCarClient  pb.SideCarClient
	GrpcConnection *grpc.ClientConn
}

func NewConfiguration() (*Configuration, error) {
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

	return &Configuration{
		Port:         port,
		FirstService: firstService > 0,
		LastService:  lastService > 0,
		ServiceName:  serviceName,
	}, nil
}

func (s *Configuration) ConnectNextService() {
	if s.LastService {
		// We are the last service, connect to the sidecar
		s.GrpcConnection = lib.GetGrpcConnection(grpcAddr+os.Getenv("SIDECAR_PORT"), serviceName)
	} else {
		// Connect to following service
		s.GrpcConnection = lib.GetGrpcConnection(grpcAddr+strconv.Itoa(s.Port+1), serviceName)
	}
}

func (s *Configuration) InitSidecarMessaging() {
	// var conn *grpc.ClientConn
	// if !s.LastService {
	// 	conn = lib.GetGrpcConnection(grpcAddr+os.Getenv("SIDECAR_PORT"), serviceName)
	// } else {
	// 	conn = s.GrpcConnection
	// }

	jobName := os.Getenv("JOB_NAME")
	if jobName == "" {
		logger.Sugar().Fatalf("Jobname not defined.")
	}

	s.SideCarClient = lib.InitializeSidecarMessaging(s.GrpcConnection, &pb.InitRequest{
		ServiceName:     jobName,
		RoutingKey:      jobName,
		QueueAutoDelete: true,
	})

	// Define a WaitGroup
	// var wg sync.WaitGroup
	// wg.Add(1)

	go func() {
		lib.StartConsumingWithRetry(serviceName, s.SideCarClient, jobName, createCallbackHandler(config), 5, 5*time.Second)

		// startConsumingWithRetry(s.SideCarClient, jobName, 5, 5*time.Second)
		// wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
	}()
	// wg.Wait() // Wait for all goroutines to finish

}

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

// Main function
func main() {
	logger.Debug("Starting algorithm service")
	var err error
	config, err = NewConfiguration()
	if err != nil {
		logger.Sugar().Fatalf("%v", err)
	}

	config.ConnectNextService()
	defer config.CloseConnection()

	var stopped <-chan struct{}
	if config.FirstService {
		config.InitSidecarMessaging()
	} else {
		// Start server to receive a SendData request from the previous service
		stopped = StartGrpcMicroserviceServer(config.Port)
	}

	<-stopped
	logger.Sugar().Infof("Exiting algorithm service")
	os.Exit(0)
}

// Register a gRPC server on our designated port
func StartGrpcMicroserviceServer(port int) <-chan struct{} {
	stopped := make(chan struct{}) // channel to tell us the server has stopped

	go func() {
		logger.Sugar().Infof("Start listening on port: %v", port)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
		if err != nil {
			logger.Sugar().Fatalw("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		serverInstance := &lib.SharedServer{}

		pb.RegisterMicroserviceServer(s, serverInstance)
		pb.RegisterHealthServer(s, serverInstance)

		serverInstance.RegisterCallback("sqlDataRequest", handleSqlDataRequest)

		go func() {
			<-stop
			logger.Info("Stopping StartGrpcMicroserviceServer")
			s.GracefulStop() // or s.Stop() if you need it to stop immediately
			close(stopped)
		}()

		if err := s.Serve(lis); err != nil {
			logger.Sugar().Fatalw("failed to serve: %v", err)
		}
	}()

	return stopped
}

// // func ReveiceData() {
// // 	// Assuming `row` is your data fetched from the database.
// // 	fields := make(map[string]*structpb.Value)
// // 	fields["name"] = structpb.NewStringValue("Jorrit")
// // 	fields["date_of_birth"] = structpb.NewStringValue("september")
// // 	fields["job_title"] = structpb.NewStringValue("IT")
// // 	fields["other"] = structpb.NewBoolValue(true)

// // 	s := &structpb.Struct{Fields: fields}

// // 	iets := s.Fields["other"].GetListValue().ProtoReflect().Type()
// // 	fmt.Println(iets)
// // 	fmt.Println("xxx")
// // 	fmt.Println(s.GetFields())
// // }

// func main() {
// 	logger.Debug("Starting algorithm service")

// 	port, err := strconv.ParseInt(os.Getenv("DESIGNATED_GRPC_PORT"), 10, 32)
// 	firstServiceInt, err1 := strconv.Atoi(os.Getenv("FIRST"))
// 	lastServiceInt, err2 := strconv.Atoi(os.Getenv("LAST"))
// 	if err != nil || err1 != nil || err2 != nil {
// 		logger.Sugar().Fatalf("Error determining port number: %v", err)
// 	}

// 	if lastServiceInt > 0 {
// 		// We are the last service, connect to the sidecar
// 		conn = lib.GetGrpcConnection(grpcAddr + os.Getenv("SIDECAR_PORT"))
// 	} else {
// 		// Connect to following service
// 		conn = lib.GetGrpcConnection(grpcAddr + strconv.Itoa(int(port+1)))
// 	}
// 	defer conn.Close()

// 	var c pb.SideCarClient
// 	if firstServiceInt > 0 {
// 		c = lib.InitializeSidecarMessaging(conn, &pb.ServiceRequest{ServiceName: fmt.Sprintf("%s-in", serviceName), RoutingKey: fmt.Sprintf("%s-in", serviceName), QueueAutoDelete: false})

// 		// Define a WaitGroup
// 		var wg sync.WaitGroup
// 		wg.Add(1)

// 		go func() {
// 			startConsumingWithRetry(c, fmt.Sprintf("%s-in", serviceName), 5, 5*time.Second)

// 			wg.Done() // Decrement the WaitGroup counter when the goroutine finishes
// 		}()

// 	}

// 	stopped := StartGrpcMicroserviceServer(port)
// 	logger.Info("started GRPC server")

// 	// Wait for GRPC server to shutdown gracefully before quiting
// 	<-stopped
// 	logger.Sugar().Infof("Exiting algorithm service")
// 	os.Exit(0)
// }

// This is the function being called by the last microservice
func handleSqlDataRequest(ctx context.Context, data *pb.MicroserviceCommunication) error {
	logger.Info("Start handleSqlDataRequest")
	switch data.Type {
	case "sqlDataRequest":
		logger.Sugar().Info("switching on sqlDataRequest")

		// Unpack the metadata
		metadata := data.Metadata

		// Print each metadata field
		logger.Sugar().Debugf("Length metadata: %s", strconv.Itoa(len(metadata)))
		for key, value := range metadata {
			fmt.Printf("Key: %s, Value: %+v\n", key, value)
		}

		// Unpack the data
		// dataStruct := data.Data
		sqlDataRequest := &pb.SqlDataRequest{}
		if err := data.OriginalRequest.UnmarshalTo(sqlDataRequest); err != nil {
			logger.Sugar().Errorf("Failed to unmarshal sqlDataRequest message: %v", err)
		}

		c := pb.NewMicroserviceClient(conn)

		// ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		// defer cancel()
		// Just pass on the data for now...
		c.SendData(ctx, data)
		close(stop)
	default:
		logger.Sugar().Errorf("Unknown message type: %v", data.Type)
	}
	return nil
}
