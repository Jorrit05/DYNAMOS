<template>
    <el-dropdown>
        <el-button type="primary" v-on:click="loginRedirect">Sign In</el-button>
        <!-- <template #dropdown>
                <el-dropdown-item v-on:click="loginRedirect">Login Redirect</el-dropdown-item>
        </template> -->
    </el-dropdown>
</template>

<script setup lang="ts">
import { useMsal } from '../composition-api/useMsal';
import { loginRequest } from "../authConfig";

const { instance } = useMsal();

const loginRedirect = () => {
  instance.loginRedirect(loginRequest);
}
</script>
