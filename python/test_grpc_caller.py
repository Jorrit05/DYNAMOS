
from dynamos.ms_init import NewConfiguration
from dynamos.signal_flow import signal_continuation, signal_wait
from dynamos.logger import InitLogger
from google.protobuf.empty_pb2 import Empty

import microserviceCommunication_pb2 as msCommTypes
import threading
import time
import sys


# --- DYNAMOS Interface code At the TOP ----------------------------------------------------
logger = InitLogger()

# Events to start the shutdown of this Microservice, can be used to call 'signal_shutdown'
stop_event = threading.Event()
stop_microservice_condition = threading.Condition()

# Events to make sure all services have started before starting to process a message
# Might be overkill, but good practice
wait_for_setup_event = threading.Event()
wait_for_setup_condition = threading.Condition()

config = None

# --- END DYNAMOS Interface code At the TOP ----------------------------------------------------


# Functionality code ----------------------------------------------------


#-------------------------------------------------------------------------------------------



# ---  DYNAMOS Interface code At the Bottom -----------------------------------------------------

def request_handler(msComm : msCommTypes.MicroserviceCommunication):
    try:
        logger.info(f"Received original request type: {msComm.request_type}")

        # Ensure all connections have finished setting up before processing data
        signal_wait(wait_for_setup_event, wait_for_setup_condition)

        if msComm.request_type == "example_request_type":
            # Do logic....

            extra_metadata = {
                "extra_key1": "extra_value1",
                "extra_key2": "extra_value2"
            }

            config.next_client.ms_comm.send_data(msComm, msComm.data, extra_metadata)
    except Exception as e:
        logger.error(f"Error in request_handler: {e}")


    signal_continuation(stop_event, stop_microservice_condition)
    return Empty()


config = NewConfiguration("caller", "localhost:", request_handler)

# Signal the message handler that all connections have been created
signal_continuation(wait_for_setup_event, wait_for_setup_condition)

# Wait for the end of processing to shutdown this Microservice
try:
    signal_wait(stop_event, stop_microservice_condition)

except KeyboardInterrupt:
    print("KeyboardInterrupt received, stopping server...")
    signal_continuation(stop_event, stop_microservice_condition)


config.grpc_server.stop()
config.next_client.close_program()
time.sleep(2)

logger.debug(f"Exiting {config.service_name}")
sys.exit(0)

# ---  END DYNAMOS Interface code At the Bottom -------------------------------------------------
