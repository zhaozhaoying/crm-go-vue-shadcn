<script setup lang="ts">
import { computed, onMounted, watchEffect } from "vue"
import { useRoute, useRouter } from "vue-router"

import { Button } from "@/components/ui/button"
import { useNotificationStore } from "@/stores/notification"

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()

const notificationId = computed(() => Number(route.params.id))
const notification = computed(() => notificationStore.getById(notificationId.value))

onMounted(() => {
  if (notificationStore.notifications.length === 0) {
    notificationStore.fetchNotifications()
  }
})

watchEffect(() => {
  if (notification.value?.id) {
    notificationStore.markAsRead(notification.value.id)
  }
})

const goBack = () => {
  router.push("/notifications")
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
  <section class="space-y-4">
    <Button variant="ghost" class-name="px-0 text-muted-foreground hover:text-foreground" @click="goBack">
      ← 返回通知中心
    </Button>

    <div v-if="notification" class="rounded-xl border bg-card p-6 shadow-sm">
      <div class="border-b pb-4">
        <h2 class="text-xl font-semibold text-card-foreground">{{ notification.title }}</h2>
        <p class="mt-2 text-xs text-muted-foreground">
          {{ notification.createdAt }} · {{ formatRelativeTime(notification.createdAt) }}
        </p>
      </div>
      <article class="mt-5 whitespace-pre-line text-sm leading-7 text-foreground">
        {{ notification.content || notification.summary }}
      </article>
    </div>

    <div v-else class="rounded-xl border bg-card p-6 text-sm text-muted-foreground shadow-sm">
      通知不存在或已被删除。
    </div>
  </section>
</template>
