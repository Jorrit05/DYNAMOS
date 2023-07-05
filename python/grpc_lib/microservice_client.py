import grpc
import os
from google.protobuf.empty_pb2 import Empty
import microserviceCommunication_pb2 as msServerTypes
import microserviceCommunication_pb2_grpc as msServer
from google.protobuf.struct_pb2 import Struct
from grpc_lib import SecureChannel

class MsCommunication(SecureChannel):
    def __init__(self, grpc_addr):
        super().__init__(grpc_addr + str(int(os.getenv("ORDER")) + 1))
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