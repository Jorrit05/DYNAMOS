import grpc
import time
from .logger import InitLogger
from .base_client import BaseClient
from .rabbit_client import RabbitClient
from opentelemetry.instrumentation.grpc import GrpcInstrumentorClient
from google.protobuf.empty_pb2 import Empty
from concurrent import futures
from opentelemetry import trace

import health_pb2_grpc as healthServer
import health_pb2 as healthTypes
import rabbitMQ_pb2_grpc as rabbitServer
import rabbitMQ_pb2 as rabbitTypes
import etcd_pb2_grpc as etcdServer
import etcd_pb2 as etcdTypes
import microserviceCommunication_pb2_grpc as msCommServer
import microserviceCommunication_pb2 as msCommTypes

logger = InitLogger()


class HealthClient(BaseClient):
    def __init__(self, channel):
        super().__init__(channel)
        self.stub = healthServer.HealthStub(channel)

    def check_health(self):
        try:
            response = self.stub.Check(healthTypes.HealthCheckRequest())
            logger.info(f"Health status: {response.status}")
            return response.status
        except grpc.RpcError as e:
            logger.error(f"Health check failed: {e.details()}")
            return None


class EtcdClient(BaseClient):
    def __init__(self, channel):
        super().__init__(channel)
        self.stub = etcdServer.EtcdStub(channel)

    def initialize_etcd(self):
        empty = Empty()
        self.client.InitEtcd(empty)

    def getDatasetMetadata(self, key):
        path = etcdTypes.EtcdKey()
        path.path = key
        return self.client.GetDatasetMetadata(path)


class MicroserviceClient(BaseClient):
    def __init__(self, channel):
        super().__init__(channel)
        self.stub = msCommServer.MicroserviceStub(channel)

    # Define microservice-specific methods here
    def send_data(self, msComm, data, metadata):
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

        logger.debug(f"Sending message to {self.stub}")
        self.stub.SendData(msComm)


class GRPCClient:
    def __init__(self, grpc_addr, service_name):
        self.grpc_addr = grpc_addr

        self.channel = self.get_grpc_connection(grpc_addr)
        self.health = HealthClient(self.channel)
        self.rabbit = RabbitClient(self.channel, service_name)
        self.etcd = EtcdClient(self.channel)
        self.ms_comm = MicroserviceClient(self.channel)

    def close_program(self):
        """Close the gRPC channel gracefully"""
        self.channel.close()
        logger.debug("Closed gRPC channel")


    def get_grpc_connection(self, grpc_addr):
        channel = grpc.insecure_channel(grpc_addr)
        grpc_server_instrumentor = GrpcInstrumentorClient()
        grpc_server_instrumentor.instrument(channel=channel)

        logger.debug(f"Try connecting to: {grpc_addr}")
        for i in range(1, 8):  # maximum of 7 retries
            try:
                health_stub = healthServer.HealthStub(channel)
                response = health_stub.Check(healthTypes.HealthCheckRequest())
                if response.status == healthTypes.HealthCheckResponse.SERVING:
                    logger.info(f"Successfully connected to gRPC server at {grpc_addr}")
                    return channel  # Return the channel
            except grpc.RpcError as e:
                logger.warning(f"Could not check: {e.details()}")

            logger.info("Sleep 1 second")
            time.sleep(1)  # Wait a second before checking again

        raise Exception(f"Could not connect with gRPC {grpc_addr} after {i} tries")
