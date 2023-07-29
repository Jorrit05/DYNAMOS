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
