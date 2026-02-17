import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from './components/Dashboard.vue'
import Join from './components/Join.vue'
import CreateBoard from './components/CreateBoard.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'start',
      component: Join,
    },
    {
      path: '/create',
      name: 'create',
      component: CreateBoard,
      beforeEnter: to => {
        if (
          !localStorage.getItem('user') ||
          !localStorage.getItem('xid') ||
          !localStorage.getItem('nickname')
        ) {
          return { path: '/', query: to.query }
        }
      },
    },
    {
      path: '/board/:board',
      name: 'dashboard',
      component: Dashboard,
      beforeEnter: to => {
        if (
          !localStorage.getItem('user') ||
          !localStorage.getItem('xid') ||
          !localStorage.getItem('nickname')
        ) {
          return `/board/${to.params.board}/join`
        }
      },
    },
    {
      path: '/board/:board/join',
      name: 'join',
      component: Join,
    },
  ],
})
