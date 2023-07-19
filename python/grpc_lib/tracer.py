
from jaeger_client import Config
from grpc_opentracing import open_tracing_client_interceptor
from opentracing import Format


class Tracer:
    def __init__(self, service_name):

        # initialize Jaeger tracer
        self.tracer = self.initialize_tracer(service_name)


    def initialize_tracer(self, service_name):
        config = Config(
            config={  # usually read from some yaml config
                'sampler': {
                    'type': 'const',
                    'param': 1,
                },
                'local_agent': {
                    'reporting_host': "collector.linkerd-jaeger",
                    'reporting_port': '55678',
                },
                'logging': True,
            },
            service_name=service_name,
            validate=True,
        )
        return config.initialize_tracer()

    def close_tracer(self):
        # call this method when your application is shutting down
        self.tracer.close()

    def get_interceptor(self):
        return open_tracing_client_interceptor(self.tracer)

    def printSpan(self, binary):
        span_context = self.tracer.extract(Format.BINARY, binary)
        print(f"Trace ID: {span_context.trace_id}")
