import os
import microserviceCommunication_pb2 as msServerTypes
import microserviceCommunication_pb2_grpc as msServer
from grpc_lib import SecureChannel

class MsCommunication(SecureChannel):
    def __init__(self, grpc_addr):
        self.next_service_port = str(int(os.getenv("DESIGNATED_GRPC_PORT")) + 1)
        super().__init__(grpc_addr, self.next_service_port)
        self.client = msServer.MicroserviceStub(self.channel)

    def SendData(self, type, data, metadata):
        # Instantiate your protobuf message
        communication = msServerTypes.MicroserviceCommunication()

        # Populate the message fields
        communication.type = type
        communication.data.CopyFrom(data)
        self.logger.debug(f"Sending message to {self.next_service_port}")
        self.client.SendData(communication)
