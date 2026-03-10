<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from "vue"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Button } from "@/components/ui/button"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Loader2, Upload, Trash2, Camera, X } from "lucide-vue-next"
import { createUser, updateUser, uploadUserAvatar } from "@/api/modules/users"
import type { Role, UserWithRole } from "@/types/user"

const props = defineProps<{
  open: boolean
  mode: "create" | "edit"
  userData?: UserWithRole | null
  roles: Role[]
  users: UserWithRole[]
}>()

const emit = defineEmits<{
  (e: "update:open", value: boolean): void
  (e: "success"): void
}>()

const form = ref({
  username: "", password: "", nickname: "", email: "",
  mobile: "", roleId: "", parentId: "none", status: "enabled", avatar: ""
})
const formError = ref("")
const formSubmitting = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const previewUrl = ref("")
const selectedAvatarFile = ref<File | null>(null)
const localPreviewObjectUrl = ref("")

const clearLocalPreviewObjectUrl = () => {
  if (localPreviewObjectUrl.value) {
    URL.revokeObjectURL(localPreviewObjectUrl.value)
    localPreviewObjectUrl.value = ""
  }
}

watch(() => props.open, (val) => {
  if (!val) {
    selectedAvatarFile.value = null
    clearLocalPreviewObjectUrl()
    return
  }
  formError.value = ""
  previewUrl.value = ""
  selectedAvatarFile.value = null
  clearLocalPreviewObjectUrl()
  if (props.mode === "create") {
    form.value = {
      username: "", password: "", nickname: "", email: "", mobile: "",
      roleId: props.roles[0]?.id ? String(props.roles[0].id) : "",
      parentId: "none", status: "enabled", avatar: ""
    }
  } else if (props.userData) {
    form.value = {
      username: props.userData.username, password: "",
      nickname: props.userData.nickname || "",
      email: props.userData.email || "",
      mobile: props.userData.mobile || "",
      roleId: props.userData.roleId ? String(props.userData.roleId) : "",
      parentId: props.userData.parentId ? String(props.userData.parentId) : "none",
      status: props.userData.status || "enabled",
      avatar: props.userData.avatar || ""
    }
    previewUrl.value = form.value.avatar ||
      `https://api.dicebear.com/7.x/notionists/svg?seed=${form.value.username}&backgroundColor=ffffff`
  }
})

const close = () => { emit("update:open", false) }

const parentOptions = () => {
  if (props.mode === "create" || !props.userData) return props.users
  return props.users.filter((u) => u.id !== props.userData!.id)
}

const triggerUpload = () => { fileInput.value?.click() }

const onFileSelected = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (!target.files?.length) return
  const file = target.files[0]
  if (!file.type.startsWith("image/")) { formError.value = "请上传图片格式的文件"; return }
  if (file.size > 2 * 1024 * 1024) { formError.value = "图片大小不能超过 2MB"; return }

  formError.value = ""
  selectedAvatarFile.value = file
  clearLocalPreviewObjectUrl()
  localPreviewObjectUrl.value = URL.createObjectURL(file)
  previewUrl.value = localPreviewObjectUrl.value
  target.value = ""
}

const removeAvatar = () => {
  selectedAvatarFile.value = null
  clearLocalPreviewObjectUrl()
  previewUrl.value = ""
  form.value.avatar = ""
}

const handleSubmit = async () => {
  formError.value = ""
  formSubmitting.value = true
  try {
    const parentId = form.value.parentId && form.value.parentId !== "none" ? Number(form.value.parentId) : null
    let avatarUrl = form.value.avatar

    if (selectedAvatarFile.value) {
      const uploadResult = await uploadUserAvatar(selectedAvatarFile.value)
      avatarUrl = uploadResult.url
    }

    if (props.mode === "create") {
      if (!form.value.username || !form.value.password) { formError.value = "请填写用户名和密码"; return }
      await createUser({
        username: form.value.username, password: form.value.password,
        nickname: form.value.nickname, email: form.value.email,
        mobile: form.value.mobile, roleId: Number(form.value.roleId),
        parentId: parentId,
        avatar: avatarUrl
      } as any)
    } else if (props.userData) {
      await updateUser(props.userData.id, {
        username: form.value.username, nickname: form.value.nickname,
        email: form.value.email, mobile: form.value.mobile,
        roleId: Number(form.value.roleId),
        parentId: parentId,
        status: form.value.status, avatar: avatarUrl
      } as any)
    }
    emit("success")
    close()
  } catch (e) {
    formError.value = e instanceof Error ? e.message : "操作失败"
  } finally {
    formSubmitting.value = false
  }
}

