<script setup lang="ts">
import { ref, onMounted } from "vue"
import { Loader2, Plus, Ellipsis } from "lucide-vue-next"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table"
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuSeparator, DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { listRoles, deleteRole } from "@/api/modules/users"
import type { Role } from "@/types/user"

import PopupForm from "./popupForm.vue"
import ConfirmDialog from "@/components/custom/ConfirmDialog.vue"

const confirmDialog = ref<InstanceType<typeof ConfirmDialog> | null>(null)

const roles = ref<Role[]>([])
const loading = ref(false)

const showDialog = ref(false)
const dialogMode = ref<"create" | "edit">("create")
const editingRole = ref<Role | null>(null)

const fetchRoles = async () => {
  loading.value = true
  try {
    roles.value = (await listRoles()) || []
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  dialogMode.value = "create"
  editingRole.value = null
  showDialog.value = true
}

const openEdit = (role: Role) => {
  dialogMode.value = "edit"
  editingRole.value = role
  showDialog.value = true
}

const handleDelete = async (role: Role) => {
  const confirmed = await confirmDialog.value?.open({
    title: "删除角色",
    description: `确定要删除角色「${role.label}」吗？此操作不可撤销。`,
    confirmText: "删除",
    variant: "danger",
  })
  if (!confirmed) return
  try {
    await deleteRole(role.id)
    await fetchRoles()
  } catch (e) {
    alert(e instanceof Error ? e.message : "删除失败")
  }
}

onMounted(fetchRoles)
</script>

<template>
  <div class="flex w-full flex-col gap-4 lg:gap-6">
    <div class="flex items-center justify-between px-4 lg:px-6">
      <div class="flex items-center gap-2">
        <Button size="sm" @click="openCreate">
          <Plus class="h-4 w-4" />
          <span>添加</span>
        </Button>
      </div>
    </div>

    <div class="px-4 lg:px-6">
      <div class="overflow-hidden rounded-lg border">
        <div v-if="loading" class="flex items-center justify-center py-24">
          <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
        </div>

        <Table v-else>
          <TableHeader class="bg-muted/50">
            <TableRow>
              <TableHead class="w-16">编号</TableHead>
              <TableHead>角色标识</TableHead>
              <TableHead>角色名称</TableHead>
              <TableHead class="w-20">排序</TableHead>
              <TableHead class="w-12" />
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="role in roles" :key="role.id">
              <TableCell class="text-muted-foreground">{{ role.id }}</TableCell>
              <TableCell>
                <Badge variant="outline" class="text-muted-foreground font-mono">{{ role.name }}</Badge>
              </TableCell>
              <TableCell class="font-medium">{{ role.label }}</TableCell>
              <TableCell class="text-muted-foreground">{{ role.sort }}</TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground">
                      <Ellipsis class="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" class="w-32">
                    <DropdownMenuItem @click="openEdit(role)">编辑</DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem class="text-destructive" @click="handleDelete(role)">删除</DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
            <TableRow v-if="roles.length === 0">
              <TableCell colspan="5" class="h-24 text-center text-muted-foreground">暂无角色数据</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </div>
  </div>

  <PopupForm v-model:open="showDialog" :mode="dialogMode" :roleData="editingRole" @success="fetchRoles" />

  <ConfirmDialog ref="confirmDialog" />
</template>
