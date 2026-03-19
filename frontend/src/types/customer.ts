export type CustomerCategory = "my" | "pool" | "search" | "potential" | "partner"
export type CustomerPoolSortBy = "dropTime" | "followTime" | "updatedAt"

export type DealStatus = "undone" | "done" | string | number

export interface CustomerPhone {
  id: number
  customerId: number
  phone: string
  phoneLabel?: string
  isPrimary: boolean
  createdAt: string
  updatedAt: string
}

export interface CustomerStatusLog {
  id: number
  customerId: number
  fromStatus: number
  toStatus: number
  triggerType: number
  reason?: string
  operatorUserId?: number
  operatorName?: string
  operateTime: string
}

export interface Customer {
  id: number
  name: string
  legalName?: string
  contactName?: string
  weixin?: string
  email?: string
  status?: string
  createdAt?: string

  customerLevelId?: number
  customerSourceId?: number
  customerLevelName?: string
  customerSourceName?: string

  province?: number
  city?: number
  area?: number
  detailAddress?: string
  lng?: number
  lat?: number

  nextTime?: string
  followTime?: string
  remark?: string

  ownerUserId?: number | null
  ownerUserName?: string
  dealStatus?: DealStatus
  dealTime?: string

  customerStatus?: number
  collectTime?: string
  dropTime?: string
  dropUserId?: number
  dropUserName?: string

  createUserId?: number
  insideSalesUserId?: number | null
  convertedAt?: string | null
  operateUserId?: number

  isLock?: boolean
  isInPool?: boolean
  historicalOwnerIds?: number[]

  phones?: CustomerPhone[]
  deleteTime?: string
}

export interface CustomerListParams {
  ownershipScope?: string
  keyword?: string
  name?: string
  contactName?: string
  phone?: string
  weixin?: string
  ownerUserName?: string
  province?: string
  city?: string
  area?: string
  excludePool?: string
  sortBy?: CustomerPoolSortBy | string
  page?: number
  pageSize?: number
}

export interface CustomerListResult {
  items: Customer[]
  total: number
  page: number
  pageSize: number
}

export interface LegacyCustomerListParams {
  category?: CustomerCategory
  keyword?: string
  page?: number
  pageSize?: number
}

export interface CustomerFormPhone {
  id?: number
  phone: string
  phoneLabel?: string
  isPrimary: boolean
}

export interface CustomerFormPayload {
  name: string
  legalName?: string
  contactName?: string
  email?: string
  weixin?: string
  province?: number
  city?: number
  area?: number
  detailAddress?: string
  nextTime?: string
  remark?: string
  phones?: CustomerFormPhone[]
}

export interface CreateCustomerRequest extends CustomerFormPayload {
  status?: string
  dealStatus?: DealStatus
  ownerUserId?: number | null
}

export type UpdateCustomerRequest = Partial<CreateCustomerRequest>

export interface AddPhoneRequest {
  phone: string
  phoneLabel?: string
  isPrimary: boolean
}

export interface UpdatePhoneRequest {
  phone: string
  phoneLabel?: string
  isPrimary: boolean
}

export interface CreateStatusLogRequest {
  toStatus: number
  reason?: string
}

export interface CustomerUniqueCheckRequest {
  excludeCustomerId?: number
  name?: string
  legalName?: string
  contactName?: string
  weixin?: string
  phones?: string[]
}

export interface CustomerUniqueCheckResult {
  nameExists: boolean
  legalNameExists: boolean
  contactNameExists: boolean
  weixinExists: boolean
  duplicatePhones: string[]
}
