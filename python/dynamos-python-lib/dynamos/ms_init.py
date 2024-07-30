import os
import logging
from .logger import InitLogger
from .grpc_client import GRPCClient
from .grpc_server import GRPCServer
import microserviceCommunication_pb2 as msCommTypes
from typing import Callable, Dict
from google.protobuf.empty_pb2 import Empty
import threading


logger = InitLogger()

class Configuration:
    def __init__(self,
                 port,
                 last_service,
                 service_name,
                 ms_message_handler : Callable[[msCommTypes.MicroserviceCommunication], Empty]):
        self.Port = port
        self.last_service = last_service
        self.service_name = service_name
        self.grpc_server = None
        self.next_client = None
        self.ms_message_handler = ms_message_handler

    def init_sidecar_messaging(self, receive_mutex):
        pass  # Implement sidecar messaging initialization


def NewConfiguration(service_name,
                      grpc_addr,
                      ms_message_handler:  Callable[[msCommTypes.MicroserviceCommunication],Empty ]):
    try:
        port = int(os.getenv("DESIGNATED_GRPC_PORT"))
    except ValueError:
        raise ValueError("Error determining port number")

    try:
        # If last service, send outgoing message to Sidecar for rabbitMQ to process
        last_service = int(os.getenv("LAST"))
    except ValueError:
        raise ValueError("Error determining last service")

    logging.debug(f"NewConfiguration {service_name}, port: {port}, lastservice: {last_service}")

    conf = Configuration(
        port=port,
        last_service=last_service > 0,
        service_name=service_name,
        ms_message_handler=ms_message_handler
    )

    next_target_port = str(conf.Port + 1)

    if conf.last_service:
        next_target_port = os.getenv("SIDECAR_PORT")

    conf.grpc_server = GRPCServer(grpc_addr + str(conf.Port), ms_message_handler)
    conf.next_client = GRPCClient(grpc_addr + next_target_port)

    return conf


def signal_continuation(event: threading.Event, condition: threading.Condition) -> None:
    with condition:
        event.set()
        condition.notify()

def signal_wait(event: threading.Event, condition: threading.Condition) -> None:
    with condition:
        while not event.is_set():
            condition.wait()  # Wait for the signal to stop

# def NewConfiguration(service_name,
#                       grpc_addr,
#                       COORDINATOR,
#                       sidecar_callback,
#                       next_callback,
#                       receive_mutex):
#     try:
#         port = int(os.getenv("DESIGNATED_GRPC_PORT"))
#     except ValueError:
#         raise ValueError("Error determining port number")

#     try:
#         # If first service, setup incoming RabbitMQ channel
#         first_service = int(os.getenv("FIRST"))
#     except ValueError:
#         raise ValueError("Error determining first service")

#     try:
#         # If last service, send outgoing message to Sidecar for rabbitMQ to process
#         last_service = int(os.getenv("LAST"))
#     except ValueError:
#         raise ValueError("Error determining last service")

#     logging.debug(f"NewConfiguration {service_name}, firstServer: {first_service}, port: {port}, lastservice: {last_service}")

#     conf = Configuration(
#         port=port,
#         first_service=first_service > 0,
#         last_service=last_service > 0,
#         service_name=service_name,
#         sidecar_callback=sidecar_callback,
#         # receive_callback=receive_callback
#     )

#     if conf.FirstService and conf.LastService:
#         # First and last, connect to sidecar for processing and final destination
#         conf.SidecarConnection = GRPCClient(grpc_addr + os.getenv("SIDECAR_PORT"))
#         conf.init_sidecar_messaging(receive_mutex)
#         conf.NextClient = conf.SidecarConnection
#     elif conf.FirstService:
#         # First service, connect to sidecar for processing and look for next MS for connecting to
#         conf.SidecarConnection = GRPCClient(grpc_addr + os.getenv("SIDECAR_PORT"))
#         conf.init_sidecar_messaging(receive_mutex)
#         conf.NextClient = GRPCClient(grpc_addr + str(conf.Port + 1))
#     elif conf.LastService:
#         # Last service, connect to sidecar as final destination and start own server to receive from previous MS
#         conf.GrpcServer = GRPCServer(grpc_addr + str(conf.Port), sidecar_callback)
#         conf.NextClient = GRPCClient(grpc_addr + os.getenv("SIDECAR_PORT"))
#     else:
#         conf.GrpcServer = GRPCServer(grpc_addr + str(conf.Port))
#         conf.NextClient = GRPCClient(grpc_addr + str(conf.Port + 1))

#     # COORDINATOR.set()  # Signal the coordinator, all connections setup
#     return conf
