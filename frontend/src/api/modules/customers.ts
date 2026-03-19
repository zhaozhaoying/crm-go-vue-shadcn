import { request } from "@/api/http"
import type {
  Customer,
  CustomerListResult,
  CustomerListParams,
  LegacyCustomerListParams,
  CreateCustomerRequest,
  UpdateCustomerRequest,
  CustomerPhone,
  AddPhoneRequest,
  UpdatePhoneRequest,
  CustomerStatusLog,
  CreateStatusLogRequest,
  CustomerUniqueCheckRequest,
  CustomerUniqueCheckResult
} from "@/types/customer"

const createEmptyListResult = (): CustomerListResult => {
  return {
    items: [],
    total: 0,
    page: 1,
    pageSize: 10
  }
}

const normalizeListResult = (data: CustomerListResult | null | undefined): CustomerListResult => {
  if (!data) return createEmptyListResult()
  return {
    items: Array.isArray(data.items) ? data.items : [],
    total: Number.isFinite(data.total) ? data.total : 0,
    page: Number.isFinite(data.page) && data.page > 0 ? data.page : 1,
    pageSize: Number.isFinite(data.pageSize) && data.pageSize > 0 ? data.pageSize : 10
  }
}

const listCustomerPage = (url: string, params?: CustomerListParams) => {
  return request<CustomerListResult | null>({
    method: "GET",
    url,
    params
  }).then((data) => normalizeListResult(data))
}

export const listCustomersPage = (params?: CustomerListParams) =>
  listCustomerPage("/v1/customers", params)

// Legacy method, returns only list data.
export const listCustomers = (params?: LegacyCustomerListParams) => {
  return request<CustomerListResult | null>({
    method: "GET",
    url: "/v1/customers",
    params
  }).then((data) => normalizeListResult(data).items)
}

export const listMyCustomers = (params?: CustomerListParams) => listCustomerPage("/v1/customers/my", params)
export const listPoolCustomers = (params?: CustomerListParams) => listCustomerPage("/v1/customers/pool", params)
export const listPotentialCustomers = (params?: CustomerListParams) => listCustomerPage("/v1/customers/potential", params)
export const listPartnerCustomers = (params?: CustomerListParams) => listCustomerPage("/v1/customers/partner", params)
export const listSearchCustomers = (params?: CustomerListParams) => listCustomerPage("/v1/customers/search", params)

export const createCustomer = (data: CreateCustomerRequest) => {
  return request<Customer>({
    method: "POST",
    url: "/v1/customers",
    data
  })
}

export const updateCustomer = (customerId: number, data: UpdateCustomerRequest) => {
  return request<Customer>({
    method: "PUT",
    url: `/v1/customers/${customerId}`,
    data
  })
}

export const claimCustomer = (customerId: number) => {
  return request<Customer>({
    method: "POST",
    url: `/v1/customers/${customerId}/claim`
  })
}

export const convertCustomer = (customerId: number) => {
  return request<Customer>({
    method: "POST",
    url: `/v1/customers/${customerId}/convert`
  })
}

export const releaseCustomer = (customerId: number) => {
  return request<Customer>({
    method: "POST",
    url: `/v1/customers/${customerId}/release`
  })
}

// Phone management APIs
export const addPhone = (customerId: number, data: AddPhoneRequest) => {
  return request<CustomerPhone>({
    method: "POST",
    url: `/v1/customers/${customerId}/phones`,
    data
  })
}

export const listPhones = (customerId: number) => {
  return request<CustomerPhone[]>({
    method: "GET",
    url: `/v1/customers/${customerId}/phones`
  })
}

export const updatePhone = (customerId: number, phoneId: number, data: UpdatePhoneRequest) => {
  return request<CustomerPhone>({
    method: "PUT",
    url: `/v1/customers/${customerId}/phones/${phoneId}`,
    data
  })
}

export const deletePhone = (customerId: number, phoneId: number) => {
  return request<void>({
    method: "DELETE",
    url: `/v1/customers/${customerId}/phones/${phoneId}`
  })
}

// Status log APIs
export const listStatusLogs = (customerId: number, params?: { page?: number; pageSize?: number }) => {
  return request<CustomerStatusLog[]>({
    method: "GET",
    url: `/v1/customers/${customerId}/status-logs`,
    params
  })
}

export const createStatusLog = (customerId: number, data: CreateStatusLogRequest) => {
  return request<CustomerStatusLog>({
    method: "POST",
    url: `/v1/customers/${customerId}/status-logs`,
    data
  })
}

export const validateCustomerUnique = (data: CustomerUniqueCheckRequest) => {
  return request<CustomerUniqueCheckResult>({
    method: "POST",
    url: "/v1/customers/validate-unique",
    data
  })
}
