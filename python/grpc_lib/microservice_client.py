import os

import microserviceCommunication_pb2 as msServerTypes
import microserviceCommunication_pb2_grpc as msServer
from google.protobuf.any_pb2 import Any
from grpc_lib import SecureChannel
# from opentelemetry.propagate import set_in_grpc_metadata
# from grpc import Metadata
from opentelemetry import trace, context

class MsCommunication(SecureChannel):
    def __init__(self, config, ctx):
        self.next_service_port = ""
        if int(os.getenv("LAST")) > 0:
            self.next_service_port = os.getenv("SIDECAR_PORT")
        else:
            self.next_service_port = str(int(os.getenv("DESIGNATED_GRPC_PORT")) + 1)
        super().__init__(config, self.next_service_port)
        self.client = msServer.MicroserviceStub(self.channel)
        self.ctx = ctx

    def SendData(self, type, data, metadata, msComm):
        # Populate the message fields
        msComm.data.CopyFrom(data)
        # Populate the metadata field
        for key, value in metadata.items():
            msComm.metadata[key] = value

        # if msComm.traces == None:
        #     self.logger.warning(" msComm.Trace == None")

        # Create metadata for gRPC request
        # metadata = Metadata()
        # metadata = [('binaryTrace', msComm.traces["binaryTrace"])]
        # metadata = [('binarytrace', serialized_span_context)]
        # Inject context into metadata
        # set_in_grpc_metadata(self.ctx, metadata)

        # Add metadata to gRPC call
        span = trace.get_current_span()
        span_context = span.get_span_context()
        print(f"Span ID: {hex(span_context.span_id)[2:].zfill(16)}")
        print(f"Span trace_id: {hex(span_context.trace_id)[2:].zfill(16)}")
        print(f"Span trace_flags: {hex(span_context.trace_flags)[2:].zfill(2)}")
        print(f"Span trace_state: {span_context.trace_state}")

        self.logger.debug(f"Sending message to {self.next_service_port}")
        self.client.SendData(msComm)

    def pack_any(msg) -> Any:
        any_msg = Any()
        any_msg.Pack(msg)
        return any_msg