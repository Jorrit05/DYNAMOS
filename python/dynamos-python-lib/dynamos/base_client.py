"""
Package dynamos, implements functionality for handling Microservice chains in Python.

File: base_client.py

Description:
This file contains the class BaseClient, this client contains default variables that a gRPC server or
client needs.

Notes:
There are some issues around tracing, hebce this is commented out pending further research.


Author: Jorrit Stutterheim
"""


from dynamos.logger import InitLogger
from dynamos.tracer import InitTracer


class BaseClient:
    """
    Base class for client implementations.

    Args:
        channel (object): The channel object used for communication.
        service_name (str): The name of the service.

    Attributes:
        service_name (str): The name of the service.
        channel (object): The channel object used for communication.
        logger (object): The logger object for logging.
    """

    def __init__(self, channel, service_name):
        self.service_name = service_name
        self.channel = channel
        self.logger = InitLogger()
        # self.logger.info("1")
        # self.tracer = InitTracer(self.service_name, "http://localhost:32003")
        # self.logger.info("2")
