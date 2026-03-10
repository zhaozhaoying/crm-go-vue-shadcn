export interface ActivityLog {
  id: number
  userId: number
  action: string
  targetType: string
  targetId: number
  targetName: string
  content: string
  createdAt: string
}

export type NotificationCategory = "contract" | "customer" | "system"

export interface NotificationItem {
  id: number
  key: string
  title: string
  summary: string
  content: string
  category: NotificationCategory
  createdAt: string
  unread: boolean
}
