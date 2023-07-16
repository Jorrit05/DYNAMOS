package main

import (
	"context"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func handleSqlDataRequest(ctx context.Context, data *pb.MicroserviceCommunication) error {
	logger.Debug("Start msCommunication handleSqlDataRequest")
	sqlDataRequest := &pb.SqlDataRequest{}
	if err := data.OriginalRequest.UnmarshalTo(sqlDataRequest); err != nil {
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
	_, err = send(ctx, message, data.RequestMetada.ReturnAddress)
	if err != nil {
		logger.Sugar().Errorf("Error sending microserviceCommunication to agent: %v", err)
		return err
	}
	close(stop)

	// Graceful exit
	return nil
}

func (s *server) SendShutdownSignal(ctx context.Context, in *pb.ShutDown) (*emptypb.Empty, error) {
	logger.Debug("Starting SendShutdownSignal")

	return &emptypb.Empty{}, nil
}
