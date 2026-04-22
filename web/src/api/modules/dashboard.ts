import { request } from "@/api/http"
import type { DashboardOverview } from "@/types/dashboard"

export const getDashboardOverview = () => {
  return request<DashboardOverview>({
    method: "GET",
    url: "/v1/dashboard/overview",
  })
}
