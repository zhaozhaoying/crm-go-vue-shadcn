import { request } from '../http'

export interface CustomerVisit {
  id: number
  operatorUserId: number
  operatorUserName?: string
  customerName: string
  inviter: string
  checkInLat: number
  checkInLng: number
  province: string
  city: string
  area: string
  detailAddress: string
  images: string
  visitPurpose: string
  remark: string
  visitDate: string
  createdAt: string
  updatedAt: string
}

export interface CustomerVisitListResponse {
  items: CustomerVisit[]
  total: number
  page: number
  pageSize: number
}

export interface CreateCustomerVisitRequest {
  customerName: string
  inviter?: string
  checkInLat: number
  checkInLng: number
  province?: string
  city?: string
  area?: string
  detailAddress: string
  images: string
  visitPurpose: string
  remark: string
}

export const getCustomerVisits = (params: {
  page?: number
  pageSize?: number
  keyword?: string
  startTime?: string
  endTime?: string
}) => {
  return request<CustomerVisitListResponse>({
    method: 'GET',
    url: '/v1/customer-visits',
    params,
  })
}

export const createCustomerVisit = (data: CreateCustomerVisitRequest) => {
  return request<{ id: number }>({
    method: 'POST',
    url: '/v1/customer-visits',
    data,
  })
}

export const uploadVisitImage = (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return request<{ url: string }>({
    method: 'POST',
    url: '/v1/users/avatar/upload',
    data: formData,
  })
}
