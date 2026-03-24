<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Loader2, Pencil } from "lucide-vue-next"
import { updateUser, uploadUserAvatar } from "@/api/modules/users"
import { getRequestErrorMessage } from "@/lib/http-error"
import { requiredString } from "@/lib/form-validation"
import { resolveUserAvatar } from "@/lib/user-avatar"
import { useAuthStore } from "@/stores/auth"
import { toast } from "vue-sonner"
import { toTypedSchema } from "@vee-validate/zod"
import { useForm, useField } from "vee-validate"
import * as z from "zod"

const authStore = useAuthStore()
const photoPreview = ref<string | null>(null)
const photoInput = ref<HTMLInputElement>()

const sanitizeHanghangCrmMobileInput = (raw: string) =>
  String(raw || "").replace(/\D/g, "").slice(0, 11)

const formSchema = toTypedSchema(
  z.object({
    nickname: requiredString("昵称"),
    email: z.string().email({ message: "请输入有效的邮箱地址" }).optional().or(z.literal("")),
    mobile: z.string().optional(),
    hanghangCrmMobile: z.string().optional().refine((value) => !value || /^1\d{10}$/.test(value), {
      message: "坐席手机号必须为11位数字",
    }),
    password: z.string().min(6, { message: "密码至少需要6位" }).optional().or(z.literal("")),
    photo: z.any().optional(),
  }),
)

const { handleSubmit, setValues, errors } = useForm({
  validationSchema: formSchema,
})

const { value: nickname } = useField<string>("nickname")
const { value: email } = useField<string>("email")
const { value: mobile } = useField<string>("mobile")
const { value: hanghangCrmMobile } = useField<string>("hanghangCrmMobile")
const { value: password } = useField<string>("password")
const { value: photo } = useField<File | null>("photo")

const formSubmitting = ref(false)

watch(hanghangCrmMobile, (value) => {
  const sanitized = sanitizeHanghangCrmMobileInput(value)
  if (sanitized !== value) {
    hanghangCrmMobile.value = sanitized
  }
})

const currentAvatar = computed(() => photoPreview.value || resolveUserAvatar(authStore.user?.avatar))
const avatarLoaded = ref(false)
const avatarLoadFailed = ref(false)
const showAvatarLoading = computed(
  () => !!currentAvatar.value && !avatarLoaded.value && !avatarLoadFailed.value,
)

watch(
  () => authStore.user,
  (nextUser) => {
    if (!nextUser) return
    setValues({
      nickname: nextUser.nickname || nextUser.username || "",
      email: nextUser.email || "",
      mobile: nextUser.mobile || "",
      hanghangCrmMobile: nextUser.hanghangCrmMobile || "",
      password: "",
      photo: null,
    })
    photoPreview.value = null
  },
  { immediate: true, deep: true },
)

onMounted(async () => {
  await authStore.fetchCurrentUserProfile(true)
})

watch(
  currentAvatar,
  () => {
    avatarLoaded.value = false
    avatarLoadFailed.value = false
  },
  { immediate: true },
)

const selectNewPhoto = () => {
  photoInput.value?.click()
}

const updatePhotoPreview = () => {
  const file = photoInput.value?.files?.[0]
  if (!file) return
  if (!file.type.startsWith("image/")) {
    toast.error("请上传图片格式文件")
    return
  }
  if (file.size > 20 * 1024 * 1024) {
    toast.error("图片大小不能超过 20MB")
    return
  }

  photo.value = file
  const reader = new FileReader()
  reader.onload = (event) => {
    photoPreview.value = (event.target?.result as string) || null
  }
  reader.readAsDataURL(file)
}

const clearPhotoFileInput = () => {
  if (photoInput.value?.value) {
    photoInput.value.value = ""
  }
}

const handleAvatarLoad = () => {
  avatarLoaded.value = true
  avatarLoadFailed.value = false
}

const handleAvatarError = () => {
  avatarLoaded.value = false
  avatarLoadFailed.value = true
}

const onSubmit = handleSubmit(async (values) => {
  const currentUser = authStore.user
  if (!currentUser) {
    toast.error("未获取到当前登录用户")
    return
  }
  if (!currentUser.roleId) {
    toast.error("当前用户角色信息缺失，请重新登录后重试")
    return
  }

  formSubmitting.value = true
  try {
    let avatarUrl = resolveUserAvatar(currentUser.avatar)
    if (values.photo) {
      const uploadResult = await uploadUserAvatar(values.photo)
      avatarUrl = uploadResult.url || ""
    }

    await updateUser(currentUser.id, {
      username: currentUser.username,
      password: values.password?.trim() || "",
      nickname: values.nickname.trim(),
      email: values.email?.trim() || "",
      mobile: values.mobile?.trim() || "",
      hanghangCrmMobile: values.hanghangCrmMobile?.trim() || "",
      avatar: avatarUrl,
      roleId: currentUser.roleId,
      parentId: currentUser.parentId,
      status: currentUser.status || "enabled",
    })

    await authStore.fetchCurrentUserProfile(true)
    password.value = ""
    photo.value = null
    photoPreview.value = null
    clearPhotoFileInput()
    toast.success("个人资料已更新")
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "更新失败"))
  } finally {
    formSubmitting.value = false
  }
})
</script>

