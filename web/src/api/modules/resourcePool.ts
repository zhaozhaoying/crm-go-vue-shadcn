import { request } from "@/api/http"
import type {
  ResourcePoolBatchConvertRequest,
  ResourcePoolBatchConvertResult,
  ResourcePoolConvertResult,
  ResourcePoolListParams,
  ResourcePoolListResult,
  ResourcePoolSearchRequest,
  ResourcePoolSearchResult
} from "@/types/resourcePool"

const createEmptyListResult = (): ResourcePoolListResult => ({
  items: [],
  total: 0,
  page: 1,
  pageSize: 10
})

const normalizeListResult = (
  data: ResourcePoolListResult | null | undefined
): ResourcePoolListResult => {
  if (!data) return createEmptyListResult()
  return {
    items: Array.isArray(data.items) ? data.items : [],
    total: Number.isFinite(data.total) ? data.total : 0,
    page: Number.isFinite(data.page) && data.page > 0 ? data.page : 1,
    pageSize: Number.isFinite(data.pageSize) && data.pageSize > 0 ? data.pageSize : 10
  }
}

export const listResourcePool = (params?: ResourcePoolListParams) => {
  return request<ResourcePoolListResult | null>({
    method: "GET",
    url: "/v1/resource-pool",
    params
  }).then((data) => normalizeListResult(data))
}

export const searchAndStoreResourcePool = (data: ResourcePoolSearchRequest) => {
  return request<ResourcePoolSearchResult>({
    method: "POST",
    url: "/v1/resource-pool/search",
    data
  })
}

export const convertResourcePoolToCustomer = (resourceId: number) => {
  return request<ResourcePoolConvertResult>({
    method: "POST",
    url: `/v1/resource-pool/${resourceId}/convert`
  })
}

export const convertResourcePoolBatchToCustomer = (data: ResourcePoolBatchConvertRequest) => {
  return request<ResourcePoolBatchConvertResult>({
    method: "POST",
    url: "/v1/resource-pool/convert/batch",
    data
  })
}
