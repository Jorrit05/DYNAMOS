
from dynamos.ms_init import NewConfiguration
from dynamos.signal_flow import signal_continuation, signal_wait
from dynamos.logger import InitLogger
import rabbitMQ_pb2 as rabbitTypes

from google.protobuf.empty_pb2 import Empty
from opentelemetry.context.context import Context

import microserviceCommunication_pb2 as msCommTypes
import threading
import time
import sys
import os

# --- DYNAMOS Interface code At the TOP ----------------------------------------------------
if os.getenv('ENV') == 'PROD':
    import config_prod as config
else:
    import config_local as config

logger = InitLogger()

# Events to start the shutdown of this Microservice, can be used to call 'signal_shutdown'
stop_event = threading.Event()
stop_microservice_condition = threading.Condition()

# Events to make sure all services have started before starting to process a message
# Might be overkill, but good practice
wait_for_setup_event = threading.Event()
wait_for_setup_condition = threading.Condition()

ms_config = None

# --- END DYNAMOS Interface code At the TOP ----------------------------------------------------



# Functionality code ----------------------------------------------------



def logic(sqlDataRequest, ctx):
    return "Hello world!", {}

#-------------------------------------------------------------------------------------------

# ---  DYNAMOS Interface code At the Bottom -----------------------------------------------------

def request_handler(msComm : msCommTypes.MicroserviceCommunication, ctx: Context):
    global ms_config
    logger.info(f"Received original request type: {msComm.request_type}")

    # Ensure all connections have finished setting up before processing data
    signal_wait(wait_for_setup_event, wait_for_setup_condition)

    try:
        if msComm.request_type == "sqlDataRequest":
            sqlDataRequest = rabbitTypes.SqlDataRequest()
            msComm.original_request.Unpack(sqlDataRequest)

            # with tracer.start_as_current_span("process_sql_data_request", context=ctx) as span1:
            data, metadata = logic(sqlDataRequest, ctx)
                # span1.set_attribute("handleMsCommunication finished:", metadata)

            logger.debug(f"Forwarding result, metadata: {metadata}")
            ms_config.next_client.ms_comm.send_data(msComm, data, metadata)
            signal_continuation(stop_event, stop_microservice_condition)

        else:
            logger.error(f"An unknown request_type: {msComm.request_type}")

        return Empty()
    except Exception as e:
        logger.error(f"An unexpected error occurred: {e}")
        return Empty()



def main():
    global config
    global ms_config

    ms_config = NewConfiguration(config.service_name, config.grpc_addr, request_handler)

    # Signal the message handler that all connections have been created
    signal_continuation(wait_for_setup_event, wait_for_setup_condition)

    # Wait for the end of processing to shutdown this Microservice
    try:
        signal_wait(stop_event, stop_microservice_condition)

    except KeyboardInterrupt:
        print("KeyboardInterrupt received, stopping server...")
        signal_continuation(stop_event, stop_microservice_condition)


    ms_config.stop(2)
    logger.debug(f"Exiting {config.service_name}")
    sys.exit(0)


if __name__ == "__main__":
    main()
# ---  END DYNAMOS Interface code At the Bottom -------------------------------------------------
