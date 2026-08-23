import { createRouter, createWebHistory } from 'vue-router'
import WorkspacePage from '../pages/WorkspacePage.vue'
import LoginPage from '../pages/LoginPage.vue'
export const router = createRouter({ history: createWebHistory(), routes: [{path:'/',redirect:'/workspace'},{ path: '/login', component:LoginPage },{ path: '/workspace', component: WorkspacePage,meta:{requiresAuth:true} }] })
router.beforeEach(to=>{if(to.meta.requiresAuth&&!sessionStorage.getItem('access_token'))return '/login';if(to.path==='/login'&&sessionStorage.getItem('access_token'))return '/workspace'})
