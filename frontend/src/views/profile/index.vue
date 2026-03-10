<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Loader2, Pencil } from "lucide-vue-next"
import { updateUser, uploadUserAvatar } from "@/api/modules/users"
import { getRequestErrorMessage } from "@/lib/http-error"
import { useAuthStore } from "@/stores/auth"
import { toast } from "vue-sonner"

const authStore = useAuthStore()
const photoPreview = ref<string | null>(null)
const photoInput = ref<HTMLInputElement>()

const form = ref({
  nickname: "",
  email: "",
  mobile: "",
  password: "",
  avatar: "",
  photo: null as File | null,
})

const formSubmitting = ref(false)

const currentAvatar = computed(() => photoPreview.value || form.value.avatar || "")
const avatarLoaded = ref(false)
const avatarLoadFailed = ref(false)
const showAvatarLoading = computed(
  () => !!currentAvatar.value && !avatarLoaded.value && !avatarLoadFailed.value,
)

watch(
  () => authStore.user,
  (nextUser) => {
    if (!nextUser) return
    form.value.nickname = nextUser.nickname || nextUser.username || ""
    form.value.email = nextUser.email || ""
    form.value.mobile = nextUser.mobile || ""
    form.value.avatar = nextUser.avatar || ""
    form.value.password = ""
    form.value.photo = null
    photoPreview.value = null
  },
  { immediate: true },
)

onMounted(async () => {
  if (!authStore.profileLoaded) {
    await authStore.fetchCurrentUserProfile()
  }
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
  const photo = photoInput.value?.files?.[0]
  if (!photo) return
  if (!photo.type.startsWith("image/")) {
    toast.error("请上传图片格式文件")
    return
  }
  if (photo.size > 2 * 1024 * 1024) {
    toast.error("图片大小不能超过 2MB")
    return
  }

  form.value.photo = photo
  const reader = new FileReader()
  reader.onload = (event) => {
    photoPreview.value = (event.target?.result as string) || null
  }
  reader.readAsDataURL(photo)
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

const handleSubmit = async () => {
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
    let avatarUrl = form.value.avatar || ""
    if (form.value.photo) {
      const uploadResult = await uploadUserAvatar(form.value.photo)
      avatarUrl = uploadResult.url || ""
    }

    await updateUser(currentUser.id, {
      username: currentUser.username,
      password: form.value.password.trim(),
      nickname: form.value.nickname.trim(),
      email: form.value.email.trim(),
      mobile: form.value.mobile.trim(),
      avatar: avatarUrl,
      roleId: currentUser.roleId,
      parentId: currentUser.parentId,
      status: currentUser.status || "enabled",
    })

    await authStore.fetchCurrentUserProfile(true)
    form.value.avatar = avatarUrl
    form.value.password = ""
    form.value.photo = null
    photoPreview.value = null
    clearPhotoFileInput()
    toast.success("个人资料已更新")
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "更新失败"))
  } finally {
    formSubmitting.value = false
  }
}
</script>

<template>
  <div class="w-full max-w-3xl mx-auto">
    <section class="space-y-6">
      <header class="flex flex-col gap-2">
        <h2 class="text-lg font-medium text-gray-900">
          个人资料
        </h2>
        <p class="text-sm text-gray-600">
          更新您的昵称、手机号、头像和密码。
        </p>
      </header>

      <form @submit.prevent="handleSubmit" class="space-y-6">
        <div>
          <input
            id="photo"
            ref="photoInput"
            name="photo"
            type="file"
            class="hidden"
            @change="updatePhotoPreview"
            accept="image/*"
          />

          <div class="flex items-center gap-4">
            <div class="group relative">
              <Avatar class="size-20 rounded-lg">
                <AvatarImage
                  v-if="currentAvatar"
                  :src="currentAvatar"
                  :alt="form.nickname"
                  @load="handleAvatarLoad"
                  @error="handleAvatarError"
                />
                <AvatarFallback class="rounded-lg text-xl">
                  <Loader2 v-if="showAvatarLoading" class="h-5 w-5 animate-spin text-muted-foreground" />
                  <span v-else>
                    {{ (form.nickname || authStore.user?.username || "用户").substring(0, 2).toUpperCase() }}
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

        <div class="space-y-1.5">
          <Label for="name">昵称</Label>
          <Input
            id="name"
            type="text"
            class="max-w-lg"
            v-model="form.nickname"
            required
            autocomplete="name"
          />
        </div>

        <div class="space-y-1.5">
          <Label for="email">邮箱</Label>
          <Input
            id="email"
            type="email"
            class="max-w-lg"
            v-model="form.email"
            autocomplete="username"
          />
        </div>

        <div class="space-y-1.5">
          <Label for="mobile">手机号</Label>
          <Input
            id="mobile"
            type="tel"
            class="max-w-lg"
            v-model="form.mobile"
            autocomplete="tel"
          />
        </div>

        <div class="space-y-1.5">
          <Label for="password">新密码（不修改请留空）</Label>
          <Input
            id="password"
            type="password"
            class="max-w-lg"
            v-model="form.password"
            placeholder="留空表示不修改密码"
            autocomplete="new-password"
          />
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
