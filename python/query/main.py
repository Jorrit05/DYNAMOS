from opentracing import Format
import pandas as pd
from pandasql import sqldf
import re
import time
import sys
import os
from tracer import Tracer
from rabbit_client import RabbitClient
from microservice_client import MsCommunication
from google.protobuf.struct_pb2 import Struct, Value, ListValue
import rabbitMQ_pb2 as rabbitTypes
import microserviceCommunication_pb2 as msCommTypes

from my_logger import InitLogger
import argparse
from jaeger_client import Config
from opentracing.propagation import Format

if os.getenv('ENV') == 'PROD':
    import config_prod as config
else:
    import config_local as config


# globals
logger = InitLogger()
rabbitClient = None
microserviceCommunicator = None
tracer_class = Tracer(config.service_name)

# Go into local test code with flag '-t'
parser = argparse.ArgumentParser()
parser.add_argument("-t", "--test", action='store_true')
args = parser.parse_args()
test = args.test

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


def process_sql_data_request(sqlDataRequest, msComm):
    logger.debug("Start process_sql_data_request")
    try:
        print(config.dataset_filepath)
        result = load_and_query_csv(config.dataset_filepath, sqlDataRequest.query)
        logger.debug("Got result")
        logger.debug(result)
        data, metadata = dataframe_to_protobuf(result)
        logger.debug("Got 1")

        microserviceCommunicator = MsCommunication(config.grpc_addr, tracer_class)
        logger.debug("Got 2")
        microserviceCommunicator.SendData("sqlDataRequest", data, metadata, msComm)
        logger.debug("Got 3")
    except FileNotFoundError:
        logger.error(f"File not found at path {config.dataset_filepath}")
    except Exception as e:
        logger.error(f"An error occurred: {str(e)}")


def handleMsCommunication(rabbitClient, msComm):
    logger.warn(type(msComm))

    logger.warn(f"response.request_type: {msComm.request_type}")

    if msComm.request_type == "sqlDataRequest":

        sqlDataRequest = rabbitTypes.SqlDataRequest()
        msComm.original_request.Unpack(sqlDataRequest)

        logger.info("Query: " + sqlDataRequest.query)
        process_sql_data_request(sqlDataRequest, msComm)
        return True

    else:
        logger.error(f"An unexpected msCommunication: {msComm.request_type}")
        return False

def handle_incoming_request(rabbitClient, msg):
    logger.debug("Start handle_incoming_request")
    logger.debug(f"msg.type is:  {msg.type}")

    print(f"TYPE if trace: {type(msg.trace)}")
    if msg.trace:
        logger.info(f"Incoming trace: has data")
        # # tracer_class.printSpan(msg.trace)
    else:
        print("NO TRACE")

    if msg.type == "microserviceCommunication":
        try:

            msComm = msCommTypes.MicroserviceCommunication()
            msg.body.Unpack(msComm)
            msComm.trace = msg.trace

            # extract trace context from msComm
            # carrier = {"trace": msComm.trace}  # replace with actual trace field name
            # span_context = rabbitClient.tracer.extract(Format.TEXT_MAP, carrier)
            # with rabbitClient.tracer.start_active_span("handle_incoming_request", child_of=span_context) as scope:
            handleMsCommunication(rabbitClient, msComm)
            logger.debug("Returning True")
            rabbitClient.close_program()
            tracer_class.close_tracer()
            return True
        except Exception as e:
            logger.error(f"Failed to unmarshal message: {e}")
        except:
            logger.error("An unexpected error occurred.")
    else:
        logger.error(f"An unexpected message arrived occurred: {msg.type}")
        return False

def test_single_query():
    # Define your SQL query
    query = """SELECT DISTINCT p.Unieknr, p.Geslacht, p.Gebdat, s.Aanst_22, s.Functcat, s.Salschal as Salary
               FROM Personen p
               JOIN Aanstellingen s
               ON p.Unieknr = s.Unieknr LIMIT 4"""

    # Load the CSV file and execute the query
    result_df = load_and_query_csv(config.dataset_filepath, query)
    data, metadata = dataframe_to_protobuf(result_df)

    print("--------------\ndata:")
    print(data)
    print("--------------\nmetadata:")
    print(metadata)


def main():
    if test:
        test_single_query()
        exit(0)

    logger.debug("Starting Query service")

    if int(os.getenv("FIRST")) > 0:
        logger.debug("First service")
        job_name = os.getenv("JOB_NAME")
        rabbitClient = RabbitClient(config.grpc_addr, job_name, job_name, True, tracer_class, handle_incoming_request)
        rabbitClient.start_consuming(job_name, 10, 2)
    else:
        #TODO: Setup listener service for Python
        # microserviceCommunicator = MsCommunication(config.grpc_addr)
        # microserviceCommunicator.
        logger.debug("Not the first service")
        exit(1)

    logger.debug("Exiting query service")
    sys.exit(0)


if __name__ == "__main__":
    main()
