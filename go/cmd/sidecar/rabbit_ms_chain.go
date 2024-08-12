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

func (s *server) StopReceivingRabbit(ctx context.Context, in *pb.StopRequest) (*emptypb.Empty, error) {
	logger.Sugar().Infow("StopReceivingRabbit Received stop request")

	// Cancel the context
	if s.consumerManager != nil {
		s.consumerManager.cancel()

		// Signal the stop channel
		close(s.consumerManager.stopChan)
	}

	return &emptypb.Empty{}, nil
}

func (s *server) InitRabbitForChain(ctx context.Context, in *pb.ChainRequest) (*emptypb.Empty, error) {
	logger.Sugar().Infow("InitRabbitForChain Received:", "Servicename", in.ServiceName, "RoutingKey", in.RoutingKey, "Port", in.Port)

	var err error
	// Call the SetupConnection function and handle the message consumption inside this function
	_, conn, channel, err = setupConnection(in.ServiceName, in.RoutingKey, in.QueueAutoDelete)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	msConnection := lib.GetGrpcConnection("localhost:" + strconv.Itoa(int(in.Port)))

	msClient := pb.NewMicroserviceClient(msConnection)

	// Create a new context because the context from InitRabbitForChain is canceled when the function returns
	consumeCtx, cancel := context.WithCancel(context.Background())
	s.consumerManager = &ConsumerManager{
		stopChan: make(chan struct{}),
		cancel:   cancel,
	}

	go func() {
		err = ChainConsume(consumeCtx, in.ServiceName, true, msClient, s.consumerManager.stopChan)

		if err != nil {
			logger.Sugar().Errorf("Error in chainconsume: ", codes.Internal, err.Error())
		}
	}()

	return &emptypb.Empty{}, nil
}

func ChainConsume(ctx context.Context, queueName string, autoAck bool, msClient pb.MicroserviceClient, stopChan chan struct{}) error {
	var err error
	logger.Sugar().Infow("Start ChainConsume")

	messages, err = channel.Consume(
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

	logger.Sugar().Infof("Started consuming from queue `%s`", queueName)

	for {
		select {
		case msg := <-messages:
			// Handle the message
			logger.Sugar().Debugw("switchin: ", "msg,Type", msg.Type)
			switch msg.Type {
			case "microserviceCommunication":
				msComm := &pb.MicroserviceCommunication{}
				msComm.RequestMetadata = &pb.RequestMetadata{}

				if err := proto.Unmarshal(msg.Body, msComm); err != nil {
					logger.Sugar().Errorf("Error unmarshalling msComm msg, %v", err)
					return err
				}

				resp, err := msClient.SendData(ctx, msComm)
				if err != nil {
					logger.Sugar().Errorf("Error calling SendData: %v", err)
				} else {
					logger.Sugar().Debugf("SendData response: %v", resp)
				}
			default:
				logger.Sugar().Errorf("Unknown message type: %s", msg.Type)
				return status.Error(codes.Unknown, fmt.Sprintf("Unknown message type: %s", msg.Type))
			}
		case <-stopChan:
			logger.Sugar().Info("Received stop signal, exiting ChainConsume")
			channel.Close()
			return nil
		}
	}
}
