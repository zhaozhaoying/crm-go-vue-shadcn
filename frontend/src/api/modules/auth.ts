import { request } from "@/api/http";
import type {
  CaptchaResponse,
  LoginResponse,
  LoginRequest,
  UserWithRole,
} from "@/types/user";

export const getLoginCaptcha = (): Promise<CaptchaResponse> => {
  return request<CaptchaResponse>({
    url: "/v1/auth/captcha",
    method: "GET",
  });
};

export const login = (data: LoginRequest): Promise<LoginResponse> => {
  return request<LoginResponse>({
    url: "/v1/auth/login",
    method: "POST",
    data,
  });
};

export const refreshToken = (refreshToken: string): Promise<LoginResponse> => {
  return request<LoginResponse>({
    url: "/v1/auth/refresh",
    method: "POST",
    data: { refreshToken },
  });
};

export const logout = (
  refreshToken?: string,
): Promise<{ success: boolean }> => {
  return request<{ success: boolean }>({
    url: "/v1/auth/logout",
    method: "POST",
    data: refreshToken ? { refreshToken } : {},
  });
};

export const getCurrentUser = (): Promise<UserWithRole> => {
  return request<UserWithRole>({
    url: "/v1/auth/me",
    method: "GET",
  });
};

export interface VerifyResetIdentityRequest {
  username: string;
  contact: string;
}

export interface VerifyResetIdentityResponse {
  resetToken: string;
}

export interface ResetPasswordRequest {
  resetToken: string;
  newPassword: string;
}

export const verifyResetIdentity = (
  data: VerifyResetIdentityRequest,
): Promise<VerifyResetIdentityResponse> => {
  return request<VerifyResetIdentityResponse>({
    url: "/v1/auth/reset-password/verify",
    method: "POST",
    data,
  });
};

export const resetPassword = (
  data: ResetPasswordRequest,
): Promise<{ success: boolean }> => {
  return request<{ success: boolean }>({
    url: "/v1/auth/reset-password/confirm",
    method: "POST",
    data,
  });
};

export interface ResetPasswordDirectRequest {
  username: string;
  contact: string;
  newPassword: string;
}

export const resetPasswordDirect = (
  data: ResetPasswordDirectRequest,
): Promise<{ success: boolean }> => {
  return request<{ success: boolean }>({
    url: "/v1/auth/reset-password",
    method: "POST",
    data,
  });
};
