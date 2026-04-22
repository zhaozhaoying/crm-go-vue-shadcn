export interface ExternalCompany {
  id: number;
  companyNo: string;
  platform: number;
  platformCompanyId: string;
  dedupeKey: string;
  companyName: string;
  companyNameEn?: string;
  companyUrl?: string;
  companyLogo?: string;
  companyImages?: string;
  companyDesc?: string;
  country?: string;
  province?: string;
  city?: string;
  address?: string;
  mainProducts?: string;
  businessType?: string;
  employeeCount?: string;
  establishedYear?: string;
  annualRevenue?: string;
  certification?: string;
  contact?: string;
  phone?: string;
  email?: string;
  dataVersion: number;
  interestStatus: number;
  isDeleted: boolean;
  rawPayload?: string;
  firstSeenAt?: string | null;
  lastSeenAt?: string | null;
  createTime: string;
  updateTime: string;
}

export interface ExternalCompanySearchTask {
  id: number;
  taskNo: string;
  platform: number;
  keyword: string;
  keywordNormalized: string;
  regionKeyword: string;
  status: number;
  priority: number;
  targetCount: number;
  pageLimit: number;
  pageNo: number;
  progressPercent: number;
  fetchedCount: number;
  savedCount: number;
  duplicateCount: number;
  failedCount: number;
  retryCount: number;
  maxRetryCount: number;
  nextRunAt?: string | null;
  lockedAt?: string | null;
  lastHeartbeatAt?: string | null;
  startedAt?: string | null;
  finishedAt?: string | null;
  workerToken?: string;
  searchOptions?: string;
  resumeCursor?: string;
  errorMessage?: string;
  createdBy: number;
  createdAt: string;
  updatedAt: string;
}

export interface ExternalCompanySearchTaskListResult {
  items: ExternalCompanySearchTask[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ExternalCompanySearchResultItem {
  id: number;
  taskId: number;
  companyId: number;
  platform: number;
  keyword: string;
  regionKeyword: string;
  pageNo: number;
  rankNo: number;
  isNewCompany: boolean;
  resultPayload?: string;
  createdAt: string;
  updatedAt: string;
  companyNo: string;
  platformCompanyId: string;
  dedupeKey: string;
  companyName: string;
  companyNameEn?: string;
  companyUrl?: string;
  companyLogo?: string;
  companyImages?: string;
  companyDesc?: string;
  country?: string;
  province?: string;
  city?: string;
  address?: string;
  mainProducts?: string;
  businessType?: string;
  employeeCount?: string;
  establishedYear?: string;
  annualRevenue?: string;
  certification?: string;
  contact?: string;
  phone?: string;
  email?: string;
  dataVersion: number;
  interestStatus: number;
  isDeleted: boolean;
  rawPayload?: string;
  firstSeenAt?: string | null;
  lastSeenAt?: string | null;
}

export interface ExternalCompanySearchResultListResult {
  items: ExternalCompanySearchResultItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ExternalCompanySearchEvent {
  id: number;
  taskId: number;
  seqNo: number;
  eventType: string;
  message?: string;
  payload?: string;
  createdAt: string;
}

export interface ExternalCompanySearchEventListResult {
  items: ExternalCompanySearchEvent[];
  nextSeq: number;
}

export interface CreateExternalCompanySearchTasksRequest {
  platforms: number[];
  keyword: string;
  regionKeyword?: string;
  pageLimit?: number;
  targetCount?: number;
  priority?: number;
  searchOptions?: Record<string, unknown> | string | null;
}

export interface CreateExternalCompanySearchTasksResponse {
  items: ExternalCompanySearchTask[];
}

export interface ListExternalCompanySearchTasksParams {
  platform?: number;
  status?: number;
  keyword?: string;
  page?: number;
  pageSize?: number;
}

export interface ListExternalCompanySearchResultsParams {
  search?: string;
  platform?: number;
  newOnly?: boolean;
  page?: number;
  pageSize?: number;
}

export const EXTERNAL_COMPANY_SEARCH_PLATFORM = {
  ALIBABA: 1,
  MADE_IN_CHINA: 2,
  GOOGLE: 3,
} as const;

export const EXTERNAL_COMPANY_SEARCH_EVENT_TYPE = {
  TASK_CREATED: "task.created",
  TASK_STARTED: "task.started",
  TASK_PROGRESS: "task.progress",
  TASK_COMPLETED: "task.completed",
  TASK_FAILED: "task.failed",
  TASK_CANCELED: "task.canceled",
  RESULT_SAVED: "result.saved",
} as const;

export const EXTERNAL_COMPANY_SEARCH_TASK_STATUS = {
  PENDING: 1,
  RUNNING: 2,
  COMPLETED: 3,
  FAILED: 4,
  CANCELED: 5,
} as const;

export interface ExternalCompanySearchTaskProgressPayload {
  taskId: number;
  status: number;
  pageNo: number;
  progressPercent: number;
  fetchedCount: number;
  savedCount: number;
  duplicateCount: number;
  failedCount: number;
}

export interface ExternalCompanySearchResultSavedPayload {
  taskId: number;
  companyId: number;
  companyName: string;
  platform: number;
  pageNo: number;
  rankNo: number;
  isNewCompany: boolean;
  duplicateCount: number;
}

export interface ExternalCompanySearchFailedPayload extends Partial<ExternalCompanySearchTaskProgressPayload> {
  errorMessage?: string;
}
