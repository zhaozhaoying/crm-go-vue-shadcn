import { defineStore } from "pinia"
import { ref, computed } from "vue"

import {
  getCurrentUser,
  login as loginApi,
  logout as logoutApi,
} from "@/api/modules/auth"
import type { LoginRequest, UserWithRole } from "@/types/user"

const TOKEN_KEY = "auth_token"
const REFRESH_TOKEN_KEY = "refresh_token"
const USER_CACHE_KEY = "auth_user"

export interface AuthUser {
  id: number
  username: string
  nickname: string
  email: string
  mobile: string
  avatar: string
  role: string
  roleId: number
  roleName: string
  roleLabel: string
  parentId: number | null
  status: string
  createdAt: string
  updatedAt: string
}

const parseJwtPayload = (token: string): Record<string, unknown> | null => {
  try {
    const base64 = token.split(".")[1]
    const json = atob(base64.replace(/-/g, "+").replace(/_/g, "/"))
    return JSON.parse(json)
  } catch {
    return null
  }
}

const userFromToken = (token: string): AuthUser | null => {
  const p = parseJwtPayload(token)
  if (!p) return null
  if (typeof p.exp === "number" && p.exp * 1000 < Date.now()) return null
  return {
    id: Number(p.sub ?? 0),
    username: p.username as string,
    nickname: (p.nickname as string) || "",
    email: (p.email as string) || "",
    mobile: "",
    avatar: "",
    role: (p.role as string) || "",
    roleId: Number(p.role_id ?? 0),
    roleName: (p.role as string) || "",
    roleLabel: "",
    parentId: null,
    status: "",
    createdAt: "",
    updatedAt: "",
  }
}

const readCachedUser = (): AuthUser | null => {
  const raw = localStorage.getItem(USER_CACHE_KEY)
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as Partial<AuthUser>
    if (!parsed || typeof parsed !== "object") return null
    if (!parsed.id || typeof parsed.username !== "string") return null
    return {
      id: Number(parsed.id),
      username: parsed.username,
      nickname: parsed.nickname || "",
      email: parsed.email || "",
      mobile: parsed.mobile || "",
      avatar: parsed.avatar || "",
      role: parsed.role || parsed.roleName || "",
      roleId: Number(parsed.roleId || 0),
      roleName: parsed.roleName || parsed.role || "",
      roleLabel: parsed.roleLabel || "",
      parentId: parsed.parentId ?? null,
      status: parsed.status || "",
      createdAt: parsed.createdAt || "",
      updatedAt: parsed.updatedAt || "",
    }
  } catch {
    localStorage.removeItem(USER_CACHE_KEY)
    return null
  }
}

const saveCachedUser = (user: AuthUser | null) => {
  if (!user) {
    localStorage.removeItem(USER_CACHE_KEY)
    return
  }
  localStorage.setItem(USER_CACHE_KEY, JSON.stringify(user))
}

const mapCurrentUser = (user: UserWithRole): AuthUser => {
  return {
    id: user.id,
    username: user.username,
    nickname: user.nickname || "",
    email: user.email || "",
    mobile: user.mobile || "",
    avatar: user.avatar || "",
    role: user.roleName || "",
    roleId: Number(user.roleId || 0),
    roleName: user.roleName || "",
    roleLabel: user.roleLabel || "",
    parentId: user.parentId ?? null,
    status: user.status || "",
    createdAt: user.createdAt || "",
    updatedAt: user.updatedAt || "",
  }
}

export const useAuthStore = defineStore("auth", () => {
  const user = ref<AuthUser | null>(readCachedUser())
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const refreshToken = ref<string | null>(localStorage.getItem(REFRESH_TOKEN_KEY))
  const loading = ref(false)
  const profileLoaded = ref(!!user.value)
  let profileRequest: Promise<AuthUser | null> | null = null

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  const setUser = (next: AuthUser | null) => {
    user.value = next
    saveCachedUser(next)
  }

  const setTokens = (accessToken: string, nextRefreshToken: string) => {
    token.value = accessToken
    refreshToken.value = nextRefreshToken
    localStorage.setItem(TOKEN_KEY, accessToken)
    localStorage.setItem(REFRESH_TOKEN_KEY, nextRefreshToken)
  }

  const clearAuth = () => {
    setUser(null)
    token.value = null
    refreshToken.value = null
    profileLoaded.value = false
    profileRequest = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    localStorage.removeItem(USER_CACHE_KEY)
  }

  const fetchCurrentUserProfile = async (force = false): Promise<AuthUser | null> => {
    if (!token.value) return null
    if (!force && profileLoaded.value && user.value) return user.value
    if (profileRequest) return profileRequest

    profileRequest = (async () => {
      loading.value = true
      try {
        const profile = await getCurrentUser()
        const next = mapCurrentUser(profile)
        const tokenUser = token.value ? userFromToken(token.value) : null
        setUser({
          ...(tokenUser ?? next),
          ...next,
          role: next.role || tokenUser?.role || "",
          roleId: next.roleId || tokenUser?.roleId || 0,
          roleName: next.roleName || tokenUser?.role || "",
        })
        profileLoaded.value = true
      } catch {
        if (!user.value && token.value) {
          const parsed = userFromToken(token.value)
          if (parsed) setUser(parsed)
        }
        // If API is unavailable, keep existing cached/token user to avoid
        // repeated profile requests on every route change.
        profileLoaded.value = !!user.value
      } finally {
        loading.value = false
        profileRequest = null
      }
      return user.value
    })()

    return profileRequest
  }

  const login = async (credentials: LoginRequest) => {
    loading.value = true
    try {
      const response = await loginApi(credentials)
      setTokens(response.token, response.refreshToken)
      setUser(userFromToken(response.token))
      profileLoaded.value = false
      await fetchCurrentUserProfile(true)
      return response
    } finally {
      loading.value = false
    }
  }

  const checkAuth = (): boolean => {
    const savedToken = localStorage.getItem(TOKEN_KEY)
    const savedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
    if (!savedToken) {
      clearAuth()
      return false
    }
    const parsed = userFromToken(savedToken)
    if (!parsed) {
      clearAuth()
      return false
    }
    token.value = savedToken
    refreshToken.value = savedRefreshToken
    const cached = readCachedUser()
    if (cached && cached.id === parsed.id) {
      setUser({
        ...parsed,
        ...cached,
        role: cached.role || cached.roleName || parsed.role,
        roleId: cached.roleId || parsed.roleId,
        roleName: cached.roleName || cached.role || parsed.role,
      })
      profileLoaded.value = true
    } else {
      setUser(parsed)
      profileLoaded.value = false
    }
    return true
  }

  const logout = async () => {
    try {
      if (token.value) {
        await logoutApi(refreshToken.value || undefined)
      }
    } catch {
      // Ignore logout request failure and clear local state anyway.
    } finally {
      clearAuth()
    }
  }

  return {
    user,
    token,
    refreshToken,
    loading,
    profileLoaded,
    isAuthenticated,
    login,
    logout,
    checkAuth,
    fetchCurrentUserProfile,
    clearAuth,
  }
})
