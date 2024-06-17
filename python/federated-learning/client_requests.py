import requests
import json
import numpy as np
import pandas as pd
import h5py

# Function to train the client model and return the parameters
def train_client(X_train, y_train):
    # Convert DataFrame and Series to list of lists and list respectively
    X_train_list = X_train.values.tolist()
    y_train_list = y_train.values.tolist()
    
    payload = {'X_train': X_train_list, 'y_train': y_train_list}
    response = requests.post("http://localhost:5002/train_client", json=payload)
    if response.status_code == 200:
        print("Training completed successfully.")
        return response.json()  # Return the client parameters
    else:
        print(f"Error: {response.status_code}")
        return None

# Function to send client parameters and aggregate them
def aggregate_models(client_params, mode='sequential'):
    payload = {'client_params': client_params, 'mode': mode}
    response = requests.post("http://localhost:5002/aggregate", json=payload)
    if response.status_code == 200:
        print("Aggregation done successfully.")
    else:
        print(f"Error: {response.status_code}")

if __name__ == '__main__':
    # Process each dataset and perform training and aggregation sequentially
    files = ['datasets/part_1_N-CMAPSS_DS01-005.h5', 'datasets/part_2_N-CMAPSS_DS01-005.h5']
    
    for file in files:
        with h5py.File(file, 'r') as hdf:
            # Load train data
            dev_data = np.array(hdf.get('dev_data'))
            column_name = [col.decode('utf-8') for col in hdf.get('column_name')]
            test_data = np.array(hdf.get('test_data'))

            # Create DataFrame for data
            df_train = pd.DataFrame(data=dev_data, columns=column_name)
            df_test = pd.DataFrame(data=test_data, columns=column_name)

            X_train = df_train.drop(columns=["RUL"])  # Features
            y_train = df_train['RUL']  # Target variable

            # Train client and get parameters
            client_params = train_client(X_train, y_train)
            
            # Aggregate client parameters after each training
            if client_params:
                aggregate_models(client_params, mode='sequential')

