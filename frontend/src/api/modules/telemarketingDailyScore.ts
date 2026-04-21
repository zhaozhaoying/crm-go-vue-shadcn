import { request } from "@/api/http"

export interface TelemarketingDailyScore {
  scoreDate: string
  seatWorkNumber: string
  seatName: string
  matchedUserId?: number
  matchedUserName: string
  serviceNumber: string
  groupName: string
  roleName: string
  callNum: number
  answeredCallCount: number
  missedCallCount: number
  answerRate: number
  callDurationSecond: number
  newCustomerCount: number
  invitationCount: number
  callScoreByCount: number
  callScoreByDuration: number
  callScoreType: string
  callScore: number
  invitationScore: number
  newCustomerScore: number
  totalScore: number
  scoreReachedAt?: string
  updatedAt: string
}

export interface TelemarketingDailyScoreRankingItem extends TelemarketingDailyScore {
  rank: number
}

export interface TelemarketingDailyScoreRankingListResult {
  scoreDate: string
  total: number
  items: TelemarketingDailyScoreRankingItem[]
}

export interface TelemarketingDailyScoreDetail {
  scoreDate: string
  rank: number
  totalUsers: number
  hasData: boolean
  score: TelemarketingDailyScore
}

export const getTelemarketingDailyScoreRankings = (params?: { scoreDate?: string; sync?: boolean }) => {
  return request<TelemarketingDailyScoreRankingListResult>({
    method: "GET",
    url: "/v1/telemarketing-rankings",
    params,
  })
}

export const getTelemarketingDailyScoreDetail = (seatWorkNumber: string, params?: { scoreDate?: string; sync?: boolean }) => {
  return request<TelemarketingDailyScoreDetail>({
    method: "GET",
    url: `/v1/telemarketing-rankings/${seatWorkNumber}`,
    params,
  })
}
