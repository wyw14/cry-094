<script setup lang="ts">
import { onMounted, reactive } from 'vue'
import { Refresh, VideoPlay, DocumentChecked } from '@element-plus/icons-vue'
import { useWorkspaceStore } from '../stores/workspace'
import AnalysisPanel from '../features/analysis/AnalysisPanel.vue'
const store = useWorkspaceStore()
const selection = reactive({ teamID: 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', libraryID: 'demo-library' })
const inventory = { target_name: 'staging-cn-01', os: 'linux', tools: { jq: { name: 'jq', version: { major: 1, minor: 7, patch: 1 } }, yq: { name: 'yq', version: { major: 4, minor: 44, patch: 3 } } }, environment_names: { HOME: true, PATH: true }, probe_build: 'local-demo-v1' }
onMounted(() => store.loadArtifacts(selection.teamID,selection.libraryID))
</script>
<template>
  <main>
    <header><div><p class="eyebrow">ScriptScope / Platform Operations</p><h1>执行前依赖工作台</h1></div><div class="actions"><el-button :icon="Refresh" @click="store.loadArtifacts(selection.teamID,selection.libraryID)"/><el-button type="primary" :icon="VideoPlay" :loading="store.loading" @click="store.runAnalysis(selection.teamID,selection.libraryID, inventory)">静态分析</el-button></div></header>
    <el-alert v-if="store.error" type="error" :title="store.error" show-icon/>
    <section class="status-strip"><div><span>脚本版本</span><strong>{{ store.artifacts.length }}</strong></div><div><span>阻断项</span><strong :class="{danger:store.blockingCount>0}">{{ store.blockingCount }}</strong></div><div><span>解析器</span><strong>{{ store.analysis?.parser_build ?? 'static-v1' }}</strong></div><div><span>编排状态</span><el-tag :type="store.analysis?.status === 'blocked' ? 'danger' : 'success'">{{ store.analysis?.status ?? '待分析' }}</el-tag></div></section>
    <section><div class="section-head"><h2>版本清单</h2><span>摘要绑定 · 只读存储</span></div><el-table :data="store.artifacts" size="small" stripe><el-table-column prop="filename" label="文件" min-width="180"/><el-table-column prop="number" label="版本" width="72"/><el-table-column prop="shell" label="解析器" width="110"/><el-table-column label="SHA-256" min-width="180"><template #default="scope"><code>{{ scope.row.digest.slice(0,18) }}…</code></template></el-table-column><el-table-column prop="size" label="字节" width="90"/></el-table></section>
    <section v-if="store.analysis"><div class="section-head"><h2>依赖图与证据链</h2><el-button :icon="DocumentChecked" :disabled="store.analysis.status === 'blocked'" @click="store.createPlan(selection.teamID)">生成检查方案</el-button></div><AnalysisPanel :analysis="store.analysis"/></section>
    <section v-if="store.plan"><div class="section-head"><h2>人工审核队列</h2><el-tag>{{ store.plan.state }}</el-tag></div><el-table :data="store.plan.steps" size="small"><el-table-column prop="description" label="检查项"/><el-table-column prop="command" label="只读命令"><template #default="scope"><code>{{ scope.row.command }}</code></template></el-table-column></el-table></section>
  </main>
</template>
<style scoped>main{max-width:1280px;margin:0 auto;padding:24px}header,.section-head{display:flex;justify-content:space-between;align-items:center;gap:16px}header{margin-bottom:18px}h1{font-size:28px;margin:2px 0}h2{font-size:17px;margin:0}.eyebrow{font-size:12px;color:#5f6b76;margin:0}.actions{display:flex;gap:8px}.status-strip{display:grid;grid-template-columns:repeat(4,1fr);border:1px solid #dfe3e8;background:#fff;margin-bottom:18px}.status-strip div{padding:12px 16px;border-right:1px solid #dfe3e8}.status-strip div:last-child{border:0}.status-strip span{display:block;font-size:12px;color:#68737d}.status-strip strong{font-size:20px}.danger{color:#b42318}section{margin-bottom:22px}.section-head{margin-bottom:9px}.section-head>span{font-size:12px;color:#68737d}code{font-family:'Cascadia Code',Consolas,monospace;font-size:12px}@media(max-width:720px){main{padding:14px}.status-strip{grid-template-columns:1fr 1fr}header{align-items:flex-start;flex-direction:column}}</style>
