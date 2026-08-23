import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../services/api'

export const useAuthStore=defineStore('auth',()=>{const loading=ref(false);const error=ref('');async function login(email:string,password:string){loading.value=true;error.value='';try{const tokens=await api.login(email,password);sessionStorage.setItem('access_token',tokens.access_token);sessionStorage.setItem('refresh_token',tokens.refresh_token)}catch(cause){error.value=cause instanceof Error?cause.message:'登录失败';throw cause}finally{loading.value=false}}function logout(){sessionStorage.removeItem('access_token');sessionStorage.removeItem('refresh_token')}return{loading,error,login,logout}})
