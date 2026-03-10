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
  transfer_customer: "转移客户",
  sales_follow: "销售跟进",
  operation_follow: "运营跟进",
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
  const actionLabel = actionLabelMap[log.action] ?? log.action
  const title = log.targetName
    ? `${actionLabel} - ${log.targetName}`
    : actionLabel

  return {
    id: log.id,
    key,
    title,
    summary: log.content || actionLabel,
    content: log.content || "",
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

  const getById = (id: number) => {
    return notifications.value.find((n) => n.id === id) ?? null
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
    getById,
    fetchNotifications,
  }
})
