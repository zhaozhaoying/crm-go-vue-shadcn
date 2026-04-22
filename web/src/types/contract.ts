export type ContractCooperationType = "domestic" | "foreign"
export type ContractPaymentStatus = "pending" | "paid" | "partial"
export type ContractAuditStatus = "pending" | "success" | "failed"
export type ContractExpiryHandlingStatus = "pending" | "renewed" | "ended"

export interface Contract {
  id: number
  contractImage: string
  paymentImage: string
  paymentStatus: ContractPaymentStatus | string
  remark: string
  userId: number
  customerId: number
  cooperationType: ContractCooperationType | string
  contractNumber: string
  contractName: string
  contractAmount: number
  paymentAmount: number
  cooperationYears: number
  nodeCount: number
  serviceUserId?: number | null
  websiteName: string
  websiteUrl: string
  websiteUsername: string
  isOnline: boolean
  startDate?: string
  endDate?: string
  auditStatus: ContractAuditStatus | string
  auditComment?: string
  auditedBy?: number | null
  auditedAt?: string | null
  expiryHandlingStatus: ContractExpiryHandlingStatus | string
  userName?: string
  customerName?: string
  serviceUserName?: string
  auditedByName?: string
  createdAt?: string
  updatedAt?: string
}

export interface ContractListParams {
  keyword?: string
  paymentStatus?: string
  cooperationType?: string
  auditStatus?: string
  expiryHandlingStatus?: string
  userId?: number
  customerId?: number
  page?: number
  pageSize?: number
}

export interface ContractListResult {
  items: Contract[]
  total: number
  page: number
  pageSize: number
}

export interface ContractFormPayload {
  contractImage: string
  paymentImage: string
  paymentStatus: ContractPaymentStatus | string
  remark: string
  customerId: number
  cooperationType: ContractCooperationType | string
  contractNumber: string
  contractNumberSuffix: string
  contractName: string
  contractAmount: number
  paymentAmount: number
  cooperationYears: number
  nodeCount: number
  serviceUserId?: number | null
  websiteName: string
  websiteUrl: string
  websiteUsername: string
  isOnline: boolean
  startDate?: number | null
  endDate?: number | null
  auditStatus: ContractAuditStatus | string
  expiryHandlingStatus: ContractExpiryHandlingStatus | string
}

export type CreateContractRequest = ContractFormPayload
export type UpdateContractRequest = ContractFormPayload

export interface AuditContractRequest extends ContractFormPayload {
  auditComment?: string
}
