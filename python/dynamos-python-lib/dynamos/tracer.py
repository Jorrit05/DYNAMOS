"""
Package dynamos, implements functionality for handling Microservice chains in Python.

File: tracer.py

Description:
Simple generic tracer initiation.

Notes:
Some problems here.

Author: Jorrit Stutterheim
"""

from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

from opentelemetry.trace.span import TraceFlags, TraceState
from opentelemetry.trace.propagation import set_span_in_context


# Service name is required for most backends
def InitTracer(service_name : str, tracing_host : str):
    resource = Resource(attributes={
        SERVICE_NAME: service_name
    })

    provider = TracerProvider(resource=resource)
    processor = BatchSpanProcessor(OTLPSpanExporter(endpoint=tracing_host, insecure=True))
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)

    return trace.get_tracer(f"{service_name}.tracer")

# Mirrors Go's StartRemoteParentSpan (see go/pkg/lib/tracing.go).
#
# Starts a new span using a remote parent span context, derived from binary data.
# This is used to maintain trace continuity across service boundaries.
#
# Parameters:
# - tracer: The OpenTelemetry tracer instance to use.
# - span_name: The name for the new span.
# - trace_map: A dictionary expected to contain a key "binaryTrace" with bytes.
#
# Returns:
# - ctx: A new context with the parent span set.
# - span: The newly started child span.
def start_remote_parent_span(tracer, span_name: str, trace_map: dict):
    binary_trace = trace_map.get("binaryTrace")

    # If no trace context was provided, start a new root span.
    if not binary_trace:
        span = tracer.start_span(span_name)
        ctx = set_span_in_context(span)
        return ctx, span

    # Attempt to extract the trace ID (16 bytes), span ID (8 bytes), and optional trace flags (1 byte).
    try:
        trace_id = int.from_bytes(binary_trace[0:16], byteorder="big")
        span_id = int.from_bytes(binary_trace[16:24], byteorder="big")
        trace_flags = TraceFlags(binary_trace[24]) if len(binary_trace) > 24 else TraceFlags(1)
    except Exception as e:
        raise ValueError(f"Invalid binaryTrace format: {e}")
    # Construct a remote SpanContext from the extracted fields.
    span_context = SpanContext(
        trace_id=trace_id,
        span_id=span_id,
        is_remote=True,
        trace_flags=trace_flags,
        trace_state=TraceState()  # Empty state unless custom headers are used
    )

    # Wrap the remote context in a NonRecordingSpan and attach it to a new context.
    parent = NonRecordingSpan(span_context)
    ctx = set_span_in_context(parent)

    # Start a new span using the parent context.
    span = tracer.start_span(span_name, context=ctx)
    return ctx, span


def pretty_print_span_context(span):
    ctx = span.get_span_context()
    print(f"Trace ID: {format(ctx.trace_id, '032x')}")
    print(f"Span ID:   {format(ctx.span_id, '016x')}")
    print(f"Sampled:   {ctx.trace_flags.sampled}")
