<script setup lang="ts">
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notification'
import { resolveUserAvatar } from '@/lib/user-avatar'
import { useRouter } from 'vue-router'
import { computed, onMounted, ref, watch } from 'vue'
import {
  BadgeCheck,
  Bell,
  ChevronsUpDown,
  Loader2,
  LogOut,
} from 'lucide-vue-next'

const authStore = useAuthStore()
const notificationStore = useNotificationStore()
const router = useRouter()
const { isMobile } = useSidebar()

const displayName = computed(() =>
  authStore.user?.nickname || authStore.user?.username || '用户'
)

const avatarUrl = computed(() => resolveUserAvatar(authStore.user?.avatar))
const avatarLoaded = ref(false)
const avatarLoadFailed = ref(false)

const initials = computed(() =>
  displayName.value.slice(0, 2).toUpperCase()
)
const showAvatarLoading = computed(
  () => !!avatarUrl.value && !avatarLoaded.value && !avatarLoadFailed.value,
)

const roleLabel = computed(() => {
  if (authStore.user?.roleLabel) return authStore.user.roleLabel
  const r = authStore.user?.role || authStore.user?.roleName
  if (!r) return ''
  const map: Record<string, string> = {
    admin: '管理员',
    finance_manager: '财务经理',
    sales_director: '销售总监',
    sales_manager: '销售经理',
    sales_staff: '销售员工',
    sales_inside: 'Inside销售',
    sales_outside: 'Outside销售',
    ops_manager: '运营经理',
    ops_staff: '运营员工',
  }
  return map[r] || r
})

const handleLogout = async () => {
  await authStore.logout()
  window.location.href = '/login'
}

const handleAvatarLoad = () => {
  avatarLoaded.value = true
  avatarLoadFailed.value = false
}

const handleAvatarError = () => {
  avatarLoaded.value = false
  avatarLoadFailed.value = true
}

watch(
  avatarUrl,
  () => {
    avatarLoaded.value = false
    avatarLoadFailed.value = false
  },
  { immediate: true },
)

onMounted(async () => {
  if (!authStore.token) return
  await authStore.fetchCurrentUserProfile(true)
  notificationStore.fetchNotifications()
})
</script>

<template>
  <SidebarMenu>
    <SidebarMenuItem>
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <SidebarMenuButton
            size="lg"
            class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
          >
            <Avatar class="h-8 w-8 rounded-lg">
              <AvatarImage
                v-if="avatarUrl"
                :src="avatarUrl"
                :alt="displayName"
                @load="handleAvatarLoad"
                @error="handleAvatarError"
              />
              <AvatarFallback class="rounded-lg">
                <Loader2 v-if="showAvatarLoading" class="h-4 w-4 animate-spin text-muted-foreground" />
                <span v-else>{{ initials }}</span>
              </AvatarFallback>
            </Avatar>
            <div class="grid flex-1 text-left text-sm leading-tight">
              <span class="truncate font-semibold">{{ displayName }}</span>
              <span class="truncate text-xs">{{ roleLabel }}</span>
            </div>
            <ChevronsUpDown class="ml-auto size-4" />
          </SidebarMenuButton>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          class="w-[--bits-dropdown-menu-anchor-width] min-w-56 rounded-lg"
          :side="isMobile ? 'bottom' : 'right'"
          align="end"
          :sideOffset="4"
        >
          <DropdownMenuLabel class="p-0 font-normal">
            <div class="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
              <Avatar class="h-8 w-8 rounded-lg">
                <AvatarImage
                  v-if="avatarUrl"
                  :src="avatarUrl"
                  :alt="displayName"
                  @load="handleAvatarLoad"
                  @error="handleAvatarError"
                />
                <AvatarFallback class="rounded-lg">
                  <Loader2 v-if="showAvatarLoading" class="h-4 w-4 animate-spin text-muted-foreground" />
                  <span v-else>{{ initials }}</span>
                </AvatarFallback>
              </Avatar>
              <div class="grid flex-1 text-left text-sm leading-tight">
                <span class="truncate font-semibold">{{ displayName }}</span>
                <span class="truncate text-xs">{{ roleLabel }}</span>
              </div>
            </div>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem class="cursor-pointer" @click="router.push('/profile')">
              <BadgeCheck class="mr-2 h-4 w-4" />
              个人资料
            </DropdownMenuItem>
            <DropdownMenuItem class="cursor-pointer" @click="router.push('/notifications')">
              <div class="relative mr-2">
                <Bell class="h-4 w-4" />
                <span
                  v-if="notificationStore.unreadCount > 0"
                  class="absolute -top-1 -right-1 h-2 w-2 rounded-full bg-red-500"
                />
              </div>
              通知中心
              <span
                v-if="notificationStore.unreadCount > 0"
                class="ml-auto rounded-full bg-red-500 px-1.5 py-0.5 text-[10px] font-medium leading-none text-white"
              >
                {{ notificationStore.unreadCount > 99 ? '99+' : notificationStore.unreadCount }}
              </span>
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuItem class="cursor-pointer" @click="handleLogout">
            <LogOut class="mr-2 h-4 w-4" />
            退出登录
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  </SidebarMenu>
</template>
