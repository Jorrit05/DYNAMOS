import grpc
import os
import time

from google.protobuf.empty_pb2 import Empty
import microserviceCommunication_pb2 as msServerTypes
import microserviceCommunication_pb2_grpc as msServer
import health_pb2_grpc as healthServer
import health_pb2 as healthTypes

import etcd_pb2_grpc as etcd
from google.protobuf.struct_pb2 import Struct
import etcd_pb2 as etcdTypes
import grpc
import time
from my_logger import InitLogger

class SecureChannel:
    def __init__(self, grpc_addr, grpc_port):
        self.logger = InitLogger()
        if grpc_port == "":
            self.logger.fatal("Grpc port is undefined")

        self.channel = None
        self.grpc_addr = grpc_addr
        self.grpc_port = grpc_port
        self.connect()

    def connect(self):
        self.channel = grpc.insecure_channel(self.grpc_addr + self.grpc_port)
        health_stub = healthServer.HealthStub(self.channel)

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


class EtcdClient(SecureChannel):
    def __init__(self, grpc_addr):
        super().__init__(grpc_addr)
        self.client = etcd.EtcdStub(self.channel)
        self.initialize_etcd()

    def initialize_etcd(self):
        empty = Empty()
        self.client.InitEtcd(empty)

    def getDatasetMetadata(self, key):
        path = etcdTypes.EtcdKey()
        path.path = key
        return self.client.GetDatasetMetadata(path)

class MsCommunication(SecureChannel):
    def __init__(self):
        super().__init__(config.grpc_addr + str(int(os.getenv("ORDER")) + 1))
        self.client = msServer.MicroserviceStub(self.channel)
        self.empty = Empty()

    def SendData(self):


        # Instantiate your protobuf message
        communication = msServerTypes.MicroserviceCommunication()

        # Populate the message fields
        communication.type = "sqlDataRequest"

        # Create Struct for complex data
        data_struct = Struct()
        data_struct.update({
            "name": "John Doe",
            "age": 30,
            "occupation": "Software Engineer"
        })
        communication.data.CopyFrom(data_struct)

        communication.metadata = "Sample Metadata"
        print(communication.metadata)
        print(communication.type)
        self.client.SendData(communication)