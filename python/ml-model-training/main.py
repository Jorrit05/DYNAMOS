import pandas as pd
import re
import time
import sys
import os
from google.protobuf.struct_pb2 import Struct, Value, ListValue
import json
import argparse
from dynamos.ms_init import NewConfiguration
from dynamos.signal_flow import signal_continuation, signal_wait
from dynamos.logger import InitLogger
from dynamos.tracer import InitTracer

from google.protobuf.empty_pb2 import Empty
import microserviceCommunication_pb2 as msCommTypes
import rabbitMQ_pb2 as rabbitTypes
import threading
import time
import sys
from opentelemetry.context.context import Context

from config_prod import service_name

# for the model training functionality
import joblib  # For saving the model
import pandas as pd
from xgboost import XGBRegressor
from sklearn.model_selection import cross_val_score, KFold
from sklearn.metrics import mean_squared_error, make_scorer

# from pathlib import Path
# __file__ is the path to the current script (main.py)
# service_name = Path(__file__).parent.name  # proxy for service name assuming

# --- DYNAMOS Interface code At the TOP ----------------------------------------------------
if os.getenv('ENV') == 'PROD':
    import config_prod as config
else:
    import config_local as config

logger = InitLogger()
# tracer = InitTracer(config.service_name, config.tracing_host)

# Events to start the shutdown of this Microservice, can be used to call 'signal_shutdown'
stop_event = threading.Event()
stop_microservice_condition = threading.Condition()

# Events to make sure all services have started before starting to process a message
# Might be overkill, but good practice
wait_for_setup_event = threading.Event()
wait_for_setup_condition = threading.Condition()

ms_config = None

# --- END DYNAMOS Interface code At the TOP ----------------------------------------------------

#---- LOCAL TEST SETUP OPTIONAL!

# Go into local test code with flag '-t'
parser = argparse.ArgumentParser()
parser.add_argument("-t", "--test", action='store_true')
args = parser.parse_args()
test = args.test

#--------------------------------

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


def protobuf_to_dataframe(data_struct: Struct, metadata: dict = None) -> pd.DataFrame:
    """
    Convert a google.protobuf.Struct and metadata dict back to a pandas DataFrame.
    Assumes data_struct was created by dataframe_to_protobuf above.
    """
    if metadata is None:
        metadata = {}
    data = {}

    for key, value in data_struct.fields.items():
        value_list = value.list_value.values
        items = [v.string_value for v in value_list]
        dtype = metadata.get(key, "object")
        if dtype.startswith("int"):
            data[key] = [int(x) for x in items]
        elif dtype.startswith("float"):
            data[key] = [float(x) for x in items]
        elif dtype == "bool":
            data[key] = [x.lower() in ("true", "1") for x in items]
        else:
            data[key] = items

    return pd.DataFrame(data)

# ---  DYNAMOS Interface code At the Bottom -----------------------------------------------------

def register_service_on_metadata(metadata:dict, service_name:str) -> dict:
    """
    Adds a JSON encoded list of the services that took place on the field "services".
    """
    if "services" in metadata:
        services = json.loads(metadata["services"])
        services.append(service_name)
        metadata["services"] = json.dumps(services)
        return metadata

    metadata["services"] = json.dumps([service_name])

    return metadata


def train_model(df:pd.DataFrame):
    """
    python -m pip install xgboost==2.1.4
    python -m pip install scikit-learn==1.3.2
    """

    # df = synthetic_data
    if "building_id" in df.columns:
        df.drop(columns=["building_id"], inplace=True)
    # Split features and target
    X = df.drop(columns=['consumption'])
    y = df['consumption']

    X = pd.get_dummies(X, columns=["primary_use"])

    # Define XGBoost regressor from sklearn API
    xgb_reg = XGBRegressor(
        n_estimators=100,
        learning_rate=0.1,
        max_depth=6,
        random_state=42,
        n_jobs=-1
    )

    # Define cross-validation strategy
    kf = KFold(n_splits=5, shuffle=True, random_state=42)

    # Use negative mean squared error for regression
    scorer = make_scorer(mean_squared_error, greater_is_better=False)

    # Perform cross-validation
    logger.info("Validating model...")
    cv_scores = cross_val_score(xgb_reg, X, y, cv=kf, scoring=scorer)

    # logger.info(f"Cross-validated MSE scores: {-cv_scores}")
    # logger.info(f"Average MSE: {-cv_scores.mean()}")
    # logger.info(f"Average RMSE: {(-cv_scores.mean())**0.5}")

    evaluation = {
        "mse_scores": [str((-cv_scores))],
        "avg_mse": [-cv_scores.mean()],
        "avg_rmse": [(-cv_scores.mean()) ** 0.5],
    }

    logger.info("evaluation: " + json.dumps(evaluation, indent=2))

    # Original merged dataset
    # Average MSE: 3315628410515.4336
    # Average RMSE: 1820886.7099617794

    # Synthetic dataset
    # Average MSE: 5197084552824.165
    # Average RMSE: 2279711.506490276

    # Both synthetic and DP:
    #  Average MSE: 3832433857370.583
    #  Average RMSE: 1957660.3018324152

    # Optional: Train final model on all data
    xgb_reg.fit(X, y)

    # Save the model to localhost (current directory)
    model_path = "xgb_regressor_model.joblib"
    joblib.dump(xgb_reg, model_path)
    logger.info(f"Model saved to {model_path}")

    return pd.DataFrame(evaluation)



def request_handler(msComm : msCommTypes.MicroserviceCommunication, ctx: Context=None):
    global ms_config
    logger.info(f"Request type: {msComm.request_type}")
    # logger.debug(f"ml: {msComm.request_type}")

    # Ensure all connections have finished setting up before processing data
    signal_wait(wait_for_setup_event, wait_for_setup_condition)

    try:
        if msComm.request_type == "sqlDataRequest":
            sqlDataRequest = rabbitTypes.SqlDataRequest()
            msComm.original_request.Unpack(sqlDataRequest)

            # logger.debug(f"msComm: {msComm}")
            # logger.debug(f"msComm.data: {msComm.data}")

            mscomm_metadata = dict(msComm.metadata)
            logger.debug(f"msComm metadata original: {str(mscomm_metadata)}")
            mscomm_metadata = register_service_on_metadata(mscomm_metadata, service_name=service_name)
            logger.debug(f"msComm metadata updated: {str(mscomm_metadata)}")

            dataframe_metadata_dict = json.loads(msComm.metadata['dataframe_metadata'])
            data_df = protobuf_to_dataframe(msComm.data, dataframe_metadata_dict)
            logger.debug(f"df head: {data_df.head()}")

            eval_metrics_df = train_model(data_df)

            data, dataframe_metadata = dataframe_to_protobuf(eval_metrics_df)
            mscomm_metadata['dataframe_metadata'] = json.dumps(dataframe_metadata)

            # # with tracer.start_as_current_span("process_sql_data_request", context=ctx) as span1:
            # data, metadata = process_sql_data_request(sqlDataRequest, ctx)
                # span1.set_attribute("handleMsCommunication finished:", metadata)

            logger.debug(f"Forwarding result, dataframe: {eval_metrics_df}")
            logger.debug(f"Forwarding result, metadata: {mscomm_metadata}")
            ms_config.next_client.ms_comm.send_data(msComm, data, mscomm_metadata)
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

    if test:
        logger.info("Running in test mode")
        return

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

# ---  END DYNAMOS Interface code At the Bottom -------------------------------------------------


if __name__ == "__main__":
    main()
