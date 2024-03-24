import { createApp } from 'vue'
import App from './App.vue'
import PrimeVue from 'primevue/config';
import Button from 'primevue/button';
import Toast from 'primevue/toast';
import ToastService from 'primevue/toastservice';
import Menubar from 'primevue/menubar';
import MultiSelect from 'primevue/multiselect';
import Editor from 'primevue/editor';
import InputSwitch from 'primevue/inputswitch';
import InputText from 'primevue/inputtext';
import Card from 'primevue/card';
import InlineMessage from 'primevue/inlinemessage';

import "primeflex/primeflex.css";
import "primevue/resources/themes/bootstrap4-light-blue/theme.css"
// import "primevue/resources/primevue.min.css"; /* Deprecated */
import "primeicons/primeicons.css";

import { router } from './router'


const app = createApp(App);
app.use(PrimeVue);
app.use(ToastService);

app.component('Button', Button);
app.component('Toast', Toast);
app.component('Menubar', Menubar)
app.component('MultiSelect', MultiSelect)
app.component('Editor', Editor)
app.component('InputSwitch', InputSwitch)
app.component('InputText', InputText)
app.component('Card', Card)
app.component('InlineMessage', InlineMessage)

app.use(router)

app.mount('#app')
