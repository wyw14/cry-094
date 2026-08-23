<script setup lang="ts">
import { WarningFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import type { Issue } from '../types/analysis'
defineProps<{ issue: Issue }>()
</script>
<template>
  <section class="issue" :class="issue.severity">
    <div class="issue-heading"><el-icon><CircleCloseFilled v-if="issue.severity === 'blocking'"/><WarningFilled v-else/></el-icon><strong>{{ issue.summary }}</strong><el-tag size="small" :type="issue.severity === 'blocking' ? 'danger' : 'warning'">{{ issue.kind }}</el-tag></div>
    <ol><li v-for="step in issue.evidence" :key="`${step.artifact_id}-${step.line}-${step.detail}`"><code>{{ step.artifact_id }}</code><span>{{ step.detail }}</span><small v-if="step.line">第 {{ step.line }} 行</small></li></ol>
    <p>{{ issue.resolution }}</p>
  </section>
</template>
<style scoped>.issue{border-left:3px solid #d3d8df;padding:10px 12px;background:#fff}.issue.blocking{border-color:#c73e3a}.issue.warning{border-color:#d68b1f}.issue-heading{display:flex;align-items:center;gap:8px}.issue ol{margin:10px 0;padding-left:22px}.issue li{display:grid;grid-template-columns:160px 1fr auto;gap:8px;padding:4px 0}.issue p{margin:6px 0 0;color:#59636e;font-size:13px}</style>
