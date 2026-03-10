<script setup lang="ts">
import { ref, computed, onMounted } from "vue"
import {
  Plus,
  Search,
  Loader2,
  CircleCheck,
  CircleX,
  Ellipsis,
} from "lucide-vue-next"

import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Checkbox } from "@/components/ui/checkbox"

import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table"
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuSeparator, DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Pagination } from "@/components/ui/pagination"

import {
  listUsers, deleteUser, updateUser, listRoles, batchDisableUsers,
} from "@/api/modules/users"
import type { UserWithRole, Role } from "@/types/user"
import { getRequestErrorMessage } from "@/lib/http-error"
import { toast } from "vue-sonner"

import PopupForm from "./popupForm.vue"
import ConfirmDialog from "@/components/custom/ConfirmDialog.vue"

const confirmDialog = ref<InstanceType<typeof ConfirmDialog> | null>(null)

const users = ref<UserWithRole[]>([])
const roles = ref<Role[]>([])
const loading = ref(false)
const batchDisabling = ref(false)
const searchQuery = ref("")

// Selection
const selectedIds = ref<number[]>([])

// Pagination
const pageIndex = ref(0)
const pageSize = ref(10)

// Dialog state
const showDialog = ref(false)
const dialogMode = ref<"create" | "edit">("create")
const editingUser = ref<UserWithRole | null>(null)


// Filtered data
const filteredUsers = computed(() => {
  if (!searchQuery.value) return users.value
  const q = searchQuery.value.toLowerCase()
  return users.value.filter(
    (u) =>
      u.username.toLowerCase().includes(q) ||
      (u.nickname && u.nickname.toLowerCase().includes(q)) ||
      (u.email && u.email.toLowerCase().includes(q)) ||
      (u.mobile && u.mobile.includes(q))
  )
})

// Paginated data
const totalPages = computed(() => Math.max(1, Math.ceil(filteredUsers.value.length / pageSize.value)))
const paginatedUsers = computed(() => {
  const start = pageIndex.value * pageSize.value
  return filteredUsers.value.slice(start, start + pageSize.value)
})

const allPageSelected = computed(() =>
  paginatedUsers.value.length > 0 && paginatedUsers.value.every((u) => selectedIds.value.includes(u.id))
)
const somePageSelected = computed(() =>
  paginatedUsers.value.some((u) => selectedIds.value.includes(u.id)) && !allPageSelected.value
)

const toggleAllPage = (val: boolean | 'indeterminate') => {
  const checked = val === true
  if (checked) {
    const ids = paginatedUsers.value.map((u) => u.id)
    selectedIds.value = [...new Set([...selectedIds.value, ...ids])]
  } else {
    const pageIds = new Set(paginatedUsers.value.map((u) => u.id))
    selectedIds.value = selectedIds.value.filter((id) => !pageIds.has(id))
  }
}

