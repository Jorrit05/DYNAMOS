import pandas as pd
from pandasql import sqldf
import re
import time
import sys
import os
from rabbit_client import RabbitClient
from microservice_client import MsCommunication
from google.protobuf.struct_pb2 import Struct, Value, ListValue
import rabbitMQ_pb2 as rabbitTypes
import microserviceCommunication_pb2 as msCommTypes
import json
from my_logger import InitLogger
import argparse
from opentelemetry import trace, context
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import SERVICE_NAME, Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.propagate import extract
from opentelemetry.trace.span import TraceFlags, TraceState
from opentelemetry.trace.propagation import set_span_in_context

if os.getenv('ENV') == 'PROD':
    import config_prod as config
else:
    import config_local as config


# globals
logger = InitLogger()
rabbitClient = None
microserviceCommunicator = None

# Service name is required for most backends
resource = Resource(attributes={
    SERVICE_NAME: config.service_name
})

provider = TracerProvider(resource=resource)
processor = BatchSpanProcessor(OTLPSpanExporter(endpoint=config.tracingHost, insecure=True))
provider.add_span_processor(processor)
trace.set_tracer_provider(provider)

tracer = trace.get_tracer("query.tracer")

# Go into local test code with flag '-t'
parser = argparse.ArgumentParser()
parser.add_argument("-t", "--test", action='store_true')
args = parser.parse_args()
test = args.test

@tracer.start_as_current_span("load_and_query_csv")
def load_and_query_csv(file_path_prefix, query):
    # Extract table names from the query
    table_names = re.findall(r'FROM (\w+)', query) + re.findall(r'JOIN (\w+)', query)
    # Create a dictionary to hold DataFrames, keyed by table name
    dfs = {}
    DATA_STEWARD_NAME = os.getenv("DATA_STEWARD_NAME")
    if DATA_STEWARD_NAME == "":
        logger.error(f"DATA_STEWARD_NAME not set.")


    for table_name in table_names:
        try:
            file_name = f"{file_path_prefix}{table_name}_{DATA_STEWARD_NAME}.csv"
            logger.debug(f"Loading file {file_name}")
            dfs[table_name] = pd.read_csv(file_name, delimiter=';')
        except FileNotFoundError:
            logger.error(f"CSV file for table {table_name}_{DATA_STEWARD_NAME} not found.")
            return None

    # Use pandasql's sqldf function to execute the SQL query
    result_df = sqldf(query, dfs)

    return result_df


@tracer.start_as_current_span("dataframe_to_protobuf")
def dataframe_to_protobuf(df):
    # Convert the DataFrame to a dictionary of lists (one for each column)
    data_dict = df.to_dict(orient='list')

    # Convert the dictionary to a Struct
    data_struct = Struct()

    # Iterate over the dictionary and add each value to the Struct
    for key, values in data_dict.items():
        # Pack each item of the list into a Value object
        value_list = [Value(string_value=str(item)) for item in values]
        # Pack these Value objects into a ListValue
        list_value = ListValue(values=value_list)
        # Add the ListValue to the Struct
        data_struct.fields[key].CopyFrom(Value(list_value=list_value))

    # Create the metadata
    # Infer the data types of each column
    data_types = df.dtypes.apply(lambda x: x.name).to_dict()
    # Convert the data types to string values
    metadata = {k: str(v) for k, v in data_types.items()}

    return data_struct, metadata

def process_sql_data_request(sqlDataRequest, msComm, microserviceCommunicator, ctx):
    logger.debug("Start process_sql_data_request")

    try:
        result = load_and_query_csv(config.dataset_filepath, sqlDataRequest.query)
        data, metadata = dataframe_to_protobuf(result)

        with tracer.start_as_current_span("SendData") as span3:
            microserviceCommunicator.SendData("sqlDataRequest", data, metadata, msComm)

        logger.debug("After sendData")
        return True
    except FileNotFoundError:
        logger.error(f"File not found at path {config.dataset_filepath}")
        return False
    except Exception as e:
        logger.error(f"An error occurred: {str(e)}")
        return False

def handleMsCommunication(msComm, microserviceCommunicator, ctx):
    if msComm.request_type == "sqlDataRequest":

        sqlDataRequest = rabbitTypes.SqlDataRequest()
        msComm.original_request.Unpack(sqlDataRequest)

        with tracer.start_as_current_span("process_sql_data_request", context=ctx) as span1:
            result = process_sql_data_request(sqlDataRequest, msComm, microserviceCommunicator, ctx)
            span1.set_attribute("handleMsCommunication finished:", result)
            return result

    else:
        logger.error(f"An unexpected msCommunication: {msComm.request_type}")
        return False


def handle_incoming_request(rabbitClient, msg):
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

    microserviceCommunicator = MsCommunication(config, ctx)
    # _, span = tracer_class.create_parent_span("handle_incoming", msg.trace)
    if msg.type == "microserviceCommunication":
        try:

            msComm = msCommTypes.MicroserviceCommunication()
            msg.body.Unpack(msComm)
            for k, v in msg.traces.items():
                msComm.traces[k] = v

            result = handleMsCommunication(msComm, microserviceCommunicator, ctx)
            logger.debug(f"Returning restult {result}")
            return result
        except Exception as e:
            logger.error(f"An unexpected error occurred: {e}")
            return False
    else:
        logger.error(f"An unexpected message arrived occurred: {msg.type}")
        return False

# @tracer.start_as_current_span("test_single_query")
def test_single_query():
    size = "100"
    # Define your SQL query
    query = f"""SELECT *
               FROM Personen p
               JOIN Aanstellingen s LIMIT {size}"""

    # print(msg)
    # Load the CSV file and execute the query
    result_df = load_and_query_csv(config.dataset_filepath, query)

    # with open("output.json", "w") as file1:
        # Writing data to a file
    start = time.time()
    result_df.to_csv(f"output_{size}.txt", sep='\t', index=False)
    # df = pd.read_csv(f"output_{size}.txt", sep='\t')
    end = time.time()
    print(f'Time elapsed for file write: {end - start} seconds')

    return True

def main():
    if test:
        job_name="Test"

        # rabbitClient = RabbitClient(config, job_name, job_name, test_single_query)
        # result = rabbitClient.start_consuming(job_name, 10, 2)
        # logger.info("lets wait a few seconds before quitting")
        # time.sleep(5)
        test_single_query()

        exit(0)

    logger.debug("Starting Query service")

    if int(os.getenv("FIRST")) > 0:
        # logger.debug("First service")
        job_name = os.getenv("JOB_NAME")
        rabbitClient = RabbitClient(config, job_name, job_name, handle_incoming_request, False)
        rabbitClient.start_consuming(job_name, 10, 2)
    else:
        #TODO: Setup listener service for Python
        # logger.debug("Not the first service")
        exit(1)



    logger.debug("Exiting query service")
    sys.exit(0)


if __name__ == "__main__":
    main()
