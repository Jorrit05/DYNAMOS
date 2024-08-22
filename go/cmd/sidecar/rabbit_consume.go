// Package main, that implements 'sidecar' functionality
//
// File: rabbit_consume.go
//
// Description:
// This file contains the functionality to stream a incoming rabbitMQ
// message to the main container via gRPC.
//
// Notes:
// This could be up for refactoring, for Microservice Chains, we switched to not streaming the messages, but just sending them.
// Depending on if we want to keep this streaming functionality, we might need to refactor this.
//
// Author: Jorrit Stutterheim

package main

import (
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Handle incoming AMQ messages to stream from a sidecar to the main container.
//
// handleResponse handles the response received from RabbitMQ and sends it to the main container via gRPC.
// It unmarshals the received message into the provided protobuf message, creates an `anypb.Any` message from it,
// and constructs a `pb.SideCarMessage` with the necessary fields. If there are any trace headers in the RabbitMQ
// message, they are added to the `pb.SideCarMessage` as well. Finally, it sends the constructed message to the
// main container via the provided gRPC stream.
//
// Parameters:
// - msg: The RabbitMQ delivery message containing the response.
// - stream: The gRPC stream used to send the response to the main container.
// - pbMsg: The protobuf message to unmarshal the RabbitMQ message into.
//
// Returns:
// - An error if there was an error unmarshalling the message or sending it via the gRPC stream, otherwise nil.
func (s *serverInstance) handleResponse(msg amqp.Delivery, stream pb.RabbitMQ_ConsumeServer, pbMsg proto.Message) error {
	proto.Reset(pbMsg)

	if err := proto.Unmarshal(msg.Body, pbMsg); err != nil {
		logger.Sugar().Errorf("Error unmarshalling proto msg, %v", err)
		return err
	}

	any, err := anypb.New(pbMsg)
	if err != nil {
		logger.Sugar().Error(err)
		return err
	}

	grpcMsg := &pb.SideCarMessage{
		Body: any,
		Type: msg.Type,
	}
	logger.Debug("stream to main container")
	if msg.Headers != nil {
		logger.Debug("msg.Headers != nil")

		grpcMsg.Traces = make(map[string][]byte)
		value, ok := msg.Headers["jsonTrace"]
		if ok {
			logger.Debug("Adding jsonTraces")

			grpcMsg.Traces["jsonTrace"] = value.([]byte)
		}
		value, ok = msg.Headers["binaryTrace"]
		if ok {
			logger.Debug("Adding binaryTrace")
			grpcMsg.Traces["binaryTrace"] = value.([]byte)
		}
	}

	sendMutex.Lock()
	err = stream.SendMsg(grpcMsg)
	sendMutex.Unlock()

	if err != nil {
		logger.Sugar().Warnf("stream error: %v", err)
	}
	return err
}

func (s *serverInstance) handleValidationResponse(msg amqp.Delivery, stream pb.RabbitMQ_ConsumeServer) error {
	return s.handleResponse(msg, stream, &pb.ValidationResponse{})
}

func (s *serverInstance) handleRequestApprovalResponse(msg amqp.Delivery, stream pb.RabbitMQ_ConsumeServer) error {
	logger.Debug("sidecar/responses: handleRequestApprovalResponse")

	return s.handleResponse(msg, stream, &pb.RequestApproval{})
}

func (s *serverInstance) handleCompositionRequestResponse(msg amqp.Delivery, stream pb.RabbitMQ_ConsumeServer) error {
	return s.handleResponse(msg, stream, &pb.CompositionRequest{})
}

func (s *serverInstance) handleSqlDataRequest(msg amqp.Delivery, stream pb.RabbitMQ_ConsumeServer) error {
	logger.Debug("Starting handleSqlDataRequest")
	return s.handleResponse(msg, stream, &pb.SqlDataRequest{})
}

func (s *serverInstance) handleMicroserviceCommunication(msg amqp.Delivery, stream pb.RabbitMQ_ConsumeServer) error {
	logger.Debug("Starting handleMicroserviceCommunication")

	return s.handleResponse(msg, stream, &pb.MicroserviceCommunication{RequestMetadata: &pb.RequestMetadata{}})
}

func (s *serverInstance) handlePolicyUpdate(msg amqp.Delivery, stream pb.RabbitMQ_ConsumeServer) error {
	logger.Debug("Starting handlePolicyUpdate")

	return s.handleResponse(msg, stream, &pb.PolicyUpdate{RequestMetadata: &pb.RequestMetadata{}})
}

func (s *serverInstance) handleRequestApprovalToApiResponse(msg amqp.Delivery, stream pb.RabbitMQ_ConsumeServer) error {
	logger.Debug("Starting handleRequestApprovalToApiResponse")

	return s.handleResponse(msg, stream, &pb.RequestApprovalResponse{})
}
