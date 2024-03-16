package main

import (
	"context"
	"os"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	"github.com/Jorrit05/DYNAMOS/pkg/msinit"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"github.com/gogo/protobuf/jsonpb"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	logger      = lib.InitLogger(logLevel)
	config      *msinit.Configuration
	COORDINATOR = make(chan struct{})
)

// Main function
func main() {
	logger.Debug("Starting algorithm service")

	oce, err := lib.InitTracer(serviceName)
	if err != nil {
		logger.Sugar().Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	config, err = msinit.NewConfiguration(serviceName, grpcAddr, COORDINATOR, sideCarMessageHandler, sendDataHandler)
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

// This is the function being called by the last microservice
func handleSqlDataRequest(ctx context.Context, msComm *pb.MicroserviceCommunication) error {
	ctx, span := trace.StartSpan(ctx, "anonymize: handleSqlDataRequest")
	defer span.End()

	logger.Info("anonymize Start handleSqlDataRequest")

	sqlDataRequest := &pb.SqlDataRequest{}
	if err := msComm.OriginalRequest.UnmarshalTo(sqlDataRequest); err != nil {
		logger.Sugar().Errorf("Failed to unmarshal sqlDataRequest message: %v", err)
	}

	anonymizeDatesInStruct(msComm.Data)

	<-COORDINATOR

	c := pb.NewMicroserviceClient(config.GrpcConnection)
	if sqlDataRequest.Options["graph"] {
		// jsonString, _ := json.Marshal(msComm.Data)
		// msComm.Result = jsonString

		m := &jsonpb.Marshaler{}
		jsonString, _ := m.MarshalToString(msComm.Data)
		msComm.Result = []byte(jsonString)

		c.SendData(ctx, msComm)
		close(config.StopServer)
		return nil
	}

	// Process all data to make this service more realistic.

	msComm.Traces["binaryTrace"] = propagation.Binary(span.SpanContext())

	c.SendData(ctx, msComm)
	// time.Sleep(2 * time.Second)
	close(config.StopServer)
	return nil
}

func anonymizeDatesInStruct(data *structpb.Struct) {
	fieldsToAnonymize := []string{"Ingdatdv", "Gebdat"}

	for _, field := range fieldsToAnonymize {
		fieldValue, ok := data.GetFields()[field]
		if !ok {
			continue
		}

		listValues := fieldValue.GetListValue().GetValues()
		for index, value := range listValues {
			stringValue := value.GetStringValue()
			if len(stringValue) >= 4 {
				// Create a new value with modified string and set it in the slice
				newValue := structpb.NewStringValue(stringValue[:4])
				listValues[index] = newValue
			}
		}
	}
}
