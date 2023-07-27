package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/Jorrit05/DYNAMOS/pkg/msinit"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	logger = lib.InitLogger(logLevel)
	config *msinit.Configuration
)

// Main function
func main() {
	logger.Debug("Starting algorithm service")

	oce, err := lib.InitTracer(serviceName)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	config, err = msinit.NewConfiguration(serviceName, grpcAddr, sideCarMessageHandler, sendDataHandler)
	if err != nil {
		logger.Sugar().Fatalf("%v", err)
	}

	<-config.Stopped
	logger.Sugar().Infof("Wait 2 seconds before ending algorithm service")

	oce.Flush()
	time.Sleep(2 * time.Second)
	oce.Stop()
	config.CloseConnection()
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
func handleSqlDataRequest(ctx context.Context, data *pb.MicroserviceCommunication, config *msinit.Configuration) error {
	ctx, span := trace.StartSpan(ctx, "handleSqlDataRequest")
	defer span.End()

	logger.Info("Start handleSqlDataRequest")
	// Unpack the metadata
	metadata := data.Metadata
	// fields := make(map[string]*structpb.Value)
	dataField := data.GetData()
	// Get the "Functcat" field from the struct
	functcatValue := dataField.Fields["HOOPgeb"]

	// Check if it's a ListValue
	if functcatValue != nil {
		if listValue, ok := functcatValue.Kind.(*structpb.Value_ListValue); ok {
			// Iterate over the Values in the ListValue
			for _, item := range listValue.ListValue.GetValues() {
				// item is a *structpb.Value, so we need to get the actual value using one of its getter methods
				switch v := item.Kind.(type) {
				case *structpb.Value_StringValue:
					fmt.Printf("String value: %s\n", v.StringValue)
				case *structpb.Value_NumberValue:
					fmt.Printf("Number value: %f\n", v.NumberValue)
				case *structpb.Value_BoolValue:
					fmt.Printf("Bool value: %v\n", v.BoolValue)
				// etc. for other possible types
				default:
					fmt.Printf("Other value: %v\n", v)
				}
			}
		}
	}

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
	// time.Sleep(1 * time.Second)
	close(config.StopServer)
	return nil
}
