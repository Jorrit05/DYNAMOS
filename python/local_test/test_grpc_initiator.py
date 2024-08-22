
# from dynamos.grpc_client import GRPCClient
from dynamos.grpc_client import GRPCClient
from dynamos.logger import InitLogger
from google.protobuf import struct_pb2, any_pb2
import microserviceCommunication_pb2_grpc as msCommServer
import microserviceCommunication_pb2 as msCommTypes
import generic_pb2 as genericTypes
import threading


# Initialize the logger (assuming InitLogger is correctly defined somewhere in dynamos.logger)
logger = InitLogger()


def getMsComm():
    # Create an instance of MicroserviceCommunication
    msComm = msCommTypes.MicroserviceCommunication()

    # Populate the fields with test data
    msComm.type = "example_type"
    msComm.request_type = "example_request_type"

    # Create and populate the Struct message for the data field
    data_struct = struct_pb2.Struct()
    data_struct.update({
        "type": "sqlDataRequest",
        "key2": 1234,
        "key3": True
    })

    # Populate the metadata field
    msComm.metadata["meta_key1"] = "meta_value1"
    msComm.metadata["meta_key2"] = "meta_value2"

    # Create and populate the Any message for the original_request field
    original_request_any = any_pb2.Any()
    original_request_any.Pack(data_struct)  # Packing the same data_struct as an example
    msComm.original_request.CopyFrom(original_request_any)

    # Create and populate the RequestMetadata message
    request_metadata = genericTypes.RequestMetadata(
        correlation_id="12345",
        destination_queue="example_queue",
        job_name="example_job",
        return_address="example_return_address",
        job_id="job_12345"
    )
    msComm.request_metadata.CopyFrom(request_metadata)
    logger.warning(f"msComm.request_metadata: {msComm.request_metadata}")
    # Populate the traces field with some test data
    msComm.traces["trace_key1"] = b'trace_value1'
    msComm.traces["trace_key2"] = b'trace_value2'

    # Populate the result field with some test data
    msComm.result = b'result_data'

    # Populate the routing_data field with some test data
    msComm.routing_data.extend(["route1", "route2", "route3"])

    # Print the message for verification
    # print(msComm)
    return data_struct, msComm




grpc_addr = "localhost:"
data, msComm = getMsComm()
client = GRPCClient("localhost:50052")
logger.debug(f"msComm type: {type(msComm)}")

client.ms_comm.send_data(msComm, data, msComm.metadata)


# client.close_program()