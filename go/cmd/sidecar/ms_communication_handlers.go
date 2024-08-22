// Package main, that implements 'sidecar' functionality
//
// File: ms_communication_handlers.go
//
// Description:
// This file contains functions to send a message to the specified AMQ target queue. This
// is specifically used for a microservice chain to send the message to the new destination after
// processing on the current data steward. The function will ensure the sidecar exits
// after sending this message
//
// Notes:
// This function can perhaps be more streamlined with the generic 'send' function, clearly delining that
// the sidecar exits afterwards.
//
// Author: Jorrit Stutterheim

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

// SendDataThroughAMQ sends a message to the specified AMQ target queue.
// After sending the message, it closes the stop channel to signal the end of the program,
// the sidecar will exit.
//
// Parameters:
// - ctx: Context
// - data: message to send in the form of a MicroserviceCommunication protobuf message.
// - s:  serverInstance pointer, used to access the AMQ channel.
//
// Returns:
// - An empty protobuf message.
// - An error if the message could not be sent, otherwise nil.
func SendDataThroughAMQ(ctx context.Context, data *pb.MicroserviceCommunication, s *serverInstance) (*emptypb.Empty, error) {
	logger.Debug("Starting lib.SendDataThroughAMQ")

	ctx, span := trace.StartSpan(ctx, "sidecar SendDataThroughAMQ/func:")

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

	err = s.channel.PublishWithContext(timeoutCtx, exchangeName, data.RequestMetadata.ReturnAddress, true, false, msg)
	if err != nil {
		logger.Sugar().Errorf("Error sending microserviceCommunication: %v", err)
		// return &emptypb.Empty{}, err
	}

	logger.Debug("Ending lib.SendDataThroughAMQ")
	return &emptypb.Empty{}, nil
}
