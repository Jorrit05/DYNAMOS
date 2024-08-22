"""
Package dynamos, implements functionality for handling Microservice chains in Python.

File: rabbit_client.py

Description:
This file contains the RabbitClient class, this class requests the sidecar to initialize a RabbitMQ queue
and can request cancellation as well.

Notes:

Author: Jorrit Stutterheim
"""

import grpc
import rabbitMQ_pb2_grpc as rabbitServer
import rabbitMQ_pb2 as rabbitTypes
import threading


class RabbitClient:
    """
    A class representing a RabbitMQ client.

    Attributes:
        channel (object): The RabbitMQ channel object.
        service_name (str): The name of the service.
        logger (object): The logger object for logging messages.
        stub (object): The RabbitMQ server stub.
        stop_event (object): The threading event for stopping the client.
        condition (object): The threading condition for synchronization.
        own_grpc_client (object): The gRPC client object.
        thread (object): The threading thread object.

    Methods:
        __init__(self, channel, service_name, logger): Initializes the RabbitClient object.
        initialize_rabbit(self, routing_key, port, queue_auto_delete=False): Initializes RabbitMQ for chain processing.
        stop(self): Stops the RabbitMQ client.
    """
    def __init__(self, channel, service_name, logger):
        self.logger = logger
        self.channel = channel
        self.service_name = service_name
        self.stub = rabbitServer.RabbitMQStub(channel)
        self.stop_event = threading.Event()
        self.condition = threading.Condition()
        self.own_grpc_client = None
        self.thread = None


    def initialize_rabbit(self, routing_key : str, port : int,  queue_auto_delete=False):
        """
        Send ChainRequest message to the sidecar to initialize a RabbitMQ queue
        for MS chain processing.

        Args:
            routing_key (str): The routing key for the RabbitMQ exchange.
            port (int): The port number on which this server starts a gRPC server for receiving the AMQ messages.
            queue_auto_delete (bool, optional): Whether to automatically delete the queue. Defaults to False.
        """
        try:
            chain_request = rabbitTypes.ChainRequest()
            chain_request.service_name = self.service_name
            chain_request.routing_key = routing_key
            chain_request.queue_auto_delete = queue_auto_delete
            chain_request.port = port

            self.stub.InitRabbitForChain(chain_request)

        except grpc.RpcError as e:
            self.logger.warning(f"Attempt : could not establish connection with RabbitMQ: {e}")


    def stop(self):
        """
        Send a StopRequest to the RabbitMQ client.
        """
        stop_request = rabbitTypes.StopRequest()

        self.stub.StopReceivingRabbit(stop_request)

