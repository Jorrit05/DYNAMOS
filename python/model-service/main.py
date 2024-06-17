from flask import Flask, jsonify, request

app = Flask(__name__)

global_model = {
    'coef': [0.0],
    'intercept': [0.0]
}

@app.route('/get_global_model', methods=['GET'])
def get_global_model():
    return jsonify(global_model)

@app.route('/set_global_model', methods=['POST'])
def set_global_model():
    global global_model
    data = request.get_json()
    global_model = data
    return "Global model updated successfully."

if __name__ == '__main__':
    app.run(debug=True, port=5004)
