<template>
	<view class="page">
		<scroll-view class="main-scroll" scroll-y enhanced="true" show-scrollbar="false">
			<!-- Main -->
			<view class="main-body">
				<!-- USER CARD -->
				<view class="user-card">
					<view class="uc-left" @click="handleUserCardClick">
						<view class="uc-avatar">
							<image v-if="userAvatar" :src="userAvatar" mode="aspectFill" class="uc-avatar-img" />
							<u-icon v-else name="account-fill" color="#c0c4cc" size="32"></u-icon>
						</view>
						<view class="uc-info">
							<text class="uc-name">{{ userDisplayName }}</text>
						</view>
					</view>
					<view class="uc-right" @click="goToRecords">
						<u-icon name="calendar" size="26" color="#333"></u-icon>
						<text class="uc-action-txt">记录</text>
					</view>
				</view>

				<!-- BIG WHITE CARD -->
				<view class="card">
					<!-- CHECK-IN BUTTON - on top -->
					<view class="ck-area">
						<!-- Before check-in: blue -->
						<view v-if="!checkedIn" class="ck-btn ck-btn-blue" :class="{ 'ck-btn-disabled': !locationGranted }"
							@click="handleCheckIn">
							<text class="ck-label">拜访签到</text>
							<text class="ck-date">{{ formattedDate }}</text>
							<text class="ck-time">{{ formattedTime }}</text>
						</view>

						<!-- After check-in: green -->
						<view v-else class="ck-btn ck-btn-green">
							<text class="ck-label">签到成功</text>
							<text class="ck-date">{{ formattedDate }}</text>
							<text class="ck-time">{{ checkedInTime }}</text>
						</view>

						<!-- Address displayed outside the circle -->
						<view class="ck-address-box">
							<view class="ck-address-inner" v-if="locationGranted || checkedIn">
								<up-icon name="map-fill" color="#64748b" size="14"></up-icon>
								<text class="address-txt">当前位置：{{ currentAddress }}</text>
							</view>
							<view class="ck-address-inner" v-else-if="locationLoading">
								<u-loading-icon color="#9ca3af" size="14" mode="circle" class="mr-1"></u-loading-icon>
								<text class="address-txt pl-1">正在定位...</text>
							</view>
						</view>
					</view>

					<!-- Location Error -->
					<view v-if="locationError" class="loc-err">
						<view class="err-txt-wrap">
							<u-icon name="info-circle-fill" color="#ef4444" size="16"></u-icon>
							<text class="err-txt">{{
              isWechat && !locationDenied
                ? "若微信中无法定位，建议跳过或者稍后重试"
                : locationError
            }}</text>
						</view>
						<view class="err-btns">
							<text class="retry-btn" @click="getLocation">重新获取</text>
							<text class="retry-btn txt-gray" @click="skipLocation">跳过定位</text>
						</view>
					</view>

					<!-- Form -->
					<view class="form-section">
						<!-- First Row: Visit Purpose and Customer Name -->
						<view class="field-row">
							<!-- Visit Purpose -->
							<view class="field purpose-field">
								<text class="lbl-txt"><text class="req-star">*</text>拜访目的</text>
								<picker @change="onPurposeChange" :value="purposeIndex" :range="visitPurposeOptions">
									<view class="picker-view" :class="{ 'placeholder-color': !visitPurpose }">
										<text class="text-truncate" style="max-width: 100px">{{
                    visitPurpose || "选择目的"
                  }}</text>
										<u-icon name="arrow-down" color="#ccc" size="14"></u-icon>
									</view>
								</picker>
							</view>

							<!-- Customer Name -->
						<view class="field name-field">
								<text class="lbl-txt"><text class="req-star">*</text>公司名称</text>
								<input v-model="customerName" type="text" class="ipt" placeholder="必填公司名称"
									placeholder-class="placeholder-color" />
							</view>
						</view>

						<view class="field">
							<text class="lbl-txt">邀约人</text>
							<picker @change="onInviterChange" :value="inviterIndex" :range="telemarketingUserNames">
								<view class="picker-view" :class="{ 'placeholder-color': !inviterDisplayName }">
									<text class="text-truncate">{{ inviterDisplayName || "选择邀约人" }}</text>
									<u-icon name="arrow-down" color="#ccc" size="14"></u-icon>
								</view>
							</picker>
						</view>

						<!-- Remark -->
						<view class="field">
							<text class="lbl-txt">备注</text>
							<textarea v-model="remark" class="ipt ipt-area" placeholder="填写备注信息..."
								placeholder-class="placeholder-color" :maxlength="600"></textarea>
						</view>

						<!-- Images -->
						<view class="field">
							<view class="lbl">
								<u-icon name="camera-fill" size="16" color="#333"></u-icon>
								<text class="lbl-txt">签到图片</text>
							</view>
							<view class="img-row">
								<view v-for="(url, idx) in images" :key="idx" class="img-thumb">
									<image :src="url" mode="aspectFill" class="img-item-bg" @click="openPreview(idx)" />
									<view class="img-mask" @click="removeImage(idx)">
										<u-icon name="close" color="#fff" size="16"></u-icon>
									</view>
								</view>

								<view class="img-add" @click="handleImageUpload" v-if="!uploadingImage">
									<u-icon name="camera" color="#999" size="24"></u-icon>
									<text class="img-add-txt">拍照/相册</text>
								</view>
								<view class="img-add" v-else>
									<u-loading-icon mode="circle" size="24" color="#999"></u-loading-icon>
								</view>
							</view>
						</view>
					</view>

					<!-- Submit Button -->
					<view class="submit-area">
						<button class="submit-btn" :class="{ 'submit-btn-disabled': !canSubmit }" @click="handleSubmit">
							<u-loading-icon v-if="submitting" mode="circle" color="#fff" size="20"></u-loading-icon>
							<text class="submit-btn-txt">{{
              submitting ? "提交中..." : "提交"
            }}</text>
						</button>
					</view>
				</view>

			</view>
		</scroll-view>

		<!-- Success Modal Overlay -->
		<view v-if="submitted" class="success-overlay" @touchmove.stop.prevent="() => {}">
			<view class="success-modal animate-pop">
				<view class="done-icon">
					<view class="icon-circle">
						<u-icon name="checkbox-mark" color="#fff" size="40"></u-icon>
					</view>
				</view>
				<text class="done-title">签到成功</text>
				<view class="done-info-box">
					<view class="info-row">
						<u-icon name="clock" color="#64748b" size="14"></u-icon>
						<text class="info-txt">{{ formattedDate }} {{ checkedInTime }}</text>
					</view>
					<view class="info-row mt-6">
						<u-icon name="map-fill" color="#64748b" size="14"></u-icon>
						<text class="info-txt line-clamp-2">{{ currentAddress }}</text>
					</view>
				</view>
				<button class="ok-btn" @click="resetAll">完成</button>
			</view>
		</view>
	</view>
