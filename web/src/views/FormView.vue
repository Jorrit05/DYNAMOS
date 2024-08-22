<script setup>
import VueJsonPretty from 'vue-json-pretty';
import { ref, computed } from 'vue';
import { requestApproval } from '../http';
import 'vue-json-pretty/lib/styles.css';
import { useToast } from "primevue/usetoast";

const algo = ref("average");
const graph = ref(false);
const aggregate = ref(false);
const sql = ref("SELECT p.Geslacht, s.Salschal FROM Personen p JOIN Aanstellingen s ON p.Unieknr = s.Unieknr LIMIT 100");
const selectedProviders = ref();
// TODO: Change from hard coded
const requestType = "sqlDataRequest"
// TODO: Make an actual user system
const user = {
    id: "12324",
    userName: "jorrit.stutterheim@cloudnation.nl"
    // Actual loginToken features....
}

const responseData = ref(null);
const toast = useToast();
// TODO: replace with actual data
const availableProviders = ref([
    { label: 'UvA', value: 'UVA' },
    { label: 'VU', value: 'VU' }
])

const submitting = ref(false);


const getSelectedProviderValues = () => {
    var selectedValues = []
    selectedProviders.value.forEach(element => {
        selectedValues.push(element.value)
    });

    return selectedValues
}

const processedData = computed(() => {
    // Process your response data here if needed
    return responseData.value; // Or perform some computation on responseData
});

const submit = () => {
    if (selectedProviders.value == null || sql.value == null) {
        toast.add({ severity: 'error', summary: 'Form submission invalid', life: 3000 })

    } else {
        submitting.value = true;
        setTimeout(() => {
            submitting.value = false;
            requestApproval(requestType,
                user,
                getSelectedProviderValues(),
                sql.value,
                algo.value,
                graph.value,
                aggregate.value)
                .then(response => {
                    // Handle successful response
                    responseData.value = response.data
                    toast.add({ severity: 'success', summary: 'Request succesful!', life: 3000 })
                })
                .catch(error => {
                    // Handle error
                    toast.add({ severity: 'error', summary: 'Request unsuccesful', life: 3000 })

                });
        }, 5000);
    }
};
</script>

<template>
    <div>
        <Toast />
        <Card class="m-4">
            <template #title>DYNAMOS Request Form</template>
            <template #content>
                <div class="field">
                    <label class="col-12" for="availProviders">
                        Available Providers
                    </label>
                    <MultiSelect v-model="selectedProviders" :options="availableProviders" optionLabel="label"
                        placeholder="Select available providers" id="availProviders"
                        aria-describedby="availProviders-help" class="col-12 mb-2" />
                    <small id="availProviders-help">
                        The currently available providers
                    </small>

                    <InlineMessage v-if="selectedProviders == null" class="col-12 mt-2" severity="warn">
                        At least one provider is required
                    </InlineMessage>

                </div>
                <div class="field">
                    <label for="sql" class="col-12">
                        SQL Query
                    </label>
                    <Editor class="col-12" id="sql" v-model="sql" editorStyle="height: 100px"
                        aria-describedby="sql-help">
                        <template v-slot:toolbar> - </template>
                    </Editor>
                    <small id="sql-help">
                        The SQL query to execute
                    </small>
                    <InlineMessage v-if="sql == null || sql == ''" class="col-12 mt-2" severity="warn">
                        SQL query is required
                    </InlineMessage>
                </div>

                <div class="field">
                    <label class="col-12" for="algo">Algorithm</label>
                    <InputText class="col-12 mb-1" id="algo" v-model="algo" aria-describedby="algo-help" />
                    <small id="algo-help">
                        Enter the algorithm to be applied.
                    </small>
                </div>

                <div class="field">
                    <div class="col-12">
                        <label for="graph">
                            Graph
                        </label>
                        <InputSwitch id="graph" v-model="graph" aria-describedby="graph-help" />
                    </div>
                    <small id="graph-help">
                        Whether a graph is generated.
                    </small>
                </div>

                <div class="field">
                    <div class="col-12">
                        <label class="" for="aggregate">
                            Aggregate
                        </label>
                        <InputSwitch class="ml-2 " id="aggregate" v-model="aggregate"
                            aria-describedby="aggregate-help" />
                    </div>
                    <small id="aggregate-help">
                        Will the data be aggregated?
                    </small>
                </div>


                <Button class="mt-5" type="button" label="Submit" icon="pi pi-search" :loading="submitting"
                    @click="submit" />
            </template>
        </Card>

        <div v-if="(responseData != null)">

            <Card class="m-4">
                <template #title>Response</template>
                <template #content>
                    <vue-json-pretty :data="processedData" />
                </template>
            </Card>
        </div>
    </div>
</template>