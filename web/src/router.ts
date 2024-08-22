import { createRouter, createWebHashHistory } from "vue-router";
import FormView from './views/FormView.vue';
const routes = [
    { path: '/', component: FormView },
    // { path: '/about', component: '<div>About</div>' },
  ]
  
export const router = createRouter({
    history: createWebHashHistory(),
    routes: routes
});