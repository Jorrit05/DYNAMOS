from dynamos.logger import InitLogger
from dynamos.tracer import InitTracer


class BaseClient:
    def __init__(self, channel, service_name):
        self.service_name = service_name
        self.channel = channel
        self.logger = InitLogger()
        self.logger.info("1")
        self.tracer = InitTracer(self.service_name, "http://localhost:32003")
        self.logger.info("2")
