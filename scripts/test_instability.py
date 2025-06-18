import requests
import time
import copy

# This script is designed to test different archetypes in the system to analyze potential issues and instability in DYNAMOS.
# Executing this script will print the results in the console, including execution times and response sizes.

# =================================================================== Constants ===================================================================
# DYNAMOS requests
HEADERS = {
    "Content-Type": "application/json",
    # Access token required for data requests in DYNAMOS
    "Authorization": "bearer 1234"
}
DATA_REQ_URL = "http://api-gateway.api-gateway.svc.cluster.local:80/api/v1/requestApproval"
DATA_REQ_HEADERS = {"Content-Type": "application/json"}
REQUEST_BODY_DATA_REQ = {
    "type": "sqlDataRequest",
    "user": {"id": "12324", "userName": "jorrit.stutterheim@cloudnation.nl"},
    "dataProviders": ["UVA"],
    "data_request": {
        "type": "sqlDataRequest",
        "query": "SELECT DISTINCT p.Unieknr, p.Geslacht, p.Gebdat, s.Aanst_22, s.Functcat, s.Salschal as Salary FROM Personen p JOIN Aanstellingen s ON p.Unieknr = s.Unieknr LIMIT 30000",
        "algorithm": "",
        "options": {"graph": False, "aggregate": False},
        "requestMetadata": {}
    },
}

# Update archetypes
ARCHETYPES = ["ComputeToData", "DataThroughTTP"]
UPDATE_ARCH_URL = "http://orchestrator.orchestrator.svc.cluster.local:80/api/v1/archetypes/agreements"
INITIAL_REQUEST_BODY_ARCH = {
    "name": "computeToData",
    "computeProvider": "dataProvider",
    "resultRecipient": "requestor",
}
WEIGHTS = {
    "ComputeToData": 100,
    "DataThroughTTP": 300
}
HEADERS_UPDATE_ARCH = { "Content-Type": "application/json" }


# =================================================================== Functions ===================================================================
def switch_archetype(archetype):
    """
    Switches the system to use a specific archetype by modifying the weight
    in the request and sending an update to the appropriate endpoint.
    """
    print(f"Switching archetypes to {archetype}...")
    
    # Create body for archetype switch request
    request_body_arch = INITIAL_REQUEST_BODY_ARCH.copy()
    request_body_arch["weight"] = WEIGHTS[archetype]  # Set new weight
    
    # Send PUT request to update archetype
    response = requests.put(UPDATE_ARCH_URL, json=request_body_arch, headers=HEADERS_UPDATE_ARCH)
    print(f"Archetype switch response: {response.status_code}, time: {response.elapsed.total_seconds()}s")
    
    # Wait briefly to ensure the change propagates before next experiment
    time.sleep(5)

def run_requests(request_body, label, count=10):
    """
    Runs 'count' iterations of request approval + data request with a specified approval body.
    Measures and returns execution times of data requests.
    """
    print(f"\nRunning {label} ({count} iterations)...")

    for i in range(count):
        print(f"Iteration {i+1}/{count}...")

        # Request data (approval and data request in one with the API-gateway)
        response_approval = requests.post(DATA_REQ_URL, json=request_body, headers=DATA_REQ_HEADERS)
        handle_request_response(response_approval)

        # Short wait to avoid overloading or tight loops, similar to full experiments execution for thesis
        time.sleep(7)

def handle_request_response(response):
    """
    Handles response from approval and data request, logs status, time, and content size.
    Returns execution time.
    """
    # Get the execution time and status code
    exec_time = response.elapsed.total_seconds()
    status = response.status_code

    # Get the size of the returned data (the data is inside the responses field of the JSON response, such as "responses": [""])
    try:
        json_data = response.json()
        responses = json_data.get("responses", [])
        response_sizes = [len(r.encode("utf-8")) for r in responses if isinstance(r, str)]
        total_size = sum(response_sizes)
    except Exception as e:
        total_size = 0
        print(f"[!] Failed to parse response JSON or measure size: {e}")

    print(f"Status: {status}, Time: {exec_time:.3f}s, Returned data size: {total_size} bytes")

def run_test():
    """
    Main test driver that runs two complete setups:
    Each archetype is tested with both standard and modified approval request bodies.
    """
    # Iterate through the first two archetypes (e.g., ComputeToData, DataThroughTTP)
    # Change below to a specific archetype to only test that one, such as removing the for loop and adding:
    # archetype = "DataThroughTTP"
    # idx = 1
    for idx, archetype in enumerate(ARCHETYPES[:2]):
        print(f"\n========== SETUP {idx+1} | Archetype: {archetype} ==========")
        
        # Switch to the selected archetype
        switch_archetype(archetype)

        # ---------- Test with standard request (one data provider) ----------
        result_key_1 = f"Arch{idx+1}_OneDataProvider"
        run_requests(REQUEST_BODY_DATA_REQ, label=result_key_1)

        # ---------- Test with modified approval request (multiple data providers) ----------
        # Deep copy to avoid modifying the original request body of nested values just to be safe
        modified_approval = copy.deepcopy(REQUEST_BODY_DATA_REQ)
        modified_approval["dataProviders"] = ["UVA", "VU"]  # Change providers
        result_key_2 = f"Arch{idx+1}_MultipleDataProviders"
        run_requests(modified_approval, label=result_key_2)

# Entry point
if __name__ == "__main__":
    run_test()