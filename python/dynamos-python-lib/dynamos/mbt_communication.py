import base64
import json
import threading
import pika
import os

class TestQueuePublisher:
    def __init__(self, queue_name="mbt_testing_queue"):
        self.queue_name = queue_name
        self.host = "rabbitmq.core.svc.cluster.local"
        self.port = 5672
        self.user = os.getenv("AMQ_USER", "guest")
        self.password = os.getenv("AMQ_PASSWORD", "guest")
        self.connection = None
        self.channel = None

    def _connect(self):
        if self.connection and self.connection.is_open:
            return

        credentials = pika.PlainCredentials(self.user, self.password)
        parameters = pika.ConnectionParameters(self.host, self.port, "/", credentials)
        self.connection = pika.BlockingConnection(parameters)
        self.channel = self.connection.channel()

        self.channel.queue_declare(queue=self.queue_name, durable=True)

    def send_message_async(self, msg_type: str, proto_msg):
        def task():
            try:
                self._connect()
                payload = self._prepare_payload(msg_type, proto_msg)
                self.channel.basic_publish(
                    exchange="",
                    routing_key=self.queue_name,
                    body=payload,
                    properties=pika.BasicProperties(content_type="application/json")
                )
            except Exception as e:
                print(f"Error sending message to RabbitMQ: {e}")

        threading.Thread(target=task).start()

    def _prepare_payload(self, msg_type: str, proto_msg) -> str:
        binary_body = proto_msg.SerializeToString()
        encoded_body = base64.b64encode(binary_body).decode("utf-8")
        message = {
            "type": msg_type,
            "body": encoded_body
        }
        return json.dumps(message)
