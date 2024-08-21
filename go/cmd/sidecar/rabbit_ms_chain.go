// Package main, that implements 'sidecar' functionality
//
// File: rabbit_ms_chain.go
//
// Description:
// This file contains the gRPC server implementations for handling Microservice Chain
// RabbitMQ consumption. Instead of streaming a message to the microservice, the message is sent
// to the microservice using the generic 'SendData' function.
//
// The flow is that a microservice needs to send a single 'InitRabbitForChain' request to the sidecar to
// create the required AMQ connection. The sidecar will then start consuming messages from
// the specified queue. After consumption the microservice can send a 'StopReceivingRabbit' request to
// exit the sidecar.
//
// Notes:
// This works at the moment for Python microservices, this needs to be the new standard so that Go microservices
// are handled the same way.Go microservices still use the old streaming method in rabbit_chain_consume.go.
//
// Author: Jorrit Stutterheim

package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *serverInstance) StopReceivingRabbit(ctx context.Context, in *pb.StopRequest) (*emptypb.Empty, error) {
	logger.Sugar().Infow("StopReceivingRabbit Received stop request")

	// Cancel the context
	if s.consumerManager != nil {
		// s.consumerManager.cancel()

		// Signal the stop channel
		close(s.consumerManager.stopChan)
	}

	return &emptypb.Empty{}, nil
}

// InitRabbitForChain sets up a RabbitMQ connection and starts consuming messages from the specified queue.
//
// Parameters:
// - ctx: Context
// - in: ChainRequest containing the service name, routing key, and port.
//
// Returns:
// - An empty protobuf message.
// - An error if the connection could not be set up, otherwise nil.
func (s *serverInstance) InitRabbitForChain(ctx context.Context, in *pb.ChainRequest) (*emptypb.Empty, error) {
	logger.Sugar().Infow("InitRabbitForChain Received:", "Servicename", in.ServiceName, "RoutingKey", in.RoutingKey, "Port", in.Port)

	// Call the SetupConnection function and handle the message consumption inside this function
	_, conn, channel, err := setupConnection(in.RoutingKey, in.RoutingKey, in.QueueAutoDelete)
	s.channel = channel
	s.conn = conn

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	go func() {

		msConnection := lib.GetGrpcConnection("localhost:" + strconv.Itoa(int(in.Port)))

		msClient := pb.NewMicroserviceClient(msConnection)

		// Create a new context because the context from InitRabbitForChain is canceled when the function returns
		consumeCtx, cancel := context.WithCancel(context.Background())

		s.consumerManager = &ConsumerManager{
			stopChan: make(chan struct{}),
			cancel:   cancel,
		}

		err = ChainConsume(consumeCtx, in.RoutingKey, true, msClient, s.consumerManager.stopChan, s)
		if err != nil {
			logger.Sugar().Errorf("Error in chainconsume: ", codes.Internal, err.Error())
		}
	}()

	return &emptypb.Empty{}, nil
}

// ChainConsume consumes 'MicroserviceCommunication' messages from a specified queue.
// Will exit on when a stop signal is received.
//
// Parameters:
// - ctx: Context
// - queueName: Name of the queue to consume from.
// - autoAck: Whether to automatically acknowledge messages.
// - msClient: MicroserviceClient to send messages to.
// - stopChan: Channel to receive stop signals on.
// - serverInstance: Pointer to the serverInstance to access the channel.
//
// Returns:
// - An error if there was a problem consuming messages or handling them, otherwise nil.
func ChainConsume(ctx context.Context, queueName string, autoAck bool, msClient pb.MicroserviceClient, stopChan chan struct{}, serverInstance *serverInstance) error {
	logger.Sugar().Infow("Start ChainConsume", "queueName:", queueName)

	messages, err := serverInstance.channel.Consume(
		queueName, // queue
		"",        // consumer
		autoAck,   // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	logger.Sugar().Infow("Started consuming", "queue", queueName)

	for {
		select {
		case msg := <-messages:
			// Handle the message
			logger.Sugar().Debugw("switchin:", "msg,Type", msg.Type)
			switch msg.Type {

			case "microserviceCommunication":
				msComm := &pb.MicroserviceCommunication{}
				msComm.RequestMetadata = &pb.RequestMetadata{}

				if err := proto.Unmarshal(msg.Body, msComm); err != nil {
					logger.Sugar().Errorf("Error unmarshalling msComm msg, %v", err)
					return err
				}

				logger.Debug("Send Mscomm to main container")
				if msg.Headers != nil {
					logger.Debug("msg.Headers != nil")

					msComm.Traces = make(map[string][]byte)
					value, ok := msg.Headers["jsonTrace"]
					if ok {
						logger.Debug("Adding jsonTraces")

						msComm.Traces["jsonTrace"] = value.([]byte)
					}

					value, ok = msg.Headers["binaryTrace"]
					if ok {
						logger.Debug("Adding binaryTrace")
						msComm.Traces["binaryTrace"] = value.([]byte)
					}
				} else {
					logger.Debug("msg.Headers == nil")
				}

				resp, err := msClient.SendData(ctx, msComm)
				if err != nil {
					logger.Sugar().Errorf("Error calling SendData: %v", err)
				} else {
					// Will normally be nil
					logger.Sugar().Debugf("SendData response: %v", resp)
				}
			default:
				logger.Sugar().Errorf("Unknown message type: %s", msg.Type)
				logger.Sugar().Errorf("Message: %v", msg)
				return status.Error(codes.Unknown, fmt.Sprintf("Unknown message: %v", msg))
			}
		case <-stopChan:
			logger.Sugar().Info("Received stop signal, exiting ChainConsume")

			close(stop)
			return nil
		}
	}
}
