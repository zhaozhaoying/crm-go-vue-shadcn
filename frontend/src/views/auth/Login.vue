<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { Eye, EyeOff, AlertCircle, RefreshCw } from "lucide-vue-next";
import { getLoginCaptcha } from "@/api/modules/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from "@/components/ui/input-group";
import { Label } from "@/components/ui/label";
import { getRequestErrorMessage } from "@/lib/http-error";
import AuthenticationLayout from "@/layouts/AuthenticationLayout.vue";
import { useAuthStore } from "@/stores/auth";

const router = useRouter();
const authStore = useAuthStore();

const username = ref("");
const password = ref("");
const showPassword = ref(false);
const errorMsg = ref("");
const loading = ref(false);
const captchaInput = ref("");
const captchaId = ref("");
const captchaImage = ref("");
const captchaLoading = ref(false);

const refreshCaptcha = async () => {
  captchaLoading.value = true;
  try {
    const captcha = await getLoginCaptcha();
    captchaId.value = captcha.captchaId;
    captchaImage.value = captcha.captchaImage;
    captchaInput.value = "";
  } catch (error: unknown) {
    captchaId.value = "";
    captchaImage.value = "";
    errorMsg.value = getRequestErrorMessage(error, "验证码加载失败");
  } finally {
    captchaLoading.value = false;
  }
};

const submit = async (e: Event) => {
  e.preventDefault();
  errorMsg.value = ""; // 提交时先清空之前的错误

  if (!username.value.trim()) {
    errorMsg.value = "账号必填";
    return;
  }
  if (!password.value.trim()) {
    errorMsg.value = "密码必填";
    return;
  }
  if (!captchaInput.value.trim()) {
    errorMsg.value = "验证码必填";
    return;
  }
  if (!captchaId.value) {
    errorMsg.value = "验证码已失效，请刷新后重试";
    await refreshCaptcha();
    return;
  }

  loading.value = true;
  try {
    await authStore.login({
      username: username.value,
      password: password.value,
      captchaId: captchaId.value,
      captchaCode: captchaInput.value.trim(),
    });
    router.push("/dashboard");
  } catch (error: unknown) {
    errorMsg.value = getRequestErrorMessage(error, "登录失败，请重试");
    await refreshCaptcha();
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  void refreshCaptcha();
});
</script>

<template>
  <AuthenticationLayout>
    <form class="flex flex-col gap-6" @submit="submit">
      <div class="flex flex-col items-center gap-2 text-center">
        <h1 class="text-2xl font-bold">登录您的账户</h1>
        <p class="text-balance text-sm text-muted-foreground">
          输入您的账号密码以登录系统
        </p>
      </div>

      <div
        v-if="errorMsg"
        class="flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800 dark:border-red-800 dark:bg-red-950/40 dark:text-red-400"
      >
        <AlertCircle class="h-4 w-4 shrink-0" />
        <span>{{ errorMsg }}</span>
      </div>

      <div class="grid gap-6">
        <div class="grid gap-2">
          <Label for="username">账号</Label>
          <Input
            id="username"
            type="text"
            v-model="username"
            autocomplete="username"
            placeholder="请输入账号"
            autofocus
          />
        </div>
        <div class="grid gap-2">
          <Label for="password">密码</Label>
          <div class="relative">
            <Input
              id="password"
              :type="showPassword ? 'text' : 'password'"
              v-model="password"
              autocomplete="current-password"
              placeholder="请输入6位数以上密码"
              class="pr-10"
            />
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
              @click="showPassword = !showPassword"
              tabindex="-1"
            >
              <EyeOff v-if="showPassword" class="h-4 w-4" />
              <Eye v-else class="h-4 w-4" />
            </button>
          </div>
        </div>
        <div class="grid gap-2">
          <div class="flex items-center justify-between">
            <Label for="captcha">验证码</Label>
            <button
              type="button"
              class="inline-flex items-center gap-1 text-xs text-muted-foreground transition-colors hover:text-foreground"
              @click="refreshCaptcha"
            >
              <RefreshCw class="h-3 w-3" />
              换一张
            </button>
          </div>
          <InputGroup>
            <InputGroupInput
              id="captcha"
              v-model="captchaInput"
              type="text"
              maxlength="4"
              placeholder="请输入验证码"
              class="uppercase"
            />
            <InputGroupAddon class="w-[120px] overflow-hidden bg-muted/40 px-0">
              <button
                type="button"
                class="h-full w-full shrink-0"
                title="点击刷新验证码"
                :disabled="captchaLoading"
                @click="refreshCaptcha"
              >
                <img
                  v-if="captchaImage"
                  :src="captchaImage"
                  alt="验证码"
                  class="h-full w-full object-cover"
                />
                <span v-else class="text-xs text-muted-foreground">
                  {{ captchaLoading ? "加载中..." : "加载失败" }}
                </span>
              </button>
            </InputGroupAddon>
          </InputGroup>
        </div>

        <Button type="submit" class="w-full" :disabled="loading">
          {{ loading ? "登录中..." : "登录" }}
        </Button>
        <div class="text-center text-sm">
          <RouterLink to="/forgot-password" class="underline underline-offset-4">忘记密码？</RouterLink>
        </div>
      </div>
    </form>
  </AuthenticationLayout>
</template>
