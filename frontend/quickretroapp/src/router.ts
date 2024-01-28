import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from './components/Dashboard.vue'
import Join from './components/Join.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/board/:board',
      component: Dashboard,
    },
    {
      path: '/board/:board/join',
      component: Join,
    },
    /*
    {
      path: '/contact',
    //   component: () => import('@/views/Contact.vue'),
    },
    */
  ],
})
