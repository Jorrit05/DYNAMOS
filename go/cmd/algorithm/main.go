package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	logger      = lib.InitLogger(logLevel)
	conn        *grpc.ClientConn
	lastService bool = false
	stop             = make(chan struct{}) // channel to tell the server to stop

)

type server struct {
	pb.UnimplementedMicroserviceServer
}

// This is the function being called by the last microservice
func (s *server) SendData(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error) {
	logger.Info("Start SendData")
	switch data.Type {
	case "sqlDataRequest":
		logger.Sugar().Info("switching on sqlDataRequest")

		// Unpack the metadata
		metadata := data.Metadata

		// Print each metadata field
		for key, value := range metadata {
			fmt.Printf("Key: %s, Value: %+v\n", key, value)
		}

		// Unpack the data
		dataStruct := data.Data

		// Print the entire data field
		fmt.Println(dataStruct)

		close(stop)
	default:
		logger.Sugar().Errorf("Unknown message type: %v", data.Type)
	}
	return &emptypb.Empty{}, nil
}

func StartGrpcMicroserviceServer(port int64) <-chan struct{} {
	stopped := make(chan struct{}) // channel to tell us the server has stopped

	go func() {
		logger.Sugar().Infof("Start listening on port: %v", port)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
		if err != nil {
			logger.Sugar().Fatalw("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		serverInstance := &lib.SharedServer{}

		pb.RegisterMicroserviceServer(s, &server{})
		pb.RegisterHealthServer(s, serverInstance)

		logger.Info("pb.RegisterMicroserviceServer(s, serverInstance)")

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

// func ReveiceData() {
// 	// Assuming `row` is your data fetched from the database.
// 	fields := make(map[string]*structpb.Value)
// 	fields["name"] = structpb.NewStringValue("Jorrit")
// 	fields["date_of_birth"] = structpb.NewStringValue("september")
// 	fields["job_title"] = structpb.NewStringValue("IT")
// 	fields["other"] = structpb.NewBoolValue(true)

// 	s := &structpb.Struct{Fields: fields}

// 	iets := s.Fields["other"].GetListValue().ProtoReflect().Type()
// 	fmt.Println(iets)
// 	fmt.Println("xxx")
// 	fmt.Println(s.GetFields())
// }

func main() {
	logger.Debug("Starting algorithm service")

	port, err := strconv.ParseInt(os.Getenv("DESIGNATED_GRPC_PORT"), 10, 32)
	lastServiceInt, err1 := strconv.ParseInt(os.Getenv("LAST"), 10, 32)
	if err != nil || err1 != nil {
		logger.Sugar().Fatalf("Error determining port number: %v", err)
	}
	stopped := StartGrpcMicroserviceServer(port)

	// Connect to following service.
	if lastServiceInt > 0 {
		// We are the last service
		lastService = true
	} else {
		conn = lib.GetGrpcConnection(grpcAddr + strconv.Itoa(int(port+1)))
		defer conn.Close()
	}
	logger.Info("started GRPC server")

	// Wait for GRPC server to shutdown gracefully before quiting
	<-stopped
	logger.Sugar().Infof("Exiting algorithm service")
	os.Exit(0)
}
