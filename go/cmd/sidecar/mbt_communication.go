package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/etcd"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MbtMessage struct {
	Type string
	Body []byte
}

func SendMSCommDataToTestingQueue(ctx context.Context, data *pb.MicroserviceCommunication, s *serverInstance) (*emptypb.Empty, error) {
	logger.Sugar().Debug("Starting lib.SendDataToTestingQueue")

	ctx, span := trace.StartSpan(ctx, "sidecar SendDataToTestingQueue/func:")

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

	sendToTestQueue(ctx, msg, s)
	logger.Sugar().Debug("Ending lib.sendToTestQueue")
	return &emptypb.Empty{}, nil
}

func sendToTestQueue(ctx context.Context, msg amqp.Publishing, s *serverInstance) *emptypb.Empty {
	logger.Sugar().Debug("Starting lib.sendToTestQueue")
	retryOpts := etcd.DefaultRetryOptions

	if !useMbtQueue {
		logger.Sugar().Debug("MBT Testing Queue is disabled, messages will not be forwarded to mbt_adapter ")
		return &emptypb.Empty{}
	}
	logger.Sugar().Debug("Using MBT Queue")

	testing_queue_exists := queueExists(s.channel, mbtTestingQueueName)
	logger.Sugar().Debugf("Testing queue exists? %v", testing_queue_exists)
	if !testing_queue_exists {
		logger.Sugar().Errorf("Testing queue %s is not present", mbtTestingQueueName)
		return &emptypb.Empty{}
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
		msg.Headers["jsonTrace"] = scJson
	}

	msg.Headers["binaryTrace"] = binarySc // Create a context with a timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	logger.Sugar().Infof("Sending message to queue %v with message %v", mbtTestingQueueName, msg)
	err := s.channel.PublishWithContext(
		timeoutCtx,
		"",
		mbtTestingQueueName,
		true,
		false,
		prepareMsgForTransfer(msg),
	)
	if err != nil {
		logger.Sugar().Errorf("Error sending sendToTestQueue: %v", err)
		// return &emptypb.Empty{}, err
	}

	logger.Sugar().Debug("Ending lib.SendMSCommDataToTestingQueue")

	return &emptypb.Empty{}
}

func prepareMsgForTransfer(msg amqp.Publishing) amqp.Publishing {
	message := MbtMessage{
		msg.Type,
		msg.Body,
	}

	marshalled_message, _ := json.Marshal(message)

	msg.Body = marshalled_message
	return msg
}

func queueExists(ch *amqp.Channel, queueName string) bool {
	_, err := ch.QueueDeclarePassive(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	return err == nil
}