onBeforeUnmount(() => {
  clearLocalPreviewObjectUrl()
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center">
        <!-- 遮罩层 -->
        <div class="absolute inset-0 bg-black/60" @click="close" />

        <!-- 弹窗内容 -->
        <div class="relative z-10 w-full max-w-[600px] mx-4 rounded-xl bg-white shadow-2xl">
          <!-- 关闭按钮 -->
          <button
            type="button"
            @click="close"
            class="absolute right-4 top-4 z-20 flex h-8 w-8 items-center justify-center rounded-full bg-slate-100 text-slate-500 transition hover:bg-slate-200 hover:text-slate-700"
          >
            <X class="h-4 w-4" />
          </button>

          <!-- 头部 -->
          <div class="px-6 py-5 border-b border-slate-100 bg-slate-50/50 rounded-t-xl pr-14">
            <h2 class="text-lg font-semibold text-slate-900">
              {{ mode === "create" ? "添加" : "编辑" }}
            </h2>
          </div>

          <form @submit.prevent="handleSubmit">
            <div class="p-6 space-y-6 max-h-[60vh] overflow-y-auto">
              <div v-if="formError" class="rounded-lg bg-red-50 p-3 text-sm text-red-600 border border-red-100">
                {{ formError }}
              </div>

              <!-- 头像 -->
              <div class="flex flex-col items-center space-y-3 pb-4 border-b border-slate-100">
                <input type="file" ref="fileInput" class="hidden" accept="image/png,image/jpeg,image/webp" @change="onFileSelected" />
                <div class="relative group cursor-pointer" @click="triggerUpload">
                  <Avatar class="h-20 w-20 border-[3px] border-white shadow-md">
                    <AvatarImage :src="previewUrl" class="object-cover" />
                    <AvatarFallback class="bg-indigo-50 text-indigo-700 text-xl font-bold">
                      {{ form.nickname ? form.nickname.charAt(0) : form.username ? form.username.charAt(0).toUpperCase() : "U" }}
                    </AvatarFallback>
                  </Avatar>
                  <div class="absolute inset-0 bg-black/40 rounded-full opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity">
                    <Camera class="h-6 w-6 text-white" />
                  </div>
                </div>
                <div class="flex gap-2">
                  <Button type="button" variant="outline" size="sm" class="h-7 text-xs" @click="triggerUpload">
                    <Upload class="h-3 w-3 mr-1.5" /> 上传头像
                  </Button>
                  <Button v-if="previewUrl && !previewUrl.includes('dicebear')" type="button" variant="ghost" size="sm" class="h-7 text-xs text-red-500 hover:text-red-600 hover:bg-red-50" @click="removeAvatar">
                    <Trash2 class="h-3 w-3 mr-1.5" /> 移除
                  </Button>
                </div>
                <p class="text-xs text-slate-400">支持 JPG, PNG, WEBP 格式，最大 2MB</p>
              </div>

              <!-- 表单 -->
              <div class="grid grid-cols-2 gap-x-6 gap-y-5">
                <div class="space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">登录账号 <span class="text-red-500">*</span></Label>
                  <Input v-model="form.username" required placeholder="如: zhangsan" class="h-10" :disabled="mode === 'edit'" />
                </div>
                <div class="space-y-1.5" v-if="mode === 'create'">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">初始密码 <span class="text-red-500">*</span></Label>
                  <Input v-model="form.password" type="password" required placeholder="至少6位" class="h-10" />
                </div>
                <div class="space-y-1.5" v-if="mode === 'edit'">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">账号状态</Label>
                  <Select v-model="form.status">
                    <SelectTrigger class="h-10"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="enabled">启用 (允许登录)</SelectItem>
                        <SelectItem value="disabled">禁用 (禁止登录)</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                <div class="space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">用户昵称</Label>
                  <Input v-model="form.nickname" placeholder="如: 张三" class="h-10" />
                </div>
                <div class="space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">系统角色 <span class="text-red-500">*</span></Label>
                  <Select v-model="form.roleId">
                    <SelectTrigger class="h-10"><SelectValue placeholder="选择角色" /></SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem v-for="r in roles" :key="r.id" :value="String(r.id)">{{ r.label }}</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                <div class="space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">电子邮箱</Label>
                  <Input v-model="form.email" type="email" placeholder="example@company.com" class="h-10" />
                </div>
                <div class="space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">手机号码</Label>
                  <Input v-model="form.mobile" placeholder="138xxxx8888" class="h-10" />
                </div>
                <div class="col-span-2 space-y-1.5 pt-4 border-t border-slate-100">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">汇报上级 (可选)</Label>
                  <Select v-model="form.parentId">
                    <SelectTrigger class="h-10"><SelectValue placeholder="无（顶层节点）" /></SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="none">无（顶层节点）</SelectItem>
                        <SelectItem v-for="u in parentOptions()" :key="u.id" :value="String(u.id)">
                          {{ u.nickname || u.username }}
                        </SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>

            <!-- 底部按钮 -->
            <div class="px-6 py-4 border-t border-slate-100 bg-slate-50/50 flex justify-end gap-3 rounded-b-xl">
              <Button type="button" variant="outline" @click="close" class="h-10 px-5">取消</Button>
              <Button type="submit" :disabled="formSubmitting" class="h-10 px-5 bg-indigo-600 hover:bg-indigo-700 text-white">
                <Loader2 v-if="formSubmitting" class="mr-2 h-4 w-4 animate-spin" />
                {{ formSubmitting ? "保存中..." : "保存" }}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active, .modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from, .modal-leave-to {
  opacity: 0;
}
</style>
