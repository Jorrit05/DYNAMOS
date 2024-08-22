"""
Package dynamos, implements functionality for handling Microservice chains in Python.

File: ms_init.py

Description:
This file contains the functionality to initialize a Microservice in the Microservice chain. Based on
environment variables this configuration takes all steps to initiate gRPC server/clients and/or connections
with RabbitMQ.

Notes:

Author: Jorrit Stutterheim
"""

import os
import time
from .logger import InitLogger
from .grpc_client import GRPCClient
from .grpc_server import GRPCServer
import microserviceCommunication_pb2 as msCommTypes
from typing import Callable, Dict
from google.protobuf.empty_pb2 import Empty

from opentelemetry import trace
import json
from opentelemetry.trace.span import TraceFlags, TraceState
from opentelemetry.trace.propagation import set_span_in_context
from opentelemetry.context.context import Context

logger = InitLogger()

# Global map to register functions
global_function_map: Dict[str, Callable[[msCommTypes.MicroserviceCommunication, Context], Empty]] = {}

def register_function(name: str, handler: Callable[[msCommTypes.MicroserviceCommunication, Context], Empty]):
    global_function_map[name] = handler

def get_and_call_function(name: str, msComm: msCommTypes.MicroserviceCommunication, ctx: Context) -> Empty:
    """
    Retrieves a function from the global function map based on the given name and calls it with the provided arguments.

    Args:
        name (str): The name of the function to retrieve and call.
        msComm (msCommTypes.MicroserviceCommunication): The MicroserviceCommunication object.
        ctx (Context): The Context object.

    Returns:
        Empty: The result of calling the function.

    Raises:
        ValueError: If no function is registered under the given name.
    """
    if name in global_function_map:
        handler = global_function_map[name]
        return handler(msComm, ctx)
    else:
        raise ValueError(f"No function registered under the name: {name}")

class Configuration:
    def __init__(self,
                 job_name,
                 port,
                 first_service,
                 last_service,
                 service_name,
                 ms_message_handler : Callable[[msCommTypes.MicroserviceCommunication], Empty]):
        self.job_name = job_name
        self.Port = port
        self.first_service = first_service
        self.last_service = last_service
        self.service_name = service_name
        self.rabbit_msg_client = None
        self.grpc_server = None
        self.next_client = None
        self.ms_message_handler = ms_message_handler

    def stop(self, sleep_time=2):
        try:
            if self.rabbit_msg_client and self.last_service:
                self.rabbit_msg_client.rabbit.stop()
            if self.grpc_server:
                self.grpc_server.stop()
            if self.next_client:
                self.next_client.close_program()
        except Exception as e:
            logger.error(f"An error occurred while stopping the configuration: {e}")

        time.sleep(sleep_time)


def request_handler(msComm : msCommTypes.MicroserviceCommunication):
    """
    Handles the incoming microservice communication message.

    Args:
        msComm (msCommTypes.MicroserviceCommunication): The microservice communication object.

    Returns:
        bool: True if the communication is handled successfully, False otherwise.
    """
    try:
        if msComm.type == "microserviceCommunication":
            try:
                # Parse the trace header back into a dictionary
                scMap = json.loads(msComm.traces["jsonTrace"])
                state = TraceState([("sampled", "1")])
                logger.debug(f"1. scMap: {scMap}")
                logger.debug(f"1. state: {state}")
                sc = trace.SpanContext(
                    trace_id=int(scMap['TraceID'], 16),
                    span_id=int(scMap['SpanID'], 16),
                    is_remote=True,
                    trace_flags=TraceFlags(TraceFlags.SAMPLED),
                    trace_state=state
                )

                # create a non-recording span with the SpanContext and set it in a Context
                span = trace.NonRecordingSpan(sc)
                logger.debug(f"2")
                ctx = set_span_in_context(span)
            except Exception as e:
                logger.error(f"A tracing error occurred: {e}")
                return False

            try:
                get_and_call_function("callback", msComm, ctx)
            except Exception as e:
                logger.error(f"A callback error occurred: {e}")
                return False
        else:
            logger.error(f"An unexpected message arrived: {msComm.type}")
            return False

    except Exception as e:
        logger.error(f"Error in ms_init request_handler: {e}")


def get_env_var(var_name):
    """
    Retrieves the value of the specified environment variable.

    Args:
        var_name (str): The name of the environment variable.

    Returns:
        str: The value of the environment variable.

    Raises:
        ValueError: If an error occurs while retrieving the environment variable.

    """
    try:
        val = os.getenv(var_name)
    except ValueError:
        raise ValueError(f"Error {var_name}")

    return val


def NewConfiguration(service_name,
                      grpc_addr,
                      ms_message_handler:  Callable[[msCommTypes.MicroserviceCommunication],Empty ]):
    """
    Creates a new configuration object for a setting up a Microservice Chain.

    Args:
        service_name (str): The name of the microservice.
        grpc_addr (str): The address of the gRPC server.
        ms_message_handler (Callable): A callable object that handles microservice communication.

    Returns:
        Configuration: The new configuration object.

    """
    port = int(get_env_var("DESIGNATED_GRPC_PORT"))

    register_function("callback", ms_message_handler)

    first_service = int(get_env_var("FIRST"))
    last_service = int(get_env_var("LAST"))
    job_name = get_env_var("JOB_NAME")
    logger.debug(f"NewConfiguration {service_name}, \njob_name: {job_name} \nport: {port},  \nfirst_service: {first_service},  \nlast_service: {last_service}")

    conf = Configuration(
        job_name=job_name,
        port=port,
        first_service=first_service > 0,
        last_service=last_service > 0,
        service_name=service_name,
        ms_message_handler=ms_message_handler
    )

    if conf.first_service:
        # First and possibly last, connect to sidecar for processing and final destination
        conf.grpc_server = GRPCServer(grpc_addr + str(conf.Port), request_handler)
        conf.rabbit_msg_client = GRPCClient(grpc_addr + os.getenv("SIDECAR_PORT"), service_name)
        if conf.last_service:
            conf.next_client = conf.rabbit_msg_client
        else:
            conf.next_client = GRPCClient(grpc_addr + str(conf.Port + 1), service_name)

        # Send a message to the RabbitMQ server to initialize the connection
        conf.rabbit_msg_client.rabbit.initialize_rabbit(job_name, conf.Port)

    elif conf.last_service:
        # Last service, connect to sidecar as final destination and start own server to receive from previous MS
        conf.grpc_server = GRPCServer(grpc_addr + str(conf.Port), ms_message_handler)
        conf.next_client = GRPCClient(grpc_addr + os.getenv("SIDECAR_PORT"), service_name)
    else:
        conf.grpc_server = GRPCServer(grpc_addr + str(conf.Port), ms_message_handler)
        conf.next_client = GRPCClient(grpc_addr + str(conf.Port + 1), service_name)

    return conf
