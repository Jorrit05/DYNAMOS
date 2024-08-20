"""
Package dynamos, implements functionality for handling Microservice chains in Python.

File: grpc_client.py

Description:
This file contains the GRPCClient, this client initiates a connection to another gRPC server and
registers all gRPC function stubs. This allows calling functions implemented on the gRPC server side.

Notes:

Author: Jorrit Stutterheim
"""

import grpc
import time
from .base_client import BaseClient
from .rabbit_client import RabbitClient
from opentelemetry.instrumentation.grpc import GrpcInstrumentorClient

import health_pb2_grpc as healthServer
import health_pb2 as healthTypes
import microserviceCommunication_pb2_grpc as msCommServer


class GRPCClient(BaseClient):
    """A client for interacting with a gRPC server."""

    def __init__(self, grpc_addr, service_name):
        """
        Initialize the GRPCClient, start a connection to the gRPC server. And register all
        gRPC function stubs.

        Args:
            grpc_addr (str): The address of the gRPC server.
            service_name (str): The name of the service.

        """
        self.grpc_addr = grpc_addr

        # Init Logger first without a channel
        super().__init__(None, service_name)
        self.channel = self.get_grpc_connection(grpc_addr)

        self.health = HealthClient(self.channel, service_name, self.logger)
        self.ms_comm = MicroserviceClient(self.channel, service_name, self.logger)
        self.rabbit = RabbitClient(self.channel, self.service_name, self.logger)

    def close_program(self):
        """Close the gRPC channel gracefully."""
        self.channel.close()
        self.logger.debug("Closed gRPC channel")

    def get_grpc_connection(self, grpc_addr):
        """
        Get a gRPC connection to the server.

        Args:
            grpc_addr (str): The address of the gRPC server.

        Returns:
            grpc.Channel: The gRPC channel.

        Raises:
            Exception: If unable to connect to the gRPC server after 7 retries.

        """
        channel = grpc.insecure_channel(grpc_addr)
        grpc_server_instrumentor = GrpcInstrumentorClient()
        grpc_server_instrumentor.instrument(channel=channel)

        self.logger.debug(f"Try connecting to: {grpc_addr}")
        for i in range(1, 8):  # maximum of 7 retries
            try:
                health_stub = healthServer.HealthStub(channel)
                response = health_stub.Check(healthTypes.HealthCheckRequest())
                if response.status == healthTypes.HealthCheckResponse.SERVING:
                    self.logger.info(f"Successfully connected to gRPC server at {grpc_addr}")
                    return channel  # Return the channel
            except grpc.RpcError as e:
                self.logger.warning(f"Could not check: {e.details()}")

            self.logger.info("Sleep 1 second")
            time.sleep(1)  # Wait a second before checking again

        raise Exception(f"Could not connect with gRPC {grpc_addr} after {i} tries")


class MicroserviceClient:
    """
    Represents a client for interacting with the next microservices in a microservice chain.

    Args:
        channel: The gRPC channel used for communication with the microservice.
        service_name: Own name of the microservice.
        logger: The logger instance used for logging.

    Attributes:
        logger: The logger instance used for logging.
        channel: The gRPC channel used for communication with the microservice.
        service_name: The name of this microservice.
        stub: The gRPC stub for making RPC calls to the microservice.
    """

    def __init__(self, channel, service_name, logger):
        self.logger = logger
        self.channel = channel
        self.service_name = service_name
        self.stub = msCommServer.MicroserviceStub(self.channel)

    # Define microservice-specific methods here
    def send_data(self, msComm, data, metadata):
        """
        Sends data to the microservice.

        Args:
            msComm: The message object used for communication with the microservice.
            data: The data to be sent.
            metadata: The metadata associated with the data.

        Returns:
            None
        """
        # Populate the message fields
        msComm.data.CopyFrom(data)

        # Populate the metadata field
        for key, value in metadata.items():
            msComm.metadata[key] = value

        # Add metadata to gRPC call
        # span = trace.get_current_span()
        # span_context = span.get_span_context()
        # # print(f"Span ID: {hex(span_context.span_id)[2:].zfill(16)}")
        # # print(f"Span trace_id: {hex(span_context.trace_id)[2:].zfill(16)}")
        # # print(f"Span trace_flags: {hex(span_context.trace_flags)[2:].zfill(2)}")
        # # print(f"Span trace_state: {span_context.trace_state}")

        self.logger.debug(f"Sending message to {self.stub}")
        self.stub.SendData(msComm)


class HealthClient:
    def __init__(self, channel, service_name, logger):
        self.logger = logger
        self.channel = channel
        self.service_name = service_name
        self.stub = healthServer.HealthStub(self.channel)

    def check_health(self):
        try:
            response = self.stub.Check(healthTypes.HealthCheckRequest())
            self.logger.info(f"Health status: {response.status}")
            return response.status
        except grpc.RpcError as e:
            self.logger.error(f"Health check failed: {e.details()}")
            return None
