import { createApp } from 'vue'
// import './style.css'
import './index.css'
import App from './App.vue'
import router from './router'
import "preline/preline"

createApp(App).use(router).mount('#app')
