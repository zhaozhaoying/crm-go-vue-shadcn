import type { Router } from "vue-router"

import { hasAnyRole, isAdminUser } from "@/lib/auth-role"
import { useAuthStore } from "@/stores/auth"

export const setupRouterGuards = (router: Router) => {
  router.beforeEach(async (to, _from, next) => {
    const authStore = useAuthStore()

    // 尝试从 localStorage 恢复认证状态
    if (!authStore.isAuthenticated && authStore.token) {
      authStore.checkAuth()
    }

    // 仅在缓存缺失时获取一次当前用户，避免每个菜单反复请求
    if (authStore.token && !authStore.profileLoaded) {
      await authStore.fetchCurrentUserProfile()
    }

    // 需要认证的路由
    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
      next({ path: "/login", query: { redirect: to.fullPath } })
      return
    }

    // 按角色可访问的路由
    const requiredRoles = Array.isArray(to.meta.requiresRoles)
      ? (to.meta.requiresRoles as string[])
      : []
    if (requiredRoles.length > 0 && !hasAnyRole(authStore.user, requiredRoles)) {
      next("/dashboard")
      return
    }

    // 仅管理员可访问的路由
    if (to.meta.requiresAdmin && !isAdminUser(authStore.user)) {
      next("/dashboard")
      return
    }

    // 已登录访问登录页，跳转到首页
    if (to.path === "/login" && authStore.isAuthenticated) {
      next("/dashboard")
      return
    }

    next()
  })
}
