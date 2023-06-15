import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import Home from '../views/Home.vue';
import Profile from '../views/Profile.vue';
import ProfileNoGuard from '../views/ProfileNoGuard.vue';
import Failed from "../views/Failed.vue";
import RequestApproval from "../views/RequestApproval.vue";
import { registerGuard } from "./Guard";

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    name: 'Home',
    component: Home,
  },
  {
    path: '/profile',
    name: 'Profile',
    component: Profile,
    meta: {
        requiresAuth: true
    }
  },
  {
    path: '/profileNoGuard',
    name: 'ProfileNoGuard',
    component: ProfileNoGuard
  },
  {
    path: '/failed',
    name: 'Failed',
    component: Failed
  },
  {
    path: '/requestApproval',
    name: 'RequestApproval',
    component: RequestApproval
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

registerGuard(router);

export default router;
