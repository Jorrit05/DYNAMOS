from flask import Flask, request, jsonify
from sklearn.linear_model import LinearRegression
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

@app.route('/train', methods=['POST'])
def train():
    try:
        # Fetch global model parameters
        coef, intercept = fetch_global_model()
        
        # Get training data from request
        data = request.get_json()
        X_train = np.array(data['X_train'])
        y_train = np.array(data['y_train'])
        
        # Initialize local model with global model parameters
        local_model = LinearRegression()
        local_model.coef_ = coef
        local_model.intercept_ = intercept
        
        # Train local model
        local_model.fit(X_train, y_train)
        
        # Prepare response with local model parameters
        params = {
            'coef': local_model.coef_.tolist(),
            'intercept': [local_model.intercept_]  # Wrap intercept in a list for consistency
        }
        return jsonify(params)
    
    except Exception as e:
        return str(e), 500

if __name__ == '__main__':
    app.run(debug=True, port=5000)
