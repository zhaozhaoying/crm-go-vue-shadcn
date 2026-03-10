import { request } from "@/api/http"
import type { HealthPayload } from "@/types/api"

export const getHealth = () => {
  return request<HealthPayload>({
    method: "GET",
    url: "/health"
  })
}
