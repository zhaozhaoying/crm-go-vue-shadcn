import { request } from "@/api/http"

export interface SalesDailyScore {
  id: number
  scoreDate: string
  userId: number
  userName: string
  roleName: string
  callNum: number
  callDurationSecond: number
  callScoreByCount: number
  callScoreByDuration: number
  callScoreType: string
  callScore: number
  visitCount: number
  visitScore: number
  newCustomerCount: number
  newCustomerScore: number
  totalScore: number
  createdAt: string
  updatedAt: string
}

export interface SalesDailyScoreRankingItem extends SalesDailyScore {
  rank: number
}

export interface SalesDailyScoreRankingListResult {
  scoreDate: string
  total: number
  items: SalesDailyScoreRankingItem[]
}

export interface SalesDailyScoreDetail {
  scoreDate: string
  rank: number
  totalUsers: number
  hasData: boolean
  score: SalesDailyScore
}

export const getSalesDailyScoreRankings = (params?: { scoreDate?: string }) => {
  return request<SalesDailyScoreRankingListResult>({
    method: "GET",
    url: "/v1/sales-daily-scores",
    params,
  })
}

export const getSalesDailyScoreDetail = (userId: number, params?: { scoreDate?: string }) => {
  return request<SalesDailyScoreDetail>({
    method: "GET",
    url: `/v1/sales-daily-scores/${userId}`,
    params,
  })
}

export const refreshTodaySalesDailyScoreRankings = () => {
  return request<unknown>({
    method: "POST",
    url: "/v1/tasks/hanghang-crm-daily-user-call-stats/run",
  })
}
