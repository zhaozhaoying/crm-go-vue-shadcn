<script setup lang="ts">
import AppBreadcrumb from '@/components/custom/AppBreadcrumb.vue'
import AppCommand from '@/components/custom/AppCommand.vue'
import AppSidebar from '@/components/custom/AppSidebar.vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import { useNotificationStore } from '@/stores/notification'
import { useRouter } from 'vue-router'
import { Bell } from 'lucide-vue-next'

const router = useRouter()
const notificationStore = useNotificationStore()
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    <SidebarInset class="min-w-0 overflow-x-hidden">
      <header
        class="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12"
      >
        <div class="flex items-center gap-2 px-4 flex-1">
          <SidebarTrigger class="-ml-1 h-4 w-4" />
          <Separator orientation="vertical" class="mr-2 h-4" />
          <AppBreadcrumb />
        </div>
        <div class="flex items-center gap-1 px-4">
          <Button
            variant="ghost"
            size="icon"
            class="relative h-8 w-8"
            @click="router.push('/notifications')"
          >
            <Bell class="h-4 w-4" />
            <span
              v-if="notificationStore.unreadCount > 0"
              class="absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-medium text-white"
            >
              {{ notificationStore.unreadCount > 99 ? '99+' : notificationStore.unreadCount }}
            </span>
          </Button>
        </div>
      </header>
      <div class="flex min-w-0 flex-1 flex-col gap-4 p-4 pt-0">
        <RouterView v-slot="{ Component, route }">
          <component :is="Component" :key="route.fullPath" />
        </RouterView>
      </div>
    </SidebarInset>
    <AppCommand />
  </SidebarProvider>
</template>
