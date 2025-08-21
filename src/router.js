import { createRouter, createWebHistory } from 'vue-router'
import Home from './Home.vue'
import Signup from './Signup.vue'
import Dashboard from './Dashboard.vue'
import NoteView from './NoteView.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },

  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: { requiresAuth: true }
  },
  {
    path: '/signup',
    name: 'Signup',
    component: Signup
  },
  {
    path: '/note/:id',
    name: 'NoteView',
    component: NoteView,
    meta: { requiresAuth: true }
  }
  // ,
  // {
  //   path: '/signup',
  //   name: 'Signup',
  //   component: Signup
  // },

]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Global auth guard
router.beforeEach((to, from, next) => {
  if (to.meta && to.meta.requiresAuth) {
    const token = localStorage.getItem('token')
    if (!token) return next({ path: '/signup' })
  }
  next()
})

export default router 