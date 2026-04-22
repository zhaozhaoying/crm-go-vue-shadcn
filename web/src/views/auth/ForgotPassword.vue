<script setup lang="ts">
import { ref, computed } from "vue";
import { useRouter } from "vue-router";
import {
  Eye,
  EyeOff,
  AlertCircle,
  Check,
  X,
  CheckCircle2,
} from "lucide-vue-next";
import { resetPasswordDirect } from "@/api/modules/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { getRequestErrorMessage } from "@/lib/http-error";
import AuthenticationLayout from "@/layouts/AuthenticationLayout.vue";

const router = useRouter();

const username = ref("");
const contact = ref("");
const newPassword = ref("");
const confirmPassword = ref("");
const showNewPassword = ref(false);
const showConfirmPassword = ref(false);
const loading = ref(false);
const errorMsg = ref("");
const success = ref(false);

// 各字段是否已触碰（blur 或提交后显示校验状态）
const touched = ref({
  username: false,
  contact: false,
  newPassword: false,
  confirmPassword: false,
});

// 密码规则：字母+数字组合，至少6位
const passwordRegex = /^(?=.*[a-zA-Z])(?=.*\d).{6,}$/;

const isUsernameValid = computed(() => username.value.trim().length > 0);
const isContactValid = computed(() => contact.value.trim().length > 0);
const isNewPasswordValid = computed(() =>
  passwordRegex.test(newPassword.value),
);
const isConfirmPasswordValid = computed(
  () =>
    confirmPassword.value.length > 0 &&
    confirmPassword.value === newPassword.value,
);

const allValid = computed(
  () =>
    isUsernameValid.value &&
    isContactValid.value &&
    isNewPasswordValid.value &&
    isConfirmPasswordValid.value,
);

