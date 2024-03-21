<script lang="ts" setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import ConnectionState from "./components/ConnectionState.vue";
import ConnectionManager from "./components/ConnectionManager.vue";
const router = useRouter();
const items = ref([
    {
        label: 'Home',
        icon: 'pi pi-home',
        command: () => {
            router.push('/');
        }
    },
    {
        label: 'Sign In',
        icon: 'pi pi-sign-in',
        command: () => {
            router.push('/aboutUs');
        }
    }
]);
</script>

<template>
  <div id="app">
    <header>
      <div>
        <Menubar :model="items"/>
        <ConnectionState/>
        <ConnectionManager/>
      </div>
    </header>
 
  <router-view/>
</div> 
</template>

<!-- <style scoped>
 -->
<!-- </style> -->
