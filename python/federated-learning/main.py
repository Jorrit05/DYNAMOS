from flask import Flask, request, jsonify
from sklearn.linear_model import LinearRegression
from sklearn.metrics import r2_score, mean_squared_error
import numpy as np
import h5py


app = Flask(__name__)

# Global model parameters
global_model = LinearRegression()
global_model.coef_ = np.zeros(1)  # Initial coefficients
global_model.intercept_ = 0.0     # Initial intercept
num_clients = 0  # Counter to keep track of the number of clients

# Storage for client parameters in aggregate mode
client_parameters = []

def evaluate(y_true, y_pred, label='test'):
    mse = mean_squared_error(y_true, y_pred)
    rmse = np.sqrt(mse)
    r2 = r2_score(y_true, y_pred)
    print(f'{label} set RMSE: {rmse}, R2: {r2}')

@app.route('/get_global_model', methods=['GET'])
def get_global_model():
    model_params = {
        'coef': global_model.coef_.tolist(), 
        'intercept': [global_model.intercept_]  # Wrap intercept in a list
    }
    return jsonify(model_params)

@app.route('/train_client', methods=['POST'])
def train_client():
    data = request.get_json()
    X_train = np.array(data['X_train'])
    y_train = np.array(data['y_train'])
    
    # Train local model
    local_model = global_model
    local_model.fit(X_train, y_train)
    
    # Send model parameters back to server
    params = {
        'coef': local_model.coef_.tolist(),
        'intercept': [local_model.intercept_]  # Wrap intercept in a list
    }
    return jsonify(params)

@app.route('/aggregate', methods=['POST'])
def aggregate_models():
    global global_model, num_clients, client_parameters
    data = request.get_json()
    client_params = data['client_params']
    mode = data.get('mode', 'sequential')  # Default to 'aggregate' mode if not specified
    
    if mode == 'sequential':
        # Extract parameters from client
        client_coef = np.array(client_params['coef'])
        client_intercept = client_params['intercept'][0]
        
        # Incrementally aggregate the client's model parameters with the global model
        if num_clients == 0:
            global_model.coef_ = client_coef
            global_model.intercept_ = client_intercept
        else:
            global_model.coef_ = (global_model.coef_ * num_clients + client_coef) / (num_clients + 1)
            global_model.intercept_ = (global_model.intercept_ * num_clients + client_intercept) / (num_clients + 1)
        
        num_clients += 1
        
        return "Model aggregated sequentially."
    
    elif mode == 'aggregate':
        # Store the client parameters
        client_parameters.append(client_params)
        
        # If you decide to aggregate only when a certain number of clients have sent their models, you can add a check here
        # Example: if len(client_parameters) >= expected_num_clients:
        
        return "Client parameters received. Waiting for other clients to aggregate."
    
    else:
        return "Invalid mode specified.", 400

@app.route('/final_aggregate', methods=['POST'])
def final_aggregate():
    global global_model, num_clients, client_parameters
    if not client_parameters:
        return "No client parameters received.", 400
    
    # Aggregate all collected client parameters
    total_clients = len(client_parameters)
    coefs = np.mean([np.array(params['coef']) for params in client_parameters], axis=0)
    intercepts = np.mean([params['intercept'][0] for params in client_parameters])
    
    global_model.coef_ = coefs
    global_model.intercept_ = intercepts
    
    # Clear client parameters after aggregation
    client_parameters = []
    num_clients = 0
    
    return "Final model aggregated."

@app.route('/evaluate', methods=['POST'])
def evaluate_model():
    data = request.get_json()
    X_test = np.array(data['X_test'])
    y_test = np.array(data['y_test'])
    
    y_pred = global_model.predict(X_test)
    evaluate(y_test, y_pred)
    
    return "Model evaluated successfully."

if __name__ == '__main__':
    app.run(debug=True,port = 5002)

    # Example to simulate training and aggregation process
    files = ['part_1_N-CMAPSS_DS01-005.h5', 'part_2_N-CMAPSS_DS01-005.h5','part_3_N-CMAPSS_DS01-005.h5']
    
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
                aggregate_client_params(client_params, mode='sequential')  # Change mode to 'aggregate' if needed

    
    # For final aggregation (only needed in aggregate mode)
    # final_aggregate()
    
    # Evaluate the global model
    evaluate_global_model(X_test, y_test)
