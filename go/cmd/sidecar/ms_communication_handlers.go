package main

import (
	"context"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

func handleSqlDataRequest(ctx context.Context, data *pb.MicroserviceCommunication) error {
	logger.Debug("Start msCommunication handleSqlDataRequest")
	sqlDataRequest := &pb.SqlDataRequest{}
	if err := data.UserRequest.UnmarshalTo(sqlDataRequest); err != nil {
		logger.Sugar().Errorf("Failed to unmarshal sqlDataRequest message: %v", err)
	}

	// Marshaling google.protobuf.Struct to Proto wire format
	body, err := proto.Marshal(data)
	if err != nil {
		logger.Sugar().Errorf("Failed to marshal struct to proto wire format: %v", err)
		return err
	}

	message := amqp.Publishing{
		CorrelationId: sqlDataRequest.RequestMetada.CorrelationId,
		Body:          body,
		Type:          "microserviceCommunication",
	}
	_, err = send(ctx, message, data.ReturnAddress)
	if err != nil {
		logger.Sugar().Errorf("Error sending microserviceCommunication to agent: %v", err)
		return err
	}
	close(stop)

	// Graceful exit
	return nil
}
