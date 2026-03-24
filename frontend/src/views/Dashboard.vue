<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue"
import {
  ArrowDownRight,
  ArrowUpRight,
  Clock3,
  CreditCard,
  DollarSign,
  FileText,
  Loader2,
  PhoneCall,
  Trophy,
  TrendingUp,
  Users,
} from "lucide-vue-next"

import { getDashboardOverview } from "@/api/modules/dashboard"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { getRequestErrorMessage } from "@/lib/http-error"
import { hasAnyRole, isAdminUser, isInsideSalesUser } from "@/lib/auth-role"
import { useAuthStore } from "@/stores/auth"
import type {
  DashboardAutoDropOverview,
  DashboardDailyCallSummary,
  DashboardRankingItem,
  DashboardSalesDailyPersonalOverview,
  DashboardSalesDailyMetricsOverview,
  DashboardSalesAdminOverview,
  DashboardMonthlyContractCount,
  DashboardMonthlyRevenue,
  DashboardOverview,
  DashboardRecentActivity,
  DashboardStat,
} from "@/types/dashboard"

type ChartPoint = {
  x: number
  y: number
  label: string
  value: number
}

const dashboardRefreshEvent = "dashboard:refresh"
const loading = ref(false)
const errorMessage = ref("")
const overview = ref<DashboardOverview | null>(null)
const lastLoadedBusinessDate = ref("")
const authStore = useAuthStore()

const chartWidth = 1180
const chartHeight = 320
const chartPadding = { top: 18, right: 24, bottom: 42, left: 24 }

const currencyFormatter = new Intl.NumberFormat("zh-CN", {
  style: "currency",
  currency: "CNY",
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})

const safeStat = (value?: DashboardStat): DashboardStat => ({
  current: Number.isFinite(value?.current) ? Number(value?.current) : 0,
  previous: Number.isFinite(value?.previous) ? Number(value?.previous) : 0,
  changeRate: Number.isFinite(value?.changeRate) ? Number(value?.changeRate) : 0,
})

const formatCurrency = (value: number) =>
  currencyFormatter.format(Number.isFinite(value) ? value : 0)

const formatCount = (value: number) =>
  Math.round(Number.isFinite(value) ? value : 0).toLocaleString("zh-CN")

const formatRate = (value: number) => {
  const current = Number.isFinite(value) ? value : 0
  return `${current.toFixed(1)}%`
}

const formatChange = (value: number) => {
  const current = Number.isFinite(value) ? value : 0
  const sign = current > 0 ? "+" : ""
  return `${sign}${current.toFixed(1)}%`
}

const trendDirection = (value: number): boolean | null => {
  if (value > 0) return true
  if (value < 0) return false
  return null
}

const metricValueClass = (up: boolean | null) => {
  if (up === true) return "text-emerald-600"
  if (up === false) return "text-red-600"
  return "text-foreground"
}

const metricChangeClass = (up: boolean | null) => {
  if (up === true) return "text-emerald-600"
  if (up === false) return "text-red-600"
  return "text-muted-foreground"
}

const metricTrendHeadline = (up: boolean | null) => {
  if (up === true) return "本月趋势上升"
  if (up === false) return "本月趋势回落"
  return "本月趋势持平"
}

const formatRelativeTime = (value: string) => {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return "-"

  const diff = Date.now() - date.getTime()
  if (diff < 60_000) return "刚刚"

  const minutes = Math.floor(diff / 60_000)
  if (minutes < 60) return `${minutes} 分钟前`

  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} 小时前`

  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`

  return date.toLocaleDateString("zh-CN", { month: "2-digit", day: "2-digit" })
}

