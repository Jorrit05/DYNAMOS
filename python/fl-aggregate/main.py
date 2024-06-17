from flask import Flask, request, jsonify
import numpy as np
import requests

app = Flask(__name__)

MODEL_SERVICE_URL = 'http://localhost:5004'  # URL of the Model Service

# Function to fetch global model parameters
def fetch_global_model():
    global_model_response = requests.get(f"{MODEL_SERVICE_URL}/get_global_model")
    if global_model_response.status_code == 200:
        global_model = global_model_response.json()
        coef = np.array(global_model['coef'])
        intercept = global_model['intercept'][0]  # Extract the float value from the list
        return coef, intercept
    else:
        raise RuntimeError("Failed to fetch global model.")


def update_global_model(coef, intercept):
    data = {
        'coef': coef,  # Convert numpy array to list
        'intercept': [intercept]  # Ensure intercept is wrapped in a list
    }
    response = requests.post(f"{MODEL_SERVICE_URL}/set_global_model", json=data)
    if response.status_code == 200:
        print("Global model updated successfully.")
    else:
        print(f"Failed to update global model: {response.status_code}")

num_clients = 0
global_model = {
    'coef': np.zeros(1).tolist(),
    'intercept': [0.0]
}

@app.route('/aggregate', methods=['POST'])
def aggregate():
    global global_model, num_clients
    coef, intercept = fetch_global_model()  # Fetch current global model parameters
    data = request.get_json()
    client_params = data['client_params']
    
    client_coef = np.array(client_params['coef'])
    client_intercept = client_params['intercept'][0]

    if num_clients == 0:
        global_model['coef'] = client_coef.tolist()
        global_model['intercept'] = [client_intercept]
    else:
        global_model['coef'] = ((np.array(global_model['coef']) * num_clients + client_coef) / (num_clients + 1)).tolist()
        global_model['intercept'] = [(global_model['intercept'][0] * num_clients + client_intercept) / (num_clients + 1)]
    
    num_clients += 1
    
    # Update the model on the Model Service
    update_global_model(global_model['coef'], global_model['intercept'])
    
    return jsonify({"message": "Model aggregated successfully", "global_model": global_model})

if __name__ == '__main__':
    app.run(debug=True, port=5002)
