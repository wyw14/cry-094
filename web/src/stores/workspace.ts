import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { api } from '../services/api'
import type { Analysis, Artifact, PrecheckPlan } from '../types/analysis'

export const useWorkspaceStore = defineStore('workspace', () => {
  const artifacts = ref<Artifact[]>([])
  const analysis = ref<Analysis | null>(null)
  const plan = ref<PrecheckPlan | null>(null)
  const loading = ref(false)
  const error = ref('')
  const blockingCount = computed(() => analysis.value?.issues.filter(i => i.severity === 'blocking').length ?? 0)
  async function loadArtifacts(teamID:string,libraryID: string) { loading.value = true; error.value = ''; try { artifacts.value = (await api.listArtifacts(teamID,libraryID)).items } catch (cause) { error.value = cause instanceof Error ? cause.message : '加载失败' } finally { loading.value = false } }
  async function runAnalysis(teamID:string,libraryID: string, inventory: unknown) { loading.value = true; error.value = ''; try { analysis.value = await api.analyze(teamID,libraryID, artifacts.value.map(a => a.id), inventory) } catch (cause) { error.value = cause instanceof Error ? cause.message : '分析失败' } finally { loading.value = false } }
  async function createPlan(teamID: string) { if (!analysis.value) return; plan.value = await api.generatePlan(analysis.value.id, teamID) }
  return { artifacts, analysis, plan, loading, error, blockingCount, loadArtifacts, runAnalysis, createPlan }
})
