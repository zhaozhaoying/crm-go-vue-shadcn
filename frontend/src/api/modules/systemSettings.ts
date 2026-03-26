import { request } from "@/api/http";

export interface CustomerLevel {
  id?: number;
  name: string;
  sort: number;
}

export interface CustomerSource {
  id?: number;
  name: string;
  sort: number;
}

export interface FollowMethod {
  id?: number;
  name: string;
  sort: number;
}

export interface SystemSettings {
  customerAutoDropEnabled: boolean;
  followUpDropDays: number;
  dealDropDays: number;
  salesAssignDealDropDays: number;
  claimFreezeDays: number;
  holidayModeEnabled: boolean;
  customerLimit: number;
  showFullContact: boolean;
  contractNumberPrefix: string;
  visitPurposes: string[];
  customerLevels: CustomerLevel[];
  customerSources: CustomerSource[];
}

export interface UpdateSystemSettingsRequest {
  customerAutoDropEnabled?: boolean;
  followUpDropDays?: number;
  dealDropDays?: number;
  salesAssignDealDropDays?: number;
  claimFreezeDays?: number;
  holidayModeEnabled?: boolean;
  customerLimit?: number;
  showFullContact?: boolean;
  contractNumberPrefix?: string;
  visitPurposes?: string[];
}

export const getSystemSettings = () => {
  return request<SystemSettings>({
    method: "GET",
    url: "/v1/settings",
  });
};

export const updateSystemSettings = (data: UpdateSystemSettingsRequest) => {
  return request<{ message: string }>({
    method: "PUT",
    url: "/v1/settings",
    data,
  });
};

export const createCustomerLevel = (data: { name: string; sort: number }) => {
  return request<CustomerLevel>({
    method: "POST",
    url: "/v1/settings/customer-levels",
    data,
  });
};

export const deleteCustomerLevel = (id: number) => {
  return request<{ message: string }>({
    method: "DELETE",
    url: `/v1/settings/customer-levels/${id}`,
  });
};

export const createCustomerSource = (data: { name: string; sort: number }) => {
  return request<CustomerSource>({
    method: "POST",
    url: "/v1/settings/customer-sources",
    data,
  });
};

export const deleteCustomerSource = (id: number) => {
  return request<{ message: string }>({
    method: "DELETE",
    url: `/v1/settings/customer-sources/${id}`,
  });
};

export const getFollowMethods = () => {
  return request<FollowMethod[]>({
    method: "GET",
    url: "/v1/follow-methods",
  });
};

export const createFollowMethod = (data: { name: string; sort: number }) => {
  return request<{ id: number }>({
    method: "POST",
    url: "/v1/follow-methods",
    data,
  });
};

export const deleteFollowMethod = (id: number) => {
  return request<null>({
    method: "DELETE",
    url: `/v1/follow-methods/${id}`,
  });
};
