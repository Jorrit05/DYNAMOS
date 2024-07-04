from flask import Flask, request, jsonify
from sklearn.linear_model import LinearRegression
import numpy as np
import requests
import logging

app = Flask(__name__)
MODEL_SERVICE_URL = 'http://localhost:5004'  # URL of the Model Service

# Configure logging
logging.basicConfig(level=logging.DEBUG)

# Function to fetch global model parameters
def fetch_global_model():
    try:
        global_model_response = requests.get(f"{MODEL_SERVICE_URL}/get_global_model")
        global_model_response.raise_for_status()
        global_model = global_model_response.json()
        coef = np.array(global_model['coef'])
        intercept = global_model['intercept'][0]  # Extract the float value from the list
        return coef, intercept
    except requests.RequestException as e:
        logging.error(f"Error fetching global model: {e}")
        raise RuntimeError("Failed to fetch global model.")

@app.route('/train', methods=['POST'])
def train():
    try:
        logging.debug('Received train request')
        
        # Fetch global model parameters
        coef, intercept = fetch_global_model()
        logging.debug(f'Global model coef: {coef}, intercept: {intercept}')
        
        # Get training data from request
        data = request.get_json()
        if not data or 'X_train' not in data or 'y_train' not in data:
            logging.error("Invalid or missing data in request")
            return jsonify({'error': 'Invalid or missing data in request'}), 400
        
        X_train = np.array(data['X_train'])
        y_train = np.array(data['y_train'])
        logging.debug(f'X_train: {X_train.shape}, y_train: {y_train.shape}')
        
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
        logging.error(f"Error in train endpoint: {e}")
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(debug=True, port=5000)
