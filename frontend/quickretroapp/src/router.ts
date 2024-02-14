import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from './components/Dashboard.vue'
import Join from './components/Join.vue'
import CreateBoard from './components/CreateBoard.vue'
import { type IStaticMethods } from "preline/preline"

declare global {
  interface Window {
    HSStaticMethods: IStaticMethods;
  }
}

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: "start",
      component: Join,
    },
    {
      path: '/create',
      name: 'create',
      component: CreateBoard,
      beforeEnter: () => {
        if (!localStorage.getItem("user") || !localStorage.getItem("xid") || !localStorage.getItem("nickname")) {
          return `/`
        }
      },
    },
    {
      path: '/board/:board',
      name: 'dashboard',
      component: Dashboard,
      beforeEnter: (to) => {
        if (!localStorage.getItem("user") || !localStorage.getItem("xid") || !localStorage.getItem("nickname")) {
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

router.afterEach((_to, _from, failure) => {
  if (!failure) {
    setTimeout(() => {
      window.HSStaticMethods.autoInit();
    }, 100)
  }
})

export default router