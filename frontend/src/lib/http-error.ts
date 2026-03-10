import axios from "axios"

type ApiErrorData = {
  code?: number
  message?: string
}

const containsChinese = (message: string): boolean => {
  return /[\u4e00-\u9fa5]/.test(message)
}

const toChineseMessage = (message: string, status?: number): string => {
  const normalized = message.trim()
  if (!normalized) {
    return status ? `请求失败（${status}）` : "请求失败"
  }
  if (containsChinese(normalized)) {
    return normalized
  }

  const lower = normalized.toLowerCase()
  if (lower.includes("network error") || lower.includes("failed to fetch")) {
    return "网络异常，无法连接后端服务"
  }
  if (lower.includes("timeout")) {
    return "请求超时，请稍后重试"
  }
  if (lower.includes("request failed with status code 404")) {
    return "接口不存在（请重启后端并确认已更新到最新代码）"
  }
  if (lower.includes("request failed with status code")) {
    return status ? `请求失败（${status}）` : "请求失败"
  }
  if (lower.includes("request failed")) {
    return status ? `请求失败（${status}）` : "请求失败"
  }
  if (status === 404) {
    return "接口不存在（请重启后端并确认已更新到最新代码）"
  }
  if (status) {
    return `请求失败（${status}）`
  }
  return "请求失败，请稍后重试"
}

export const getRequestErrorMessage = (error: unknown, fallback = "请求失败"): string => {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status
    const data = error.response?.data as ApiErrorData | string | undefined

    if (typeof data === "object" && data && typeof data.message === "string" && data.message.trim()) {
      return toChineseMessage(data.message, status)
    }

    if (typeof data === "string" && data.trim()) {
      return toChineseMessage(data, status)
    }

    if (status === 404) {
      return "接口不存在（请重启后端并确认已更新到最新代码）"
    }

    if (!status) {
      return "无法连接后端服务，请确认后端已启动（默认 8080）"
    }

    return toChineseMessage(error.message, status)
  }

  if (error instanceof Error && error.message) {
    return toChineseMessage(error.message)
  }

  return toChineseMessage(fallback)
}
