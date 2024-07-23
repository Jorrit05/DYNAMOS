import os
import logging
from threading import Lock
from .logger import InitLogger
import grpc
import time
from opentelemetry.instrumentation.grpc import GrpcInstrumentorClient
import health_pb2_grpc as healthServer
import health_pb2 as healthTypes


logger = InitLogger()

class Configuration:
    def __init__(self, port, first_service, last_service, service_name, sidecar_callback, grpc_callback):
        self.Port = port
        self.FirstService = first_service
        self.LastService = last_service
        self.ServiceName = service_name
        self.SidecarConnection = None
        self.NextConnection = None
        self.SideCarCallback = sidecar_callback
        self.GrpcCallback = grpc_callback
        self.StopMicroservice = None  # Placeholder for stopping the service
        self.GrpcServer = None  # Placeholder for gRPC server

    def init_sidecar_messaging(self, receive_mutex):
        pass  # Implement sidecar messaging initialization

    def start_grpc_server(self):
        pass  # Implement gRPC server start


def get_grpc_connection(grpc_addr):
        channel = grpc.insecure_channel(grpc_addr)
        grpc_server_instrumentor = GrpcInstrumentorClient()
        grpc_server_instrumentor.instrument(channel=channel)

        health_stub = healthServer.HealthStub(channel)
        logger.debug(f"Try connecting to: {grpc_addr}")
        for i in range(1, 8):  # maximum of 7 retries
            try:

                response = health_stub.Check(healthTypes.HealthCheckRequest())
                if response.status == healthTypes.HealthCheckResponse.SERVING:
                    break  # The connection is ready, so break the loop.
            except grpc.RpcError as e:
                logger.warning(f"Could not check: {e.details()}")

            logger.info("Sleep 1 second")
            time.sleep(1)  # Wait a second before checking again.

            if i == 7:
                raise Exception(f"Could not connect with gRPC {grpc_addr} after {i} tries")



def start_grpc_server():
    logging.info(f"Start listening on port: {port}")
    server_instance = lib.SharedServer()

    pb.add_MicroserviceServerServicer_to_server(server_instance, grpc_server)
    pb.add_HealthServerServicer_to_server(server_instance, grpc_server)
    server_instance.register_callback("microserviceCommunication", grpc_callback)

    grpc_server.add_insecure_port(f'[::]:{port}')
    grpc_server.start()
    grpc_server.wait_for_termination()



def NewConfiguration(service_name,
                      grpc_addr,
                      COORDINATOR,
                      sidecar_callback,
                      grpc_callback,
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
        grpc_callback=grpc_callback
    )

    if conf.FirstService and conf.LastService:
        # First and last, connect to sidecar for processing and final destination
        conf.SidecarConnection = get_grpc_connection(grpc_addr + os.getenv("SIDECAR_PORT"))
        conf.init_sidecar_messaging(receive_mutex)
        conf.NextConnection = conf.SidecarConnection
    elif conf.FirstService:
        # First service, connect to sidecar for processing and look for next MS for connecting to
        conf.SidecarConnection = get_grpc_connection(grpc_addr + os.getenv("SIDECAR_PORT"))
        conf.init_sidecar_messaging(receive_mutex)
        conf.NextConnection = get_grpc_connection(grpc_addr + str(conf.Port + 1))
    elif conf.LastService:
        # Last service, connect to sidecar as final destination and start own server to receive from previous MS
        conf.GrpcServer = start_grpc_server()
        conf.NextConnection = get_grpc_connection(grpc_addr + os.getenv("SIDECAR_PORT"))
    else:
        conf.GrpcServer = start_grpc_server()
        conf.NextConnection = get_grpc_connection(grpc_addr + str(conf.Port + 1))

    COORDINATOR.set()  # Signal the coordinator, all connections setup
    return conf


# Example usage
COORDINATOR = threading.Event()
receive_mutex = Lock()

