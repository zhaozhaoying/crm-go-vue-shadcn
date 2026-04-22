import axios from "axios"

type ApiErrorData = {
  code?: number
  message?: string
}

const rawEnglishTracePattern =
  /(error \d+ \([^)]+\)|sqlstate|column\s+'[^']+'\s+cannot\s+be\s+null|duplicate entry|foreign key constraint fails|request failed with status code|network error|timeout of \d+ms exceeded)/i

const containsChinese = (message: string): boolean => {
  return /[\u4e00-\u9fa5]/.test(message)
}

const stripMixedLanguageTrace = (message: string): string => {
  const normalized = message.trim()
  if (!normalized || !containsChinese(normalized) || !rawEnglishTracePattern.test(normalized)) {
    return normalized
  }

  const parts = normalized.split(/[:：]/).map((part) => part.trim()).filter(Boolean)
  const chineseParts = parts.filter((part) => containsChinese(part) && !rawEnglishTracePattern.test(part))
  if (chineseParts.length > 0) {
    return chineseParts.join("：")
  }
  return normalized
}

const getChineseStatusMessage = (status?: number): string => {
  switch (status) {
    case 400:
      return "请求参数错误"
    case 401:
      return "登录已失效，请重新登录"
    case 403:
      return "无权限执行该操作"
    case 404:
      return "接口不存在"
    case 409:
      return "请求冲突"
    case 422:
      return "请求数据校验失败"
    case 429:
      return "请求过于频繁，请稍后重试"
    case 500:
      return "服务器内部错误"
    case 502:
      return "网关错误"
    case 503:
      return "服务暂不可用"
    case 504:
      return "服务响应超时"
    default:
      return ""
  }
}

const toChineseMessage = (message: string, fallback: string, status?: number): string => {
  const normalized = message.trim()
  if (!normalized) {
    return getChineseStatusMessage(status) || fallback || "请求失败"
  }
  const stripped = stripMixedLanguageTrace(normalized)
  if (containsChinese(stripped) && !rawEnglishTracePattern.test(stripped)) {
    return stripped
  }

  const statusMessage = getChineseStatusMessage(status)
  if (statusMessage) {
    return statusMessage
  }
  return fallback || "请求失败，请稍后重试"
}

export const getRequestErrorMessage = (error: unknown, fallback = "请求失败"): string => {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status
    const data = error.response?.data as ApiErrorData | string | undefined

    if (typeof data === "object" && data && typeof data.message === "string" && data.message.trim()) {
      return toChineseMessage(data.message, fallback, status)
    }

    if (typeof data === "string" && data.trim()) {
      return toChineseMessage(data, fallback, status)
    }

    if (!status) {
      return "无法连接后端服务"
    }

    return toChineseMessage(error.message, fallback, status)
  }

  if (error instanceof Error && error.message) {
    return toChineseMessage(error.message, fallback)
  }

  return toChineseMessage(fallback, fallback)
}
