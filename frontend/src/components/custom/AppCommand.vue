<script setup lang="ts">
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command'
import { hasAnyRole } from '@/lib/auth-role'
import { useAuthStore } from '@/stores/auth'
import { useRoute, useRouter } from 'vue-router'
import { LayoutDashboard, Users, UserCog, Bell, Trophy, Navigation, PhoneCall } from 'lucide-vue-next'
import { computed, onMounted, onUnmounted, ref } from 'vue'

interface NavigationItem {
  title: string
  href: string
  icon: typeof LayoutDashboard
  allowedRoles?: string[]
}

const navigationItems: NavigationItem[] = [
  { title: '仪表盘', href: '/dashboard', icon: LayoutDashboard },
  { title: '客户管理', href: '/customers', icon: Users },
  { title: '用户管理', href: '/users', icon: UserCog, allowedRoles: ['admin', 'finance_manager', 'finance', '财务经理', '财务'] },
  {
    title: '销售每日排名',
    href: '/sales-daily-scores',
    icon: Trophy,
  },
  {
    title: '电销每日排名',
    href: '/telemarketing-daily-scores',
    icon: Trophy,
  },
  {
    title: '排名榜单',
    href: '/ranking-leaderboard',
    icon: Trophy,
  },
  {
    title: '通话录音',
    href: '/call-recordings',
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
  {
    title: '客户分配',
    href: '/custom/customer-assignments',
    icon: Users,
    allowedRoles: ['admin', 'finance_manager', 'finance', '财务经理', '财务'],
  },
  { title: '上门拜访', href: '/custom/visits', icon: Navigation },
  { title: '通知中心', href: '/notifications', icon: Bell },
]

const isOpen = ref(false)
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const visibleNavigationItems = computed(() =>
  navigationItems.filter((item) => !item.allowedRoles?.length || hasAnyRole(authStore.user, item.allowedRoles)),
)

const onKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
    e.preventDefault()
    isOpen.value = !isOpen.value
  }
}

onMounted(() => document.addEventListener('keydown', onKeyDown))
onUnmounted(() => document.removeEventListener('keydown', onKeyDown))

const goTo = (href: string) => {
  isOpen.value = false
  if (href === '/dashboard' && route.path === '/dashboard') {
    window.dispatchEvent(new Event('dashboard:refresh'))
  }
  if (href === '/sales-daily-scores' && route.path === '/sales-daily-scores') {
    window.dispatchEvent(new Event('sales-daily-scores:refresh'))
  }
  if (href === '/telemarketing-daily-scores' && route.path === '/telemarketing-daily-scores') {
    window.dispatchEvent(new Event('telemarketing-daily-scores:refresh'))
  }
  router.push(href)
}
</script>

<template>
  <CommandDialog v-model:open="isOpen">
    <Command>
      <CommandInput placeholder="搜索命令..." />
      <CommandList>
        <CommandEmpty>未找到结果。</CommandEmpty>
        <CommandGroup heading="前往...">
          <CommandItem
            v-for="item in visibleNavigationItems"
            :key="item.href"
            :value="item.href"
            @select="() => goTo(item.href)"
          >
            <div class="flex items-center gap-2">
              <component :is="item.icon" class="size-4" />
              <span>{{ item.title }}</span>
            </div>
          </CommandItem>
        </CommandGroup>
      </CommandList>
    </Command>
  </CommandDialog>
</template>
