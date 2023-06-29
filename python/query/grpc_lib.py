import grpc
import os
import time

from google.protobuf.empty_pb2 import Empty
import rabbitMQ_pb2_grpc as rabbit
import rabbitMQ_pb2 as rabbitTypes
import microserviceCommunication_pb2 as msServerTypes
import microserviceCommunication_pb2_grpc as msServer

import etcd_pb2_grpc as etcd
from google.protobuf.struct_pb2 import Struct
import etcd_pb2 as etcdTypes

if os.getenv('ENV') == 'PROD':
    import config_prod as config
else:
    import config_local as config

import grpc
import time

class SecureChannel:
    def __init__(self, grpc_addr, max_retries=5, retry_delay=2):
        self.channel = None
        self.grpc_addr = grpc_addr
        self.max_retries = max_retries
        self.retry_delay = retry_delay
        self.connect()

    def connect(self):
        retries = 0
        while retries < self.max_retries:
            try:
                self.channel = grpc.insecure_channel(self.grpc_addr)
                if self.channel:  # Check if the connection is successful
                    print("Succesfully connected to: " + self.grpc_addr)
                    break
            except Exception as e:  # Catch and print the exception if there is one
                print(f"Failed to establish a secure channel. Attempt: {retries+1}")
                print(f"Exception: {str(e)}")
                retries += 1
                time.sleep(self.retry_delay)  # Wait before the next retry
        else:  # The 'else' clause in the 'while' loop executes when the loop finishes without break
            raise Exception("Failed to establish a secure channel after maximum retries.")



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

class RabbitClient(SecureChannel):
    def __init__(self, grpc_addr, service_request):
        super().__init__(grpc_addr)
        self.client = rabbit.SideCarStub(self.channel)
        self.initialize_rabbit(service_request)

    def initialize_rabbit(self, service_request):
        try:
            self.client.InitRabbitMq(service_request)
            print("Service started successfully")
        except grpc.RpcError as e:
            print(f"Attempt : could not establish connection with RabbitMQ: {e}")

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
        self.client.SendData(data_struct)