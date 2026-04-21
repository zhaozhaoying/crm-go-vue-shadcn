<script setup lang="ts">
import { CONTRACT_PENDING_COUNT_REFRESH_EVENT, getPendingContractCount } from '@/api/modules/contracts'
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
import { Globe2, LayoutDashboard, Users, UserCog, Settings, ClipboardList, FileText, Bell, Trophy } from 'lucide-vue-next'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

interface MainNavigationItem {
  title: string
  url: string
  icon: any
  badge?: string
  allowedRoles?: string[]
  items?: { title: string; url: string; allowedRoles?: string[] }[]
}

const navMain: MainNavigationItem[] = [
  { title: '仪表盘', url: '/dashboard', icon: LayoutDashboard },
  {
    title: '用户管理',
    url: '/users',
    icon: UserCog,
    allowedRoles: ['admin', 'finance_manager', 'finance', '财务经理', '财务'],
    items: [
      { title: '用户列表', url: '/users' },
      { title: '角色管理', url: '/users/roles', allowedRoles: ['admin'] },
    ],
  },
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
  {
    title: '跟进管理',
    url: '/follow-records/sales',
    icon: ClipboardList,
    items: [
      { title: '销售跟进', url: '/follow-records/sales', allowedRoles: ['admin'] },
      { title: '运营跟进', url: '/follow-records/operation', allowedRoles: ['admin'] },
    ],
  },
    { title: '合同管理', url: '/contracts', icon: FileText },
  {
    title: '每日排名',
    url: '/sales-daily-scores',
    icon: Trophy,
    items: [
      { title: '销售排名', url: '/sales-daily-scores' },
      { title: '电销排名', url: '/ranking-leaderboard' },
    ],
  },
  {
    title: '通话录音',
    url: '/call-recordings',
    icon: Trophy,
    items: [
      { title: '销售录音', url: '/call-recordings' },
      { title: '电销录音', url: '/telemarketing-recordings' },
    ],
  },
  {
    title: '资源获取',
    url: '/resource-acquisition',
    icon: Globe2,
    items: [
      { title: '地图资源', url: '/resource-pool' },
      { title: '资源获取', url: '/resource-acquisition' },
    ],
  },
]

const authStore = useAuthStore()
const notificationStore = useNotificationStore()
const route = useRoute()
const router = useRouter()
const pendingContractCount = ref(0)
const pendingContractCountLoaded = ref(false)
let pendingContractCountRequestId = 0

const fetchPendingContractCount = async () => {
  const requestId = ++pendingContractCountRequestId
  try {
    const total = await getPendingContractCount()
    if (requestId !== pendingContractCountRequestId) return
    pendingContractCount.value = Number.isFinite(total) ? total : 0
  } catch {
    if (requestId !== pendingContractCountRequestId) return
    pendingContractCount.value = 0
  } finally {
    if (requestId === pendingContractCountRequestId) {
      pendingContractCountLoaded.value = true
    }
  }
}

const pendingContractBadge = computed(() => {
  if (!pendingContractCountLoaded.value || pendingContractCount.value <= 0) return ''
  return pendingContractCount.value > 99 ? '99+' : String(pendingContractCount.value)
})

const handlePendingContractCountRefresh = () => {
  void fetchPendingContractCount()
}

const isAdmin = computed(() => isAdminUser(authStore.user))
const visibleNavMain = computed(() =>
  navMain
    .filter((item) => !item.allowedRoles?.length || hasAnyRole(authStore.user, item.allowedRoles))
    .map((item) => ({
      ...item,
      badge: item.url === '/contracts' ? pendingContractBadge.value : undefined,
      items: item.items?.filter(
        (subItem) => !subItem.allowedRoles?.length || hasAnyRole(authStore.user, subItem.allowedRoles),
      ),
    })),
)

watch(
  () => route.path,
  () => {
    void fetchPendingContractCount()
  },
)

onMounted(() => {
  void fetchPendingContractCount()
  window.addEventListener('focus', handlePendingContractCountRefresh)
  window.addEventListener(CONTRACT_PENDING_COUNT_REFRESH_EVENT, handlePendingContractCountRefresh)
})

onBeforeUnmount(() => {
  window.removeEventListener('focus', handlePendingContractCountRefresh)
  window.removeEventListener(CONTRACT_PENDING_COUNT_REFRESH_EVENT, handlePendingContractCountRefresh)
})
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
