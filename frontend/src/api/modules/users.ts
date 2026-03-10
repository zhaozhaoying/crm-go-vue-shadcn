import { request } from "@/api/http"
import type { User, UserWithRole, Role } from "@/types/user"

export const listUsers = () => {
  return request<UserWithRole[]>({ method: "GET", url: "/v1/users" })
}

export const getUserById = (id: number) => {
  return request<User>({ method: "GET", url: `/v1/users/${id}` })
}

export const createUser = (data: {
  username: string; password: string; nickname?: string;
  email?: string; mobile?: string; avatar?: string;
  roleId: number; parentId?: number | null
}) => {
  return request<User>({ method: "POST", url: "/v1/users", data })
}

export const updateUser = (id: number, data: {
  username: string; password?: string; nickname?: string; email?: string;
  mobile?: string; avatar?: string; roleId: number;
  parentId?: number | null; status?: string
}) => {
  return request<User>({ method: "PUT", url: `/v1/users/${id}`, data })
}

export const batchDisableUsers = (userIds: number[]) => {
  return request<{ affected: number }>({
    method: "PUT",
    url: "/v1/users/batch/disable",
    data: { userIds },
  })
}

export const uploadUserAvatar = (file: File) => {
  const formData = new FormData()
  formData.append("file", file)
  return request<{ url: string }>({ method: "POST", url: "/v1/users/avatar/upload", data: formData })
}

export const resetPassword = (id: number, password: string) => {
  return request<void>({ method: "PUT", url: `/v1/users/${id}/password`, data: { password } })
}

export const deleteUser = (id: number) => {
  return request<void>({ method: "DELETE", url: `/v1/users/${id}` })
}

export const listRoles = () => {
  return request<Role[]>({ method: "GET", url: "/v1/roles" })
}

export const createRole = (data: { name: string; label: string; sort: number }) => {
  return request<Role>({ method: "POST", url: "/v1/roles", data })
}

export const updateRole = (id: number, data: { name: string; label: string; sort: number }) => {
  return request<Role>({ method: "PUT", url: `/v1/roles/${id}`, data })
}

export const deleteRole = (id: number) => {
  return request<void>({ method: "DELETE", url: `/v1/roles/${id}` })
}
