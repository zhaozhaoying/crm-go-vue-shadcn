import { request } from "@/api/http"
import type { CaptchaResponse, LoginResponse, LoginRequest, UserWithRole } from "@/types/user"

export const getLoginCaptcha = (): Promise<CaptchaResponse> => {
  return request<CaptchaResponse>({
    url: "/v1/auth/captcha",
    method: "GET"
  })
}

export const login = (data: LoginRequest): Promise<LoginResponse> => {
  return request<LoginResponse>({
    url: "/v1/auth/login",
    method: "POST",
    data
  })
}

export const refreshToken = (refreshToken: string): Promise<LoginResponse> => {
  return request<LoginResponse>({
    url: "/v1/auth/refresh",
    method: "POST",
    data: { refreshToken }
  })
}

export const logout = (refreshToken?: string): Promise<{ success: boolean }> => {
  return request<{ success: boolean }>({
    url: "/v1/auth/logout",
    method: "POST",
    data: refreshToken ? { refreshToken } : {}
  })
}

export const getCurrentUser = (): Promise<UserWithRole> => {
  return request<UserWithRole>({
    url: "/v1/auth/me",
    method: "GET"
  })
}
