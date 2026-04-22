import { request } from "@/api/http"
import type { ActivityLog } from "@/types/notification"

export const getActivityLogs = (limit = 50) => {
  return request<ActivityLog[]>({
    method: "GET",
    url: "/v1/notifications/activity-logs",
    params: { limit },
  })
}

export const getNotificationReadKeys = () => {
  return request<{ keys: string[] }>({
    method: "GET",
    url: "/v1/notifications/read-keys",
  })
}

export const markNotificationsRead = (keys: string[]) => {
  return request<null>({
    method: "POST",
    url: "/v1/notifications/mark-read",
    data: { keys },
  })
}

export const getUnreadCount = () => {
  return request<{ count: number }>({
    method: "GET",
    url: "/v1/notifications/unread-count",
  })
}
