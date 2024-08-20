package main

import (
	"context"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"github.com/gogo/protobuf/jsonpb"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/protobuf/types/known/structpb"
)

// This is the function being called by the last microservice
func handleSqlDataRequest(ctx context.Context, msComm *pb.MicroserviceCommunication) (context.Context, error) {
	ctx, span := trace.StartSpan(ctx, "anonymize: handleSqlDataRequest")
	defer span.End()

	logger.Info("anonymize Start handleSqlDataRequest")

	sqlDataRequest := &pb.SqlDataRequest{}
	if err := msComm.OriginalRequest.UnmarshalTo(sqlDataRequest); err != nil {
		logger.Sugar().Errorf("Failed to unmarshal sqlDataRequest message: %v", err)
		return ctx, err
	}

	anonymizeDatesInStruct(msComm.Data)

	if sqlDataRequest.Options["graph"] {
		// jsonString, _ := json.Marshal(msComm.Data)
		// msComm.Result = jsonString

		m := &jsonpb.Marshaler{}
		jsonString, _ := m.MarshalToString(msComm.Data)
		msComm.Result = []byte(jsonString)

		return ctx, nil
	}

	msComm.Traces["binaryTrace"] = propagation.Binary(span.SpanContext())
	return ctx, nil
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