const safeMonthlyRevenue = computed(() => overview.value?.monthlyRevenue ?? [])
const safeMonthlyContracts = computed(() => overview.value?.monthlyContracts ?? [])
const recentDeals = computed(() => overview.value?.recentDeals ?? [])
const recentActivities = computed(() => overview.value?.recentActivities ?? [])
const salesAdminOverview = computed<DashboardSalesAdminOverview | null>(
  () => overview.value?.salesAdminOverview ?? null,
)
const salesDailyPersonalOverview = computed<DashboardSalesDailyPersonalOverview | null>(
  () => overview.value?.salesDailyPersonalOverview ?? null,
)
const dailyCallSummary = computed<DashboardDailyCallSummary | null>(
  () => overview.value?.dailyCallSummary ?? null,
)
const salesDailyMetricsOverview = computed<DashboardSalesDailyMetricsOverview | null>(
  () => overview.value?.salesDailyMetricsOverview ?? null,
)
const autoDropOverview = computed<DashboardAutoDropOverview>(() => ({
  followUpDueSoonCount: Number(overview.value?.autoDropOverview?.followUpDueSoonCount ?? 0),
  dealDueSoonCount: Number(overview.value?.autoDropOverview?.dealDueSoonCount ?? 0),
  monthlyFollowUpDropped: Number(overview.value?.autoDropOverview?.monthlyFollowUpDropped ?? 0),
  monthlyDealDropped: Number(overview.value?.autoDropOverview?.monthlyDealDropped ?? 0),
}))
const isSalesRole = computed(() =>
  hasAnyRole(authStore.user, [
    "sales_director",
    "销售总监",
    "sales_manager",
    "销售经理",
    "sales_staff",
    "销售员工",
    "sales_inside",
    "sale_inside",
    "销售",
    "inside销售",
    "电销员工",
    "sales_outside",
    "sale_outside",
    "outside销售",
  ]),
)
const isSalesLeader = computed(() =>
  hasAnyRole(authStore.user, [
    "sales_director",
    "销售总监",
    "sales_manager",
    "销售经理",
  ]),
)
const isInsideSalesRole = computed(() => isInsideSalesUser(authStore.user))
const isGlobalSalesMetricViewer = computed(() =>
  isAdminUser(authStore.user) ||
  hasAnyRole(authStore.user, ["finance_manager", "finance", "财务经理", "财务"]),
)
const showSalesDailyMetricsOverview = computed(
  () =>
    (isGlobalSalesMetricViewer.value ||
      (isSalesRole.value && !isInsideSalesRole.value)) &&
    !!salesDailyMetricsOverview.value,
)
const showSalesAdminOverview = computed(
  () => (isAdminUser(authStore.user) || isSalesRole.value) && !!salesAdminOverview.value,
)
const showSalesDailyPersonalOverview = computed(
  () => isSalesRole.value && !!salesDailyPersonalOverview.value,
)
const showDailyCallSummary = computed(
  () => isAdminUser(authStore.user) && !!dailyCallSummary.value,
)
const showSalesRankings = computed(
  () => (isAdminUser(authStore.user) || isSalesLeader.value) && !!salesAdminOverview.value,
)
const todayNewCustomerRanks = computed<DashboardRankingItem[]>(
  () => salesAdminOverview.value?.todayNewCustomerRanks ?? [],
)
const todayFollowRecordRanks = computed<DashboardRankingItem[]>(
  () => salesAdminOverview.value?.todayFollowRecordRanks ?? [],
)
const salesCustomerRankTitle = computed(() =>
  isAdminUser(authStore.user) ? "销售今日新增客户排名" : "部门今日新增客户排名",
)
const salesCustomerRankDescription = computed(() =>
  isAdminUser(authStore.user)
    ? "按客户归属销售统计今日新增客户数量"
    : "按当前部门销售统计今日新增客户数量",
)
const salesFollowRankTitle = computed(() =>
  isAdminUser(authStore.user) ? "今日跟进记录数量排名" : "部门今日跟进记录排名",
)
const salesFollowRankDescription = computed(() =>
  isAdminUser(authStore.user)
    ? "按销售今日新增的跟进记录数量排序"
    : "按当前部门销售今日新增的跟进记录数量排序",
)

