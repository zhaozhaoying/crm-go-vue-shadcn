import { request, requestBlob } from "@/api/http"

export interface CallRecording {
  id: string
  agentCode: number
  callStatus: number
  callStatusName: string
  callType: number
  calleeAttr: string
  callerAttr: string
  createTime: number
  deptName: string
  duration: number
  endTime: number
  enterpriseName: string
  finishStatus: number
  finishStatusName: string
  handle: number
  interfaceId: string
  interfaceName: string
  lineName: string
  mobile: string
  mode: number
  moveBatchCode?: string | null
  octCustomerId?: string | null
  phone: string
  postage: number
  preRecordUrl: string
  realName: string
  startTime: number
  status: number
  telA: string
  telB: string
  telX: string
  tenantCode: string
  updateTime: number
  userId: string
  workNum?: string | null
}

export interface CallRecordingListResponse {
  items: CallRecording[]
  total: number
  page: number
  pageSize: number
}

export interface SyncCallRecordingsRequest {
  startTimeBegin?: string
  startTimeFinish?: string
  minTime?: string
  limit?: number
}

export interface SyncCallRecordingsResponse {
  startTimeBegin: string
  startTimeFinish: string
  minTime: string
  pageCount: number
  totalFetched: number
  totalSaved: number
  items: CallRecording[]
}

export const getCallRecordings = (params?: {
  page?: number
  pageSize?: number
  keyword?: string
  minDuration?: number
  maxDuration?: number
}) => {
  return request<CallRecordingListResponse>({
    method: "GET",
    url: "/v1/call-recordings",
    params,
  })
}

export const getCallRecordingAudio = (id: string) => {
  return requestBlob({
    method: "GET",
    url: `/v1/call-recordings/${id}/audio`,
  })
}

export const syncCallRecordings = (data?: SyncCallRecordingsRequest) => {
  return request<SyncCallRecordingsResponse>({
    method: "POST",
    url: "/v1/call-recordings/sync",
    data,
  })
}
