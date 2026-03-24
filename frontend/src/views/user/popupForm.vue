<script setup lang="ts">
import { ref, watch, onBeforeUnmount, computed } from "vue"
import { toTypedSchema } from "@vee-validate/zod"
import { useForm, useField } from "vee-validate"
import * as z from "zod"
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
import { requiredString } from "@/lib/form-validation"
import { DEFAULT_USER_AVATAR, resolveUserAvatar } from "@/lib/user-avatar"
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

const sanitizeHanghangCrmMobileInput = (raw: string) =>
  String(raw || "").replace(/\D/g, "").slice(0, 11)

const formSchema = computed(() => toTypedSchema(z.object({
  username: requiredString("登录账号"),
  password: props.mode === "create"
    ? requiredString("密码").min(6, { message: "密码至少需要6位" })
    : z.string().optional(),
  nickname: z.string().optional(),
  email: z.string().email({ message: "请输入有效的邮箱地址" }).optional().or(z.literal('')),
  mobile: requiredString("手机号码").regex(/^1\d{10}$/, { message: "手机号必须为11位数字" }),
  hanghangCrmMobile: z.string().optional().refine((value) => !value || /^1\d{10}$/.test(value), {
    message: "坐席手机号必须为11位数字",
  }),
  roleId: requiredString("系统角色"),
  parentId: z.string(),
  status: z.string(),
  avatar: z.string().optional(),
})))

const { handleSubmit, resetForm, errors } = useForm({
  validationSchema: formSchema,
})

const { value: username } = useField<string>('username')
const { value: password } = useField<string>('password')
const { value: nickname } = useField<string>('nickname')
const { value: email } = useField<string>('email')
const { value: mobile } = useField<string>('mobile')
const { value: hanghangCrmMobile } = useField<string>('hanghangCrmMobile')
const { value: roleId } = useField<string>('roleId')
const { value: parentId } = useField<string>('parentId')
const { value: status } = useField<string>('status')
const { value: avatar } = useField<string>('avatar')

const formSubmitting = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const previewUrl = ref("")
const selectedAvatarFile = ref<File | null>(null)
const localPreviewObjectUrl = ref("")

watch(hanghangCrmMobile, (value) => {
  const sanitized = sanitizeHanghangCrmMobileInput(value)
  if (sanitized !== value) {
    hanghangCrmMobile.value = sanitized
  }
})

const clearLocalPreviewObjectUrl = () => {
  if (localPreviewObjectUrl.value) {
    URL.revokeObjectURL(localPreviewObjectUrl.value)
    localPreviewObjectUrl.value = ""
  }
}

const initialValues = computed(() => {
  if (props.mode === "create") {
    return {
      username: "", password: "", nickname: "", email: "", mobile: "",
      hanghangCrmMobile: "",
      roleId: props.roles[0]?.id ? String(props.roles[0].id) : "",
      parentId: "none", status: "enabled", avatar: DEFAULT_USER_AVATAR
    }
  } else if (props.userData) {
    return {
      username: props.userData.username, password: "",
      nickname: props.userData.nickname || "",
      email: props.userData.email || "",
      mobile: props.userData.mobile || "",
      hanghangCrmMobile: props.userData.hanghangCrmMobile || "",
      roleId: props.userData.roleId ? String(props.userData.roleId) : "",
      parentId: props.userData.parentId ? String(props.userData.parentId) : "none",
      status: props.userData.status || "enabled",
      avatar: resolveUserAvatar(props.userData.avatar)
    }
  }
  return {}
})

