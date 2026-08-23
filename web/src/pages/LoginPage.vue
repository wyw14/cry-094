<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { Lock, User } from '@element-plus/icons-vue'
import { useAuthStore } from '../stores/auth'
const auth=useAuthStore();const router=useRouter();const form=reactive({email:'admin@example.test',password:'scriptscope-demo'})
async function submit(){try{await auth.login(form.email,form.password);await router.push('/workspace')}catch{/* store exposes the safe message */}}
</script>
<template><main class="login"><section><p>ScriptScope</p><h1>运维依赖工作台</h1><el-alert v-if="auth.error" :title="auth.error" type="error" show-icon/><el-form label-position="top" @submit.prevent="submit"><el-form-item label="邮箱"><el-input v-model="form.email" :prefix-icon="User" autocomplete="username"/></el-form-item><el-form-item label="密码"><el-input v-model="form.password" :prefix-icon="Lock" type="password" show-password autocomplete="current-password"/></el-form-item><el-button type="primary" native-type="submit" :loading="auth.loading">登录</el-button></el-form></section></main></template>
<style scoped>.login{min-height:100vh;display:grid;place-items:center;padding:20px;background:#e9edef}.login section{width:min(390px,100%);background:#fff;border:1px solid #d9dfe4;padding:28px}.login p{color:#16635b;font-weight:700;margin:0}.login h1{font-size:24px;margin:6px 0 22px}.el-button{width:100%}</style>
