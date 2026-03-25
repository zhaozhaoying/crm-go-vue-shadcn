<script setup lang="ts">
import NavMain from '@/components/custom/NavMain.vue'
import NavUser from '@/components/custom/NavUser.vue'
import { Button } from '@/components/ui/button'
import { hasAnyRole, isAdminUser } from '@/lib/auth-role'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notification'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from '@/components/ui/sidebar'
import { Globe2, LayoutDashboard, Users, UserCog, Shield, Settings, ClipboardList, Headphones, FileText, MapPinned, Bell, Trophy, PhoneCall } from 'lucide-vue-next'
import { computed } from 'vue'
import { useRouter } from 'vue-router'

interface MainNavigationItem {
  title: string
  url: string
  icon: any
  allowedRoles?: string[]
  items?: { title: string; url: string; allowedRoles?: string[] }[]
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
      {
        title: '客户分配',
        url: '/custom/customer-assignments',
        allowedRoles: ['admin', 'finance_manager', 'finance', '财务经理', '财务'],
      },
      { title: '上门拜访', url: '/custom/visits' },
    ],
  },
  { title: '销售跟进', url: '/follow-records/sales', icon: ClipboardList, allowedRoles: ['admin'] },
  { title: '运营跟进', url: '/follow-records/operation', icon: Headphones, allowedRoles: ['admin'] },
  {
    title: '每日排名',
    url: '/sales-daily-scores',
    icon: Trophy,
  },
  {
    title: '通话录音',
    url: '/call-recordings',
    icon: PhoneCall,
    allowedRoles: [
      'admin',
      'finance_manager',
      'finance',
      '财务经理',
      '财务',
      'sales_director',
      '销售总监',
      'sales_manager',
      '销售经理',
      'sales_staff',
      '销售员工',
      'sales_inside',
      'sale_inside',
      'inside销售',
      '电销员工',
      'sales_outside',
      'sale_outside',
      'outside销售',
    ],
  },
  { title: '合同管理', url: '/contracts', icon: FileText },
  { title: '地图资源', url: '/resource-pool', icon: MapPinned },
  { title: '资源获取', url: '/resource-acquisition', icon: Globe2 },
]

const authStore = useAuthStore()
const notificationStore = useNotificationStore()
const router = useRouter()
const isAdmin = computed(() => isAdminUser(authStore.user))
const visibleNavMain = computed(() =>
  navMain
    .filter((item) => !item.allowedRoles?.length || hasAnyRole(authStore.user, item.allowedRoles))
    .map((item) => ({
      ...item,
      items: item.items?.filter(
        (subItem) => !subItem.allowedRoles?.length || hasAnyRole(authStore.user, subItem.allowedRoles),
      ),
    })),
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
        <Button
          variant="ghost"
          size="icon"
          class="relative ml-auto h-8 w-8 shrink-0"
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
