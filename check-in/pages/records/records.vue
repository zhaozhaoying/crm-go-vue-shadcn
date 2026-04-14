<template>
  <view class="page">
    <view class="timeline" v-if="visitList && visitList.length > 0">
      <view
        class="timeline-item"
        v-for="(item, index) in visitList"
        :key="item.id"
      >
        <!-- 左侧时间线 -->
        <view class="tl-left">
          <view class="dot"></view>
          <view class="line" v-if="index !== visitList.length - 1"></view>
        </view>

        <!-- 右侧卡片 -->
        <view class="tl-right">
          <view class="card">
            <view class="row">
              <text class="label">打卡时间：</text>
              <text class="value time-val">{{
                formatDate(item.createdAt)
              }}</text>
            </view>
            <view class="row">
              <text class="label">打卡位置：</text>
              <text class="value">{{ item.detailAddress || "未知地点" }}</text>
            </view>
            <view class="row" v-if="item.visitPurpose">
              <text class="label">打卡目的：</text>
              <text class="value">{{ item.visitPurpose }}</text>
            </view>
            <view class="row" v-if="item.customerName">
              <text class="label">公司名称：</text>
              <text class="value">{{ item.customerName }}</text>
            </view>
            <view class="row" v-if="item.inviter">
              <text class="label">邀约人：</text>
              <text class="value">{{ item.inviter }}</text>
            </view>
            <view class="row" v-if="item.remark">
              <text class="label">备注信息：</text>
              <text class="value">{{ item.remark }}</text>
            </view>
          </view>
        </view>
      </view>
    </view>

    <!-- 加载状态 -->
    <view style="padding: 20rpx 0">
      <up-loadmore :status="loadStatus" v-if="visitList.length > 0" />
    </view>

    <view class="empty" v-if="!loading && visitList.length === 0">
      <up-empty mode="data" text="暂无打卡记录"></up-empty>
    </view>
  </view>
</template>

<script setup>
import { ref } from "vue";
import { onLoad, onShow, onReachBottom, onPullDownRefresh } from "@dcloudio/uni-app";
import { getCustomerVisits } from "@/api/index.js";
import { ensureRouteAccess } from "@/utils/auth.js";

const visitList = ref([]);
const page = ref(1);
const pageSize = ref(10);
const loading = ref(false);
const loadStatus = ref("loadmore"); // loadmore, loading, nomore
const pageUrl = "/pages/records/records";

const formatDate = (dateStr) => {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  const YYYY = d.getFullYear();
  const MM = String(d.getMonth() + 1).padStart(2, "0");
  const DD = String(d.getDate()).padStart(2, "0");
  const HH = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  const ss = String(d.getSeconds()).padStart(2, "0");
  return `${YYYY}-${MM}-${DD} ${HH}:${mm}:${ss}`;
};

const fetchList = async (isRefresh = false) => {
  if (isRefresh) {
    page.value = 1;
    loadStatus.value = "loadmore";
  }
  if (loadStatus.value === "nomore" || loading.value) return;

  loading.value = true;
  loadStatus.value = "loading";

  try {
    const res = await getCustomerVisits({
      page: page.value,
      pageSize: pageSize.value,
    });

    const newData = res.items || [];

    if (isRefresh) {
      visitList.value = newData;
      uni.stopPullDownRefresh();
    } else {
      visitList.value = [...visitList.value, ...newData];
    }

    if (newData.length < pageSize.value) {
      loadStatus.value = "nomore";
    } else {
      loadStatus.value = "loadmore";
    }
    page.value++;
  } catch (err) {
    console.error("加载记录失败", err);
    uni.showToast({ title: "加载失败", icon: "none" });
    loadStatus.value = "loadmore";
    if (isRefresh) uni.stopPullDownRefresh();
  } finally {
    loading.value = false;
  }
};

onLoad(() => {
  ensureRouteAccess(pageUrl);
});

onShow(() => {
  if (!ensureRouteAccess(pageUrl)) {
    return;
  }
  fetchList(true);
});

onPullDownRefresh(() => {
  fetchList(true);
});

onReachBottom(() => {
  fetchList(false);
});
</script>

<style lang="scss" scoped>
.page {
  min-height: 100vh;
  background-color: #ffffff;
  padding: 30rpx 40rpx;
  box-sizing: border-box;
  background: url("/static/login.png") no-repeat top center;
}

.header {
  padding: 10rpx 0 40rpx;
}

/* Timeline */
.timeline {
  display: flex;
  flex-direction: column;
}

.timeline-item {
  display: flex;
  position: relative;
  min-height: 220rpx;
}

/* Left axis line and dot */
.tl-left {
  width: 60rpx;
  padding-top: 32rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  flex-shrink: 0;
}

.dot {
  width: 20rpx;
  height: 20rpx;
  background-color: #d97706; /* 橙色圆点 */
  border-radius: 50%;
  position: relative;
  z-index: 2;
  /* 圆点外圈的发光/边框效果 */
  box-shadow: 0 0 0 8rpx rgba(217, 119, 6, 0.2);
}

.line {
  width: 2rpx;
  background-color: #a8a29e;
  position: absolute;
  top: 56rpx; /* Start strictly below the dot's shadow */
  bottom: -32rpx; /* Push all the way through this item's height */
  left: 50%;
  transform: translateX(-50%);
  z-index: 1;
}

/* Right content card */
.tl-right {
  flex: 1;
  padding-left: 20rpx;
  padding-bottom: 40rpx;
  min-width: 0; /* Prevents flex children from overflowing */
}

.card {
  background-color: #ffffff;
  border-radius: 12rpx;
  padding: 24rpx 30rpx;
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.2);
}

.row {
  display: flex;
  font-size: 28rpx;
  line-height: 1.6;
}

.label {
  color: #555;
  white-space: nowrap;
}

.value {
  color: #555;
  flex: 1;
  word-break: break-all;
}

.empty {
  padding-top: 150rpx;
}
</style>
