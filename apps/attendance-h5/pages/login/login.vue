<template>
  <view class="page">
    <!-- 顶部企业 Logo，适配状态栏安全视区 -->
    <view class="logo-wrap">
      <image
        src="/static/logo-header.png"
        mode="heightFix"
        class="logo-img"
      ></image>
    </view>

    <view class="title">
      <view>您好，欢迎登录！</view>
    </view>

    <view class="form">
      <view class="inputs">
        <view class="input-wrap account">
          <up-input
            v-model="form.username"
            placeholder="请输入账号"
            border="none"
            :adjust-position="false"
            placeholderStyle="color: #ccc; font-size: 16px;"
            customStyle="padding: 0;"
          ></up-input>
        </view>

        <view class="input-wrap password">
          <up-input
            v-model="form.password"
            :type="pwdShow ? 'text' : 'password'"
            :passwordIcon="false"
            placeholder="请输入密码"
            border="none"
            :adjust-position="false"
            placeholderStyle="color: #ccc; font-size: 16px;"
            customStyle="padding: 0;"
          >
            <template #suffix>
              <view
                @click="pwdShow = !pwdShow"
                style="padding: 4px"
                class="my-eye-icon"
              >
                <up-icon
                  :name="pwdShow ? 'eye-off' : 'eye'"
                  color="#cccccc"
                  size="24"
                ></up-icon>
              </view>
            </template>
          </up-input>
        </view>

        <!-- Captcha -->
        <view class="input-wrap captcha">
          <up-input
            v-model="form.captchaCode"
            maxlength="4"
            placeholder="请输入验证码"
            border="none"
            :adjust-position="false"
            placeholderStyle="color: #ccc; font-size: 16px;"
            customStyle="padding: 0;"
          >
            <template #suffix>
              <view class="captcha-img-wrap" @click="refreshCaptcha">
                <image
                  v-if="captchaImage"
                  :src="captchaImage"
                  mode="aspectFill"
                  class="captcha-img"
                />
                <view v-else class="captcha-fallback">
                  <up-loading-icon
                    v-if="captchaLoading"
                    size="14"
                  ></up-loading-icon>
                  <text v-else style="font-size: 10px; color: #999"
                    >加载失败</text
                  >
                </view>
              </view>
            </template>
          </up-input>
        </view>

        <!-- Error Msg -->
        <view v-if="errorMsg" class="error-msg">
          <up-icon name="info-circle" color="#ef4444" size="14"></up-icon>
          <text>{{ errorMsg }}</text>
        </view>
      </view>

      <view class="button">
        <view @click="handleLogin" :class="{ 'disabled-btn': loading }">
          <up-loading-icon
            v-if="loading"
            mode="circle"
            color="#333"
            size="20"
            style="margin-right: 8px"
          ></up-loading-icon>
          {{ loading ? "登录中..." : "登录" }}
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, reactive } from "vue";
import { onLoad } from "@dcloudio/uni-app";
import { login, getUserInfo, getLoginCaptcha } from "@/api/index.js";

const form = reactive({
  username: "",
  password: "",
  captchaCode: "",
});

const captchaId = ref("");
const captchaImage = ref("");
const pwdShow = ref(false);
const loading = ref(false);
const captchaLoading = ref(false);
const errorMsg = ref("");

const refreshCaptcha = async () => {
  if (captchaLoading.value) return;
  captchaLoading.value = true;
  try {
    const captcha = await getLoginCaptcha();
    captchaId.value = captcha.captchaId;
    captchaImage.value = captcha.captchaImage;
    form.captchaCode = "";
  } catch (error) {
    captchaId.value = "";
    captchaImage.value = "";
    errorMsg.value = error.message || "验证码加载失败";
  } finally {
    captchaLoading.value = false;
  }
};

