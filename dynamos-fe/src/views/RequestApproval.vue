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
            <h2>Received Data:</h2>
            <!-- Sample display; customize based on your response structure -->
            <pre>{{ JSON.stringify(responseData, null, 2) }}</pre>
        </div>
        </el-form>
    </div>
</template>

<script lang="ts">
import { ref, computed } from 'vue';
import axios from 'axios'; // import axios
import { msalInstance } from "../authConfig";


const responseData = ref(null);
const isLoading = ref(false);
const isError = ref(false);

export default {
    setup() {
        const form = ref({
            dataProviders: ""
        });

        async function submitForm() {
            const dataProvidersArray = form.value.dataProviders.split(',').map(value => value.trim());

            isLoading.value = true;
            isError.value = false;
            // const account = msalInstance.getActiveAccount();
            // const uniqueId = account?.localAccountId;
            // const name = account?.username;

            const account = "jorrit.stutterheim@cloudnation.nl";
            const uniqueId = "124314ou3uo424";
            const name = "Jorrit Stutterheim";
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
                    // data: JSON.stringify(body),
                    data: body,
                    headers: {
                        'Content-Type': 'application/json',
                        Accept: 'application/json',
                    },
                });
                // const response = await axios.post('http://localhost:8081/requestapproval', body);
                console.log(response.data);
                responseData.value = response.data;
            } catch (error) {
                console.error(error);
                isError.value = true;
            } finally {
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
    color: #007BFF;  /* You can adjust the color */
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
