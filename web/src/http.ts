// src/httpService.js
import axios from 'axios';

const http = axios.create({
    baseURL: 'http://api-gateway.api-gateway.svc.cluster.local:80', // Replace with your backend URL
    timeout: 25000, // Adjust timeout as needed
    headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
        Authorization: 'bearer 1234' //TODO use token
    }
});

const requestApproval = async (requestType: string,
    user: Map<String, String>,
    selectedProviders: string[],
    sql: string, algo: string,
    graph: boolean,
    aggregate: boolean) => {
    var payload = {
        type: requestType,
        user: user,
        dataProviders: selectedProviders,
        data_request: {
            type: requestType,
            query: sql,
            algorithm: algo,
            options: {
                graph: graph,
                aggregate: aggregate
            }
        }
    }

    try {
        const response = await http.post("/api/v1/requestApproval", payload)

        return response
    } catch (error) {
        console.error('Error:', error);
    }

}

export { requestApproval };