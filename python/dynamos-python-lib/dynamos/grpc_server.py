import grpc
from .logger import InitLogger
from .base_client import BaseClient
from .rabbit_client import RabbitClient
from concurrent import futures
import time
import threading
from opentelemetry import trace
from google.protobuf.empty_pb2 import Empty
from typing import Callable, Dict

import health_pb2_grpc as healthServer
import health_pb2 as healthTypes
import microserviceCommunication_pb2_grpc as msCommServer
import microserviceCommunication_pb2 as msCommTypes


CallbackType = Callable[[grpc.ServicerContext, msCommTypes.MicroserviceCommunication], Empty]

class HealthServicer(healthServer.HealthServicer):
    def __init__(self, logger):
        self.logger = logger

    def Check(self, request, context):
        self.logger.info("Received health check request")
        return healthTypes.HealthCheckResponse(
            status=healthTypes.HealthCheckResponse.SERVING
        )

    def Watch(self, request, context):
        self.logger.info("Received health watch request")
        yield healthTypes.HealthCheckResponse(
            status=healthTypes.HealthCheckResponse.SERVING
        )


class MicroserviceServicer(msCommServer.MicroserviceServicer):
    def __init__(self, msCommHandler: Callable[[msCommTypes.MicroserviceCommunication], Empty()], logger): # type: ignore
        self.callback: CallbackType = msCommHandler
        self.logger = logger

    def SendData(self, msComm: msCommTypes.MicroserviceCommunication, context):
        self.logger.debug(f"Starting MicroserviceServicer grpc_server.py/SendData: {msComm.request_metadata.destination_queue}")

        span = trace.get_current_span()
        try:
            # Start a new span
            with trace.get_tracer(__name__).start_as_current_span("grpc_server.py/SendData") as span:
                pass
        except Exception as err:
            self.logger.warn(f"Error starting span: {err}")
            span.end()

        try:
            self.logger.debug(f"msComm type: {type(msComm)}")
            if not isinstance(msComm, msCommTypes.MicroserviceCommunication):
                raise TypeError(f"Expected msComm to be of type msCommTypes.MicroserviceCommunication, got {type(msComm)}")

            self.callback(msComm)
        except TypeError as e:
            self.logger.error(f"TypeError: {e}")
            return Empty()
        except Exception as err:
            self.logger.error(f"SendData Error: {err}")
            return Empty()


        return Empty()


class GRPCServer(BaseClient):
    def __init__(self, grpc_addr, msCommHandler: Callable[[msCommTypes.MicroserviceCommunication, Callable[[msCommTypes.MicroserviceCommunication],Empty ]], None]):
        self.grpc_addr = grpc_addr
        super().__init__(None, None)
        self.callback: CallbackType = msCommHandler

        self.server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
        healthServer.add_HealthServicer_to_server(HealthServicer(self.logger), self.server)
        msCommServer.add_MicroserviceServicer_to_server(MicroserviceServicer(msCommHandler, self.logger), self.server)
        # rabbitServer.add_RabbitServicer_to_server(RabbitServicer(), self.server)
    #     # etcdServer.add_EtcdServicer_to_server(EtcdServicer(), self.server)

        self.server.add_insecure_port(self.grpc_addr)
        self.stop_event = threading.Event()
        self.condition = threading.Condition()
        self.start()

    def start_server(self):
        self.server.start()
        self.logger.info(f"gRPC server started on {self.grpc_addr}")
        with self.condition:
            while not self.stop_event.is_set():
                self.condition.wait()  # Wait for the signal to stop
        self.server.stop(0)
        self.logger.info("Server stopped")


    def start(self):
        self.thread = threading.Thread(target=self.start_server)
        self.thread.daemon = True
        self.thread.start()


    def stop(self):
        self.logger.info("Stopping gRPC server...")
        with self.condition:
            self.stop_event.set()
            self.condition.notify()  # Notify the condition to wake up the server
        self.thread.join()
        self.logger.info("gRPC server stopped")