
from jaeger_client import Config
# from grpc_opentracing import open_tracing_client_interceptor

from opencensus.trace.tracer import Tracer as OcTracer
from opencensus.trace.samplers import AlwaysOnSampler
from opencensus.ext.jaeger.trace_exporter import JaegerExporter

class Tracer:
    def __init__(self, my_service_name):
        # initialize Jaeger tracer
        self.tracer =  OcTracer(exporter=JaegerExporter(
            service_name=my_service_name,
            agent_host_name='collector.linkerd-jaeger',
            agent_port=55678,
        ))


    def tracer(self):
        return self.tracer


    # def close_tracer(self):
    #     # call this method when your application is shutting down
    #     self.tracer.close()

    # def get_interceptor(self):
    #     return open_tracing_client_interceptor(self.tracer)

    # def printSpan(self, binary):
    #     span_context = self.tracer.extract(Format.BINARY, binary)
    #     print(f"Trace ID: {span_context.trace_id}")

    # def create_parent_span(self, operation_name, binary):
    #     # Extract the parent span context from the binary
    #     parent_span_context = self.tracer.extract(Format.BINARY, binary)

    #     # Create a new span as a child of the parent span
    #     span = self.tracer.start_span(operation_name, child_of=parent_span_context)

    #     # Inject the new span context into the carrier dictionary
    #     carrier = {}
    #     self.tracer.inject(span.context, Format.BINARY, carrier)
    #     return carrier, span

    # def create_child_span(self, operation_name, parent):
    #     # Create a new span
    #     # span = self.tracer.start_span(operation_name)
    #     span = self.tracer.start_span(operation_name, child_of=parent.context)

    #     # Inject the span context into the carrier dictionary
    #     carrier = {}
    #     self.tracer.inject(span.context, Format.BINARY, carrier)
    #     return carrier, span

    # def end_span(span):
    #     # Finish the span
    #     span.finish()


    # def extract_span(self, binary):
    #     # Extract the span context from the carrier
    #     span_context = self.tracer.extract(Format.BINARY, binary)
    #     return span_context
