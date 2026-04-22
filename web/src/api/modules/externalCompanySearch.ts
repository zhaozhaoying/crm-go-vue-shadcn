import { request } from "@/api/http";
import type {
  CreateExternalCompanySearchTasksRequest,
  CreateExternalCompanySearchTasksResponse,
  ExternalCompanySearchEvent,
  ExternalCompanySearchEventListResult,
  ExternalCompanySearchResultListResult,
  ExternalCompanySearchTask,
  ExternalCompanySearchTaskListResult,
  ListExternalCompanySearchResultsParams,
  ListExternalCompanySearchTasksParams,
} from "@/types/externalCompanySearch";

export function createExternalCompanySearchTasks(
  data: CreateExternalCompanySearchTasksRequest,
) {
  return request<CreateExternalCompanySearchTasksResponse>({
    method: "POST",
    url: "/v1/external-company-search/tasks",
    data,
  });
}

export function listExternalCompanySearchTasks(
  params: ListExternalCompanySearchTasksParams = {},
) {
  return request<ExternalCompanySearchTaskListResult>({
    method: "GET",
    url: "/v1/external-company-search/tasks",
    params,
  });
}

export function getExternalCompanySearchTask(taskId: number) {
  return request<ExternalCompanySearchTask>({
    method: "GET",
    url: `/v1/external-company-search/tasks/${taskId}`,
  });
}

export function cancelExternalCompanySearchTask(taskId: number) {
  return request<{ taskId: number }>({
    method: "POST",
    url: `/v1/external-company-search/tasks/${taskId}/cancel`,
  });
}

export function listExternalCompanySearchResults(
  taskId: number,
  params: ListExternalCompanySearchResultsParams = {},
) {
  return request<ExternalCompanySearchResultListResult>({
    method: "GET",
    url: `/v1/external-company-search/tasks/${taskId}/results`,
    params,
  });
}

export function listAllExternalCompanySearchResults(
  params: ListExternalCompanySearchResultsParams = {},
) {
  return request<ExternalCompanySearchResultListResult>({
    method: "GET",
    url: "/v1/external-company-search/results",
    params,
  });
}

export function enrichExternalCompany(companyId: number) {
  return request<import("@/types/externalCompanySearch").ExternalCompany>({
    method: "POST",
    url: `/v1/external-company-search/companies/${companyId}/enrich`,
  });
}

export function listExternalCompanySearchEvents(
  taskId: number,
  afterSeq = 0,
  limit = 100,
) {
  return request<ExternalCompanySearchEventListResult>({
    method: "GET",
    url: `/v1/external-company-search/tasks/${taskId}/events`,
    params: { afterSeq, limit },
  });
}

export interface ExternalCompanySearchStreamOptions {
  taskId: number;
  afterSeq?: number;
  token?: string | null;
  onEvent: (event: ExternalCompanySearchEvent) => void;
  onError?: (event: Event) => void;
  onClose?: (event: CloseEvent) => void;
}

export function connectExternalCompanySearchStream(
  options: ExternalCompanySearchStreamOptions,
) {
  const apiBase =
    (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim() ||
    (typeof window !== "undefined" &&
    (window.location.hostname === "localhost" ||
      window.location.hostname === "127.0.0.1")
      ? "http://localhost:8080/api"
      : `${window.location.origin}/api`);
  const wsBase = apiBase.replace(/^http/i, "ws");
  const url = new URL(
    `${wsBase}/v1/external-company-search/tasks/${options.taskId}/stream`,
  );
  if (options.afterSeq && options.afterSeq > 0) {
    url.searchParams.set("afterSeq", String(options.afterSeq));
  }
  const token = options.token ?? localStorage.getItem("auth_token");
  if (token) {
    url.searchParams.set("token", token);
  }

  const socket = new WebSocket(url.toString());
  socket.onmessage = (messageEvent) => {
    try {
      const parsed = JSON.parse(
        String(messageEvent.data),
      ) as ExternalCompanySearchEvent;
      options.onEvent(parsed);
    } catch {
      // ignore malformed frames
    }
  };
  if (options.onError) {
    socket.onerror = options.onError;
  }
  if (options.onClose) {
    socket.onclose = options.onClose;
  }
  return socket;
}
