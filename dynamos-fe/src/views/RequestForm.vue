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
            <el-form-item class='form-item' label="Algorithm:">
                <el-input v-model="form.algorithm" placeholder="average"></el-input>
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
    <div>
    <!-- Display table if responseData is an array and has data -->
    <table class="styled-table" v-if="Array.isArray(responseData) && responseData.length">
      <thead>
        <tr>
          <th v-for="(header, index) in responseData[0]" :key="index">{{ header }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, rowIndex) in responseData.slice(1)" :key="rowIndex">
          <td v-for="(cell, cellIndex) in row" :key="cellIndex">{{ cell }}</td>
        </tr>
      </tbody>
    </table>

    <!-- Display JSON data if responseData is an object -->
    <div v-else-if="typeof responseData === 'object'">
      <div v-for="(value, key) in responseData" :key="key">
        {{ key }}: {{ value }}
      </div>
    </div>

    <!-- Handle other cases (like empty data or other formats) -->
    <div v-else>
      No data available.
    </div>
  </div>
</div>
    </div>
</template>


<script lang="ts">
import { ref, computed } from 'vue';
import axios from 'axios'; // import axios
import { msalInstance } from "../authConfig";
const responseData: any = ref();
const isLoading = ref(false);
const isError = ref(false);

export default {

    setup() {
        const form = ref({
            sqlQuery: "",
            jobId: "",
            urlType: "Surf" || "UVA",
            graph: false,
            algorithm: "average" || ""
        });

        // const account = msalInstance.getActiveAccount();
        // const uniqueId = account?.localAccountId;
        // const name = account?.username;

        const account = "jorrit.stutterheim@cloudnation.nl";
        const uniqueId = "124314ou3uo424";
        const name = "jorrit.stutterheim@cloudnation.nl";

        async function submitForm() {
            isLoading.value = true;
            isError.value = false;


            // Construct the request body
            const body = {
                type: "sqlDataRequest",
                query: form.value.sqlQuery,
                graph: form.value.graph,
                algorithm: form.value.algorithm,
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
    min-width: 600px;
    max-width: 85vw;
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

.styled-table {
    width: 80%;
    border-collapse: collapse;
    margin: 25px 30px;
    font-size: 0.9em;
    font-family: sans-serif;
    min-width: 400px;
    box-shadow: 0 0 20px rgba(0, 0, 0, 0.15);
}
.styled-table thead tr {
    background-color: #009879;
    color: #ffffff;
    text-align: left;
}

.styled-table th,
.styled-table td {
    padding: 12px 15px;
}

.styled-table tbody tr {
    border-bottom: 1px solid #dddddd;
}

.styled-table tbody tr:nth-of-type(even) {
    background-color: #f3f3f3;
}

.styled-table tbody tr:last-of-type {
    border-bottom: 2px solid #009879;
}

.styled-table tbody tr.active-row {
    font-weight: bold;
    color: #009879;
}
</style>
