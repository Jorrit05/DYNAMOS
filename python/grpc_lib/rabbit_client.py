import grpc
import os
import time

from google.protobuf import any_pb2
import rabbitMQ_pb2_grpc as rabbit
import rabbitMQ_pb2 as rabbitTypes
from grpc_lib import SecureChannel

class RabbitClient(SecureChannel):
    def __init__(self, grpc_addr, service_name, routing_key, auto_delete_queue):
        super().__init__(grpc_addr, os.getenv("SIDECAR_PORT"))
        self.client = rabbit.SideCarStub(self.channel)
        service_request = rabbitTypes.ServiceRequest()
        service_request.service_name = service_name
        service_request.routing_key = routing_key
        service_request.queue_auto_delete = auto_delete_queue

        self.initialize_rabbit(service_request)

    def initialize_rabbit(self, service_request):
        try:
            self.client.InitRabbitMq(service_request)
            self.logger.debug("Rabbit service started successfully")
        except grpc.RpcError as e:
            self.logger.warning(f"Attempt : could not establish connection with RabbitMQ: {e}")

    def start_consuming(self, queue_name, max_retries=5, wait_time=1):
        # create a new thread for the consuming function
        consumer_thread = threading.Thread(target=self._consume_with_retry,
                                           args=(queue_name, max_retries, wait_time))
        consumer_thread.start()

    def _consume_with_retry(self, queue_name, max_retries, wait_time):
        for i in range(max_retries):
            try:
                self._consume(queue_name)
                return
            except grpc.RpcError as e:
                self.logger.error(f"Failed to start consuming (attempt {i+1}/{max_retries}): {e}")
                time.sleep(wait_time)

    def _consume(self, queue_name):
        consume_request = rabbitTypes.ConsumeRequest()
        consume_request.queue_name = queue_name
        consume_request.auto_ack = True

        try:
            responses = self.client.Consume(consume_request)
            for response in responses:
                self._handle_response(response)
        except grpc.RpcError as e:
            self.logger.error(f"Error on consume: {e}")
            raise e


    def _handle_response(self, response):
        # here you'd handle your various message types, similar to the Go implementation
        # you may want to adjust this to match your actual message types and their handling
        if response.type == "sqlDataRequest":
            self.logger.debug("response.type is sqlDataRequest")
            try:
                print(type(response.body))
                # any_message = any_pb2.Any()
                # any_message.ParseFromString(response.body)
                sqlDataRequest = rabbitTypes.SqlDataRequest()
                response.body.Unpack(sqlDataRequest)
                self.logger.info("1")

                self.logger.info("Query: " + sqlDataRequest.query)

                # if any_message.Is(sqlDataRequest.DESCRIPTOR):
                #     self.logger.debug("Descriptor is sqlDataRequest")
                #     any_message.Unpack(sqlDataRequest)
                #     # process the sqlDataRequest
                #     self.logger.info("Query: " + sqlDataRequest.query)
                # else:
                #     self.logger.error("Unexpected message type")
            except Exception as e:
                self.logger.error(f"Failed to unmarshal message: {e}")
