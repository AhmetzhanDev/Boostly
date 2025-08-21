import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './style.css'
import { initAllAnimations } from './animations.js'

const app = createApp(App)
app.use(router)
app.mount('#app')

// Инициализация анимаций после монтирования приложения
document.addEventListener('DOMContentLoaded', () => {
  initAllAnimations();
}); 