const submit = async (e: Event) => {
  e.preventDefault();
  // 提交时标记所有字段已触碰，显示全部校验状态
  touched.value = {
    username: true,
    contact: true,
    newPassword: true,
    confirmPassword: true,
  };
  if (!allValid.value) return;

  errorMsg.value = "";
  loading.value = true;
  try {
    await resetPasswordDirect({
      username: username.value.trim(),
      contact: contact.value.trim(),
      newPassword: newPassword.value,
    });
    success.value = true;
  } catch (error: unknown) {
    errorMsg.value = getRequestErrorMessage(error, "重置失败，请重试");
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <AuthenticationLayout>
    <!-- 成功状态 -->
    <div v-if="success" class="flex flex-col gap-6">
      <div class="flex flex-col items-center gap-4 text-center">
        <div
          class="flex h-14 w-14 items-center justify-center rounded-full bg-green-100 dark:bg-green-950/40"
        >
          <CheckCircle2 class="h-7 w-7 text-green-600 dark:text-green-400" />
        </div>
        <div class="flex flex-col gap-1">
          <h1 class="text-2xl font-bold">密码已重置</h1>
          <p class="text-balance text-sm text-muted-foreground">
            您的密码已成功更新，请使用新密码登录
          </p>
        </div>
      </div>
      <Button class="w-full" @click="router.push('/login')">前往登录</Button>
    </div>

    <!-- 重置表单 -->
    <form v-else class="flex flex-col gap-6" @submit="submit">
      <div class="flex flex-col items-center gap-2 text-center">
        <h1 class="text-2xl font-bold">重置密码</h1>
        <p class="text-balance text-sm text-muted-foreground">
          验证身份后设置新密码
        </p>
      </div>

      <!-- 全局错误提示 -->
      <div
        v-if="errorMsg"
        class="flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800 dark:border-red-800 dark:bg-red-950/40 dark:text-red-400"
      >
        <AlertCircle class="h-4 w-4 shrink-0" />
        <span>{{ errorMsg }}</span>
      </div>

      <div class="grid gap-5">
        <!-- 账号 -->
        <div class="grid gap-2">
          <Label for="fp-username">账号</Label>
          <div class="relative">
            <Input
              id="fp-username"
              type="text"
              v-model="username"
              autocomplete="username"
              placeholder="请输入账号"
              autofocus
              class="pr-10"
              :class="{
                'border-green-500 focus-visible:ring-green-500/20':
                  touched.username && isUsernameValid,
                'border-red-500 focus-visible:ring-red-500/20':
                  touched.username && !isUsernameValid,
              }"
              @blur="touched.username = true"
            />
            <span
              v-if="touched.username"
              class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2"
            >
              <Check
                v-if="isUsernameValid"
                class="h-4 w-4 text-green-500"
              />
              <X v-else class="h-4 w-4 text-red-500" />
            </span>
          </div>
          <p
            v-if="touched.username && !isUsernameValid"
            class="text-xs text-red-500"
          >
            账号必填
          </p>
        </div>

        <!-- 邮箱 / 手机号 -->
        <div class="grid gap-2">
          <Label for="fp-contact">邮箱 / 手机号</Label>
          <div class="relative">
            <Input
              id="fp-contact"
              type="text"
              v-model="contact"
              placeholder="请输入绑定的邮箱或手机号"
              class="pr-10"
              :class="{
                'border-green-500 focus-visible:ring-green-500/20':
                  touched.contact && isContactValid,
                'border-red-500 focus-visible:ring-red-500/20':
                  touched.contact && !isContactValid,
              }"
              @blur="touched.contact = true"
            />
            <span
              v-if="touched.contact"
              class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2"
            >
              <Check
                v-if="isContactValid"
                class="h-4 w-4 text-green-500"
              />
              <X v-else class="h-4 w-4 text-red-500" />
            </span>
          </div>
          <p
            v-if="touched.contact && !isContactValid"
            class="text-xs text-red-500"
          >
            邮箱或手机号必填
          </p>
        </div>

        <!-- 新密码 -->
        <div class="grid gap-2">
          <Label for="fp-new-password">新密码</Label>
          <div class="relative">
            <Input
              id="fp-new-password"
              :type="showNewPassword ? 'text' : 'password'"
              v-model="newPassword"
              autocomplete="new-password"
              placeholder="字母+数字组合，至少6位"
              class="pr-16"
              :class="{
                'border-green-500 focus-visible:ring-green-500/20':
                  touched.newPassword && isNewPasswordValid,
                'border-red-500 focus-visible:ring-red-500/20':
                  touched.newPassword && !isNewPasswordValid,
              }"
              @blur="touched.newPassword = true"
            />
            <!-- 校验图标 -->
            <span
              v-if="touched.newPassword"
              class="pointer-events-none absolute right-9 top-1/2 -translate-y-1/2"
            >
              <Check
                v-if="isNewPasswordValid"
                class="h-4 w-4 text-green-500"
              />
              <X v-else class="h-4 w-4 text-red-500" />
            </span>
            <!-- 显示/隐藏密码 -->
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
              @click="showNewPassword = !showNewPassword"
              tabindex="-1"
            >
              <EyeOff v-if="showNewPassword" class="h-4 w-4" />
              <Eye v-else class="h-4 w-4" />
            </button>
          </div>
          <p
            v-if="touched.newPassword && !isNewPasswordValid"
            class="text-xs text-red-500"
          >
            {{ newPassword.length === 0 ? "新密码必填" : "密码须包含字母和数字，至少6位" }}
          </p>
        </div>

        <!-- 确认密码 -->
        <div class="grid gap-2">
          <Label for="fp-confirm-password">确认密码</Label>
          <div class="relative">
            <Input
              id="fp-confirm-password"
              :type="showConfirmPassword ? 'text' : 'password'"
              v-model="confirmPassword"
              autocomplete="new-password"
              placeholder="请再次输入新密码"
              class="pr-16"
              :class="{
                'border-green-500 focus-visible:ring-green-500/20':
                  touched.confirmPassword && isConfirmPasswordValid,
                'border-red-500 focus-visible:ring-red-500/20':
                  touched.confirmPassword && !isConfirmPasswordValid,
              }"
              @blur="touched.confirmPassword = true"
            />
            <!-- 校验图标 -->
            <span
              v-if="touched.confirmPassword"
              class="pointer-events-none absolute right-9 top-1/2 -translate-y-1/2"
            >
              <Check
                v-if="isConfirmPasswordValid"
                class="h-4 w-4 text-green-500"
              />
              <X v-else class="h-4 w-4 text-red-500" />
            </span>
            <!-- 显示/隐藏密码 -->
            <button
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
              @click="showConfirmPassword = !showConfirmPassword"
              tabindex="-1"
            >
              <EyeOff v-if="showConfirmPassword" class="h-4 w-4" />
              <Eye v-else class="h-4 w-4" />
            </button>
          </div>
          <p
            v-if="touched.confirmPassword && !isConfirmPasswordValid"
            class="text-xs text-red-500"
          >
            {{
              confirmPassword.length === 0 ? "确认密码必填" : "两次密码不一致"
            }}
          </p>
        </div>

        <Button type="submit" class="w-full" :disabled="loading">
          {{ loading ? "重置中..." : "重置密码" }}
        </Button>

        <div class="text-center text-sm">
          <RouterLink to="/login" class="underline underline-offset-4">
            返回登录
          </RouterLink>
        </div>
      </div>
    </form>
  </AuthenticationLayout>
</template>
