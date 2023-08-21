<!-- <template>
    <div class="request-approval">
        <h1 class="title">Data Request</h1>
        <p class="info">Please enter the required information below:</p>

        <el-form @submit.prevent="submitForm" class="approval-form">
            <input type="radio" id="surf" value="Surf" v-model="form.urlType" />
            <label for="one">Surf</label>

            <input type="radio" id="uva" value="UVA" v-model="form.urlType" />
            <label for="two">UVA</label>

            <br />
            <br />
            <br />

            <el-form-item class='form-item' label="SQL Query:">
                <el-input v-model="form.sqlQuery" placeholder="SELECT * FROM Personen"></el-input>
            </el-form-item>
            <el-form-item class='form-item' label="Job ID:">
                <el-input v-model="form.jobId" placeholder="jorrit-stutterheim-<jobid>"></el-input>
            </el-form-item>

            <input type="checkbox" id="graph" v-model="form.graph" />
            <label for="graph">Graph</label>

            Submit button
            <el-form-item>
                <el-button type="primary" native-type="submit">Submit</el-button>
            </el-form-item>
        </el-form>
    </div>
</template> -->

<template>
    <div class="request-approval">
        <h1 class="title">Data Request</h1>
        <p class="info">Please enter the required information below:</p>

        <el-form @submit.prevent="submitForm" class="approval-form">
            <input type="radio" id="surf" value="Surf" v-model="form.urlType" />
            <label for="one">Surf</label>

            <input type="radio" id="uva" value="UVA" v-model="form.urlType" />
            <label for="two">UVA</label>

            <br />
            <br />
            <br />

            <el-form-item class='form-item' label="SQL Query:">
                <el-input v-model="form.sqlQuery" placeholder="SELECT * FROM Personen"></el-input>
            </el-form-item>
            <el-form-item class='form-item' label="Job ID:">
                <el-input v-model="form.jobId" placeholder="jorrit-stutterheim-<jobid>"></el-input>
            </el-form-item>

            <input type="checkbox" id="graph" v-model="form.graph" />
            <label for="graph">Graph</label>

            <!-- Submit button -->
            <el-form-item>
                <el-button type="primary" native-type="submit">Submit</el-button>
            </el-form-item>
        </el-form>

        <!-- Loading Indicator -->
        <div v-if="isLoading" class="loading-section">
            Loading...
        </div>

        <!-- Display Error -->
        <div v-if="isError" class="error-message">
            An error occurred while fetching data.
        </div>

        <!-- Display Response Data -->
        <!-- <div v-if="responseData" class="response-section">
            <h2>Received Data:</h2> -->
            <!-- Sample display; customize based on your response structure -->
            <!-- <pre>{{ JSON.stringify(responseData, null, 2) }}</pre>
        </div> -->
        <div v-if="responseData" class="response-section">
    <h2>Received Data:</h2>
    <ul class="response-list">
        <li v-for="(item, index) in responseData" :key="index">
            {{ JSON.stringify(item) }}
        </li>
    </ul>
</div>
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
            sqlQuery: "",
            jobId: "",
            urlType: "Surf" || "UVA",
            graph: false
        });

        const account = msalInstance.getActiveAccount();
        const uniqueId = account?.localAccountId;
        const name = account?.username;

        async function submitForm() {
            isLoading.value = true;
            isError.value = false;


            // Construct the request body
            const body = {
                type: "sqlDataRequest",
                query: form.value.sqlQuery,
                graph: form.value.graph,
                algorithm: "average",
                // algorithmColumns: {
                //     Geslacht: "Aanst_22, Gebdat"
                // },
                user: {
                    id: uniqueId,
                    userName: name,
                },
                requestMetadata: {
                    jobId: form.value.jobId
                }
            };

            console.log(body)
            try {
                // Send the API request
                const response = await axios({
                    method: 'POST',
                    url: form.value.urlType === 'Surf' ? 'http://surf.surf.svc.cluster.local:80/agent/v1/sqlDataRequest/surf' : 'http://uva.uva.svc.cluster.local:80/agent/v1/sqlDataRequest/uva',
                    // data: JSON.stringify(body),
                    data: body,
                    headers: {
                        'Content-Type': 'application/json',
                        Accept: 'application/json',
                        Authorization: 'bearer 1234' //TODO use token
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

.form-item > label {
        width: 30%;
        /* Waarom werkt dit niet :( */
}
.response-list {
    list-style-type: none; /* Remove bullet points */
    padding: 0;
    width: 100%; /* Take the full width */
}

.response-list li {
    border: 1px solid #e5e5e5; /* Add a border to each item for better separation */
    padding: 10px;
    margin: 5px 0; /* Some margin between items */
    word-wrap: break-word; /* Break long words */
    max-width: 100%;
    overflow-x: auto; /* Add horizontal scroll for very long content */
}
</style>
