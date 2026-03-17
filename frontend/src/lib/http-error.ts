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

  if (status) {
    return `${normalized} (${status})`
  }
  return normalized || "请求失败，请稍后重试"
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
      return "接口不存在 (404)"
    }

    if (!status) {
      return "无法连接后端服务"
    }

    return toChineseMessage(error.message, status)
  }

  if (error instanceof Error && error.message) {
    return toChineseMessage(error.message)
  }

  return toChineseMessage(fallback)
}
