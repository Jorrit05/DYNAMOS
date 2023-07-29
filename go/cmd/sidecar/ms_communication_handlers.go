package main

import (
	"context"
	"time"

	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opencensus.io/trace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func SendDataThroughAMQ(ctx context.Context, data *pb.MicroserviceCommunication) (*emptypb.Empty, error) {
	logger.Debug("Starting lib.SendDataThroughAMQ")

	// ctx, span, err := lib.StartRemoteParentSpan(ctx, "sidecar SendDataThroughAMQ/func:", data.Traces)
	// if err != nil {
	// 	logger.Sugar().Warnf("Error starting span: %v", err)
	// }

	ctx, span := trace.StartSpan(ctx, "sidecar SendDataThroughAMQ/func:")

	// TODO: This go function is mostly to get an accurate feel for data transfer speeds.
	// It's probably better to just remove the Go func in the long run
	// go func(data *pb.MicroserviceCommunication, stop chan struct{}) {
	// Marshaling google.protobuf.Struct to Proto wire format
	body, err := proto.Marshal(data)
	if err != nil {
		logger.Sugar().Errorf("Failed to marshal struct to proto wire format: %v", err)
		return &emptypb.Empty{}, nil
	}

	msg := amqp.Publishing{
		CorrelationId: data.RequestMetadata.CorrelationId,
		Body:          body,
		Type:          "microserviceCommunication",
		Headers:       amqp.Table{},
	}
	span.End()

	// Here I am assuming that the msCommunication will take care of tracing requirements.
	// I don't trust context cause this might be called from a Python or other language microservice
	value, ok := data.Traces["jsonTrace"]
	if ok {
		msg.Headers["jsonTrace"] = value
	}

	value, ok = data.Traces["binaryTrace"]
	if ok {
		msg.Headers["binaryTrace"] = value
	}

	// Create a context with a timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	// send(ctx, msg, data.RequestMetadata.ReturnAddress)
	err = channel.PublishWithContext(timeoutCtx, exchangeName, data.RequestMetadata.ReturnAddress, true, false, msg)
	if err != nil {
		logger.Sugar().Errorf("Error sending microserviceCommunication: %v", err)
		// return &emptypb.Empty{}, err
	}

	close(stop)
	// }(data, stop)
	// go send(ctx, msg, data.RequestMetadata.ReturnAddress)

	return &emptypb.Empty{}, nil
}
