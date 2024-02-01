<script setup>

import { ref } from 'vue';

const algo = ref(null);
const graph = ref(false);
const aggregate = ref(false);
const sql = ref(null);
const selectedProviders = ref();

// TODO: replace with actual data
const availableProviders = ref([
    {label: 'UvA', value: 'uva'},
    {label: 'VU', value: 'vu'}
])

const submitting = ref(false);

const submit = () => {
    submitting.value = true;
    setTimeout(() => {
        submitting.value = false;
    }, 2000);
};
</script>

<template>
    <div>
        <Card class="m-4">
            <template #title>DYNAMOS Request Form</template>
            <template #content>
            <div class="field">
                <label
                    class="col-12"
                    for="availProviders"
                >
                    Available Providers
                </label>
                <MultiSelect 
                    v-model="selectedProviders"
                    :options="availableProviders"
                    optionLabel="label"
                    placeholder="Select available providers"
                    id="availProviders"
                    aria-describedby="availProviders-help"
                    class="col-12 mb-2"
                />
                <small id="availProviders-help">
                    The currently available providers
                </small>

                <InlineMessage 
                    v-if="selectedProviders == null"
                    class="col-12 mt-2"
                    severity="warn">
                    At least one provider is required
                </InlineMessage>
              
            </div>
            <div class="field">
                <label 
                    for="sql"
                    class="col-12"
                >
                    SQL Query
                </label>
                <Editor
                    class="col-12" 
                    id="sql"
                    v-model="sql"
                    editorStyle="height: 100px"
                    aria-describedby="sql-help"
                >
                    <template v-slot:toolbar> - </template>
                </Editor>
                <small id="sql-help">
                    The SQL query to execute
                </small>
                <InlineMessage 
                    v-if="sql == null || sql == ''"
                    class="col-12 mt-2"
                    severity="warn">
                    SQL query is required
                </InlineMessage>
            </div>

            <div class="field">
                <label
                    class="col-12" 
                    for="algo">Algorithm</label>
                <InputText
                    class="col-12 mb-1"
                    id="algo"
                    v-model="algo"
                    aria-describedby="algo-help"
                />
                <small id="algo-help">
                    Enter the algorithm to be applied.
                </small>
            </div>
            
            <div class="field">
                <div class="col-12">
                    <label
                        for="graph"
                    >
                        Graph
                    </label>
                    <InputSwitch
                        id="graph"
                        v-model="graph"
                        aria-describedby="graph-help"
                    />
                </div>
                <small id="graph-help">
                    Whether a graph is generated.
                </small>
            </div>

            <div class="field">
                <div class="col-12">
                    <label
                        class=""
                        for="aggregate"
                    >
                        Aggregate
                    </label>
                    <InputSwitch
                        class="ml-2 "
                        id="aggregate"
                        v-model="aggregate"
                        aria-describedby="aggregate-help"
                    />
                </div>  
                <small id="aggregate-help">
                    Will the data be aggregated?
                </small>
            </div>


            <Button 
                class="mt-5"
                type="button"
                label="Submit"
                icon="pi pi-search"
                :loading="submitting" 
                @click="submit"
            />
        </template>
    </Card>  
    </div>
</template>