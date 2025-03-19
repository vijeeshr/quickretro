import { createApp } from 'vue'
// import './style.css'
import './index.css'
import App from './App.vue'
import router from './router'
import ToastPlugin from 'vue-toast-notification'
import i18n from './i18n'

createApp(App).use(i18n).use(router).use(ToastPlugin).mount('#app')
