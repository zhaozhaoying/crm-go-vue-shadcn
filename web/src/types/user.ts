export interface User {
  id: number
  username: string
  nickname: string
  email: string
  mobile: string
  hanghangCrmMobile?: string
  mihuaWorkNumber?: string
  avatar: string
  roleId: number
  parentId: number | null
  status: string
  createdAt: string
  updatedAt: string
}

export interface UserWithRole extends User {
  roleName: string
  roleLabel: string
}

export interface Role {
  id: number
  name: string
  label: string
  sort: number
  createdAt: string
}

export interface LoginRequest {
  username: string
  password: string
  captchaId: string
  captchaCode: string
}

export interface LoginResponse {
  token: string
  refreshToken: string
  expiresInSeconds: number
}

export interface CaptchaResponse {
  captchaId: string
  captchaImage: string
  expiresAt: number
}
