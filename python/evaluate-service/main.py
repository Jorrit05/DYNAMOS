from flask import Flask, jsonify, request
from sklearn.metrics import mean_squared_error, r2_score
import requests
import numpy as np

app = Flask(__name__)

MODEL_SERVICE_URL = 'http://localhost:5004'  # URL of the Model Service

@app.route('/evaluate', methods=['POST'])
def evaluate():
    data = request.get_json()
    X_test = np.array(data['X_test'])
    y_test = np.array(data['y_test'])
    
    # Fetch global model from Model Service
    global_model_response = requests.get(f"{MODEL_SERVICE_URL}/get_global_model")
    if global_model_response.status_code == 200:
        global_model = global_model_response.json()
        coef = np.array(global_model['coef']).reshape(-1)  # Reshape to a 1D array if necessary
        intercept = global_model['intercept'][0]
        
        # Print or check the shape of intercept
        print(f"Shape of intercept: {np.array(intercept).shape}")
        print(f"Intercept value: {intercept}")
        
        # Use global_model to make predictions and evaluate
        y_pred = np.dot(X_test, coef) + intercept
        mse = mean_squared_error(y_test, y_pred)
        rmse = np.sqrt(mse)
        r2 = r2_score(y_test, y_pred)
        
        evaluation_results = {'rmse': rmse, 'r2': r2}
        return jsonify(evaluation_results)
    else:
        return "Failed to fetch global model.", 500

if __name__ == '__main__':
    app.run(debug=True, port=5003)
