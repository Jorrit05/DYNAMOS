package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
)

func createCallbackHandler(config *Configuration) func(ctx context.Context, grpcMsg *pb.RabbitMQMessage) error {
	return func(ctx context.Context, grpcMsg *pb.RabbitMQMessage) error {

		ctx, span, err := lib.StartRemoteParentSpan(ctx, serviceName+"/func: callbackHandler, procces MS", grpcMsg.Trace)
		if err != nil {
			logger.Sugar().Errorf("Error starting span: %v", err)
			return err
		}
		defer span.End()

		// logger.Debug("algorhithm span: ----------------")
		// lib.PrettyPrintSpanContext(span.SpanContext())
		// // logger.Sugar().Debugw("Type:", "MessageType", grpcMsg.Type)
		// logger.Debug("----------------")
		switch grpcMsg.Type {

		case "microserviceCommunication":

			logger.Sugar().Info("switching on microserviceCommunication")
			msComm := &pb.MicroserviceCommunication{}
			msComm.RequestMetada = &pb.RequestMetada{}

			if err := grpcMsg.Body.UnmarshalTo(msComm); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal msComm message: %v", err)
			} // Unpack the metadata

			metadata := msComm.Metadata

			// Print each metadata field
			logger.Sugar().Debugf("Length metadata: %s", strconv.Itoa(len(metadata)))
			for key, value := range metadata {
				fmt.Printf("Key: %s, Value: %+v\n", key, value)
			}

			sqlDataRequest := &pb.SqlDataRequest{}
			if err := msComm.OriginalRequest.UnmarshalTo(sqlDataRequest); err != nil {
				logger.Sugar().Errorf("Failed to unmarshal sqlDataRequest message: %v", err)
			}

			logger.Debug("---------msComm.Trace------------")

			logger.Debug(string(msComm.Trace))
			logger.Debug("---------------------")

			c := pb.NewMicroserviceClient(config.GrpcConnection)
			logger.Sugar().Debugf("desitnation queue: %v", msComm.RequestMetada.DestinationQueue)
			logger.Sugar().Debugf("ReturnAddress queue: %v", msComm.RequestMetada.ReturnAddress)

			if c == nil {
				logger.Error("C IS NUL MANNEN")
			}
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			// span.End()
			// Just pass on the data for now...
			c.SendData(ctx, msComm)
			close(stop)
		default:
			logger.Sugar().Errorf("Unknown message type: %v", grpcMsg.Type)
			return fmt.Errorf("unknown message type: %s", grpcMsg.Type)
		}

		return nil
	}
}
