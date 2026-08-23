import { setActivePinia, createPinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useWorkspaceStore } from '../src/stores/workspace'

describe('workspace store', () => {
  beforeEach(() => setActivePinia(createPinia()))
  it('counts only blocking evidence', () => {
    const store = useWorkspaceStore()
    store.analysis = { id:'a', library_id:'l', nodes:[], edges:[], order:[], parser_build:'v1', completed_at:'2026-01-01T00:00:00Z', status:'blocked', issues:[
      {kind:'cycle',severity:'blocking',summary:'cycle',evidence:[],resolution:'break cycle'},
      {kind:'missing',severity:'warning',summary:'env',evidence:[],resolution:'configure'}
    ] }
    expect(store.blockingCount).toBe(1)
  })
})
