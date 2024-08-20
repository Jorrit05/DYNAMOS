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