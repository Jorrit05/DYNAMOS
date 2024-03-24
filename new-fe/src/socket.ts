import { reactive } from "vue";
import { io } from "socket.io-client";

export const state = reactive({
  connected: false,
  availableProviders: []
});

// "undefined" means the URL will be computed from the `window.location` object
// const URL = process.env.NODE_ENV === "production" ? undefined : "http://api-gateway.api-gateway.svc.cluster.local:80/api/v1/ws";
const URL = "http://api-gateway.api-gateway.svc.cluster.local:80";

export const socket = io(URL, {
  reconnectionDelay: 1000,
  reconnection: true,
  transports: ['websocket'],
  agent: false,
  upgrade: false,
  rejectUnauthorized: false
});

socket.on("connect", () => {
  state.connected = true;
});

socket.on("disconnect", () => {
  state.connected = false;
});

socket.on("availableProviders", (...args: any) => {
  state.availableProviders.push(args);
});
