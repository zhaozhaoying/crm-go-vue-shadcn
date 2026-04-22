import { computed, ref } from "vue"
import { defineStore } from "pinia"

import type { ActivityLog, NotificationCategory, NotificationItem } from "@/types/notification"
import {
  getActivityLogs,
  getNotificationReadKeys,
  markNotificationsRead,
} from "@/api/modules/notification"

const actionLabelMap: Record<string, string> = {
  create_contract: "创建合同",
  audit_contract: "审核合同",
  create_customer: "创建客户",
  import_customer: "导入客户",
  claim_customer: "领取客户",
  release_customer: "丢弃客户",
  auto_drop_follow_up: "未跟进掉库通知",
  auto_drop_deal: "未签单掉库通知",
  transfer_customer: "转移客户",
  sales_follow: "销售跟进",
  operation_follow: "运营跟进",
}

function normalizeText(value?: string | null): string {
  return String(value ?? "").trim()
}

function wrapName(value: string, fallback: string): string {
  return `【${normalizeText(value) || fallback}】`
}

function formatNotificationDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, "0")
  const day = String(date.getDate()).padStart(2, "0")
  const hours = String(date.getHours()).padStart(2, "0")
  const minutes = String(date.getMinutes()).padStart(2, "0")
  const seconds = String(date.getSeconds()).padStart(2, "0")
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

function buildNotificationContent(log: ActivityLog): Pick<NotificationItem, "title" | "summary" | "content"> {
  const actionLabel = actionLabelMap[log.action] ?? log.action
  const actorName = normalizeText(log.userName) || (log.userId ? `用户${log.userId}` : "")
  const targetName = normalizeText(log.targetName)
  const logContent = normalizeText(log.content)
  const timeText = formatNotificationDateTime(log.createdAt)
  const actorDisplay = wrapName(actorName, actorName || "未知用户")
  const customerDisplay = wrapName(targetName, "未知客户")

  switch (log.action) {
    case "release_customer": {
      const summary = `销售${actorDisplay}于${timeText}丢弃了客户${customerDisplay}`
      return {
        title: "丢弃客户",
        summary,
        content: summary,
      }
    }
    case "claim_customer": {
      const summary = `销售${actorDisplay}于${timeText}领取了客户${customerDisplay}`
      return {
        title: "领取客户",
        summary,
        content: summary,
      }
    }
    case "auto_drop_follow_up": {
      const detail =
        logContent ||
        `客户${customerDisplay}因销售${actorDisplay}未跟进，系统于${timeText}自动触发掉库。`
      return {
        title: "未跟进掉库通知",
        summary: detail,
        content: detail,
      }
    }
    case "auto_drop_deal": {
      const detail =
        logContent ||
        `客户${customerDisplay}因销售${actorDisplay}未签单，系统于${timeText}自动触发掉库。`
      return {
        title: "未签单掉库通知",
        summary: detail,
        content: detail,
      }
    }
    case "sales_follow": {
      const base = `销售${actorDisplay}于${timeText}跟进了客户${customerDisplay}`
      const detail = logContent || "未填写跟进内容"
      return {
        title: "销售跟进",
        summary: `${base}，跟进内容：${detail}`,
        content: `${base}\n跟进内容：${detail}`,
      }
    }
    case "operation_follow": {
      const base = `运营${actorDisplay}于${timeText}跟进了客户${customerDisplay}`
      const detail = logContent || "未填写跟进内容"
      return {
        title: "运营跟进",
        summary: `${base}，跟进内容：${detail}`,
        content: `${base}\n跟进内容：${detail}`,
      }
    }
    default: {
      const targetDisplay = targetName ? wrapName(targetName, "未命名目标") : ""
      const base = targetDisplay
        ? `${actorDisplay}于${timeText}${actionLabel}${targetDisplay}`
        : `${actorDisplay}于${timeText}${actionLabel}`
      return {
        title: actionLabel,
        summary: logContent ? `${base}，说明：${logContent}` : base,
        content: logContent ? `${base}\n说明：${logContent}` : base,
      }
    }
  }
}

function resolveCategory(targetType: string): NotificationCategory {
  if (targetType === "contract") return "contract"
  if (targetType === "customer") return "customer"
  return "system"
}

function activityToNotification(
  log: ActivityLog,
  readKeys: Set<string>,
): NotificationItem {
  const key = `activity-${log.id}`
  const { title, summary, content } = buildNotificationContent(log)

  return {
    id: log.id,
    key,
    title,
    summary,
    content,
    category: resolveCategory(log.targetType),
    createdAt: log.createdAt,
    unread: !readKeys.has(key),
  }
}

export const useNotificationStore = defineStore("notification", () => {
  const notifications = ref<NotificationItem[]>([])
  const loading = ref(false)
  const readKeys = ref<Set<string>>(new Set())

  const orderedNotifications = computed(() =>
    [...notifications.value].sort((a, b) => b.id - a.id),
  )

  const unreadCount = computed(
    () => notifications.value.filter((n) => n.unread).length,
  )

  const markAsRead = async (id: number) => {
    const target = notifications.value.find((n) => n.id === id)
    if (!target || !target.unread) return

    target.unread = false
    readKeys.value.add(target.key)

    try {
      await markNotificationsRead([target.key])
    } catch {
      // revert on failure
      target.unread = true
      readKeys.value.delete(target.key)
    }
  }

  const markAllAsRead = async () => {
    const unreadItems = notifications.value.filter((n) => n.unread)
    if (unreadItems.length === 0) return

    const keys = unreadItems.map((n) => n.key)
    unreadItems.forEach((n) => {
      n.unread = false
      readKeys.value.add(n.key)
    })

    try {
      await markNotificationsRead(keys)
    } catch {
      unreadItems.forEach((n) => {
        n.unread = true
        readKeys.value.delete(n.key)
      })
    }
  }

  const fetchNotifications = async () => {
    if (loading.value) return
    loading.value = true
    try {
      const [logs, readResult] = await Promise.all([
        getActivityLogs(100),
        getNotificationReadKeys(),
      ])
      const keys = new Set(readResult?.keys ?? [])
      readKeys.value = keys
      notifications.value = (logs ?? []).map((log) =>
        activityToNotification(log, keys),
      )
    } finally {
      loading.value = false
    }
  }

  return {
    notifications,
    loading,
    orderedNotifications,
    unreadCount,
    markAsRead,
    markAllAsRead,
    fetchNotifications,
  }
})
