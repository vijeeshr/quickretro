import { createApp } from 'vue'
// import './style.css'
import './index.css'
import App from './App.vue'
import router from './router'
import ToastPlugin from 'vue-toast-notification'

createApp(App).use(router).use(ToastPlugin).mount('#app')
