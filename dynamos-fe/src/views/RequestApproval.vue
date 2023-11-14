<template>
    <div class="request-approval">
        <h1 class="title">Request Approval Form</h1>
        <p class="info">Please enter the required information below:</p>

        <el-form @submit.prevent="submitForm" class="approval-form">
            <el-form-item label="Enter Data Providers:">
                <el-input v-model="form.dataProviders" placeholder="Enter data providers"></el-input>
            </el-form-item>

            <!-- Submit button -->
            <el-form-item>
                <el-button type="primary" native-type="submit">Submit</el-button>
            </el-form-item>

            <!-- Display Response Data -->
            <div v-if="responseData" class="response-section">
                <h2>Response Data</h2>
                <div>
                    <strong>Job ID:</strong>
                    <p>{{ responseData.job_id }}</p>
                </div>
                <div>
                    <strong>Authorized Providers:</strong>
                    <ul>
                        <li v-for="(value, key) in responseData.authorized_providers" :key="key">
                            {{ key }}: {{ value }}
                        </li>
                    </ul>
                </div>
            </div>
            <div v-if="isError" class="error-section">
                <h2>Error:</h2>
                <pre>{{ isError }}</pre>
            </div>

        </el-form>
    </div>
</template>

<script lang="ts">
import { ref, computed } from 'vue';
import { msalInstance } from "../authConfig";
import axios, { AxiosError } from 'axios';

interface ResponseData {
    authorized_providers: Record<string, string>;
    job_id: string;
}

const responseData = ref<ResponseData | null>(null);
const isLoading = ref(false);
const isError = ref<string | null>(null);
let errorKey = 0;
export default {
    setup() {
        const form = ref({
            dataProviders: ""
        });

        async function submitForm() {
            const dataProvidersArray = form.value.dataProviders.split(',').map(value => value.trim());

            isLoading.value = true;
            // const account = msalInstance.getActiveAccount();
            // const uniqueId = account?.localAccountId;
            // const name = account?.username;

            const account = "jorrit.stutterheim@cloudnation.nl";
            const uniqueId = "124314ou3uo424";
            const name = "jorrit.stutterheim@cloudnation.nl";
            // Construct the request body
            const body = {
                type: "sqlDataRequest",
                user: {
                    ID: uniqueId,
                    userName: name,
                },
                dataProviders: dataProvidersArray,
                syncServices: true,
            };
            console.log(body)

            try {
                // Send the API request
                const response = await axios({
                    method: 'POST',
                    url: 'http://orchestrator.orchestrator.svc.cluster.local:80/api/v1/requestapproval',
                    data: body,
                    headers: {
                        'Content-Type': 'application/json',
                        Accept: 'application/json',
                    },
                });
                console.log(response.data);
                // Extract only authorized_providers and job_id from the response
                const { authorized_providers, job_id } = response.data;
                responseData.value = { authorized_providers, job_id };
            } catch (error: unknown) {

                if (axios.isAxiosError(error)) {
                    // Now TypeScript knows this is an AxiosError

                    const axiosError = error as AxiosError;
                    if (axiosError.response) {
                        console.error(`Server Response Error: ${axiosError.response.status} - ${axiosError.response.data}`);
                        // let str = axiosError.response.data;
                        isError.value = axiosError.message || 'An unexpected error occurred';
                        console.log("Error message set:", isError.value);

                    } else {
                        isError.value = axiosError.message || 'An unexpected error occurred';
                    }
                } else {
                    console.log("4")
                    // Handle non-Axios errors
                    if (error instanceof Error) {
                        isError.value = error.message || 'An unexpected error occurred';
                    } else {
                        isError.value = 'An unknown error occurred';
                    }
                }
            }
            finally {
                isLoading.value = false;
            }
        }

        return {
            form,
            submitForm,
            responseData,
            isLoading,
            isError
        };
    },
};
</script>


<style scoped>
.request-approval {
    max-width: 600px;
    margin: 0 auto;
    padding: 20px;
    box-shadow: 0px 0px 10px 0px rgba(0, 0, 0, 0.1);
}

.title {
    text-align: center;
    margin-bottom: 20px;
}

.info {
    text-align: center;
    margin-bottom: 30px;
    color: #888;
}

.approval-form {
    margin-top: 20px;
}

.loading-section {
    font-weight: bold;
    color: #007BFF;
    /* You can adjust the color */
}

.error-message {
    color: red;
    border: 1px solid red;
    padding: 10px;
    margin-top: 10px;
}

.response-section {
    margin-top: 20px;
    background-color: #f8f9fa;
    padding: 15px;
    border: 1px solid #dee2e6;
}
</style>
