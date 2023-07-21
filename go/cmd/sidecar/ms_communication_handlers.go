package main

import (
	"context"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// TODO: Maybe I can remove this in favor of generic send functoin? I think I already added everything in
// the MicroserviceCommunication metadata./
func handleSqlDataRequest(ctx context.Context, data *pb.MicroserviceCommunication) error {
	logger.Debug("Start msCommunication handleSqlDataRequest")

	// ctx, span, err := lib.StartRemoteParentSpan(ctx, "/func: handleSqlDataRequest", data.Trace)
	// if err != nil {
	// 	logger.Sugar().Errorf("Error starting span: %v", err)
	// 	return err
	// }

	// defer span.End()
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

	msg := amqp.Publishing{
		CorrelationId: sqlDataRequest.RequestMetada.CorrelationId,
		Body:          body,
		Type:          "microserviceCommunication",
		Headers:       amqp.Table{},
	}

	if data.Trace != nil {
		logger.Debug("handleSqlDataRequest: adding trace data to request")
		msg.Headers["trace"] = data.TraceTwo
		// spanContext, _ := propagation.FromBinary(data.Trace)
		// lib.PrettyPrintSpanContext(spanContext)
	}

	logger.Sugar().Debugf("Send to sqlREquest in msComm to: %v", data.RequestMetada.ReturnAddress)

	// Create a context with a timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	// TODO: THis is a hot mess
	err = channel.PublishWithContext(timeoutCtx, exchangeName, data.RequestMetada.ReturnAddress, true, false, msg)
	if err != nil {
		logger.Sugar().Debugf("In error chan: %v", err)
		return err
	}

	// _, err = send(ctx, msg, data.RequestMetada.ReturnAddress)
	// if err != nil {
	// 	logger.Sugar().Errorf("Error sending microserviceCommunication to agent: %v", err)
	// 	return err
	// }
	close(stop)

	// Graceful exit
	return nil
}

func (s *server) SendShutdownSignal(ctx context.Context, in *pb.ShutDown) (*emptypb.Empty, error) {
	logger.Debug("Starting SendShutdownSignal")

	return &emptypb.Empty{}, nil
}
