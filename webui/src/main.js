import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import axios from 'axios'

import './assets/dashboard.css'
import './assets/main.css'

const app = createApp(App)

// MODIFICA QUI: Lascia una stringa vuota o "/". 
// Le chiamate partiranno relative all'origine corrente (es. http://localhost:8080/session)
axios.defaults.baseURL = "" 

app.config.globalProperties.$axios = axios
app.use(router)
app.mount('#app')