const handleLogin = async () => {
  errorMsg.value = "";
  if (!form.username.trim()) {
    errorMsg.value = "账号必填";
    return;
  }
  if (!form.password.trim()) {
    errorMsg.value = "密码必填";
    return;
  }
  if (!form.captchaCode.trim()) {
    errorMsg.value = "验证码必填";
    return;
  }
  if (!captchaId.value) {
    errorMsg.value = "验证码已失效，请刷新后重试";
    refreshCaptcha();
    return;
  }

  loading.value = true;
  try {
    const res = await login({
      username: form.username.trim(),
      password: form.password.trim(),
      captchaId: captchaId.value,
      captchaCode: form.captchaCode.trim(),
    });
    uni.setStorageSync("token", res.token);

    const user = await getUserInfo();
    uni.setStorageSync("user", user);

    uni.showToast({ title: "登录成功", icon: "success" });
    setTimeout(() => {
      uni.reLaunch({ url: "/pages/index/index" });
    }, 1000);
  } catch (err) {
    errorMsg.value = err.message || "登录失败，请重试";
    refreshCaptcha();
  } finally {
    loading.value = false;
  }
};

onLoad(() => {
  refreshCaptcha();
});
</script>

<style lang="scss" scoped>
.page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: #fff;
  position: relative;

  .logo-wrap {
    position: absolute;
    top: calc(max(64rpx, env(safe-area-inset-top) + 20rpx));
    left: 64rpx;
    z-index: 10;

    .logo-img {
      height: 48rpx;
      width: auto;
    }
  }

  .title {
    padding-top: 500rpx;
    padding-bottom: 120rpx;
    background: url("../../static/login.png") no-repeat top center;
    background-size: 100%;

    view {
      padding-left: 64rpx;
      font-size: 44rpx;
      font-weight: 700;
      line-height: 1;
    }
  }

  .form {
    flex: 1;
    padding: 0 64rpx;

    .inputs {
      margin-bottom: 80rpx;

      .input-wrap {
        min-height: 96rpx;
        border-radius: 48rpx;
        border: 1rpx solid rgba(204, 204, 204, 1);
        padding: 0 48rpx;
        display: flex;
        align-items: center;
        margin-bottom: 48rpx;
        background: #fff;
      }

      .captcha {
        padding-right: 20rpx; /* less right padding for image */
      }

      .captcha-img-wrap {
        width: 180rpx;
        height: 64rpx;
        border-radius: 32rpx;
        overflow: hidden;
        background-color: #f8fafc;
        display: flex;
        align-items: center;
        justify-content: center;
        border: 1rpx solid #e2e8f0;
      }

      .captcha-img {
        width: 100%;
        height: 100%;
      }

      .captcha-fallback {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 100%;
        height: 100%;
      }

      .error-msg {
        display: flex;
        align-items: center;
        gap: 8rpx;
        color: #ef4444;
        font-size: 26rpx;
        justify-content: center;
        margin-top: -16rpx;
      }
    }

    .button {
      view {
        display: flex;
        align-items: center;
        justify-content: center;
        line-height: 96rpx;
        height: 96rpx;
        border-radius: 48rpx;
        font-size: 32rpx;
        color: #333;
        font-weight: bolder;
        background: linear-gradient(
          90deg,
          rgba(255, 222, 102, 1) 0%,
          rgba(202, 245, 253, 1) 100%
        );
      }
      .disabled-btn {
        opacity: 0.7;
      }
    }
  }
}

/* 隐藏浏览器原生的密码输入框小眼睛 */
:deep(input::-ms-reveal),
:deep(input::-ms-clear),
:deep(input::-webkit-contacts-auto-fill-button),
:deep(input::-webkit-credentials-auto-fill-button) {
  display: none !important;
}

/* 强制隐藏 up-input 默认自作主张生成的属性密码眼，只保留我们自定义插槽里的眼 */
.password :deep(.u-icon:not(.my-eye-icon .u-icon)),
.password :deep(.up-icon:not(.my-eye-icon .up-icon)) {
  display: none !important;
}
</style>
