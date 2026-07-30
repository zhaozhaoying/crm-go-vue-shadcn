<script setup lang="ts">
import { ref } from 'vue'
import AppBreadcrumb from '@/components/custom/AppBreadcrumb.vue'
import AppCommand from '@/components/custom/AppCommand.vue'
import AppSidebar from '@/components/custom/AppSidebar.vue'
import AssessmentAnnouncementModal from '@/components/custom/AssessmentAnnouncementModal.vue'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import { Megaphone } from 'lucide-vue-next'

const announcementOpen = ref(false)
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
            size="sm"
            class="gap-1.5"
            @click="announcementOpen = true"
          >
            <Megaphone class="h-4 w-4" />
            考核公告
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
    <AssessmentAnnouncementModal v-model:open="announcementOpen" />
  </SidebarProvider>
</template>
