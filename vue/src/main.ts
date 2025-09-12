import { createApp } from 'vue'
import { createPinia } from 'pinia'

// Import Element Plus style
import 'element-plus/dist/index.css'

// @ts-ignore
import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