watch(() => props.open, (val) => {
  if (!val) {
    selectedAvatarFile.value = null
    clearLocalPreviewObjectUrl()
    return
  }
  resetForm({ values: initialValues.value })
  previewUrl.value = resolveUserAvatar(initialValues.value.avatar)
  selectedAvatarFile.value = null
  clearLocalPreviewObjectUrl()
  if (props.mode === "edit" && props.userData) {
    previewUrl.value = resolveUserAvatar(props.userData.avatar)
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
  if (!file.type.startsWith("image/")) {
    // TODO: show error
    return
  }
  if (file.size > 20 * 1024 * 1024) {
    // TODO: show error
    return
  }

  selectedAvatarFile.value = file
  clearLocalPreviewObjectUrl()
  localPreviewObjectUrl.value = URL.createObjectURL(file)
  previewUrl.value = localPreviewObjectUrl.value
  target.value = ""
  avatar.value = ""
}

const removeAvatar = () => {
  selectedAvatarFile.value = null
  clearLocalPreviewObjectUrl()
  previewUrl.value = DEFAULT_USER_AVATAR
  avatar.value = DEFAULT_USER_AVATAR
}

const onSubmit = handleSubmit(async (values) => {
  formSubmitting.value = true
  try {
    const roleIdValue = Number(values.roleId)
    const parentIdValue = values.parentId && values.parentId !== "none" ? Number(values.parentId) : null
    let avatarUrl = resolveUserAvatar(values.avatar)

    if (selectedAvatarFile.value) {
      const uploadResult = await uploadUserAvatar(selectedAvatarFile.value)
      avatarUrl = uploadResult.url
    }

    const payload = {
      ...values,
      hanghangCrmMobile: values.hanghangCrmMobile,
      roleId: roleIdValue,
      parentId: parentIdValue,
      avatar: avatarUrl,
    }

    if (props.mode === "create") {
      await createUser(payload as any)
    } else if (props.userData) {
      await updateUser(props.userData.id, payload as any)
    }
    emit("success")
    close()
  } catch (e) {
    // TODO: show error
  } finally {
    formSubmitting.value = false
  }
})

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

          <form @submit="onSubmit">
            <div class="p-6 space-y-6 max-h-[60vh] overflow-y-auto">
              <!-- 头像 -->
              <div class="flex flex-col items-center space-y-3 pb-4 border-b border-slate-100">
                <input type="file" ref="fileInput" class="hidden" accept="image/png,image/jpeg,image/webp" @change="onFileSelected" />
                <div class="relative group cursor-pointer" @click="triggerUpload">
                  <Avatar class="h-20 w-20 border-[3px] border-white shadow-md">
                    <AvatarImage :src="previewUrl" class="object-cover" />
                    <AvatarFallback class="bg-indigo-50 text-indigo-700 text-xl font-bold">
                      {{ nickname ? nickname.charAt(0) : username ? username.charAt(0).toUpperCase() : "U" }}
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
                  <Button v-if="previewUrl && previewUrl !== DEFAULT_USER_AVATAR" type="button" variant="ghost" size="sm" class="h-7 text-xs text-red-500 hover:text-red-600 hover:bg-red-50" @click="removeAvatar">
                    <Trash2 class="h-3 w-3 mr-1.5" /> 移除
                  </Button>
                </div>
                <p class="text-xs text-slate-400">支持 JPG, PNG, WEBP 格式，最大 20MB</p>
              </div>

              <!-- 表单 -->
              <div class="grid grid-cols-2 gap-x-6 gap-y-5">
                <div class="space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider" for="username">登录账号 <span class="text-red-500">*</span></Label>
                  <Input v-model="username" id="username" placeholder="如: zhangsan" class="h-10" :disabled="mode === 'edit'" />
                  <p v-if="errors.username" class="text-sm text-red-600">{{ errors.username }}</p>
                </div>
                <div class="space-y-1.5" v-if="mode === 'create'">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider" for="password">初始密码 <span class="text-red-500">*</span></Label>
                  <Input v-model="password" id="password" type="password" placeholder="至少6位" class="h-10" />
                  <p v-if="errors.password" class="text-sm text-red-600">{{ errors.password }}</p>
                </div>
                <div class="space-y-1.5" v-if="mode === 'edit'">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">账号状态</Label>
                  <Select v-model="status">
                    <SelectTrigger class="h-10"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem value="enabled">启用 (允许登录)</SelectItem>
                        <SelectItem value="disabled">禁用 (禁止登录)</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <p v-if="errors.status" class="text-sm text-red-600">{{ errors.status }}</p>
                </div>
                <div class="space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider" for="nickname">用户昵称</Label>
                  <Input v-model="nickname" id="nickname" placeholder="如: 张三" class="h-10" />
                  <p v-if="errors.nickname" class="text-sm text-red-600">{{ errors.nickname }}</p>
                </div>
                <div class="space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider" for="roleId">系统角色 <span class="text-red-500">*</span></Label>
                  <Select v-model="roleId">
                    <SelectTrigger class="h-10"><SelectValue placeholder="选择角色" /></SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem v-for="r in roles" :key="r.id" :value="String(r.id)">{{ r.label }}</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <p v-if="errors.roleId" class="text-sm text-red-600">{{ errors.roleId }}</p>
                </div>
                <div class="space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider" for="email">电子邮箱</Label>
                  <Input v-model="email" id="email" type="email" placeholder="example@company.com" class="h-10" />
                  <p v-if="errors.email" class="text-sm text-red-600">{{ errors.email }}</p>
                </div>
                <div class="space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider" for="mobile">手机号码 <span class="text-red-500">*</span></Label>
                  <Input v-model="mobile" id="mobile" placeholder="138xxxx8888" class="h-10" />
                  <p v-if="errors.mobile" class="text-sm text-red-600">{{ errors.mobile }}</p>
                </div>
                <div class="col-span-2 space-y-1.5">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider" for="hanghangCrmMobile">坐席手机号</Label>
                  <Input
                    v-model="hanghangCrmMobile"
                    id="hanghangCrmMobile"
                    placeholder="请输入坐席手机号"
                    class="h-10"
                  />
                  <p v-if="errors.hanghangCrmMobile" class="text-sm text-red-600">{{ errors.hanghangCrmMobile }}</p>
                </div>
                <div class="col-span-2 space-y-1.5 pt-4 border-t border-slate-100">
                  <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">汇报上级 (可选)</Label>
                  <Select v-model="parentId">
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
                  <p v-if="errors.parentId" class="text-sm text-red-600">{{ errors.parentId }}</p>
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