const stats = computed(() => {
  const data = overview.value
  const revenue = safeStat(data?.revenue)
  const customers = safeStat(data?.newCustomers)
  const opportunities = safeStat(data?.newOpportunities)
  const conversion = safeStat(data?.conversionRate)

  return [
    {
      title: "本月签约额",
      value: formatCurrency(revenue.current),
      change: formatChange(revenue.changeRate),
      up: trendDirection(revenue.changeRate),
      icon: DollarSign,
      desc: `上月 ${formatCurrency(revenue.previous)}`,
    },
    {
      title: "本月新客户",
      value: formatCount(customers.current),
      change: formatChange(customers.changeRate),
      up: trendDirection(customers.changeRate),
      icon: Users,
      desc: `上月 ${formatCount(customers.previous)}`,
    },
    {
      title: "本月新增商机",
      value: formatCount(opportunities.current),
      change: formatChange(opportunities.changeRate),
      up: trendDirection(opportunities.changeRate),
      icon: CreditCard,
      desc: `上月 ${formatCount(opportunities.previous)}`,
    },
    {
      title: "本月转化率",
      value: formatRate(conversion.current),
      change: formatChange(conversion.changeRate),
      up: trendDirection(conversion.changeRate),
      icon: TrendingUp,
      desc: `上月 ${formatRate(conversion.previous)}`,
    },
  ]
})

const salesAdminStats = computed(() => {
  const data = salesAdminOverview.value
  if (!data) return []

  const todayNewCustomers = safeStat(data.todayNewCustomers)
  const todayFollowRecords = safeStat(data.todayFollowRecords)
  const monthlyNewCustomers = safeStat(data.monthlyNewCustomers)
  const monthlyFollowRecords = safeStat(data.monthlyFollowRecords)

  return [
    {
      title: "今日新增客户",
      value: formatCount(todayNewCustomers.current),
      change: formatChange(todayNewCustomers.changeRate),
      up: trendDirection(todayNewCustomers.changeRate),
      icon: Users,
      desc: `昨日 ${formatCount(todayNewCustomers.previous)}`,
    },
    {
      title: "今日跟进数量",
      value: formatCount(todayFollowRecords.current),
      change: formatChange(todayFollowRecords.changeRate),
      up: trendDirection(todayFollowRecords.changeRate),
      icon: FileText,
      desc: `昨日 ${formatCount(todayFollowRecords.previous)}`,
    },
    {
      title: "本月客户数量",
      value: formatCount(monthlyNewCustomers.current),
      change: formatChange(monthlyNewCustomers.changeRate),
      up: trendDirection(monthlyNewCustomers.changeRate),
      icon: Users,
      desc: `上月 ${formatCount(monthlyNewCustomers.previous)}`,
    },
    {
      title: "本月跟进数量",
      value: formatCount(monthlyFollowRecords.current),
      change: formatChange(monthlyFollowRecords.changeRate),
      up: trendDirection(monthlyFollowRecords.changeRate),
      icon: FileText,
      desc: `上月 ${formatCount(monthlyFollowRecords.previous)}`,
    },
  ]
})

const formatDuration = (seconds: number) => {
  const safe = Math.max(0, Math.floor(Number(seconds) || 0))
  const hours = Math.floor(safe / 3600)
  const minutes = Math.floor((safe % 3600) / 60)
  const remain = safe % 60
  if (hours > 0) return `${hours}小时${minutes}分${remain}秒`
  if (minutes > 0) return `${minutes}分${remain}秒`
  return `${remain}秒`
}

const salesDailyMetricsStats = computed(() => {
  const data = salesDailyMetricsOverview.value
  if (!data) return []

  const callCount = safeStat(data.callCount)
  const callDuration = safeStat(data.callDuration)
  const visitCount = safeStat(data.visitCount)
  const totalScore = safeStat(data.totalScore)
  const visitTitle = isGlobalSalesMetricViewer.value ? "今日总拜访数量" : "当日总拜访数量"
  const scoreTitle = isGlobalSalesMetricViewer.value ? "今日总积分" : "当日积分"

  return [
    {
      title: "总通话量",
      value: formatCount(callCount.current),
      change: formatChange(callCount.changeRate),
      up: trendDirection(callCount.changeRate),
      icon: PhoneCall,
      desc: `昨日 ${formatCount(callCount.previous)}`,
    },
    {
      title: "总通话时长",
      value: formatDuration(callDuration.current),
      change: formatChange(callDuration.changeRate),
      up: trendDirection(callDuration.changeRate),
      icon: Clock3,
      desc: `昨日 ${formatDuration(callDuration.previous)}`,
    },
    {
      title: visitTitle,
      value: formatCount(visitCount.current),
      change: formatChange(visitCount.changeRate),
      up: trendDirection(visitCount.changeRate),
      icon: FileText,
      desc: `昨日 ${formatCount(visitCount.previous)}`,
    },
    {
      title: scoreTitle,
      value: formatCount(totalScore.current),
      change: formatChange(totalScore.changeRate),
      up: trendDirection(totalScore.changeRate),
      icon: Trophy,
      desc: `昨日 ${formatCount(totalScore.previous)}`,
    },
  ]
})

