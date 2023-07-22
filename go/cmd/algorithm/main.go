package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/Jorrit05/DYNAMOS/pkg/msinit"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
)

var (
	logger = lib.InitLogger(logLevel)
	config *msinit.Configuration
)

// Main function
func main() {
	logger.Debug("Starting algorithm service")

	_, err := lib.InitTracer(serviceName)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	config, err = msinit.NewConfiguration(serviceName, grpcAddr, sideCarMessageHandler, sendDataHandler)
	if err != nil {
		logger.Sugar().Fatalf("%v", err)
	}
	defer config.CloseConnection()

	<-config.Stopped
	logger.Sugar().Infof("Exiting algorithm service")
	os.Exit(0)
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

// This is the function being called by the last microservice
func handleSqlDataRequest(ctx context.Context, data *pb.MicroserviceCommunication) error {
	ctx, span := trace.StartSpan(ctx, "handleSqlDataRequest")
	defer span.End()

	logger.Info("Start handleSqlDataRequest")
	// Unpack the metadata
	metadata := data.Metadata

	// Print each metadata field
	logger.Sugar().Debugf("Length metadata: %s", strconv.Itoa(len(metadata)))
	for key, value := range metadata {
		fmt.Printf("Key: %s, Value: %+v\n", key, value)
	}

	sqlDataRequest := &pb.SqlDataRequest{}
	if err := data.OriginalRequest.UnmarshalTo(sqlDataRequest); err != nil {
		logger.Sugar().Errorf("Failed to unmarshal sqlDataRequest message: %v", err)
	}

	c := pb.NewMicroserviceClient(config.GrpcConnection)
	// // Just pass on the data for now...
	c.SendData(ctx, data)

	close(config.StopServer)
	return nil
}
