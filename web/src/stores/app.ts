import { defineStore } from "pinia"

import { getHealth } from "@/api/modules/health"
import { getRequestErrorMessage } from "@/lib/http-error"
import type { HealthPayload } from "@/types/api"

interface AppState {
  health: HealthPayload | null
  loadingHealth: boolean
  healthError: string
}

export const useAppStore = defineStore("app", {
  state: (): AppState => ({
    health: null,
    loadingHealth: false,
    healthError: ""
  }),
  actions: {
    async fetchHealth() {
      this.loadingHealth = true
      this.healthError = ""

      try {
        this.health = await getHealth()
      } catch (error) {
        this.healthError = getRequestErrorMessage(error, "健康检查失败")
      } finally {
        this.loadingHealth = false
      }
    }
  }
})
