<script setup lang="ts">
import NavMain from '@/components/custom/NavMain.vue'
import NavUser from '@/components/custom/NavUser.vue'
import { hasAnyRole, isAdminUser } from '@/lib/auth-role'
import { useAuthStore } from '@/stores/auth'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from '@/components/ui/sidebar'
import { Globe2, LayoutDashboard, Users, UserCog, Shield, Settings, ClipboardList, Headphones, FileText, MapPinned } from 'lucide-vue-next'
import { computed } from 'vue'

interface MainNavigationItem {
  title: string
  url: string
  icon: any
  allowedRoles?: string[]
  items?: { title: string; url: string }[]
}

const navMain: MainNavigationItem[] = [
  { title: '仪表盘', url: '/dashboard', icon: LayoutDashboard },
  { title: '角色管理', url: '/roles', icon: Shield, allowedRoles: ['admin'] },
  { title: '用户管理', url: '/users', icon: UserCog, allowedRoles: ['admin', 'finance_manager', 'finance', '财务经理', '财务'] },
  {
    title: '客户管理',
    url: '/customers',
    icon: Users,
    items: [
      { title: '我的客户', url: '/customers/my' },
      { title: '公海客户', url: '/customers/pool' },
      { title: '查找客户', url: '/customers/search' },
      { title: '潜在客户', url: '/customers/potential' },
      { title: '合作客户', url: '/customers/partner' },
    ],
  },
  { title: '销售跟进', url: '/follow-records/sales', icon: ClipboardList, allowedRoles: ['admin'] },
  { title: '运营跟进', url: '/follow-records/operation', icon: Headphones, allowedRoles: ['admin'] },
  { title: '合同管理', url: '/contracts', icon: FileText },
  { title: '地图资源', url: '/resource-pool', icon: MapPinned },
  { title: '资源获取', url: '/resource-acquisition', icon: Globe2 },
]

const authStore = useAuthStore()
const isAdmin = computed(() => isAdminUser(authStore.user))
const visibleNavMain = computed(() =>
  navMain.filter((item) => !item.allowedRoles?.length || hasAnyRole(authStore.user, item.allowedRoles)),
)
</script>

<template>
  <Sidebar variant="inset" collapsible="icon">
    <SidebarHeader>
      <div class="flex items-center gap-2 px-2 py-1">
        <div class="flex h-6 w-6 items-center justify-center rounded-md">
          <img src="@/assets/favicon.ico" alt="">
        </div>
        <span class="font-semibold truncate group-data-[collapsible=icon]:hidden">招招营 CRM</span>
      </div>
    </SidebarHeader>
    <SidebarContent>
      <NavMain :items="visibleNavMain" />
    </SidebarContent>
    <SidebarFooter>
      <SidebarMenu>
        <SidebarMenuItem v-if="isAdmin">
          <SidebarMenuButton as-child>
            <a href="/settings" class="cursor-pointer">
              <Settings class="h-4 w-4" />
              <span>系统设置</span>
            </a>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
      <NavUser />
    </SidebarFooter>
  </Sidebar>
</template>
