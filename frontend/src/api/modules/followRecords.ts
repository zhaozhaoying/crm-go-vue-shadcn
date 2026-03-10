import { request } from '../http'

export interface FollowMethod {
  id: number
  name: string
  sort: number
  createdAt: string
}

export interface CustomerLevel {
  id: number
  name: string
  sort: number
}

export interface CustomerSource {
  id: number
  name: string
  sort: number
}

export interface OperationFollowRecord {
  id: number
  customerId: number
  customer: {
    id: number
    name?: string
    legalName?: string
    contactName?: string
    weixin?: string
    email?: string
    primaryPhone?: string
    province?: number
    city?: number
    area?: number
    detailAddress?: string
    remark?: string
    status?: string
    dealStatus?: string
    ownerUserId?: number
    ownerUserName?: string
    nextTime?: string
    followTime?: string
    collectTime?: string
    levelId?: number
    levelName?: string
    sourceId?: number
    sourceName?: string
  }
  content: string
  nextFollowTime?: string
  appointmentTime?: string
  shootingTime?: string
  customerLevelId: number
  customerLevelName?: string
  customerSourceId?: number
  customerSourceName?: string
  followMethodId: number
  followMethodName?: string
  operatorUserId: number
  operatorUserName?: string
  createdAt: string
  updatedAt: string
}

export interface SalesFollowRecord {
  id: number
  customerId: number
  customer: {
    id: number
    name?: string
    legalName?: string
    contactName?: string
    weixin?: string
    email?: string
    primaryPhone?: string
    province?: number
    city?: number
    area?: number
    detailAddress?: string
    remark?: string
    status?: string
    dealStatus?: string
    ownerUserId?: number
    ownerUserName?: string
    nextTime?: string
    followTime?: string
    collectTime?: string
    levelId?: number
    levelName?: string
    sourceId?: number
    sourceName?: string
  }
  content: string
  nextFollowTime?: string
  customerLevelId?: number
  customerLevelName?: string
  customerSourceId?: number
  customerSourceName?: string
  followMethodId?: number
  followMethodName?: string
  operatorUserId: number
  operatorUserName?: string
  createdAt: string
  updatedAt: string
}

export interface CreateFollowRecordRequest {
  customerId: number
  content: string
  nextFollowTime?: string
  appointmentTime?: string
  shootingTime?: string
  customerLevelId?: number
  customerSourceId?: number
  followMethodId?: number
}

export interface FollowRecordListResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

// 跟进方式相关
export const getFollowMethods = () => {
  return request<FollowMethod[]>({
    method: 'GET',
    url: '/v1/follow-methods'
  })
}

export const createFollowMethod = (data: { name: string; sort: number }) => {
  return request({
    method: 'POST',
    url: '/v1/follow-methods',
    data
  })
}

export const updateFollowMethod = (id: number, data: { name: string; sort: number }) => {
  return request({
    method: 'PUT',
    url: `/v1/follow-methods/${id}`,
    data
  })
}

export const deleteFollowMethod = (id: number) => {
  return request({
    method: 'DELETE',
    url: `/v1/follow-methods/${id}`
  })
}

// 运营跟进记录相关
export const getOperationFollowRecords = (customerId: number, page = 1, pageSize = 20) => {
  return request<FollowRecordListResponse<OperationFollowRecord>>({
    method: 'GET',
    url: '/v1/operation-follow-records',
    params: { customerId, page, pageSize }
  })
}

export const getAllOperationFollowRecords = (page = 1, pageSize = 20) => {
  return request<FollowRecordListResponse<OperationFollowRecord>>({
    method: 'GET',
    url: '/v1/operation-follow-records/all',
    params: { page, pageSize }
  })
}

export const createOperationFollowRecord = (data: CreateFollowRecordRequest) => {
  return request({
    method: 'POST',
    url: '/v1/operation-follow-records',
    data
  })
}

// 销售跟进记录相关
export const getSalesFollowRecords = (customerId: number, page = 1, pageSize = 20) => {
  return request<FollowRecordListResponse<SalesFollowRecord>>({
    method: 'GET',
    url: '/v1/sales-follow-records',
    params: { customerId, page, pageSize }
  })
}

export const getAllSalesFollowRecords = (page = 1, pageSize = 20) => {
  return request<FollowRecordListResponse<SalesFollowRecord>>({
    method: 'GET',
    url: '/v1/sales-follow-records/all',
    params: { page, pageSize }
  })
}

export const createSalesFollowRecord = (data: CreateFollowRecordRequest) => {
  return request({
    method: 'POST',
    url: '/v1/sales-follow-records',
    data
  })
}

// 客户级别相关
export const getCustomerLevels = () => {
  return request<CustomerLevel[]>({
    method: 'GET',
    url: '/v1/settings/customer-levels'
  })
}

// 客户来源相关
export const getCustomerSources = () => {
  return request<CustomerSource[]>({
    method: 'GET',
    url: '/v1/settings/customer-sources'
  })
}
