<script setup lang="ts">
import { computed, onMounted } from "vue"
import { useRouter } from "vue-router"

import { Button } from "@/components/ui/button"
import { useNotificationStore } from "@/stores/notification"

const router = useRouter()
const notificationStore = useNotificationStore()

const notifications = computed(() => notificationStore.orderedNotifications)

onMounted(() => {
  notificationStore.fetchNotifications()
})

const openDetail = (id: number) => {
  notificationStore.markAsRead(id)
  router.push(`/notifications/${id}`)
}

const getCategoryLabel = (category: string) => {
  switch (category) {
    case "contract":
      return "合同"
    case "customer":
      return "客户"
    default:
      return "系统"
  }
}

const formatRelativeTime = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return "-"

  const diff = Date.now() - date.getTime()
  if (diff < 60_000) return "刚刚"

  const minutes = Math.floor(diff / 60_000)
  if (minutes < 60) return `${minutes} 分钟前`

  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} 小时前`

  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`

  return date.toLocaleDateString("zh-CN", { month: "2-digit", day: "2-digit" })
}
</script>

<template>
  <section class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h2 class="text-xl font-semibold text-card-foreground">通知中心</h2>
        <p class="mt-1 text-sm text-muted-foreground">
          全部操作日志都在这里，点击可查看完整详情。
        </p>
      </div>
      <Button
        variant="outline"
        :disabled="notificationStore.unreadCount === 0"
        @click="notificationStore.markAllAsRead()"
      >
        全部标记已读
      </Button>
    </div>

    <div class="overflow-hidden rounded-xl border bg-card shadow-sm">
      <div v-if="notificationStore.loading" class="px-4 py-10 text-center text-sm text-muted-foreground">
        加载中...
      </div>
      <div v-else-if="notifications.length === 0" class="px-4 py-10 text-center text-sm text-muted-foreground">
        暂无通知
      </div>
      <ul v-else class="divide-y">
        <li v-for="item in notifications" :key="item.id">
          <button
            class="w-full cursor-pointer px-4 py-4 text-left transition-colors hover:bg-accent/40"
            @click="openDetail(item.id)"
          >
            <div class="flex items-start gap-3">
              <span
                class="mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full"
                :class="item.unread ? 'bg-primary' : 'bg-muted'"
              />
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                  <p class="truncate text-sm font-semibold text-foreground">{{ item.title }}</p>
                  <span class="rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground">
                    {{ getCategoryLabel(item.category) }}
                  </span>
                </div>
                <p class="mt-1 text-sm text-muted-foreground">{{ item.summary }}</p>
              </div>
              <span class="shrink-0 text-xs text-muted-foreground">{{ formatRelativeTime(item.createdAt) }}</span>
            </div>
          </button>
        </li>
      </ul>
    </div>
  </section>
</template>
