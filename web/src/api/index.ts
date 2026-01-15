import axios from 'axios'
import { ref } from 'vue'

const TOKEN_KEY = 'go2short_token'

export const tokenRef = ref<string | null>(localStorage.getItem(TOKEN_KEY))

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
  tokenRef.value = token
}

export function removeToken() {
  localStorage.removeItem(TOKEN_KEY)
  tokenRef.value = null
}

const api = axios.create({
  baseURL: '/api',
})

api.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      removeToken()
      window.location.href = '/admin/login'
    }
    return Promise.reject(error)
  }
)

export interface LoginResponse {
  token: string
}

export interface Link {
  code: string
  short_url: string
  long_url: string
  created_at: string
  expires_at?: string
  is_disabled: boolean
}

export interface LinksResponse {
  links: Link[]
  total: number
  page: number
  limit: number
}

export interface OverviewStats {
  total_links: number
  active_links: number
  total_clicks: number
  today_clicks: number
}

export interface DayClick {
  date: string
  clicks: number
}

export interface LinkStats {
  total_clicks: number
  daily_clicks: DayClick[]
}

// Super admin login
export async function login(username: string, password: string): Promise<LoginResponse> {
  const { data } = await api.post<LoginResponse>('/admin/login', { username, password })
  return data
}

// User login (email + password)
export async function userLogin(email: string, password: string): Promise<LoginResponse> {
  const { data } = await api.post<LoginResponse>('/auth/login', { email, password })
  return data
}

// User registration
export async function register(email: string, password: string): Promise<LoginResponse> {
  const { data } = await api.post<LoginResponse>('/auth/register', { email, password })
  return data
}

// Get available OAuth providers
export async function getOAuthProviders(): Promise<{ providers: string[] }> {
  const { data } = await api.get<{ providers: string[] }>('/auth/providers')
  return data
}

export async function logout(): Promise<void> {
  await api.post('/admin/logout')
  removeToken()
}

export async function getLinks(page = 1, limit = 20, search = ''): Promise<LinksResponse> {
  const params: Record<string, string | number> = { page, limit }
  if (search) params.search = search
  const { data } = await api.get<LinksResponse>('/admin/links', { params })
  return data
}

export async function createLink(longUrl: string, customCode?: string, expiresAt?: string): Promise<{ short_url: string }> {
  const payload: Record<string, string> = { long_url: longUrl }
  if (customCode) payload.custom_code = customCode
  if (expiresAt) payload.expires_at = expiresAt
  const { data } = await api.post('/admin/links', payload)
  return data
}

export async function updateLink(code: string, longUrl: string, expiresAt?: string): Promise<void> {
  const payload: Record<string, string | null> = { long_url: longUrl }
  payload.expires_at = expiresAt || null
  await api.put(`/admin/links/${code}`, payload)
}

export async function deleteLink(code: string): Promise<void> {
  await api.delete(`/admin/links/${code}`)
}

export async function setLinkDisabled(code: string, disabled: boolean): Promise<void> {
  await api.patch(`/admin/links/${code}/disable`, { disabled })
}

export async function getLinkStats(code: string, days = 30): Promise<LinkStats> {
  const { data } = await api.get<LinkStats>(`/admin/links/${code}/stats`, { params: { days } })
  return data
}

export async function getOverviewStats(): Promise<OverviewStats> {
  const { data } = await api.get<OverviewStats>('/admin/stats/overview')
  return data
}

export interface TopLink {
  code: string
  short_url: string
  long_url: string
  click_count: number
}

export async function getTopLinks(limit = 10, days = 30): Promise<{ links: TopLink[] }> {
  const { data } = await api.get<{ links: TopLink[] }>('/admin/stats/top-links', { params: { limit, days } })
  return data
}

export async function getClickTrend(days = 30): Promise<{ trend: DayClick[] }> {
  const { data } = await api.get<{ trend: DayClick[] }>('/admin/stats/trend', { params: { days } })
  return data
}

// API Tokens
export interface APIToken {
  id: number
  name: string
  created_at: string
  last_used_at?: string
}

export interface CreateTokenResponse {
  id: number
  name: string
  token: string
}

export async function getAPITokens(): Promise<{ tokens: APIToken[] }> {
  const { data } = await api.get<{ tokens: APIToken[] }>('/admin/tokens')
  return data
}

export async function createAPIToken(name: string): Promise<CreateTokenResponse> {
  const { data } = await api.post<CreateTokenResponse>('/admin/tokens', { name })
  return data
}

export async function deleteAPIToken(id: number): Promise<void> {
  await api.delete(`/admin/tokens/${id}`)
}
