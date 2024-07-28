
# from dynamos.grpc_client import GRPCClient
from dynamos.ms_init import NewConfiguration
from dynamos.logger import InitLogger
from google.protobuf import struct_pb2, any_pb2
import microserviceCommunication_pb2_grpc as msCommServer
import microserviceCommunication_pb2 as msCommTypes
import generic_pb2 as genericTypes
import threading


# Initialize the logger (assuming InitLogger is correctly defined somewhere in dynamos.logger)
logger = InitLogger()

stop_event = threading.Event()
stop_microservice_condition = threading.Condition()


def request_handler(msComm):
    logger.info(f"Received request: {msComm.request_type}")

    # do logic...


    # client.ms_comm.send_data(data_struct, msComm.metadata,  msComm)



grpc_addr = "localhost:"

conf = NewConfiguration("caller", grpc_addr, "COORDINATOR", request_handler, None, None)


with stop_microservice_condition:
    while not stop_event.is_set():
        stop_microservice_condition.wait()  # Wait for the signal to stop

# data_struct, msComm = getMsComm()

# conf.NextClient.ms_comm.send_data(data_struct, msComm.metadata,  msComm)


# client.close_program()