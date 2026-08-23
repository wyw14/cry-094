<script setup lang="ts">
import IssueEvidence from '../../components/IssueEvidence.vue'
import DependencyGraph from '../../components/DependencyGraph.vue'
import type { Analysis } from '../../types/analysis'
defineProps<{ analysis: Analysis }>()
</script>
<template><div class="analysis-grid"><DependencyGraph :nodes="analysis.nodes" :edges="analysis.edges" :order="analysis.order"/><div class="issues"><IssueEvidence v-for="issue in analysis.issues" :key="`${issue.kind}-${issue.summary}`" :issue="issue"/><el-empty v-if="analysis.issues.length===0" description="未发现依赖风险" :image-size="52"/></div></div></template>
<style scoped>.analysis-grid{display:grid;grid-template-columns:minmax(320px,1fr) minmax(360px,1.25fr);gap:14px}.issues{display:flex;flex-direction:column;gap:8px}@media(max-width:900px){.analysis-grid{grid-template-columns:1fr}}</style>