const toggleRow = (id: number, val: boolean | 'indeterminate') => {
  const checked = val === true
  if (checked) {
    if (!selectedIds.value.includes(id)) {
      selectedIds.value = [...selectedIds.value, id]
    }
  } else {
    selectedIds.value = selectedIds.value.filter((i) => i !== id)
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const [u, r] = await Promise.all([listUsers(), listRoles()])
    users.value = u || []
    roles.value = r || []
    const validIds = new Set(users.value.map((item) => item.id))
    selectedIds.value = selectedIds.value.filter((id) => validIds.has(id))
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  dialogMode.value = "create"
  editingUser.value = null
  showDialog.value = true
}

const openEdit = (user: UserWithRole) => {
  dialogMode.value = "edit"
  editingUser.value = user
  showDialog.value = true
}


const handleDelete = async (user: UserWithRole) => {
  const confirmed = await confirmDialog.value?.open({
    title: "删除用户",
    description: `确定要删除用户「${user.nickname || user.username}」吗？此操作不可撤销。`,
    confirmText: "删除",
    variant: "danger",
  })
  if (!confirmed) return
  try {
    await deleteUser(user.id)
    await fetchData()
    toast.success("删除成功")
  } catch (e) {
    toast.error(getRequestErrorMessage(e, "删除失败"))
  }
}

const toggleStatus = async (user: UserWithRole) => {
  const newStatus = user.status === "enabled" ? "disabled" : "enabled"
  const isDisabling = newStatus === "disabled"
  const confirmed = await confirmDialog.value?.open({
    title: isDisabling ? "禁用用户" : "启用用户",
    description: isDisabling
      ? `禁用后，用户「${user.nickname || user.username}」将无法登录系统。`
      : `确定要重新启用用户「${user.nickname || user.username}」吗？`,
    confirmText: isDisabling ? "禁用" : "启用",
    variant: isDisabling ? "warning" : "info",
  })
  if (!confirmed) return
  try {
    await updateUser(user.id, {
      username: user.username,
      nickname: user.nickname,
      email: user.email,
      mobile: user.mobile,
      roleId: user.roleId,
      parentId: user.parentId,
      status: newStatus,
    })
    await fetchData()
    toast.success(isDisabling ? "用户已禁用" : "用户已启用")
  } catch (e) {
    toast.error(getRequestErrorMessage(e, "操作失败"))
  }
}

const disableableSelectedIds = computed(() =>
  users.value
    .filter((u) => selectedIds.value.includes(u.id) && u.status !== "disabled")
    .map((u) => u.id),
)

const handleBatchDisable = async () => {
  if (selectedIds.value.length === 0 || disableableSelectedIds.value.length === 0) return
  const total = selectedIds.value.length
  const toDisable = disableableSelectedIds.value.length
  const alreadyDisabled = total - toDisable
  const description = alreadyDisabled > 0
    ? `已选择 ${total} 个用户，其中 ${alreadyDisabled} 个已禁用，将禁用其余 ${toDisable} 个用户。`
    : `确定禁用已选择的 ${toDisable} 个用户吗？禁用后将无法登录系统。`

  const confirmed = await confirmDialog.value?.open({
    title: "批量禁用用户",
    description,
    confirmText: "批量禁用",
    variant: "warning",
  })
  if (!confirmed) return

  batchDisabling.value = true
  try {
    await batchDisableUsers(disableableSelectedIds.value)
    await fetchData()
    selectedIds.value = []
    toast.success("批量禁用成功")
  } catch (e) {
    toast.error(getRequestErrorMessage(e, "批量禁用失败"))
  } finally {
    batchDisabling.value = false
  }
}


onMounted(fetchData)
</script>

<template>
  <div class="w-full flex flex-col gap-4 lg:gap-6">
    <!-- Toolbar: 左边按钮 + 右边搜索 -->
    <div class="flex items-center justify-between px-4 lg:px-6">
      <div class="flex items-center gap-2">
        <Button size="sm" @click="openCreate">
          <Plus class="h-4 w-4" />
          <span>添加</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          class="border-destructive/40 text-destructive hover:bg-destructive/10 hover:text-destructive"
          :disabled="loading || batchDisabling || disableableSelectedIds.length === 0"
          @click="handleBatchDisable"
        >
          <Loader2 v-if="batchDisabling" class="h-4 w-4 animate-spin" />
          <span>{{ batchDisabling ? "禁用中" : `批量禁用${selectedIds.length ? `(${selectedIds.length})` : ""}` }}</span>
        </Button>
      </div>

      <div class="relative w-full max-w-sm">
        <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input v-model="searchQuery" type="search" placeholder="搜索用户..." class="pl-8 h-9" @input="pageIndex = 0" />
      </div>
    </div>

    <!-- Users list -->
    <div class="relative flex flex-col gap-4 px-4 lg:px-6">

      <!-- Table -->
      <div class="overflow-hidden rounded-lg border">
        <div v-if="loading" class="flex items-center justify-center py-24">
          <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
        </div>

        <Table v-else>
          <TableHeader class="bg-muted/50 sticky top-0 z-10">
            <TableRow>
              <TableHead class="w-12">
                <div class="flex items-center justify-center">
                  <Checkbox :checked="allPageSelected || (somePageSelected && 'indeterminate')"
                    class="border-black/70 data-[state=checked]:border-black data-[state=checked]:bg-black data-[state=checked]:text-white data-[state=indeterminate]:border-black data-[state=indeterminate]:bg-black data-[state=indeterminate]:text-white focus-visible:ring-black/30"
                    @update:checked="toggleAllPage" aria-label="全选" />
                </div>
              </TableHead>
              <TableHead class="w-16">编号</TableHead>
              <TableHead>用户信息</TableHead>
              <TableHead>系统角色</TableHead>
              <TableHead>联系方式</TableHead>
              <TableHead>状态</TableHead>
              <TableHead class="w-12" />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="user in paginatedUsers" :key="user.id"
              :data-state="selectedIds.includes(user.id) ? 'selected' : undefined">
              <!-- Checkbox -->
              <TableCell>
                <div class="flex items-center justify-center">
                  <Checkbox :checked="selectedIds.includes(user.id)"
                    class="border-black/70 data-[state=checked]:border-black data-[state=checked]:bg-black data-[state=checked]:text-white data-[state=indeterminate]:border-black data-[state=indeterminate]:bg-black data-[state=indeterminate]:text-white focus-visible:ring-black/30"
                    @update:checked="(val: any) => toggleRow(user.id, val)" aria-label="选择行" />
                </div>
              </TableCell>

              <!-- ID -->
              <TableCell class="text-muted-foreground">
                {{ user.id }}
              </TableCell>

              <!-- User info -->
              <TableCell>
                <div class="flex items-center gap-3 text-left">
                  <Avatar class="h-8 w-8 shrink-0">
                    <AvatarImage :src="user.avatar" class="object-cover" />
                    <AvatarFallback class="bg-primary/10 text-primary text-xs font-semibold">
                      {{ (user.nickname || user.username).charAt(0).toUpperCase() }}
                    </AvatarFallback>
                  </Avatar>
                  <div class="min-w-0">
                    <p class="text-sm font-medium truncate">{{ user.nickname || user.username }}</p>
                    <p class="text-xs text-muted-foreground truncate">{{ user.username }}</p>
                  </div>
                </div>
              </TableCell>

              <!-- Role -->
              <TableCell>
                <Badge variant="outline" class="text-muted-foreground">
                  {{ user.roleLabel || user.roleName }}
                </Badge>
              </TableCell>

              <!-- Contact -->
              <TableCell>
                <div class="text-sm text-muted-foreground">
                  <p v-if="user.email" class="truncate max-w-[180px]">{{ user.email }}</p>
                  <p v-if="user.mobile" class="truncate max-w-[180px]">{{ user.mobile }}</p>
                  <p v-if="!user.email && !user.mobile" class="italic">—</p>
                </div>
              </TableCell>

              <!-- Status -->
              <TableCell>
                <Badge variant="outline" class="text-muted-foreground gap-1">
                  <CircleCheck v-if="user.status === 'enabled'" class="h-3.5 w-3.5 fill-emerald-500 text-white" />
                  <CircleX v-else class="h-3.5 w-3.5 fill-muted-foreground/40 text-white" />
                  {{ user.status === "enabled" ? "正常" : "已禁用" }}
                </Badge>
              </TableCell>

              <!-- Actions -->
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground">
                      <Ellipsis class="h-4 w-4" />
                      <span class="sr-only">操作菜单</span>
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" class="w-36">
                    <DropdownMenuItem @click="openEdit(user)">编辑</DropdownMenuItem>
                    <DropdownMenuItem @click="toggleStatus(user)">
                      {{ user.status === 'enabled' ? '禁用' : '启用' }}
                    </DropdownMenuItem>
                    <template v-if="user.roleName !== 'admin'">
                      <DropdownMenuSeparator />
                      <DropdownMenuItem class="text-destructive" @click="handleDelete(user)">删除</DropdownMenuItem>
                    </template>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>

            <!-- Empty state -->
            <TableRow v-if="!loading && paginatedUsers.length === 0">
              <TableCell :colspan="7" class="h-24 text-center text-muted-foreground">
                暂无数据
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>

      <!-- Pagination -->
      <Pagination :current-page="pageIndex" :total-pages="totalPages" :page-size="pageSize"
        :selected-count="selectedIds.length" :total-count="filteredUsers.length"
        @update:current-page="pageIndex = $event" @update:page-size="pageSize = $event" />
    </div>
  </div>

  <PopupForm v-model:open="showDialog" :mode="dialogMode" :userData="editingUser" :roles="roles" :users="users"
    @success="fetchData" />

  <ConfirmDialog ref="confirmDialog" />

</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
