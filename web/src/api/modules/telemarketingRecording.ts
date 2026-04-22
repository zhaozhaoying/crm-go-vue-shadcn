import { request } from "@/api/http"

export interface TelemarketingRecording {
  id: string
  ccNumber: string
  sid: number
  seid: number
  ccgeid: number
  callType: number
  outlineNumber: string
  encryptedOutlineNumber: string
  switchNumber: string
  initiator: string
  initiatorCallId: string
  serviceNumber: string
  serviceUid: number
  serviceSeatName: string
  serviceSeatWorkNumber: string
  serviceGroupName: string
  initiateTime: number
  ringTime: number
  confirmTime: number
  disconnectTime: number
  conversationTime: number
  durationSecond: number
  durationText: string
  validDurationText: string
  customerRingDuration: number
  seatRingDuration: number
  recordStatus: number
  recordFilename: string
  recordResToken: string
  evaluateValue: string
  cmResult: string
  cmDescription: string
  attribution: string
  stopReason: number
  customerFailReason: string
  customerName: string
  customerCompany: string
  groupNames: string
  seatNames: string
  seatNumbers: string
  seatWorkNumbers: string
  enterpriseName: string
  districtName: string
  serviceDeviceNumber: string
  callAnswerResult: number
  callHangupParty: number
  matchedUserId?: number
  matchedUserName: string
  roleName: string
  remoteCreatedAt?: string
  remoteUpdatedAt?: string
}

export interface TelemarketingRecordingListResponse {
  items: TelemarketingRecording[]
  total: number
  page: number
  pageSize: number
}

export interface TelemarketingRecordingDetailResponse {
  recording: TelemarketingRecording
  playbackUrl: string
  playbackFilename: string
  playbackExpiresAt: number
}

export interface SyncTelemarketingRecordingsRequest {
  pageSize?: number
  timePeriod?: string
}

export interface SyncTelemarketingRecordingsResponse {
  pageCount: number
  totalFetched: number
  totalSaved: number
  timePeriod: string
  items: TelemarketingRecording[]
}

export const getTelemarketingRecordings = (params?: {
  page?: number
  pageSize?: number
  keyword?: string
  startDate?: string
  endDate?: string
  minDuration?: number
  maxDuration?: number
}) => {
  return request<TelemarketingRecordingListResponse>({
    method: "GET",
    url: "/v1/telemarketing-recordings",
    params,
  })
}

export const getTelemarketingRecordingDetail = (id: string) => {
  return request<TelemarketingRecordingDetailResponse>({
    method: "GET",
    url: `/v1/telemarketing-recordings/${id}`,
  })
}

export const syncTelemarketingRecordings = (data?: SyncTelemarketingRecordingsRequest) => {
  return request<SyncTelemarketingRecordingsResponse>({
    method: "POST",
    url: "/v1/telemarketing-recordings/sync",
    data,
  })
}