</template>

<script setup>
	import {
		ref,
		computed,
		onUnmounted
	} from "vue";
	import {
		onLoad,
		onUnload
	} from "@dcloudio/uni-app";
	import {
		ensureRouteAccess,
		hasLoginToken,
		redirectToLogin
	} from "@/utils/auth.js";
	import {
		uploadVisitImg,
		createCustomerVisit,
		getCustomerVisits,
		getSystemSettings,
		getTelemarketingUsers,
		reverseGeocodeByApihz,
	} from "@/api/index.js";

	// === User Options ===
	const user = ref(null);

	const visitPurposeOptions = ref(["初次拜访"]);
	const purposeIndex = ref(-1);

	const telemarketingUsers = ref([]);
	const inviterIndex = ref(-1);

	const fetchTelemarketingUsers = async () => {
		try {
			const res = await getTelemarketingUsers();
			if (res && Array.isArray(res)) {
				telemarketingUsers.value = res;
			}
		} catch (err) {
			console.error("加载电销用户列表失败", err);
		}
	};

	const fetchSystemSettings = async () => {
		try {
			const res = await getSystemSettings();
			if (res && res.visitPurposes && res.visitPurposes.length > 0) {
				visitPurposeOptions.value = res.visitPurposes;
			}
		} catch (err) {
			console.error("加载系统配置失败", err);
		}
	};

	// === Location State ===
	const locationLoading = ref(true);
	const locationError = ref("");
	const locationDenied = ref(false);
	const currentLat = ref(0);
	const currentLng = ref(0);
	const currentAddress = ref("");
	const currentProvince = ref("");
	const currentCity = ref("");
	const currentArea = ref("");
	const currentDetailAddress = ref("");
	const locationGranted = ref(false);

	// Wechat environment check
	const isWechat = ref(false);

	// === Check-in State ===
	const checkedIn = ref(false);
	const checkedInTime = ref("");

	// === Form State ===
	const customerName = ref("");
	const inviter = ref("");
	const visitPurpose = ref("");
	const remark = ref("");
	const images = ref([]);
	const uploadingImage = ref(false);

	// === Submit State ===
	const submitting = ref(false);
	const submitted = ref(false);

	const currentTimeStr = ref("");
	const currentDateStr = ref("");
	let timer = null;

	// === History State ===
	const visitList = ref([]);

	// === Computed ===
	const userDisplayName = computed(() => {
		if (!user.value) return "未登录";
		return user.value.nickname || user.value.username || "用户";
	});

	const userAvatar = computed(() => {
		return user.value?.avatar || "";
	});

	const canSubmit = computed(() => {
		return checkedIn.value && !submitting.value && !uploadingImage.value && Boolean(user.value);
	});

	const formattedTime = computed(() => {
		return currentTimeStr.value;
	});

	const formattedDate = computed(() => {
		return currentDateStr.value;
	});

	const telemarketingUserNames = computed(() => {
		return telemarketingUsers.value.map((u) => u.nickname || u.username || "");
	});

	const inviterDisplayName = computed(() => {
		if (inviterIndex.value < 0 || inviterIndex.value >= telemarketingUsers.value.length) {
			return "";
		}
		const u = telemarketingUsers.value[inviterIndex.value];
		return u.nickname || u.username || "";
	});

	// === Lifecycle ===
	onLoad(() => {
		if (!ensureRouteAccess("/pages/index/index")) {
			return;
		}

		const storedUser = uni.getStorageSync("user");
		user.value = storedUser;

		// Check platform
		// #ifdef H5
		isWechat.value = /MicroMessenger/i.test(navigator.userAgent);
		// #endif

		startClock();
		getLocation();
		fetchHistory();
		fetchSystemSettings();
		fetchTelemarketingUsers();
	});

	onUnload(() => {
		if (timer) {
			clearInterval(timer);
		}
	});

	onUnmounted(() => {
		if (timer) {
			clearInterval(timer);
		}
	});

	// === Methods ===
	const startClock = () => {
		const updateTime = () => {
			const now = new Date();
			const hh = String(now.getHours()).padStart(2, "0");
			const mm = String(now.getMinutes()).padStart(2, "0");
			const ss = String(now.getSeconds()).padStart(2, "0");
			currentTimeStr.value = `${hh}:${mm}:${ss}`;

			const yyyy = now.getFullYear();
			const MMMM = now.getMonth() + 1 + "月";
			const dd = now.getDate() + "日";
			const weeks = ["日", "一", "二", "三", "四", "五", "六"];
			const week = "星期" + weeks[now.getDay()];
			currentDateStr.value = `${yyyy}年${MMMM}${dd} ${week}`;
		};
		updateTime();
		timer = setInterval(updateTime, 1000);
	};

	const firstNonEmpty = (...values) => {
		for (const value of values) {
			const trimmed = String(value || "").trim();
			if (trimmed) return trimmed;
		}
		return "";
	};

	const isDirectAdminMunicipality = (value) => {
		const trimmed = String(value || "").trim();
		return ["北京市", "天津市", "上海市", "重庆市"].includes(trimmed);
	};

	const looksLikeChineseDistrict = (value) => {
		const trimmed = String(value || "").trim();
		return !!trimmed && (trimmed.endsWith("区") || trimmed.endsWith("县") || trimmed.endsWith("旗"));
	};

	const clearResolvedAddress = () => {
		currentProvince.value = "";
		currentCity.value = "";
		currentArea.value = "";
		currentDetailAddress.value = "";
	};

	const fetchAddress = (lat, lon) => {
		return new Promise((resolve) => {
			reverseGeocodeByApihz({ lat, lon })
				.then((data) => {
					const province = firstNonEmpty(data.province);
					let city = firstNonEmpty(data.city);
					let area = firstNonEmpty(data.county);
					if (isDirectAdminMunicipality(province)) {
						city = province;
					} else if (!city) {
						city = province;
					}
					if (!area && looksLikeChineseDistrict(city)) {
						area = city;
					}

					const detailAddress = String(data.address || "").trim();
					const displayAddress = detailAddress || `${lat.toFixed(6)}, ${lon.toFixed(6)}`;

					resolve({
						province,
						city,
						area,
						detailAddress,
						displayAddress: displayAddress || `${lat.toFixed(6)}, ${lon.toFixed(6)}`,
					});
				})
				.catch(() => {
					resolve({
						province: "",
						city: "",
						area: "",
						detailAddress: "",
						displayAddress: `${lat.toFixed(6)}, ${lon.toFixed(6)}`,
					});
				});
		});
	};

	const getLocation = () => {
		locationLoading.value = true;
		locationError.value = "";
		locationDenied.value = false;

		uni.getLocation({
			type: "wgs84",
			isHighAccuracy: true,
			success: async (res) => {
				currentLat.value = res.latitude;
				currentLng.value = res.longitude;
				locationGranted.value = false;
				locationDenied.value = false;
				clearResolvedAddress();
				currentAddress.value = "正在定位中...";

				const address = await fetchAddress(res.latitude, res.longitude);
				currentProvince.value = address.province || "";
				currentCity.value = address.city || "";
				currentArea.value = address.area || "";
				currentDetailAddress.value = address.detailAddress || "";
				currentAddress.value = address.displayAddress || `${res.latitude.toFixed(6)}, ${res.longitude.toFixed(6)}`;
				locationLoading.value = false;
				locationGranted.value = true;
			},
			fail: (err) => {
				locationLoading.value = false;
				locationGranted.value = false;

				if (err.errMsg && err.errMsg.includes("auth deny")) {
					locationDenied.value = true;
					locationError.value = "定位权限被拒绝，请去设置中心开启。";
				} else {
					locationError.value = "无法获取位置：" + (err.errMsg || "未知错误");
				}
			},
		});
	};

	const skipLocation = () => {
		locationError.value = "";
		locationLoading.value = false;
		locationGranted.value = true;
		currentLat.value = 0;
		currentLng.value = 0;
		clearResolvedAddress();
		currentAddress.value = isWechat.value ?
			"微信环境未获取到位置" :
			"未获取到位置(已跳过)";
	};

	const handleCheckIn = () => {
		if (!locationGranted.value) return;
		checkedIn.value = true;
		checkedInTime.value = formattedTime.value;
	};

	const onPurposeChange = (e) => {
		purposeIndex.value = e.detail.value;
		visitPurpose.value = visitPurposeOptions.value[purposeIndex.value];
	};

	const onInviterChange = (e) => {
		inviterIndex.value = e.detail.value;
		const u = telemarketingUsers.value[inviterIndex.value];
		inviter.value = u ? (u.nickname || u.username || "") : "";
	};

	const handleImageUpload = () => {
		if (!ensureRouteAccess("/pages/index/index")) {
			return;
		}

		uni.chooseImage({
			count: 5,
			sizeType: ["compressed"],
			sourceType: ["camera", "album"],
			success: async (res) => {
				uploadingImage.value = true;
				const tempFilePaths = res.tempFilePaths;

				try {
					for (let i = 0; i < tempFilePaths.length; i++) {
						const result = await uploadVisitImg(tempFilePaths[i]);
						images.value.push(result.url);
					}
				} catch (err) {
					uni.showToast({
						title: "部分图片上传失败",
						icon: "none"
					});
				} finally {
					uploadingImage.value = false;
				}
			},
		});
	};

	const removeImage = (idx) => {
		images.value.splice(idx, 1);
	};

	const openPreview = (idx) => {
		uni.previewImage({
			urls: images.value,
			current: idx,
		});
	};

	const handleSubmit = async () => {
		if (!ensureRouteAccess("/pages/index/index")) {
			return;
		}

		if (!checkedIn.value) {
			uni.showToast({
				title: "请先签到打卡",
				icon: "none"
			});
			return;
		}
		if (submitting.value) return;
		if (uploadingImage.value) {
			uni.showToast({
				title: "图片上传中，请稍后",
				icon: "none"
			});
			return;
		}

		if (!visitPurpose.value) {
			uni.showToast({
				title: "请选择拜访目的",
				icon: "none"
			});
			return;
		}
		if (!customerName.value.trim()) {
			uni.showToast({
				title: "请填写公司名称",
				icon: "none"
			});
			return;
		}
		submitting.value = true;

		try {
			await createCustomerVisit({
				customerName: customerName.value.trim(),
				inviter: inviter.value.trim(),
				checkInLat: currentLat.value,
				checkInLng: currentLng.value,
				province: currentProvince.value,
				city: currentCity.value,
				area: currentArea.value,
				detailAddress: currentDetailAddress.value,
				images: JSON.stringify(images.value),
				visitPurpose: visitPurpose.value.trim(),
				remark: remark.value.trim(),
			});
			submitted.value = true;
		} catch (err) {
			uni.showToast({
				title: err.message || "提交失败",
				icon: "none"
			});
		} finally {
			submitting.value = false;
		}
	};

	const fetchHistory = async () => {
		try {
			const res = await getCustomerVisits({
				page: 1,
				pageSize: 5,
			});
			if (res && res.items) {
				visitList.value = res.items;
			}
		} catch (err) {
			console.error("加载记录失败", err);
		}
	};

	const formatTimeOnly = (dateStr) => {
		if (!dateStr) return "";
		const d = new Date(dateStr);
		return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
	};

	const resetAll = () => {
		checkedIn.value = false;
		checkedInTime.value = "";
		submitted.value = false;
		customerName.value = "";
		inviter.value = "";
		inviterIndex.value = -1;
		visitPurpose.value = "";
		purposeIndex.value = -1;
		remark.value = "";
		images.value = [];
		getLocation();
		fetchHistory();
	};

	const goToRecords = () => {
		uni.navigateTo({
			url: "/pages/records/records",
		});
	};

	const handleUserCardClick = () => {
		if (!user.value || !hasLoginToken()) {
			redirectToLogin();
		}
	};
