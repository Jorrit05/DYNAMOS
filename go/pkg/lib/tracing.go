package lib

import (
	"context"
	"fmt"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

func InitTracer(serviceName string) (*ocagent.Exporter, error) {
	ocagentHost := os.Getenv("OC_AGENT_HOST")
	if ocagentHost == "" {
		return nil, fmt.Errorf("env OC_AGENT_HOST not declared")
	}

	oce, err := ocagent.NewExporter(
		ocagent.WithInsecure(),
		ocagent.WithReconnectionPeriod(5*time.Second),
		ocagent.WithAddress(ocagentHost),
		ocagent.WithServiceName(serviceName),
	)

	trace.RegisterExporter(oce)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	return oce, err
}

// StartRemoteParentSpan starts a new span with a remote parent span context.
// If the parentTraceMap contains a "binaryTrace" key, it deserializes the span context
// and starts a new span with the remote parent span context. Otherwise, it starts a new
// span without a parent.
//
// Parameters:
//   - ctx: The context.Context to use for starting the span.
//   - name: The name of the span.
//   - parentTraceMap: A map containing the parent trace information.
//
// Returns:
//   - context.Context: The updated context with the new span.
//   - *trace.Span: The newly started span.
//   - error: An error if the span context is invalid.
//
// Example usage:
//
//	ctx, span, err := StartRemoteParentSpan(ctx, "mySpan", parentTraceMap)
//	if err != nil {
//	  log.Fatal(err)
//	}
//	defer span.End()
func StartRemoteParentSpan(ctx context.Context, name string, parentTraceMap map[string][]byte) (context.Context, *trace.Span, error) {
	parentTrace, ok := parentTraceMap["binaryTrace"]
	if !ok {
		logger.Warn("no binaryTrace in map")
		ctx, span := trace.StartSpan(ctx, name)
		return ctx, span, nil
	}

	// Deserialize the span context
	spanContext, ok := propagation.FromBinary(parentTrace)
	if !ok {
		return nil, nil, fmt.Errorf("invalid span context")
	}

	ctx, span := trace.StartSpanWithRemoteParent(ctx, name, spanContext)
	// logger.Sugar().Debugf("Trace ID remote parent span: %v", span.SpanContext().TraceID)
	return ctx, span, nil
}

func PrettyPrintSpanContext(ctx trace.SpanContext) {
	fmt.Printf("Trace ID: %s\n", ctx.TraceID.String())
	fmt.Printf("Span ID: %s\n", ctx.SpanID.String())
	fmt.Printf("Trace options: %v\n", ctx.TraceOptions)
	fmt.Printf("Trace IsSampled: %v\n", ctx.TraceOptions.IsSampled())
}
