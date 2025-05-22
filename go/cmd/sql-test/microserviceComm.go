package main

import (
	"context"
	"encoding/json"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func sendMicroserviceComm(c pb.RabbitMQClient) (context.Context, error) {
	ctx := context.Background()
	data := []byte(`{
    "type": "sqlDataRequest",
    "query" : "SELECT p.Geslacht, s.Salschal FROM Personen p JOIN Aanstellingen s ON p.Unieknr = s.Unieknr",
    "graph" : false,
    "algorithm" : "average",
    "algorithmColumns" : {
        "Geslacht" : "Aanst_22, Gebdat"
    },
    "user": {
        "id": "1234",
        "userName": "jorrit.stutterheim@cloudnation.nl"
    },
    "requestMetadata": {
        "jobId": "test"
    }
}`)

	var sqlDataRequest *pb.SqlDataRequest
	err := json.Unmarshal(data, &sqlDataRequest)
	if err != nil {
		logger.Sugar().Errorf("error unmarshalling JSON: %v", err)
	}

	msComm := &pb.MicroserviceCommunication{}
	msComm.RequestMetadata = &pb.RequestMetadata{}

	msComm.Type = "microserviceCommunication"
	msComm.RequestMetadata.DestinationQueue = "caller"
	msComm.RequestMetadata.ReturnAddress = serviceName
	msComm.RequestType = "sqlDataRequest"

	any, err := anypb.New(sqlDataRequest)
	if err != nil {
		logger.Sugar().Error(err)
		return ctx, err
	}

	msComm.OriginalRequest = any
	msComm.RequestMetadata.CorrelationId = "correlationId"
	logger.Sugar().Debugf("msComm.RequestMetadata.DestinationQueue: %v", msComm.RequestMetadata.DestinationQueue)
	c.SendMicroserviceComm(ctx, msComm)

	return ctx, nil
}
