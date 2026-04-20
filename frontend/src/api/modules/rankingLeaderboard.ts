import { request } from "@/api/http"

export interface RankingLeaderboardItem {
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
  total: number
  items: RankingLeaderboardItem[]
}

export const getRankingLeaderboard = (params?: { period?: string }) => {
  return request<RankingLeaderboardResult>({
    method: "GET",
    url: "/v1/ranking-leaderboard",
    params,
  })
}
