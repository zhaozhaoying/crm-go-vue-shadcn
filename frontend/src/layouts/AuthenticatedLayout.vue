<script setup lang="ts">
import AppBreadcrumb from '@/components/custom/AppBreadcrumb.vue'
import AppCommand from '@/components/custom/AppCommand.vue'
import AppSidebar from '@/components/custom/AppSidebar.vue'
import { Separator } from '@/components/ui/separator'
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from '@/components/ui/sidebar'
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
