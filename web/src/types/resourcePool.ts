export interface ResourcePoolItem {
  id: number
  name: string
  phone: string
  address: string
  province: string
  city: string
  area: string
  latitude: number
  longitude: number
  source: string
  sourceUid: string
  searchKeyword: string
  searchRadius: number
  searchRegion: string
  queryAddress: string
  centerLatitude: number
  centerLongitude: number
  createdBy: number
  converted: boolean
  convertedCustomerId?: number | null
  convertedAt?: string | null
  convertedBy?: number | null
  createdAt: string
  updatedAt: string
}

export interface ResourcePoolListParams {
  keyword?: string
  hasPhone?: string
  page?: number
  pageSize?: number
}

export interface ResourcePoolListResult {
  items: ResourcePoolItem[]
  total: number
  page: number
  pageSize: number
}

export interface ResourcePoolSearchRequest {
  region?: string
  address?: string
  radius?: number
  keyword?: string
  centerLatitude?: number
  centerLongitude?: number
}

export interface ResourcePoolSearchResult {
  centerLatitude: number
  centerLongitude: number
  totalFetched: number
  totalSaved: number
  items: ResourcePoolItem[]
}

export interface ResourcePoolConvertResult {
  resourceId: number
  customerId: number
  alreadyLinked: boolean
}

export interface ResourcePoolBatchConvertRequest {
  resourceIds: number[]
}

export interface ResourcePoolBatchConvertItemResult {
  resourceId: number
  customerId: number
  alreadyLinked: boolean
  success: boolean
  error?: string
}

export interface ResourcePoolBatchConvertResult {
  total: number
  success: number
  failed: number
  items: ResourcePoolBatchConvertItemResult[]
}
