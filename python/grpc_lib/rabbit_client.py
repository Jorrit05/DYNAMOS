import grpc
import os
import time

import rabbitMQ_pb2_grpc as rabbit
import rabbitMQ_pb2 as rabbitTypes
from grpc_lib import SecureChannel


class RabbitClient(SecureChannel):
    def __init__(self, config, service_name, routing_key, callback=None, queue_auto_delete=False):
        super().__init__(config, os.getenv("SIDECAR_PORT"))
        self.sidecar = rabbit.SideCarStub(self.channel)
        self.callback = callback
        self.stop_consuming = False
        self.initialize_rabbit(service_name, routing_key, queue_auto_delete)

    def initialize_rabbit(self, service_name, routing_key, queue_auto_delete=False):
        try:
            service_request = rabbitTypes.InitRequest()
            service_request.service_name = service_name
            service_request.routing_key = routing_key
            service_request.queue_auto_delete = queue_auto_delete
            self.sidecar.InitRabbitMq(service_request)
            self.logger.debug("Rabbit service started successfully")
        except grpc.RpcError as e:
            self.logger.warning(f"Attempt : could not establish connection with RabbitMQ: {e}")


    def start_consuming(self, queue_name, max_retries=5, wait_time=1):
        return self._consume_with_retry(queue_name, max_retries, wait_time)

    def _consume_with_retry(self, queue_name, max_retries, wait_time):
        for i in range(max_retries):
            try:
                return self._consume(queue_name)
            except grpc.RpcError as e:
                self.logger.error(f"Failed to start consuming (attempt {i+1}/{max_retries}): {e}")
                if self.stop_consuming:  # Check if we've been told to stop before handling each response
                    return
                time.sleep(wait_time)

    def _consume(self, queue_name):
        consume_request = rabbitTypes.ConsumeRequest()
        consume_request.queue_name = queue_name
        consume_request.auto_ack = False

        # Handle 1 response only
        try:
            responses = self.sidecar.ChainConsume(consume_request)

            for response in responses:
                if self.callback:
                    if self.callback(self, response):
                        self.logger.info("query service handled callback successfully")
                        self.close_program()
                        return True
                    else:
                        self.logger.info("Error in query service callback handling")
                        self.close_program()
                        return False
                else:
                    self.logger.warning("no rabbitMq callback registered")
                    return False
        except grpc.RpcError as e:
            self.logger.error(f"Error on consume: {e}")
            raise e

    def _handle_response(self, response):
        # Call the callback function, if it exists, and pass the RabbitClient instance (self)
        if self.callback:
            self.stop_consuming = self.callback(self, response)
