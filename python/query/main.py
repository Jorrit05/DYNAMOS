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

from opentelemetry.instrumentation.grpc import GrpcInstrumentorClient
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

grpc_server_instrumentor = GrpcInstrumentorClient()
grpc_server_instrumentor.instrument()
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

    for table_name in table_names:
        try:
            dfs[table_name] = pd.read_csv(f"{file_path_prefix}{table_name}.csv", delimiter=';')
        except FileNotFoundError:
            logger.error(f"CSV file for table {table_name} not found.")
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


        with tracer.start_as_current_span("load_and_query_csv") as span1:
            result = load_and_query_csv(config.dataset_filepath, sqlDataRequest.query)

        with tracer.start_as_current_span("dataframe_to_protobuf") as span2:
            data, metadata = dataframe_to_protobuf(result)
        logger.debug("Got 1")
        with tracer.start_as_current_span("SendData") as span3:
            microserviceCommunicator.SendData("sqlDataRequest", data, metadata, msComm)

        logger.debug("Got 2")
    except FileNotFoundError:
        logger.error(f"File not found at path {config.dataset_filepath}")
    except Exception as e:
        logger.error(f"An error occurred: {str(e)}")


def handleMsCommunication(msComm, microserviceCommunicator, ctx):
    logger.info(type(msComm))

    logger.info(f"response.request_type: {msComm.request_type}")

    if msComm.request_type == "sqlDataRequest":

        sqlDataRequest = rabbitTypes.SqlDataRequest()
        msComm.original_request.Unpack(sqlDataRequest)

        logger.info("Query: " + sqlDataRequest.query)

        with tracer.start_as_current_span("process_sql_data_request", context=ctx) as span1:
            process_sql_data_request(sqlDataRequest, msComm, microserviceCommunicator, ctx)
        return True

    else:
        logger.error(f"An unexpected msCommunication: {msComm.request_type}")
        return False


def handle_incoming_request(rabbitClient, msg):
    print(f"TYPE if trace: {type(msg.traces)}")

    logger.info(len(msg.traces))
    # Parse the trace header back into a dictionary
    scMap = json.loads(msg.traces["jsonTrace"])
    logger.warning(scMap)
    state = TraceState([("sampled", "1")])
    sc = trace.SpanContext(
        trace_id=int(scMap['TraceID'], 16),
        span_id=int(scMap['SpanID'], 16),
        is_remote=True,
        trace_flags=TraceFlags(TraceFlags.SAMPLED),
        trace_state=state
    )

    logger.warning(f"sc.trace_id: {sc.trace_id}")

    # ctx = set_span_in_context(sc)
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

            handleMsCommunication(msComm, microserviceCommunicator, ctx)
            logger.debug("Returning True")
            rabbitClient.close_program()
            return True
        except Exception as e:
            logger.error(f"Failed to unmarshal message: {e}")
            return False
        except:
            logger.error("An unexpected error occurred.")
            return False
    else:
        logger.error(f"An unexpected message arrived occurred: {msg.type}")
        return False

# @tracer.start_as_current_span("test_single_query")
def test_single_query():

    # # Define your SQL query
    # query = """SELECT DISTINCT p.Unieknr, p.Geslacht, p.Gebdat, s.Aanst_22, s.Functcat, s.Salschal as Salary
    #            FROM Personen p
    #            JOIN Aanstellingen s
    #            ON p.Unieknr = s.Unieknr LIMIT 4"""

    # Define your SQL query
    query = """SELECT *
               FROM Personen p
               JOIN Aanstellingen s LIMIT 30000"""

    # Load the CSV file and execute the query
    result_df = load_and_query_csv(config.dataset_filepath, query)
    # data, metadata = dataframe_to_protobuf(result_df)

    # print("--------------\ndata:")
    # print(data)
    # print("--------------\nmetadata:")
    # print(metadata)

    # with open("output.json", "w") as file1:
        # Writing data to a file
    start = time.time()
    result_df.to_csv('output.txt', sep='\t', index=False)
    end = time.time()
    print(f'Time elapsed for file write: {end - start} seconds')

        # file1.write(df.to_csv('output.txt', sep='\t', index=False))
        # file1.writelines(L)

def main():
    if test:
        job_name="Test"

        test_single_query()

        exit(0)

    logger.debug("Starting Query service")

    if int(os.getenv("FIRST")) > 0:
        logger.debug("First service")
        job_name = os.getenv("JOB_NAME")
        rabbitClient = RabbitClient(config, job_name, job_name, False, handle_incoming_request)
        rabbitClient.start_consuming(job_name, 10, 2)
    else:
        #TODO: Setup listener service for Python
        logger.debug("Not the first service")
        exit(1)



    logger.debug("Exiting query service")
    sys.exit(0)


if __name__ == "__main__":
    main()