</script>

<style lang="scss" scoped>
	page {
		height: 100%;
		min-height: 100%;
		overflow-y: auto;
		-webkit-overflow-scrolling: touch;
	}

	.page {
		height: 100vh;
		height: 100dvh;
		min-height: 100vh;
		min-height: 100dvh;
		background: url("/static/login.png") no-repeat top center;
		background-size: cover;
	}

	.main-scroll {
		height: 100%;
	}

	/* User Card */
	.user-card {
		margin-top: calc(var(--status-bar-height) + 16px);
		border-radius: 12px;
		padding: 16px 20px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
	}

	.uc-left {
		display: flex;
		align-items: center;
		gap: 14px;
	}

	.uc-avatar {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		background: #f0f2f5;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.uc-avatar-img {
		width: 100%;
		height: 100%;
	}

	.uc-info {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.uc-name {
		font-size: 18px;
		font-weight: bold;
		color: #333;
	}

	.uc-sub {
		font-size: 13px;
		color: #3b82f6;
	}

	.uc-right {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 6px;
	}

	.uc-action-txt {
		font-size: 12px;
		color: #333;
	}

	/* Main Body */
	.main-body {
		display: block;
		min-height: 100%;
		padding: 0 16px;
		padding-bottom: calc(20px + constant(safe-area-inset-bottom));
		padding-bottom: calc(20px + env(safe-area-inset-bottom));
		box-sizing: border-box;
	}

	.center-all {
		justify-content: center;
		align-items: center;
	}

	/* Check-in Area */
	.ck-area {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 28px 0 16px;
	}

	.ck-btn {
		width: 150px;
		height: 150px;
		border-radius: 200px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 12px;
	}

	.ck-btn-blue {
		background: linear-gradient(180deg, #24b2ff 0%, #0060ff 100%);
		box-shadow: 0 8px 24px rgba(0, 96, 255, 0.35);
	}

	.ck-btn-blue:active {
		transform: scale(0.96);
	}

	.ck-btn-green {
		background: linear-gradient(180deg, #18e3a2 0%, #05b47a 100%);
		box-shadow: 0 8px 24px rgba(5, 180, 122, 0.35);
	}

	.ck-btn-disabled {
		background: linear-gradient(135deg, #9ca3af 0%, #6b7280 100%);
		box-shadow: 0 4px 16px rgba(107, 114, 128, 0.2);
	}

	.ck-label {
		font-size: 18px;
		font-weight: bold;
		margin-top: 4px;
		color: #fff;
		letter-spacing: 2px;
	}

	.ck-date {
		margin-top: 36rpx;
		font-size: 13px;
		color: #dbdbdb;
	}

	.ck-time {
		margin-top: 12rpx;
		font-size: 12px;
		font-weight: 600;
		color: #dbdbdb;
	}

	.ck-address-box {
		margin-top: 18px;
		display: flex;
		justify-content: center;
		align-items: center;
		width: 100%;
	}

	.ck-address-inner {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 16px;
		border-radius: 20px;
	}

	.address-txt {
		font-size: 13px;
		color: #64748b;
		flex: 1;
		min-width: 0;
		line-height: 1.5;
		white-space: normal;
		word-break: break-all;
	}



	/* Loc Err */
	.loc-err {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 12px 14px;
		background: #fef2f2;
		border: 1px solid #fecaca;
		border-radius: 10px;
		margin-bottom: 12px;
	}

	.err-txt-wrap {
		display: flex;
		align-items: flex-start;
		gap: 6px;
	}

	.err-txt {
		font-size: 13px;
		color: #b91c1c;
		flex: 1;
		line-height: 1.4;
	}

	.err-btns {
		display: flex;
		align-items: center;
		gap: 12px;
		justify-content: flex-end;
	}

	.retry-btn {
		font-size: 13px;
		color: #2563eb;
		text-decoration: underline;
	}

	.txt-gray {
		color: #666;
		text-decoration: none;
	}

	/* Card & Form */
	.card {
		margin-top: 16px;
		display: flex;
		flex-direction: column;
		gap: 20px;
		border-radius: 14px;
		padding: 24px 24rpx;
	}

	.form-section {
		display: flex;
		flex-direction: column;
		gap: 16px;
		width: 100%;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.field-row {
		display: flex;
		gap: 12px;
		width: 100%;
	}

	.purpose-field {
		flex: 0.9;
		min-width: 0;
	}

	.name-field {
		flex: 1.6;
		min-width: 0;
	}

	.lbl {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.lbl-txt {
		font-size: 14px;
		font-weight: bold;
		color: #333;
	}

	.req-star {
		color: #ef4444;
		margin-right: 2px;
	}

	/* Picker & Input */
	.picker-view {
		width: 100%;
		height: 44px;
		border-radius: 10px;
		border: 1px solid #e5e7eb;
		background: #fafafa;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 14px;
		font-size: 14px;
		color: #333;
		box-sizing: border-box;
	}

	.ipt {
		width: 100%;
		height: 44px;
		border-radius: 10px;
		border: 1px solid #e5e7eb;
		background: #fafafa;
		padding: 0 14px;
		font-size: 14px;
		color: #333;
		box-sizing: border-box;
	}

	.ipt-area {
		height: 80px;
		padding: 10px 14px;
		width: 100%;
	}

	.placeholder-color {
		color: #999;
	}

	/* Images */
	.img-row {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
	}

	.img-thumb {
		position: relative;
		width: 72px;
		height: 72px;
		border-radius: 10px;
		border: 1px solid #e5e7eb;
		overflow: hidden;
	}

	.img-item-bg {
		width: 100%;
		height: 100%;
	}

	.img-mask {
		position: absolute;
		top: 0;
		right: 0;
		width: 24px;
		height: 24px;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		border-bottom-left-radius: 8px;
	}

	.img-add {
		width: 72px;
		height: 72px;
		border-radius: 10px;
		border: 2px dashed #d1d5db;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 4px;
		background: #fafafa;
	}

	.img-add:active {
		background: #f0f6ff;
		border-color: #2563eb;
	}

	.img-add-txt {
		font-size: 10px;
		color: #999;
	}



	/* Submit */
	.submit-area {
		padding: 20px 0 12px;
		margin-top: 10px;
	}

	.submit-btn {
		width: 100%;
		height: 48px;
		border-radius: 24px;
		background: linear-gradient(180deg, #24b2ff 0%, #0060ff 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
	}

	.submit-btn:active {
		background: linear-gradient(180deg, #18e3a2 0%, #05b47a 100%);
	}

	.submit-btn-disabled {
		opacity: 0.5;
	}

	.submit-btn-txt {
		color: #fff;
		font-size: 16px;
		font-weight: bold;
	}

	/* Success Modal overlay */
	.success-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		z-index: 999;
		display: flex;
		align-items: center;
		justify-content: center;
		backdrop-filter: blur(4px);
	}

	.success-modal {
		width: 300px;
		background: #fff;
		border-radius: 20px;
		padding: 32px 24px;
		display: flex;
		flex-direction: column;
		align-items: center;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
	}

	.animate-pop {
		animation: popIn 0.35s cubic-bezier(0.175, 0.885, 0.32, 1.275) forwards;
	}

	@keyframes popIn {
		from {
			opacity: 0;
			transform: scale(0.8) translateY(20px);
		}

		to {
			opacity: 1;
			transform: scale(1) translateY(0);
		}
	}

	.icon-circle {
		width: 72px;
		height: 72px;
		border-radius: 36px;
		background: linear-gradient(135deg, #10b981 0%, #059669 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 8px 24px rgba(16, 185, 129, 0.3);
		margin-bottom: 20px;
	}

	.done-title {
		font-size: 20px;
		font-weight: bold;
		color: #0f172a;
		margin-bottom: 24px;
	}

	.done-info-box {
		width: 100%;
		background: #f8fafc;
		border-radius: 12px;
		padding: 16px;
		margin-bottom: 28px;
	}

	.info-row {
		display: flex;
		align-items: flex-start;
		gap: 8px;
	}

	.mt-6 {
		margin-top: 10px;
	}

	.info-txt {
		font-size: 13px;
		color: #475569;
		flex: 1;
		line-height: 1.4;
	}

	.line-clamp-2 {
		display: -webkit-box;
		-webkit-box-orient: vertical;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		overflow: hidden;
	}

	.ok-btn {
		width: 100%;
		height: 44px;
		border-radius: 22px;
		background: #0f172a;
		color: #fff;
		font-size: 15px;
		font-weight: 500;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.ok-btn:active {
		background: #1e293b;
	}

	/* History Section */
	.history-section {
		margin-top: 10px;
		padding: 0 4px;
	}

	.history-title {
		font-size: 15px;
		font-weight: bold;
		color: #111;
		margin-bottom: 16px;
		display: block;
	}

	.history-list {
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	.history-item {
		display: flex;
		gap: 12px;
	}

	.hi-left {
		display: flex;
		flex-direction: column;
		align-items: center;
		width: 40px;
	}

	.hi-time {
		font-size: 13px;
		font-weight: bold;
		color: #64748b;
	}

	.hi-line {
		width: 2px;
		flex: 1;
		background: #e2e8f0;
		margin: 4px 0 12px;
		border-radius: 1px;
	}

	.hi-right {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding-bottom: 20px;
	}

	.hi-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.hi-purpose {
		font-size: 14px;
		font-weight: bold;
		color: #0f172a;
		max-width: 60%;
	}

	.hi-comp {
		font-size: 12px;
		color: #3b82f6;
		background: #eff6ff;
		padding: 2px 8px;
		border-radius: 4px;
		max-width: 40%;
	}

	.hi-loc {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.hi-addr {
		font-size: 12px;
		color: #64748b;
	}

	.text-truncate {
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
</style>
