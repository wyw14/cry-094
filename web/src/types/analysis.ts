export type Severity = 'info' | 'warning' | 'blocking'
export interface EvidenceStep { artifact_id: string; dependency: string; detail: string; line?: number }
export interface Issue { kind: string; severity: Severity; summary: string; evidence: EvidenceStep[]; resolution: string }
export interface GraphNode { artifact_id: string; digest: string }
export interface GraphEdge { from: string; to: string; contract: string; implicit: boolean; evidence: string }
export interface Analysis { id: string; library_id: string; nodes: GraphNode[]; edges: GraphEdge[]; order: string[]; issues: Issue[]; status: 'ready' | 'blocked'; parser_build: string; completed_at: string }
export interface Artifact { id: string; filename: string; number: number; shell: 'shell' | 'powershell'; digest: string; size: number; uploaded_at: string }
export interface PrecheckPlan { id: string; analysis_id: string; state: 'draft'|'in_review'|'approved'|'rejected'|'signed'; steps: Array<{description:string;command:string;read_only:boolean}>; reviewed_by?: string; signature?: string }
