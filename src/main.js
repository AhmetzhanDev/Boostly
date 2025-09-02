import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './style.css'
import { initAllAnimations } from './animations.js'

const app = createApp(App)
app.use(router)
app.mount('#app')

// Initialize animations after the app is mounted
document.addEventListener('DOMContentLoaded', () => {
  initAllAnimations();
}); 