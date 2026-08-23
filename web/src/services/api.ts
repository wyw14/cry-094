import type { Analysis, Artifact, PrecheckPlan } from '../types/analysis'

export class APIError extends Error { constructor(public code: string, message: string, public requestID: string) { super(message) } }
async function request<T>(path: string, init?: RequestInit): Promise<T> { const token=sessionStorage.getItem('access_token'); const response = await fetch(`/api/v1${path}`, { ...init, headers: { 'Content-Type': 'application/json', ...(token ? {Authorization:`Bearer ${token}`} : {}), ...(init?.headers ?? {}) } }); const body = await response.json().catch(() => ({})); if (!response.ok) { const error = body.error ?? {}; throw new APIError(error.code ?? 'HTTP_ERROR', error.message ?? 'Request failed', error.request_id ?? '') } return body.data as T }
export const api = {
  login: (email:string,password:string) => request<{access_token:string;refresh_token:string;expires_in:number}>('/auth/login',{method:'POST',body:JSON.stringify({email,password})}),
  listArtifacts: (teamID:string,libraryID: string) => request<{items: Artifact[], next_cursor: string}>(`/artifacts?team_id=${encodeURIComponent(teamID)}&library_id=${encodeURIComponent(libraryID)}&limit=50&sort=uploaded_at`),
  analyze: (teamID:string,libraryID: string, artifactIDs: string[], inventory: unknown) => request<Analysis>('/analyses', { method: 'POST', body: JSON.stringify({ team_id:teamID, library_id: libraryID, artifact_ids: artifactIDs, inventory }) }),
  generatePlan: (analysisID: string, teamID: string) => request<PrecheckPlan>('/precheck-plans', { method: 'POST', body: JSON.stringify({ analysis_id: analysisID, team_id: teamID }) }),
  reviewPlan: (id: string, approve: boolean, reason: string) => request<void>(`/precheck-plans/${id}/review?decision=${approve ? 'approve' : 'reject'}&reason=${encodeURIComponent(reason)}`, { method: 'POST' })
}
