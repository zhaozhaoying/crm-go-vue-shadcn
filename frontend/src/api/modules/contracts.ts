import { request } from "@/api/http"
import type {
  AuditContractRequest,
  Contract,
  ContractListParams,
  ContractListResult,
  CreateContractRequest,
  UpdateContractRequest,
} from "@/types/contract"

const createEmptyListResult = (): ContractListResult => {
  return {
    items: [],
    total: 0,
    page: 1,
    pageSize: 10,
  }
}

const normalizeListResult = (data: ContractListResult | null | undefined): ContractListResult => {
  if (!data) return createEmptyListResult()
  return {
    items: Array.isArray(data.items) ? data.items : [],
    total: Number.isFinite(data.total) ? data.total : 0,
    page: Number.isFinite(data.page) && data.page > 0 ? data.page : 1,
    pageSize: Number.isFinite(data.pageSize) && data.pageSize > 0 ? data.pageSize : 10,
  }
}

export const listContracts = (params?: ContractListParams) => {
  return request<ContractListResult | null>({
    method: "GET",
    url: "/v1/contracts",
    params,
  }).then((data) => normalizeListResult(data))
}

export const getContract = (id: number) => {
  return request<Contract>({
    method: "GET",
    url: `/v1/contracts/${id}`,
  })
}

export const checkContractNumberAvailable = (contractNumber: string, excludeId?: number) => {
  return request<{ available: boolean }>({
    method: "GET",
    url: "/v1/contracts/check-number",
    params: {
      contractNumber,
      excludeId,
    },
  })
}

export const uploadContractImage = (file: File) => {
  const formData = new FormData()
  formData.append("file", file)
  return request<{ url: string }>({
    method: "POST",
    url: "/v1/users/avatar/upload",
    data: formData,
  })
}

export const createContract = (data: CreateContractRequest) => {
  return request<Contract>({
    method: "POST",
    url: "/v1/contracts",
    data,
  })
}

export const updateContract = (id: number, data: UpdateContractRequest) => {
  return request<Contract>({
    method: "PUT",
    url: `/v1/contracts/${id}`,
    data,
  })
}

export const auditContract = (id: number, data: AuditContractRequest) => {
  return request<Contract>({
    method: "POST",
    url: `/v1/contracts/${id}/audit`,
    data,
  })
}

export const deleteContract = (id: number) => {
  return request<void>({
    method: "DELETE",
    url: `/v1/contracts/${id}`,
  })
}
