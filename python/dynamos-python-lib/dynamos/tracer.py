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
from opentelemetry.sdk.trace.sampling import ALWAYS_ON

import json
from opentelemetry.trace import SpanContext, TraceFlags, TraceState, NonRecordingSpan
from opentelemetry.trace.propagation import set_span_in_context

# Service name is required for most backends
# Service to initialize the tracer for a specific microservice.
def InitTracer(service_name : str, tracing_host : str):
    # Define the service-level resource (metadata for traces)
    resource = Resource(attributes={
        SERVICE_NAME: service_name
    })

    # Set up the TracerProvider 
    provider = TracerProvider(resource=resource)

    # Configure the OTLP gRPC exporter and batch processor
    processor = BatchSpanProcessor(
        OTLPSpanExporter(endpoint=tracing_host, insecure=True)
    )
    provider.add_span_processor(processor)

    # Register this provider as the global tracer provider
    trace.set_tracer_provider(provider)

    # Return a tracer scoped to this service
    return trace.get_tracer(f"{service_name}.tracer")

# Function used to debug a span
def pretty_print_span_context(span):
    ctx = span.get_span_context()
    print(f"Trace ID: {format(ctx.trace_id, '032x')}")
    print(f"Span ID:   {format(ctx.span_id, '016x')}")
    print(f"Sampled:   {ctx.trace_flags.sampled}")
