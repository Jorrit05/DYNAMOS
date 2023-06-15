<template>
  <div>
    <NavBar />
    <img class="logo" alt="Vue logo" src="./assets/simple_logo.png">
    <h1>Welcome to Dynamos</h1>
    <router-view />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import NavBar from "./components/NavBar.vue";

export default defineComponent({
  name: 'App',
  components: {
    NavBar
  }
});
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  text-align: center;
  color: #2c3e50;
}
.logo {
  width: 400px; /* Or whatever width you want */
  height: auto; /* Keep aspect ratio */
}
</style>
