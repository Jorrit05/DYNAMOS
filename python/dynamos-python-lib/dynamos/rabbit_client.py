import grpc
import os
import time

from .base_client import BaseClient
import rabbitMQ_pb2_grpc as rabbitServer
import rabbitMQ_pb2 as rabbitTypes
import threading
import json
from opentelemetry.trace.span import TraceFlags, TraceState
from opentelemetry import trace, context
from opentelemetry.trace.propagation import set_span_in_context
import microserviceCommunication_pb2 as msCommTypes


class RabbitClient:
    def __init__(self, channel, service_name, logger):
        self.logger = logger
        self.channel = channel
        self.service_name = service_name
        self.stub = rabbitServer.SideCarStub(channel)
        self.stop_event = threading.Event()
        self.condition = threading.Condition()
        self.own_grpc_client = None
        self.thread = None

    def initialize_rabbit(self, routing_key, own_grpc_client, queue_auto_delete=False):
        try:
            service_request = rabbitTypes.InitRequest()
            service_request.service_name = self.service_name
            service_request.routing_key = routing_key
            service_request.queue_auto_delete = queue_auto_delete
            self.stub.InitRabbitMq(service_request)
            self.own_grpc_client = own_grpc_client
            self.logger.debug("Rabbit own_grpc_client started successfully")
        except grpc.RpcError as e:
            self.logger.warning(f"Attempt : could not establish connection with RabbitMQ: {e}")


    def handle_incoming_request(self, msg):
        # Parse the trace header back into a dictionary
        scMap = json.loads(msg.traces["jsonTrace"])
        state = TraceState([("sampled", "1")])
        sc = trace.SpanContext(
            trace_id=int(scMap['TraceID'], 16),
            span_id=int(scMap['SpanID'], 16),
            is_remote=True,
            trace_flags=TraceFlags(TraceFlags.SAMPLED),
            trace_state=state
        )

        # create a non-recording span with the SpanContext and set it in a Context
        span = trace.NonRecordingSpan(sc)
        ctx = set_span_in_context(span)

        if msg.type == "microserviceCommunication":
            try:

                msComm = msCommTypes.MicroserviceCommunication()
                msg.body.Unpack(msComm)
                for k, v in msg.traces.items():
                    msComm.traces[k] = v

                self.own_grpc_client.ms_comm.send_data(msComm, msComm.data, msComm.metadata)

                # result = handleMsCommunication(msComm, microserviceCommunicator, ctx)
                self.logger.debug(f"After send_data")
                return True
            except Exception as e:
                self.logger.error(f"An unexpected error occurred: {e}")
                return False
        else:
            self.logger.error(f"An unexpected message arrived occurred: {msg.type}")
            return False


    def _consume_with_retry(self, queue_name, max_retries, wait_time):
        for i in range(max_retries):
            try:
                return self._consume(queue_name)
            except grpc.RpcError as e:
                self.logger.error(f"Failed to start consuming (attempt {i+1}/{max_retries}): {e}")
                if self.stop():  # Check if we've been told to stop before handling each response
                    return
                time.sleep(wait_time)


    def _consume(self, queue_name):
        consume_request = rabbitTypes.ConsumeRequest()
        consume_request.queue_name = queue_name
        consume_request.auto_ack = False

        # Handle 1 response only
        try:
            responses = self.stub.ChainConsume(consume_request)

            for response in responses:
                self.handle_incoming_request(response)
                # if self.callback:
                #     if self.callback(self, response):
                #         self.logger.info("query service handled callback successfully")
                #         self.stop()
                #         return True
                #     else:
                #         self.logger.info("Error in query service callback handling")
                #         self.stop()
                #         return False
                # else:
                #     self.logger.warning("no rabbitMq callback registered")
                #     return False
        except grpc.RpcError as e:
            self.logger.error(f"Error on consume: {e}")
            raise e


    def start_consuming(self):
        self.thread = threading.Thread(target=self._start_consuming)
        self.thread.start()

    def _start_consuming(self):
        self._consume_with_retry(self.service_name, 10, 5)

    def stop(self):
        self.logger.info("Stopping RabbitClient...")
        self.thread.join()
        self.logger.info("Rabbit Client stopped")