const salesDailyPersonalBanner = computed(() => {
  const data = salesDailyPersonalOverview.value
  if (!data) return null
  const targetScore = 80
  const scoreGap = Math.max(0, targetScore - Number(data.totalScore || 0))
  const scoreText = data.hasData ? `当前 ${data.totalScore} 分` : "当日积分待生成"
  const rankText =
    data.rank > 0 ? `部门第 ${data.rank} 名` : "部门排名待生成"
  const gapText = data.hasData
    ? scoreGap > 0
      ? `距 80 分还差 ${scoreGap} 分`
      : "已达到 80 分门槛"
    : "等待今日积分生成"
  const breakdownText = data.hasData
    ? `电话 ${data.callScore} + 拜访 ${data.visitScore} + 新增客户 ${data.newCustomerScore}`
    : "今日积分尚未生成"

  let encouragement = "今天继续稳扎稳打，分数很快就会抬起来。"
  if (data.rank > 0 && scoreGap <= 0) {
    encouragement = `今天状态很好，已经冲进部门前列，继续把优势拉开。`
  } else if (data.rank > 0 && scoreGap > 0) {
    encouragement = `再冲 ${scoreGap} 分就到 80 分门槛了，离上榜更近一步。`
  } else if (!data.hasData) {
    encouragement = "今天的数据还在生成中，先把跟进和新增客户做起来。"
  }

  return {
    scoreText,
    rankText,
    gapText,
    breakdownText,
    totalUsersText: data.totalUsers > 0 ? `本部门共 ${data.totalUsers} 人参与排名` : "暂无排名数据",
    encouragement,
  }
})

const dailyCallSummaryStats = computed(() => {
  const data = dailyCallSummary.value
  if (!data) return []

  return [
    {
      title: "每日电话总时长",
      value: formatDuration(data.totalCallDurationSecond),
      desc: `统计日期 ${data.scoreDate}`,
      icon: Clock3,
    },
    {
      title: "总接通时长",
      value: formatDuration(data.totalConnectedDurationSecond),
      desc: `统计日期 ${data.scoreDate}`,
      icon: PhoneCall,
    },
  ]
})

const mergedTrend = computed(() => {
  const revenueMap = new Map(
    safeMonthlyRevenue.value.map((item: DashboardMonthlyRevenue) => [
      item.label,
      item.amount,
    ]),
  )
  const contractMap = new Map(
    safeMonthlyContracts.value.map((item: DashboardMonthlyContractCount) => [
      item.label,
      item.count,
    ]),
  )
  const labels = safeMonthlyRevenue.value.map((item) => item.label)
  return labels.map((label) => ({
    label,
    amount: Number(revenueMap.get(label) || 0),
    count: Number(contractMap.get(label) || 0),
  }))
})

const maxAmount = computed(() =>
  Math.max(...mergedTrend.value.map((item) => item.amount), 0),
)
const maxCount = computed(() =>
  Math.max(...mergedTrend.value.map((item) => item.count), 0),
)

