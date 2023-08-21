<template>
  <div class="home">
    <span v-if="!isAuthenticated">Please sign-in to see your profile information.</span>
    <span v-else>Logged in.</span>
    <!-- <div v-else v-on:change="goToProfile()"></div> -->
    <!-- <router-link v-else to="/requestApproval">Go to Form</router-link> -->
    <RequestApproval></RequestApproval>`
    <RequestForm></RequestForm>
    <!-- <el-button v-else type="primary" v-on:click="goToProfile">Request Profile Information</el-button> -->
  </div>
</template>

<script setup lang="ts">
// import { useRouter } from "vue-router";
import { useIsAuthenticated } from "../composition-api/useIsAuthenticated";
// import { reactive, onMounted, watch } from 'vue'
import RequestForm from "./RequestForm.vue";
import RequestApproval from "./RequestApproval.vue";

const isAuthenticated = useIsAuthenticated();

// const router = useRouter();

// function goToProfile() {
//   router.push("/requestApproval");
// }
</script>
