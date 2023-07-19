import os

from opentracing import Format
import microserviceCommunication_pb2 as msServerTypes
import microserviceCommunication_pb2_grpc as msServer
from google.protobuf.any_pb2 import Any
from grpc_lib import SecureChannel

class MsCommunication(SecureChannel):
    def __init__(self, grpc_addr, tracer):
        self.next_service_port = ""
        if int(os.getenv("LAST")) > 0:
            self.next_service_port = os.getenv("SIDECAR_PORT")
        else:
            self.next_service_port = str(int(os.getenv("DESIGNATED_GRPC_PORT")) + 1)
        super().__init__(grpc_addr, self.next_service_port, tracer)
        self.client = msServer.MicroserviceStub(self.channel)

    def SendData(self, type, data, metadata, msComm):
        # Populate the message fields
        msComm.data.CopyFrom(data)
        # Populate the metadata field
        for key, value in metadata.items():
            msComm.metadata[key] = value

        # carrier = {}
        # self.tracer.inject(span.context, Format.TEXT_MAP, carrier)
        # msComm.trace = carrier["trace"]  # replace with actual trace field name
        if msComm.trace == None:
            self.logger.warning(" msComm.Trace == None")

        self.logger.debug(f"Sending message to {self.next_service_port}")
        self.client.SendData(msComm)

    def pack_any(msg) -> Any:
        any_msg = Any()
        any_msg.Pack(msg)
        return any_msg