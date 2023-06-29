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
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	logger      = lib.InitLogger(logLevel)
	conn        *grpc.ClientConn
	lastService bool = false
)

type server struct {
	pb.UnimplementedSideCarServer
	pb.UnimplementedEtcdServer
	pb.UnimplementedMicroserviceServer
}

func StartGrpcMicroserviceServer(port int64) {
	logger.Sugar().Infof("Start listening on port: %v", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		logger.Sugar().Fatalw("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	serverInstance := &server{}

	pb.RegisterMicroserviceServer(s, serverInstance)

	if err := s.Serve(lis); err != nil {
		logger.Sugar().Fatalw("failed to serve: %v", err)
	}
}

func (s *server) SendData(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error) {
	switch data.Type {
	case "sqlDataRequest":
		logger.Sugar().Debug("switching on sqlDataRequest")
		// Callback funcs
		os.Exit(0)

	default:
		// Respond with a 405 'Method Not Allowed' HTTP response if the method isn't supported
		logger.Sugar().Errorf("Unknown message type: %v", data.Type)
	}
	return &emptypb.Empty{}, nil
}

func ReveiceData() {
	// Assuming `row` is your data fetched from the database.
	fields := make(map[string]*structpb.Value)
	fields["name"] = structpb.NewStringValue("Jorrit")
	fields["date_of_birth"] = structpb.NewStringValue("september")
	fields["job_title"] = structpb.NewStringValue("IT")
	fields["other"] = structpb.NewBoolValue(true)

	s := &structpb.Struct{Fields: fields}

	iets := s.Fields["other"].GetListValue().ProtoReflect().Type()
	fmt.Println(iets)
	fmt.Println("xxx")
	fmt.Println(s.GetFields())
}

func main() {
	logger.Debug("Starting algorithm service")

	port, err := strconv.ParseInt(os.Getenv("ORDER"), 10, 32)
	lastServiceInt, err1 := strconv.ParseInt(os.Getenv("LAST"), 10, 32)
	if err != nil || err1 != nil {
		logger.Sugar().Fatalf("Error determining port number: %v", err)
	}
	go StartGrpcMicroserviceServer(port)

	// Connect to following service.
	if lastServiceInt > 0 {
		// We are the last service
		lastService = true
	} else {
		conn = lib.GetGrpcConnection(grpcAddr + strconv.Itoa(int(port+1)))
		defer conn.Close()
	}

	// Do work

	time.Sleep(20 * time.Second)
	logger.Info("Exiting algorithm service")
	os.Exit(0)
}
