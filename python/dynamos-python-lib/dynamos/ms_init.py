import os
import logging
from .logger import InitLogger
from .grpc_client import GRPCClient
from .grpc_server import GRPCServer

logger = InitLogger()

class Configuration:
    def __init__(self, port, first_service, last_service, service_name, sidecar_callback):
        self.Port = port
        self.FirstService = first_service
        self.LastService = last_service
        self.ServiceName = service_name
        self.SidecarConnection = None
        self.NextClient = None
        self.SideCarCallback = sidecar_callback
        # self.GrpcCallback = grpc_callback
        # self.StopMicroservice = None  # Placeholder for stopping the service
        # self.GrpcServer = None  # Placeholder for gRPC server

    def init_sidecar_messaging(self, receive_mutex):
        pass  # Implement sidecar messaging initialization


def NewConfiguration(service_name,
                      grpc_addr,
                      COORDINATOR,
                      sidecar_callback,
                      next_callback,
                      receive_mutex):
    try:
        port = int(os.getenv("DESIGNATED_GRPC_PORT"))
    except ValueError:
        raise ValueError("Error determining port number")

    try:
        # If first service, setup incoming RabbitMQ channel
        first_service = int(os.getenv("FIRST"))
    except ValueError:
        raise ValueError("Error determining first service")

    try:
        # If last service, send outgoing message to Sidecar for rabbitMQ to process
        last_service = int(os.getenv("LAST"))
    except ValueError:
        raise ValueError("Error determining last service")

    logging.debug(f"NewConfiguration {service_name}, firstServer: {first_service}, port: {port}, lastservice: {last_service}")

    conf = Configuration(
        port=port,
        first_service=first_service > 0,
        last_service=last_service > 0,
        service_name=service_name,
        sidecar_callback=sidecar_callback,
        # receive_callback=receive_callback
    )

    if conf.FirstService and conf.LastService:
        # First and last, connect to sidecar for processing and final destination
        conf.SidecarConnection = GRPCClient(grpc_addr + os.getenv("SIDECAR_PORT"))
        conf.init_sidecar_messaging(receive_mutex)
        conf.NextClient = conf.SidecarConnection
    elif conf.FirstService:
        # First service, connect to sidecar for processing and look for next MS for connecting to
        # conf.SidecarConnection = GRPCClient(grpc_addr + os.getenv("SIDECAR_PORT"))
        # conf.init_sidecar_messaging(receive_mutex)
        conf.NextClient = GRPCClient(grpc_addr + str(conf.Port + 1))
    elif conf.LastService:
        # Last service, connect to sidecar as final destination and start own server to receive from previous MS
        conf.GrpcServer = GRPCServer(conf.Port)
        conf.NextClient = GRPCClient(grpc_addr + os.getenv("SIDECAR_PORT"))
    else:
        conf.GrpcServer = GRPCServer(conf.Port)
        conf.NextClient = GRPCClient(grpc_addr + str(conf.Port + 1))

    # COORDINATOR.set()  # Signal the coordinator, all connections setup
    return conf
