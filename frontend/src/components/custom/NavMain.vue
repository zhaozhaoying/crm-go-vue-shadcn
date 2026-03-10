<script setup lang="ts">
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from '@/components/ui/sidebar'
import { ChevronRight } from 'lucide-vue-next'
import { useRoute } from 'vue-router'
import { computed } from 'vue'

interface NavItem {
  title: string
  url: string
  icon: any
  isActive?: boolean
  items?: { title: string; url: string }[]
}

interface Props {
  items: NavItem[]
}

defineProps<Props>()

const currentRoute = useRoute()
const currentPath = computed(() => currentRoute.path)

const isActive = (url: string) => {
  return currentPath.value === url || currentPath.value.startsWith(`${url}/`)
}
</script>

<template>
  <SidebarGroup>
    <SidebarGroupLabel>主导航</SidebarGroupLabel>
    <SidebarMenu>
      <template v-for="mainItem in items" :key="mainItem.url">
        <Collapsible :defaultOpen="isActive(mainItem.url)" class="group/collapsible">
          <SidebarMenuItem>
            <template v-if="mainItem.items">
              <CollapsibleTrigger as-child>
                <SidebarMenuButton
                  :is-active="isActive(mainItem.url)"
                  class="font-medium"
                >
                  <component v-if="mainItem.icon" :is="mainItem.icon" />
                  <span>{{ mainItem.title }}</span>
                  <ChevronRight class="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                </SidebarMenuButton>
              </CollapsibleTrigger>
              <CollapsibleContent>
                <SidebarMenuSub>
                  <template v-for="subItem in mainItem.items" :key="subItem.url">
                    <SidebarMenuSubItem>
                      <SidebarMenuSubButton as-child :is-active="isActive(subItem.url)">
                        <RouterLink :to="subItem.url">
                          <span>{{ subItem.title }}</span>
                        </RouterLink>
                      </SidebarMenuSubButton>
                    </SidebarMenuSubItem>
                  </template>
                </SidebarMenuSub>
              </CollapsibleContent>
            </template>
            <template v-else>
              <SidebarMenuButton
                as-child
                :is-active="isActive(mainItem.url)"
                class="font-medium"
              >
                <RouterLink :to="mainItem.url">
                  <component v-if="mainItem.icon" :is="mainItem.icon" />
                  <span>{{ mainItem.title }}</span>
                </RouterLink>
              </SidebarMenuButton>
            </template>
          </SidebarMenuItem>
        </Collapsible>
      </template>
    </SidebarMenu>
  </SidebarGroup>
</template>
