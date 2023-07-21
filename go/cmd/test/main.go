package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Jorrit05/DYNAMOS/pkg/lib"
	pb "github.com/Jorrit05/DYNAMOS/pkg/proto"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	logger      = lib.InitLogger(zap.DebugLevel)
	conn        *grpc.ClientConn
	serviceName string = "test"
	grpcAddr           = "localhost:50052"
)

func main() {
	defer logger.Sync() // flushes buffer, if any
	lib.InitTracer("test")
	//
	ctx, span := trace.StartSpan(context.Background(), "test span")
	defer span.End()
	sc := trace.FromContext(ctx).SpanContext()
	lib.PrettyPrintSpanContext(sc)
	conn = lib.GetGrpcConnection(grpcAddr, serviceName)
	defer conn.Close()
	c := lib.InitializeSidecarMessaging(conn, &pb.InitRequest{ServiceName: fmt.Sprintf("%s-in", serviceName), RoutingKey: fmt.Sprintf("%s-in", serviceName), QueueAutoDelete: false})

	msComm := &pb.MicroserviceCommunication{}
	msComm.RequestMetada = &pb.RequestMetada{}
	msComm.RequestMetada.DestinationQueue = "Test"
	msComm.Type = "microserviceCommunication"
	// Create a map to hold the span context values
	scMap := map[string]string{
		"TraceID": sc.TraceID.String(),
		"SpanID":  sc.SpanID.String(),
		// "TraceOptions": fmt.Sprintf("%02x", sc.TraceOptions.IsSampled()),
	}
	scJson, err := json.Marshal(scMap)
	if err != nil {
		logger.Debug("ERRROR scJson MAP")
	}
	msComm.Trace = scJson
	c.SendMicroserviceComm(ctx, msComm)
	time.Sleep(8 * time.Second)
	logger.Info("exit test")
}

// MapCarrier is a type that can carry context in a map and
// it implements propagation.TextMapCarrier
type MapCarrier map[string]string

// Get returns the value associated with the passed key.
func (c MapCarrier) Get(key string) string {
	return c[key]
}

// Set stores the key-value pair.
func (c MapCarrier) Set(key string, value string) {
	c[key] = value
}

// Keys lists the keys stored in this carrier.
func (c MapCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
