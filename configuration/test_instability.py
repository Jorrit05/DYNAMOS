import requests
import time
import constants
import argparse
import statistics

# This script is designed to test different archetypes in the system to analyze potential issues and instability in DYNAMOS.

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
    request_body_arch = constants.INITIAL_REQUEST_BODY_ARCH
    request_body_arch["weight"] = constants.WEIGHTS[archetype]  # Set new weight
    
    # Send PUT request to update archetype
    response = requests.put(constants.UPDATE_ARCH_URL, json=request_body_arch, headers=constants.HEADERS_UPDATE_ARCH)
    print(f"Archetype switch response: {response.status_code}, time: {response.elapsed.total_seconds()}s")
    
    # Wait briefly to ensure the change propagates before next experiment
    time.sleep(5)

def run_requests(data_steward, request_body, label, count=10):
    """
    Runs 'count' iterations of request approval + data request with a specified approval body.
    Measures and returns execution times of data requests.
    """
    print(f"\nRunning {label} ({count} iterations)...")
    exec_times = []

    for i in range(count):
        print(f"Iteration {i+1}/{count}...")

        # Request data (approval and data request in one with the API-gateway)
        response_approval = requests.post(constants.DATA_REQ_URL, json=request_body, headers=constants.DATA_REQ_HEADERS)
        job_id = handle_request_approval_response(response_approval)

        # ============ STEP 2: Data Request ============
        request_body = constants.INITIAL_REQUEST_BODY
        request_body["requestMetadata"] = {"jobId": f"{job_id}"}  # Embed job ID into metadata

        # Set correct host header based on steward (important for routing in Kubernetes)
        headers = constants.HEADERS.copy()
        headers["Host"] = f"{data_steward}.{data_steward}.svc.cluster.local"

        # Send data request and record execution time
        response_data = requests.post(data_request_url, json=request_body, headers=headers)
        exec_time = handle_data_request_response(response_data)
        exec_times.append(exec_time)

        # Short wait to avoid overloading or tight loops, similar to full experiments execution for thesis
        time.sleep(7)

    return exec_times

def handle_request_approval_response(response):
    """
    Handles response from approval request, logs status and time,
    and extracts the job ID needed for the data request.
    """
    print(f"Approval: Status {response.status_code}, Time {response.elapsed.total_seconds()}s")
    return response.json()["jobId"]

def handle_data_request_response(response):
    """
    Handles response from data request, logs status, time, and content size.
    Returns execution time.
    """
    exec_time = response.elapsed.total_seconds()
    print(f"Data: Status {response.status_code}, Time {exec_time}s, Size {len(response.content)} bytes")
    return exec_time

def run_test():
    """
    Main test driver that runs two complete setups:
    Each archetype is tested with both standard and modified approval request bodies.
    """
    all_results = {}  # Dictionary to store all results

    # Iterate through the first two archetypes (e.g., ComputeToData, DataThroughTTP)
    for idx, archetype in enumerate(constants.ARCHETYPES[:2]):
        print(f"\n========== SETUP {idx+1} | Archetype: {archetype} ==========")
        
        # Switch to the selected archetype
        switch_archetype(archetype)

        # ---------- Test with standard request (one data provider) ----------
        result_key_1 = f"Arch{idx+1}_OneDataProvider"
        all_results[result_key_1] = run_requests(constants.REQUEST_BODY_DATA_REQ, label=result_key_1)

        # ---------- Test with modified approval request (multiple data providers) ----------
        modified_approval = constants.REQUEST_BODY_DATA_REQ.copy()
        modified_approval["dataProviders"] = ["UVA", "VU"]  # Change providers
        result_key_2 = f"Arch{idx+1}_MultipleDataProviders"
        all_results[result_key_2] = run_requests(modified_approval, label=result_key_2)

# Entry point
if __name__ == "__main__":
    run_test()