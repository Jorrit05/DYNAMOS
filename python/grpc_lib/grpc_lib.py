import grpc
import time

import health_pb2_grpc as healthServer
import health_pb2 as healthTypes

import grpc
import time
from my_logger import InitLogger
from opentelemetry.instrumentation.grpc import GrpcInstrumentorClient

class SecureChannel:
    def __init__(self, config, grpc_port):
        self.logger = InitLogger()
        if grpc_port == "":
            self.logger.fatal("Grpc port is undefined")

        self.channel = None
        self.grpc_addr = config.grpc_addr
        self.grpc_port = grpc_port

        self.connect()

    def connect(self):

        # tracer = context_tracer.ContextTracer()
        # execution_context.set_opencensus_tracer(tracer)

        # tracer_interceptor = client_interceptor.OpenCensusClientInterceptor(
        #     tracer,
        #     host_port='collector.linkerd-jaeger:55678'
        # )
        # intercept the channel
        self.channel = grpc.insecure_channel(self.grpc_addr + self.grpc_port)
        grpc_server_instrumentor = GrpcInstrumentorClient()
        grpc_server_instrumentor.instrument(channel=self.channel)
        # self.channel = grpc.intercept_channel(self.channel, tracer_interceptor)

        # intercept
        # tracer_interceptor = tracer.get_interceptor()
        # # intercept the channel
        # self.channel = grpc.insecure_channel(self.grpc_addr + self.grpc_port)
        # self.channel = intercept_channel(self.channel, tracer_interceptor)


        health_stub = healthServer.HealthStub(self.channel)
        self.logger.debug(f"Try connecting to: {self.grpc_addr + self.grpc_port}")
        for i in range(1, 8):  # maximum of 7 retries
            try:

                response = health_stub.Check(healthTypes.HealthCheckRequest())
                if response.status == healthTypes.HealthCheckResponse.SERVING:
                    break  # The sidecar is ready, so break the loop.
            except grpc.RpcError as e:
                self.logger.warning(f"Could not check: {e.details()}")

            self.logger.info("Sleep 1 second")
            time.sleep(1)  # Wait a second before checking again.

            if i == 7:
                raise Exception(f"Could not connect with gRPC after {i} tries")

    def close_program(self):
        # Call this function to close gRPC channel gracefully
        self.channel.close()
        self.logger.debug("Closed gRPC channel")