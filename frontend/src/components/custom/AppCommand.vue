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
import { useRouter } from 'vue-router'
import { LayoutDashboard, Users, UserCog, Bell } from 'lucide-vue-next'
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
  { title: '通知中心', href: '/notifications', icon: Bell },
]

const isOpen = ref(false)
const router = useRouter()
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