<template>
  <div class="w-full max-w-4xl mx-auto">
    <section class="space-y-6">
      <form @submit="onSubmit" class="space-y-6">
        <div class="grid gap-2 md:grid-cols-[140px_minmax(0,1fr)] md:items-start">
          <input
            id="photo"
            ref="photoInput"
            type="file"
            class="hidden"
            @change="updatePhotoPreview"
            accept="image/*"
          />

          <Label for="photo" class="pt-2 text-sm text-gray-700">头像</Label>

          <div class="flex items-center gap-4">
            <div class="group relative">
              <Avatar class="size-20 rounded-lg">
                <AvatarImage
                  v-if="currentAvatar"
                  :src="currentAvatar"
                  :alt="nickname"
                  @load="handleAvatarLoad"
                  @error="handleAvatarError"
                />
                <AvatarFallback class="rounded-lg text-xl">
                  <Loader2 v-if="showAvatarLoading" class="h-5 w-5 animate-spin text-muted-foreground" />
                  <span v-else>
                    {{ (nickname || authStore.user?.username || "用户").substring(0, 2).toUpperCase() }}
                  </span>
                </AvatarFallback>
              </Avatar>

              <button
                type="button"
                class="absolute inset-0 flex size-20 cursor-pointer items-center justify-center rounded-lg bg-black/50 opacity-0 transition-opacity group-hover:opacity-100"
                @click="selectNewPhoto"
              >
                <Pencil class="h-5 w-5 text-white" />
              </button>
            </div>
          </div>
        </div>

        <div class="grid gap-2 md:grid-cols-[140px_minmax(0,1fr)] md:items-start">
          <Label for="name" class="pt-2 text-sm text-gray-700">昵称</Label>
          <div class="space-y-1.5">
            <Input id="name" type="text" class="max-w-lg" v-model="nickname" autocomplete="name" />
            <p v-if="errors.nickname" class="text-xs text-destructive">{{ errors.nickname }}</p>
          </div>
        </div>

        <div class="grid gap-2 md:grid-cols-[140px_minmax(0,1fr)] md:items-start">
          <Label for="email" class="pt-2 text-sm text-gray-700">邮箱</Label>
          <div class="space-y-1.5">
            <Input id="email" type="email" class="max-w-lg" v-model="email" autocomplete="username" />
            <p v-if="errors.email" class="text-xs text-destructive">{{ errors.email }}</p>
          </div>
        </div>

        <div class="grid gap-2 md:grid-cols-[140px_minmax(0,1fr)] md:items-start">
          <Label for="mobile" class="pt-2 text-sm text-gray-700">手机号</Label>
          <div class="space-y-1.5">
            <Input id="mobile" type="tel" class="max-w-lg" v-model="mobile" autocomplete="tel" />
            <p v-if="errors.mobile" class="text-xs text-destructive">{{ errors.mobile }}</p>
          </div>
        </div>

        <div class="grid gap-2 md:grid-cols-[140px_minmax(0,1fr)] md:items-start">
          <Label for="hanghang-crm-mobile" class="pt-2 text-sm text-gray-700">坐席手机号</Label>
          <div class="space-y-1.5">
            <Input
              id="hanghang-crm-mobile"
              class="max-w-lg"
              v-model="hanghangCrmMobile"
              placeholder="请输入坐席手机号"
            />
            <p v-if="errors.hanghangCrmMobile" class="text-xs text-destructive">{{ errors.hanghangCrmMobile }}</p>
          </div>
        </div>

        <div class="grid gap-2 md:grid-cols-[140px_minmax(0,1fr)] md:items-start">
          <Label for="password" class="pt-2 text-sm text-gray-700">新密码</Label>
          <div class="space-y-1.5">
            <Input
              id="password"
              type="password"
              class="max-w-lg"
              v-model="password"
              placeholder="留空表示不修改密码"
              autocomplete="new-password"
            />
            <p class="text-xs text-gray-500">不修改请留空</p>
            <p v-if="errors.password" class="text-xs text-destructive">{{ errors.password }}</p>
          </div>
        </div>

        <div class="flex items-center gap-4">
          <Button type="submit" :disabled="formSubmitting">
            {{ formSubmitting ? "保存中..." : "保存" }}
          </Button>
        </div>
      </form>
    </section>
  </div>
</template>
