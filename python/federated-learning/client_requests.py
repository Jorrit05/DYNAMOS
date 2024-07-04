import requests
import json
import numpy as np
import pandas as pd
import h5py

def train_client(X_train, y_train):
    X_train_list = X_train.values.tolist()
    y_train_list = y_train.values.tolist()
    
    payload = {'X_train': X_train_list, 'y_train': y_train_list}
    response = requests.post("http://localhost:5000/train", json=payload)
    if response.status_code == 200:
        print("Training completed successfully.")
        return response.json()
    else:
        print(f"Error: {response.status_code}")
        return None

def aggregate_models(client_params, mode='sequential'):
    payload = {'client_params': client_params, 'mode': mode}
    response = requests.post("http://localhost:5002/aggregate", json=payload)
    if response.status_code == 200:
        print("Aggregation done successfully.")
    else:
        print(f"Error: {response.status_code}")

def evaluate_global_model(X_test, y_test):
    global_model_response = requests.get("http://localhost:5004/get_global_model")
    if global_model_response.status_code == 200:
        global_model = global_model_response.json()
        payload = {'X_test': X_test.values.tolist(), 'y_test': y_test.values.tolist(), 'global_model': global_model}
        response = requests.post("http://localhost:5003/evaluate", json=payload)
        if response.status_code == 200:
            print("Evaluation done successfully.")
            print(response.json())
        else:
            print(f"Error: {response.status_code}")
    else:
        print(f"Error fetching global model: {global_model_response.status_code}")

if __name__ == '__main__':
    files = ['part_1_N-CMAPSS_DS01-005.h5', 'part_2_N-CMAPSS_DS01-005.h5', 'part_3_N-CMAPSS_DS01-005.h5']
    
    for file in files:
        with h5py.File(file, 'r') as hdf:
            dev_data = np.array(hdf.get('dev_data'))
            column_name = [col.decode('utf-8') for col in hdf.get('column_name')]
            test_data = np.array(hdf.get('test_data'))

            df_train = pd.DataFrame(data=dev_data, columns=column_name)
            df_test = pd.DataFrame(data=test_data, columns=column_name)

            X_train = df_train.drop(columns=["RUL"])
            y_train = df_train['RUL']
            X_test = df_test.drop(columns=["RUL"])
            y_test = df_test['RUL']

            client_params = train_client(X_train, y_train)
            
            if client_params:
                aggregate_models(client_params, mode='sequential')
                evaluate_global_model(X_test, y_test)