const buildPoints = (
  series: Array<{ label: string; value: number }>,
  maxValue: number,
) => {
  if (!series.length) return [] as ChartPoint[]
  const usableWidth = chartWidth - chartPadding.left - chartPadding.right
  const usableHeight = chartHeight - chartPadding.top - chartPadding.bottom
  const divisor = series.length > 1 ? series.length - 1 : 1
  const normalizedMax = maxValue > 0 ? maxValue : 1

  return series.map((item, index) => {
    const x = chartPadding.left + (usableWidth * index) / divisor
    const y =
      chartHeight -
      chartPadding.bottom -
      (Math.max(item.value, 0) / normalizedMax) * usableHeight
    return {
      x,
      y,
      label: item.label,
      value: item.value,
    }
  })
}

const buildSmoothPath = (points: ChartPoint[]) => {
  if (!points.length) return ""
  if (points.length === 1) return `M ${points[0].x} ${points[0].y}`

  let path = `M ${points[0].x} ${points[0].y}`
  for (let i = 0; i < points.length - 1; i += 1) {
    const current = points[i]
    const next = points[i + 1]
    const midX = (current.x + next.x) / 2
    path += ` C ${midX} ${current.y}, ${midX} ${next.y}, ${next.x} ${next.y}`
  }
  return path
}

const buildAreaPath = (points: ChartPoint[]) => {
  if (!points.length) return ""
  const baseline = chartHeight - chartPadding.bottom
  const linePath = buildSmoothPath(points)
  const first = points[0]
  const last = points[points.length - 1]
  return `${linePath} L ${last.x} ${baseline} L ${first.x} ${baseline} Z`
}

const amountPoints = computed(() =>
  buildPoints(
    mergedTrend.value.map((item) => ({ label: item.label, value: item.amount })),
    maxAmount.value,
  ),
)

const countPoints = computed(() =>
  buildPoints(
    mergedTrend.value.map((item) => ({ label: item.label, value: item.count })),
    maxCount.value,
  ),
)

const amountPath = computed(() => buildSmoothPath(amountPoints.value))
const countPath = computed(() => buildSmoothPath(countPoints.value))
const amountAreaPath = computed(() => buildAreaPath(amountPoints.value))
const countAreaPath = computed(() => buildAreaPath(countPoints.value))

const activeTrendIndex = ref<number | null>(null)

const setActiveTrendIndex = (index: number | null) => {
  activeTrendIndex.value = index
}

const activeTrendData = computed(() => {
  if (
    activeTrendIndex.value === null ||
    activeTrendIndex.value < 0 ||
    activeTrendIndex.value >= mergedTrend.value.length
  ) {
    return null
  }

  return {
    item: mergedTrend.value[activeTrendIndex.value],
    amountPoint: amountPoints.value[activeTrendIndex.value] ?? null,
    countPoint: countPoints.value[activeTrendIndex.value] ?? null,
  }
})

const trendTooltipStyle = computed(() => {
  const point = activeTrendData.value?.amountPoint
  if (!point) return {}
  const tooltipWidth = 236
  const horizontalPadding = 24
  const clampedX = Math.min(
    Math.max(point.x, tooltipWidth / 2 + horizontalPadding),
    chartWidth - tooltipWidth / 2 - horizontalPadding,
  )
  const placeBelow = point.y < 88
  const targetY = placeBelow ? point.y + 18 : point.y - 16

  return {
    left: `calc(${(clampedX / chartWidth) * 100}% + 0px)`,
    top: `calc(${(targetY / chartHeight) * 100}% + 0px)`,
    transform: placeBelow ? "translate(-50%, 0)" : "translate(-50%, -100%)",
  }
})

const amountSummary = computed(() =>
  formatCurrency(mergedTrend.value.reduce((sum, item) => sum + item.amount, 0)),
)
const countSummary = computed(() =>
  formatCount(mergedTrend.value.reduce((sum, item) => sum + item.count, 0)),
)

const fetchDashboardData = async () => {
  loading.value = true
  errorMessage.value = ""
  try {
    overview.value = await getDashboardOverview()
    lastLoadedBusinessDate.value = getCurrentBusinessDate()
  } catch (error) {
    errorMessage.value = getRequestErrorMessage(error, "加载仪表盘数据失败")
  } finally {
    loading.value = false
  }
}

const getCurrentBusinessDate = () =>
  new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Shanghai" }).format(new Date())

