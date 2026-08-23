<script setup lang="ts">
import type { GraphEdge, GraphNode } from '../types/analysis'
defineProps<{ nodes: GraphNode[]; edges: GraphEdge[]; order: string[] }>()
</script>
<template>
  <div class="graph" role="img" aria-label="脚本依赖执行顺序">
    <div v-for="(id,index) in order" :key="id" class="node"><span class="index">{{ index + 1 }}</span><div><strong>{{ id }}</strong><small>{{ nodes.find(n => n.artifact_id === id)?.digest.slice(0,12) }}</small></div><span v-if="index < order.length-1" class="arrow">→</span></div>
    <el-empty v-if="order.length === 0" description="存在阻断，无法生成执行顺序" :image-size="52"/>
    <div class="edges"><span v-for="edge in edges" :key="`${edge.from}-${edge.to}`">{{ edge.from }} → {{ edge.to }} · {{ edge.contract }}</span></div>
  </div>
</template>
<style scoped>.graph{min-height:160px;padding:12px;background:#f5f7f8;border:1px solid #dfe3e8}.node{display:flex;align-items:center;gap:10px;min-height:42px}.node div{display:flex;flex-direction:column}.node small{color:#7d8792}.index{width:24px;height:24px;border-radius:50%;display:grid;place-items:center;background:#16635b;color:#fff}.arrow{margin-left:auto;color:#7d8792}.edges{border-top:1px solid #dfe3e8;margin-top:10px;padding-top:8px;display:flex;flex-direction:column;font-size:12px;color:#59636e}</style>
