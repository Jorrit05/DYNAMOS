import sys
import os

from dynamos.grpc_client import GRPCClient
from dynamos.logger import InitLogger
from dynamos.tracer import InitTracer
from dynamos.ms_init import NewConfiguration

if os.getenv('ENV') == 'PROD':
    import config_prod as config
else:
    import config_local as config


#---- GLOBALS
logger = InitLogger()
tracer = InitTracer(config.service_name, config.tracing_host)

#------------

client = GRPCClient(config.grpc_addr)

# Check health status
health_status = client.health.check_health()
print(f"Health status: {health_status}")

client.close_program()


def main():
    logger.debug("Starting Test service")


    conf = NewConfiguration(
        service_name=config.service_name,
        grpc_addr="localhost:",
        COORDINATOR="COORDINATOR",
        sidecar_callback=lambda: lambda ctx, grpc_msg: None,
        grpc_callback=lambda ctx, data: None,
        receive_mutex="receive_mutex"
    )


    logger.debug("Exiting Test service")
    sys.exit(0)


if __name__ == "__main__":
    main()
