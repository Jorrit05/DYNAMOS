import requests
import json
import numpy as np
import pandas as pd
import numpy as np
import h5py

# Example usage to send training data to train_client1 endpoint
def train_client(X_train, y_train):
    # Convert DataFrame and Series to list of lists and list respectively
    X_train_list = X_train.values.tolist()
    y_train_list = y_train.values.tolist()
    
    payload = {'X_train': X_train_list, 'y_train': y_train_list}
    response = requests.post("http://localhost:5002/train_client", json=payload)
    if response.status_code == 200:
        print("Training completed successfully.")
    else:
        print(f"Error: {response.status_code}")



# Example usage to send test data and get evaluation metrics
def aggregate_models(X_test, y_test):
    payload = {'X_test': X_test.values.tolist(), 'y_test': y_test.values.tolist()}
    response = requests.post("http://localhost:5002/aggregate", json=payload)
    if response.status_code == 200:
        data = response.json()
        print("Evaluation Done!")
    else:
        print(f"Error: {response.status_code}")

if __name__ == '__main__':


# Open the H5 file
    with h5py.File('part_1_N-CMAPSS_DS01-005.h5', 'r') as hdf:
        # Load test data
        dev_data = np.array(hdf.get('dev_data'))
        column_name = [col.decode('utf-8') for col in hdf.get('column_name')]
        test_data = np.array(hdf.get('test_data'))

        # Create DataFrame for data
        df_train = pd.DataFrame(data=dev_data, columns=column_name)
        df_test = pd.DataFrame(data=test_data, columns=column_name)

        X_train1 = df_train.drop(columns=["RUL"]) # Features
        y_train1 = df_train['RUL']   # Target variable
        train_client(X_train1, y_train1)

    with h5py.File('part_2_N-CMAPSS_DS01-005.h5', 'r') as hdf:
        # Load test data
        dev_data = np.array(hdf.get('dev_data'))
        column_name = [col.decode('utf-8') for col in hdf.get('column_name')]
        test_data = np.array(hdf.get('test_data'))

        # Create DataFrame for data
        df_train = pd.DataFrame(data=dev_data, columns=column_name)
        df_test = pd.DataFrame(data=test_data, columns=column_name)

        X_train2 = df_train.drop(columns=["RUL"]) # Features
        y_train2 = df_train['RUL']   # Target variable
        X_test = df_test.drop(columns=["RUL"])  # Features
        y_test = df_test['RUL']   # Target variable

        train_client(X_train2, y_train2)

        aggregate_models(X_test, y_test)

