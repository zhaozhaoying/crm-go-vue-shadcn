import { request } from "@/api/http"

export type RankingLeaderboardPeriod = "month" | "week" | "day" | "all"

export interface RankingLeaderboardItem {
  identityKey: string
  rank: number
  seatWorkNumber: string
  seatName: string
  matchedUserId?: number
  matchedUserName: string
  groupName: string
  roleName: string
  callNum: number
  answeredCallCount: number
  answerRate: number
  callDurationSecond: number
  newCustomerCount: number
  invitationCount: number
  callScore: number
  invitationScore: number
  newCustomerScore: number
  totalScore: number
  scoreDays: number
}

export interface RankingLeaderboardResult {
  period: string
  startDate: string
  endDate: string
  total: number
  items: RankingLeaderboardItem[]
}

export interface RankingLeaderboardDetail {
  period: string
  startDate: string
  endDate: string
  rank: number
  totalUsers: number
  hasData: boolean
  score: RankingLeaderboardItem
}

export const getRankingLeaderboard = (params?: {
  period?: RankingLeaderboardPeriod | string
  startDate?: string
  endDate?: string
}) => {
  return request<RankingLeaderboardResult>({
    method: "GET",
    url: "/v1/ranking-leaderboard",
    params,
  })
}

export const getRankingLeaderboardDetail = (
  identityKey: string,
  params?: {
    period?: RankingLeaderboardPeriod | string
    startDate?: string
    endDate?: string
  },
) => {
  return request<RankingLeaderboardDetail>({
    method: "GET",
    url: `/v1/ranking-leaderboard/${encodeURIComponent(identityKey)}`,
    params,
  })
}
