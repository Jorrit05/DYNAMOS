import pandas as pd
from pandasql import sqldf
import re
import os
from grpc_lib import EtcdClient, MsCommunication
from my_logger import InitLogger
if os.getenv('ENV') == 'PROD':
    import config_prod as config
else:
    import config_local as config

logger = InitLogger()

def load_and_query_csv(file_path_prefix, query):
    # Extract table names from the query
    table_names = re.findall(r'FROM (\w+)', query) + re.findall(r'JOIN (\w+)', query)
    # Create a dictionary to hold DataFrames, keyed by table name
    dfs = {}

    for table_name in table_names:
        try:
            dfs[table_name] = pd.read_csv(f"{file_path_prefix}{table_name}.csv", delimiter=';')
        except FileNotFoundError:
            print(f"CSV file for table {table_name} not found.")
            return None

    # Use pandasql's sqldf function to execute the SQL query
    result_df = sqldf(query, dfs)

    return result_df


def main():
    logger.info("Starting Query service")
    if int(os.getenv("FIRST")) > 0:
        logger.info("First service")
    else:
        logger.info("Not the first service")

    # microserviceCommunicator = MsCommunication()
    # response = microserviceCommunicator.SendData()

    logger.info("Finishing work, exiting")
    exit(0)
    # Define the prefix of your CSV files' paths
    file_path_prefix = '/Users/jorrit/Documents/master-software-engineering/thesis/DYNAMOS/configuration/datasets/'

    # Define your SQL query
    query = """SELECT p.Unieknr, p.Geslacht, p.Gebdat, s.Aanst_22 as Salary
               FROM Personen p
               JOIN Aanstellingen s
               ON p.Unieknr = s.Unieknr"""

    # Load the CSV file and execute the query
    result_df = load_and_query_csv(file_path_prefix, query)

    # Print the resulting DataFrame
    print(result_df)

if __name__ == "__main__":
    main()
