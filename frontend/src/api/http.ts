import axios, { type AxiosRequestConfig } from "axios"

import type { ApiEnvelope } from "@/types/api"

const envBaseURL = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim()
const isLocalHost =
  typeof window !== "undefined" &&
  (window.location.hostname === "localhost" || window.location.hostname === "127.0.0.1")
const defaultBaseURL = isLocalHost ? "http://localhost:8080/api" : "/api"

const http = axios.create({
  baseURL: envBaseURL || defaultBaseURL,
  timeout: 10000
})

// 请求拦截器：自动附加 token
http.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("auth_token")
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器：处理 401 错误
http.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("auth_token")
      localStorage.removeItem("refresh_token")
      localStorage.removeItem("auth_user")
      window.location.href = "/login"
    }
    return Promise.reject(error)
  }
)

export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const response = await http.request<ApiEnvelope<T>>(config)
  const payload = response.data as unknown

  // Compatibility: support both envelope payload ({ code, message, data })
  // and raw JSON payload returned by older handlers.
  if (
    payload &&
    typeof payload === "object" &&
    "code" in payload &&
    "message" in payload
  ) {
    const envelope = payload as ApiEnvelope<T>
    if (envelope.code !== 0) {
      throw new Error(envelope.message || "请求失败")
    }
    return envelope.data
  }

  return payload as T
}
