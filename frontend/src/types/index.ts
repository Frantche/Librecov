export interface User {
  id: number
  email: string
  name: string
  admin: boolean
  token?: string
  created_at: string
  updated_at: string
}

export interface AuthConfig {
  oidc_enabled: boolean
  oidc?: {
    issuer: string
    client_id: string
    redirect_url: string
  }
}

export interface Project {
  id: number
  name: string
  token: string
  current_branch: string
  base_url: string
  coverage_rate: number
  user_id: number
  user?: User
  builds?: Build[]
  created_at: string
  updated_at: string
}

export interface Build {
  id: number
  project_id: number
  build_num: number
  branch: string
  commit_sha: string
  commit_msg: string
  coverage_rate: number
  project?: Project
  jobs?: Job[]
  created_at: string
  updated_at: string
}

export interface Job {
  id: number
  build_id: number
  job_number: string
  coverage_rate: number
  data: string
  build?: Build
  files?: JobFile[]
  created_at: string
  updated_at: string
}

export interface JobFile {
  id: number
  job_id: number
  name: string
  coverage: string
  source: string
  coverage_rate: number
  job?: Job
  created_at: string
  updated_at: string
}

export interface RefreshSessionResponse {
  token: string
}
