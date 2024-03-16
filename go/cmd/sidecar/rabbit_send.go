package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	bo "github.com/cenkalti/backoff/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	channel    *amqp.Channel
	conn       *amqp.Connection
	messages   <-chan amqp.Delivery
	routingKey string
)

func send(ctx context.Context, message amqp.Publishing, target string, opts ...etcd.Option) (*emptypb.Empty, error) {
	// Start with default options
	retryOpts := etcd.DefaultRetryOptions

	// Apply any specified options
	for _, opt := range opts {
		opt(&retryOpts)
	}

	// Create a returns channel
	returns := make(chan amqp.Return)
	channel.NotifyReturn(returns)

	if message.Headers == nil {
		message.Headers = amqp.Table{}
	}

	sc := trace.FromContext(ctx).SpanContext()
	binarySc := propagation.Binary(sc)

	if retryOpts.AddJsonTrace {
		// Create a map to hold the span context values
		scMap := map[string]string{
			"TraceID": sc.TraceID.String(),
			"SpanID":  sc.SpanID.String(),
			// "TraceOptions": fmt.Sprintf("%02x", sc.TraceOptions.IsSampled()),
		}
		// Serialize the map to a JSON string
		scJson, err := json.Marshal(scMap)
		if err != nil {
			logger.Debug("ERRROR scJson MAP")
		}
		message.Headers["jsonTrace"] = scJson
	}

	message.Headers["binaryTrace"] = binarySc

	operation := func() error {
		// Log before sending message
		logger.Sugar().Infow("Sending message: ", "My routingKey", routingKey, "exchangeName", exchangeName, "target", target)

		// Create a context with a timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		defer cancel()

		err := channel.PublishWithContext(timeoutCtx, exchangeName, target, true, false, message)
		if err != nil {
			logger.Sugar().Debugf("In error chan: %v", err)
			return err
		}
		select {
		case r := <-returns:
			if r.ReplyText == "NO_ROUTE" {
				logger.Sugar().Infof("No route to job yet: %v", target)
				// This will trigger a retry
				return errors.New("no route to target")
			} else {
				logger.Sugar().Errorf("Unknown reason message returned: %v", r)
				return bo.Permanent(errors.New("unknown error"))
			}
		case <-time.After(8 * time.Second): // Timeout if no message is received in 3 seconds
			logger.Sugar().Debugf("8 seconds have passed for target: %v", target)

		}

		return nil
	}

	// Create a new exponential backoff
	backoff := bo.NewExponentialBackOff()
	backoff.InitialInterval = retryOpts.InitialInterval
	backoff.MaxInterval = retryOpts.MaxInterval
	backoff.MaxElapsedTime = retryOpts.MaxElapsedTime

	err := bo.Retry(operation, backoff)
	if err != nil {
		logger.Sugar().Errorf("Publish failed after %v seconds, err: %s", backoff.MaxElapsedTime, err)
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (s *server) SendRequestApproval(ctx context.Context, in *pb.RequestApproval) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal requestApproval failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		Body: data,
		Type: "requestApproval",
	}

	go send(ctx, message, in.DestinationQueue)
	return &emptypb.Empty{}, nil
}

func (s *server) SendValidationResponse(ctx context.Context, in *pb.ValidationResponse) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal ValidationResponse failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		Body: data,
		Type: "validationResponse",
	}

	go send(ctx, message, "orchestrator-in")
	return &emptypb.Empty{}, nil

}

func (s *server) SendCompositionRequest(ctx context.Context, in *pb.CompositionRequest) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal CompositionRequest failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		Body: data,
		Type: "compositionRequest",
	}

	go send(ctx, message, in.DestinationQueue)
	return &emptypb.Empty{}, nil

}

func (s *server) SendSqlDataRequest(ctx context.Context, in *pb.SqlDataRequest) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal requestApproval failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		CorrelationId: in.RequestMetadata.CorrelationId,
		Body:          data,
		Type:          "sqlDataRequest",
	}

	logger.Sugar().Debugf("SendSqlDataRequest destination queue: %v", in.RequestMetadata.DestinationQueue)
	go send(ctx, message, in.RequestMetadata.DestinationQueue, etcd.WithMaxElapsedTime(10*time.Second))
	return &emptypb.Empty{}, nil

}

func (s *server) SendPolicyUpdate(ctx context.Context, in *pb.PolicyUpdate) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal PolicyUpdate failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		CorrelationId: in.RequestMetadata.CorrelationId,
		Body:          data,
		Type:          "policyUpdate",
	}

	logger.Sugar().Debugf("PolicyUpdate destination queue: %s", in.RequestMetadata.DestinationQueue)
	go send(ctx, message, in.RequestMetadata.DestinationQueue, etcd.WithMaxElapsedTime(10*time.Second))
	return &emptypb.Empty{}, nil

}

// TODO: This go function is mostly to get an accurate feel for data transfer speeds.
// It's probably better to just remove the Go func in the long run
func (s *server) SendMicroserviceComm(ctx context.Context, in *pb.MicroserviceCommunication) (*emptypb.Empty, error) {
	go func(in *pb.MicroserviceCommunication) {
		data, err := proto.Marshal(in)
		if err != nil {
			logger.Sugar().Errorf("Marshal SendMicroserviceComm failed: %s", err)
			return
			// return nil, status.Error(codes.Internal, err.Error())
		}

		message := amqp.Publishing{
			CorrelationId: in.RequestMetadata.CorrelationId,
			Body:          data,
			Type:          in.Type,
		}
		logger.Sugar().Debugf("SendMicroserviceComm destination queue: %s", in.RequestMetadata.DestinationQueue)
		go send(ctx, message, in.RequestMetadata.DestinationQueue, etcd.WithMaxElapsedTime(10*time.Second), etcd.WithJsonTrace())
	}(in)
	return &emptypb.Empty{}, nil
}

func (s *server) SendTest(ctx context.Context, in *pb.SqlDataRequest) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal SendMicroserviceComm failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		CorrelationId: "in.RequestMetadata.CorrelationId",
		Body:          data,
		Type:          "testSet",
	}
	go send(ctx, message, "no existss", etcd.WithMaxElapsedTime(10*time.Second))
	return &emptypb.Empty{}, nil
}

func (s *server) SendRequestApprovalResponse(ctx context.Context, in *pb.RequestApprovalResponse) (*emptypb.Empty, error) {
	data, err := proto.Marshal(in)
	if err != nil {
		logger.Sugar().Errorf("Marshal SendRequestApprovalResponse failed: %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	// Do other stuff
	message := amqp.Publishing{
		Body: data,
		Type: "requestApprovalResponse",
	}

	go send(ctx, message, in.RequestMetadata.DestinationQueue)
	return &emptypb.Empty{}, nil
}
