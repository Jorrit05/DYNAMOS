package main

import (
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Handle incoming AMQ messages.
// Stream functionality to send responses to the 'main' container. Create a SideCarMessage type by making sure the type is easily accessible
// and package the body.
func (s *server) handleResponse(msg amqp.Delivery, stream pb.SideCar_ConsumeServer, pbMsg proto.Message) error {
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

func (s *server) handleValidationResponse(msg amqp.Delivery, stream pb.SideCar_ConsumeServer) error {
	return s.handleResponse(msg, stream, &pb.ValidationResponse{})
}

func (s *server) handleRequestApprovalResponse(msg amqp.Delivery, stream pb.SideCar_ConsumeServer) error {
	logger.Debug("sidecar/responses: handleRequestApprovalResponse")

	return s.handleResponse(msg, stream, &pb.RequestApproval{})
}

func (s *server) handleCompositionRequestResponse(msg amqp.Delivery, stream pb.SideCar_ConsumeServer) error {
	return s.handleResponse(msg, stream, &pb.CompositionRequest{})
}

func (s *server) handleSqlDataRequest(msg amqp.Delivery, stream pb.SideCar_ConsumeServer) error {
	logger.Debug("Starting handleSqlDataRequest")
	return s.handleResponse(msg, stream, &pb.SqlDataRequest{})
}

func (s *server) handleMicroserviceCommunication(msg amqp.Delivery, stream pb.SideCar_ConsumeServer) error {
	logger.Debug("Starting handleMicroserviceCommunication")

	return s.handleResponse(msg, stream, &pb.MicroserviceCommunication{RequestMetadata: &pb.RequestMetadata{}})
}

func (s *server) handlePolicyUpdate(msg amqp.Delivery, stream pb.SideCar_ConsumeServer) error {
	logger.Debug("Starting handlePolicyUpdate")

	return s.handleResponse(msg, stream, &pb.PolicyUpdate{RequestMetadata: &pb.RequestMetadata{}})
}

func (s *server) handleRequestApprovalToApiResponse(msg amqp.Delivery, stream pb.SideCar_ConsumeServer) error {
	logger.Debug("Starting handleRequestApprovalToApiResponse")

	return s.handleResponse(msg, stream, &pb.RequestApprovalResponse{})
}