const refreshDashboard = () => {
  if (loading.value) return
  void fetchDashboardData()
}

const refreshDashboardIfStale = () => {
  if (!lastLoadedBusinessDate.value || lastLoadedBusinessDate.value !== getCurrentBusinessDate()) {
    refreshDashboard()
  }
}

const handleWindowFocus = () => {
  refreshDashboardIfStale()
}

const handleVisibilityChange = () => {
  if (document.visibilityState === "visible") {
    refreshDashboardIfStale()
  }
}

const handleDashboardRefreshEvent = () => {
  refreshDashboard()
}

onMounted(() => {
  void fetchDashboardData()
  window.addEventListener("focus", handleWindowFocus)
  window.addEventListener(dashboardRefreshEvent, handleDashboardRefreshEvent)
  document.addEventListener("visibilitychange", handleVisibilityChange)
})

onBeforeUnmount(() => {
  window.removeEventListener("focus", handleWindowFocus)
  window.removeEventListener(dashboardRefreshEvent, handleDashboardRefreshEvent)
  document.removeEventListener("visibilitychange", handleVisibilityChange)
})
</script>

<template>
  <div class="space-y-6">
    <Card v-if="errorMessage" class="border-red-200 bg-red-50/40">
      <CardContent class="pt-6">
        <p class="text-sm text-red-600">{{ errorMessage }}</p>
      </CardContent>
    </Card>

     <div
      v-if="showSalesDailyPersonalOverview && salesDailyPersonalBanner"
      class="overflow-hidden rounded-[18px] border border-[#e7dcc3] bg-[linear-gradient(135deg,#fff7e6_0%,#fffdf7_48%,#f8fbff_100%)] shadow-[0_8px_24px_rgba(15,23,42,0.06)]"
    >
      <div class="flex flex-col gap-4 px-5 py-4 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex items-start gap-3">
          <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-2xl bg-[#fff1c7] text-[#8a6a12] shadow-[inset_0_1px_0_rgba(255,255,255,0.9)]">
            <Trophy class="h-5 w-5" />
          </div>
          <div class="space-y-1">
            <p class="text-sm font-semibold tracking-[0.08em] text-[#8b6f2a]">今日积分播报</p>
            <p class="text-sm text-[#5f584a]">{{ salesDailyPersonalBanner.encouragement }}</p>
            <p class="text-xs text-[#7a7568]">{{ salesDailyPersonalBanner.breakdownText }}</p>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2 lg:justify-end">
          <div class="rounded-full border border-[#eadfca] bg-white/85 px-3 py-2 shadow-sm">
            <p class="text-[11px] text-[#8b816b]">当前积分</p>
            <p class="text-sm font-semibold text-[#171717]">{{ salesDailyPersonalBanner.scoreText }}</p>
          </div>
          <div class="rounded-full border border-[#eadfca] bg-white/85 px-3 py-2 shadow-sm">
            <p class="text-[11px] text-[#8b816b]">部门排名</p>
            <p class="text-sm font-semibold text-[#171717]">{{ salesDailyPersonalBanner.rankText }}</p>
          </div>
          <div class="rounded-full border border-[#eadfca] bg-white/85 px-3 py-2 shadow-sm">
            <p class="text-[11px] text-[#8b816b]">距离 80 分</p>
            <p class="text-sm font-semibold text-[#171717]">{{ salesDailyPersonalBanner.gapText }}</p>
          </div>
          <div class="rounded-full border border-transparent bg-[#fff3d7] px-3 py-2 text-[11px] text-[#8b6f2a]">
            {{ salesDailyPersonalBanner.totalUsersText }}
          </div>
        </div>
      </div>
    </div>

    <div class="space-y-4">
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card v-for="stat in stats" :key="stat.title" class="shadow-sm">
          <CardHeader class="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle class="text-sm font-medium text-muted-foreground">
              {{ stat.title }}
            </CardTitle>
            <div
            class="inline-flex items-center gap-1 rounded-full bg-background text-[13px] font-semibold"
            :class="metricChangeClass(stat.up)"
          >
            <ArrowUpRight v-if="stat.up === true" class="h-3.5 w-3.5" />
            <ArrowDownRight v-else-if="stat.up === false" class="h-3.5 w-3.5" />
            <span v-else class="text-muted-foreground">-</span>
            <span>{{ stat.change }}</span>
          </div>
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold" :class="metricValueClass(stat.up)">{{ stat.value }}</div>
            <div class="mt-2 flex items-center gap-1">
              <span class="text-xs text-muted-foreground">{{ stat.desc }}</span>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>

     <div v-if="showSalesAdminOverview" class="space-y-4">
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card v-for="stat in salesAdminStats" :key="stat.title" class="shadow-sm">
          <CardHeader class="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle class="text-sm font-medium text-muted-foreground">
              {{ stat.title }}
            </CardTitle>
            <div
              class="inline-flex items-center gap-1 rounded-full bg-background text-[13px] font-semibold"
              :class="metricChangeClass(stat.up)"
            >
              <ArrowUpRight v-if="stat.up === true" class="h-3.5 w-3.5" />
              <ArrowDownRight v-else-if="stat.up === false" class="h-3.5 w-3.5" />
              <span v-else class="text-muted-foreground">-</span>
              <span>{{ stat.change }}</span>
            </div>
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold" :class="metricValueClass(stat.up)">{{ stat.value }}</div>
            <div class="mt-1 flex items-center gap-1">
              <span class="text-xs text-muted-foreground">{{ stat.desc }}</span>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>

    <div v-if="showSalesDailyMetricsOverview" class="space-y-4">
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card v-for="stat in salesDailyMetricsStats" :key="stat.title" class="shadow-sm">
          <CardHeader class="flex flex-row items-center justify-between pb-2 space-y-0">
            <CardTitle class="text-sm font-medium text-muted-foreground">
              {{ stat.title }}
            </CardTitle>
            <div
              class="inline-flex items-center gap-1 rounded-full bg-background text-[13px] font-semibold"
              :class="metricChangeClass(stat.up)"
            >
              <ArrowUpRight v-if="stat.up === true" class="h-3.5 w-3.5" />
              <ArrowDownRight v-else-if="stat.up === false" class="h-3.5 w-3.5" />
              <span v-else class="text-muted-foreground">-</span>
              <span>{{ stat.change }}</span>
            </div>
          </CardHeader>
          <CardContent>
            <div class="text-2xl font-bold" :class="metricValueClass(stat.up)">{{ stat.value }}</div>
            <div class="mt-1 flex items-center gap-1">
              <span class="text-xs text-muted-foreground">{{ stat.desc }}</span>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>

   

   

    <Card class="overflow-hidden rounded-[12px] border border-[#dddddd] bg-white pt-0 shadow-[0_10px_30px_rgba(15,23,42,0.08)]">
      <CardHeader class="flex items-start gap-2 space-y-0 border-b border-[#e9e9e9] px-8 py-6 sm:flex-row">
        <CardTitle class="text-lg font-semibold">业绩概览</CardTitle>
      </CardHeader>
      <CardContent class="px-2 pt-4 sm:px-6 sm:pt-6">
        <div
          v-if="loading && !overview"
          class="flex h-[320px] items-center justify-center"
        >
          <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
        </div>

        <div
          v-else-if="mergedTrend.length === 0"
          class="flex h-[320px] items-center justify-center text-sm text-muted-foreground"
        >
          暂无趋势数据
        </div>

        <div v-else>
          <div
            class="relative overflow-hidden bg-white"
            @mouseleave="setActiveTrendIndex(null)"
          >
            <div class="pointer-events-none absolute inset-x-0 top-4 bottom-12">
              <div
                v-for="index in 4"
                :key="index"
                class="absolute inset-x-0 border-t border-[#ebebeb]"
                :style="{ top: `${(index - 1) * 25}%` }"
              />
            </div>

            <svg
              :viewBox="`0 0 ${chartWidth} ${chartHeight}`"
              class="aspect-auto h-[280px] w-full sm:h-[320px]"
              fill="none"
              preserveAspectRatio="none"
            >
              <defs>
                <linearGradient id="dashboardAmountFill" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#8f8f8f" stop-opacity="0.68" />
                  <stop offset="100%" stop-color="#8f8f8f" stop-opacity="0.06" />
                </linearGradient>
                <linearGradient id="dashboardCountFill" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="#b7b7b7" stop-opacity="0.42" />
                  <stop offset="100%" stop-color="#b7b7b7" stop-opacity="0.04" />
                </linearGradient>
              </defs>

              <path :d="amountAreaPath" fill="url(#dashboardAmountFill)" />
              <path :d="countAreaPath" fill="url(#dashboardCountFill)" />
              <path
                :d="amountPath"
                stroke="#202020"
                stroke-width="2.3"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                :d="countPath"
                stroke="#2c2c2c"
                stroke-width="2.1"
                stroke-linecap="round"
                stroke-linejoin="round"
              />

              <g v-for="point in countPoints" :key="`count-${point.label}`">
                <circle
                  :cx="point.x"
                  :cy="point.y"
                  :r="activeTrendData?.item.label === point.label ? 5.6 : 0"
                  fill="#1f1f1f"
                />
              </g>

              <g v-for="point in amountPoints" :key="`amount-${point.label}`">
                <circle
                  :cx="point.x"
                  :cy="point.y"
                  :r="activeTrendData?.item.label === point.label ? 6.6 : 0"
                  fill="#1f1f1f"
                />
              </g>

              <g v-for="(point, index) in amountPoints" :key="`label-${point.label}`">
                <text
                  :x="point.x"
                  :y="chartHeight - 8"
                  text-anchor="middle"
                  class="fill-[#878787] text-[12px]"
                >
                  {{ point.label }}
                </text>
                <rect
                  :x="point.x - 34"
                  y="0"
                  width="68"
                  :height="chartHeight"
                  fill="transparent"
                  class="cursor-crosshair"
                  @mouseenter="setActiveTrendIndex(index)"
                />
                <line
                  v-if="activeTrendIndex === index"
                  :x1="point.x"
                  :y1="chartPadding.top"
                  :x2="point.x"
                  :y2="chartHeight - chartPadding.bottom"
                  stroke="rgba(125, 125, 125, 0.5)"
                  stroke-width="1"
                  stroke-dasharray="4 6"
                  opacity="0"
                />
              </g>
            </svg>

            <div class="mt-1 flex items-center justify-center gap-6 pb-2 text-sm">
              <div class="flex items-center gap-2 text-[#666666]">
                <span class="h-2.5 w-2.5 rounded-full bg-[#202020]" />
                <span>合同金额({{ amountSummary }})</span>
              </div>
              <div class="flex items-center gap-2 text-[#666666]">
                <span class="h-2.5 w-2.5 rounded-full bg-[#767676]" />
                <span>合同数量({{ countSummary }})</span>
              </div>
            </div>

            <div
              v-if="activeTrendData"
              class="pointer-events-none absolute z-10 w-[236px] rounded-[12px] border border-[#e8e8e8] bg-white p-4"
              :style="trendTooltipStyle"
            >
              <p class="mb-3 text-[15px] font-semibold text-[#171717]">
                {{ activeTrendData.item.label }}
              </p>
              <div class="space-y-2">
                <div class="flex items-center justify-between gap-3">
                  <span class="flex items-center gap-3 text-[14px] text-[#7d7d7d]">
                    <span class="h-5 w-5 rounded-[4px] bg-[#1f1f1f]" />
                    合同金额
                  </span>
                  <span class="text-[15px] font-semibold text-[#171717]">
                    {{ formatCurrency(activeTrendData.item.amount) }}
                  </span>
                </div>
                <div class="flex items-center justify-between gap-3">
                  <span class="flex items-center gap-3 text-[14px] text-[#7d7d7d]">
                    <span class="h-5 w-5 rounded-[4px] bg-[#1f1f1f]" />
                    合同数量
                  </span>
                  <span class="text-[15px] font-semibold text-[#171717]">
                    {{ formatCount(activeTrendData.item.count) }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
