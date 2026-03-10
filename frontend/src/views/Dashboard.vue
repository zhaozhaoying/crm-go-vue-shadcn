<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import {
  ArrowDownRight,
  ArrowUpRight,
  CreditCard,
  DollarSign,
  FileText,
  Loader2,
  TrendingUp,
  Users,
} from "lucide-vue-next"

import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { getDashboardOverview } from "@/api/modules/dashboard"
import { getRequestErrorMessage } from "@/lib/http-error"
import type {
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

const loading = ref(false)
const errorMessage = ref("")
const overview = ref<DashboardOverview | null>(null)

const chartWidth = 900
const chartHeight = 280
const chartPadding = { top: 28, right: 30, bottom: 38, left: 30 }

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
      up: revenue.changeRate >= 0,
      icon: DollarSign,
      desc: `上月 ${formatCurrency(revenue.previous)}`,
    },
    {
      title: "本月新客户",
      value: formatCount(customers.current),
      change: formatChange(customers.changeRate),
      up: customers.changeRate >= 0,
      icon: Users,
      desc: `上月 ${formatCount(customers.previous)}`,
    },
    {
      title: "本月新增商机",
      value: formatCount(opportunities.current),
      change: formatChange(opportunities.changeRate),
      up: opportunities.changeRate >= 0,
      icon: CreditCard,
      desc: `上月 ${formatCount(opportunities.previous)}`,
    },
    {
      title: "本月转化率",
      value: formatRate(conversion.current),
      change: formatChange(conversion.changeRate),
      up: conversion.changeRate >= 0,
      icon: TrendingUp,
      desc: `上月 ${formatRate(conversion.previous)}`,
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

const countRenderPoints = computed(() =>
  countPoints.value.map((point, index) => {
    const amountPoint = amountPoints.value[index]
    if (!amountPoint) return point

    const isOverlapping = Math.abs(point.y - amountPoint.y) < 10
    if (!isOverlapping) return point

    return {
      ...point,
      y: Math.max(chartPadding.top + 6, point.y - 10),
    }
  }),
)

const amountPath = computed(() => buildSmoothPath(amountPoints.value))
const countPath = computed(() => buildSmoothPath(countRenderPoints.value))

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
  const tooltipWidth = 220
  const horizontalPadding = 20
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
  } catch (error) {
    errorMessage.value = getRequestErrorMessage(error, "加载仪表盘数据失败")
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void fetchDashboardData()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">仪表盘</h1>
      <p class="text-sm text-muted-foreground">业务数据概览与近期动态</p>
    </div>

    <Card v-if="errorMessage" class="border-red-200 bg-red-50/40">
      <CardContent class="pt-6">
        <p class="text-sm text-red-600">{{ errorMessage }}</p>
      </CardContent>
    </Card>

    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <Card v-for="stat in stats" :key="stat.title" class="shadow-sm">
        <CardHeader class="flex flex-row items-center justify-between pb-2 space-y-0">
          <CardTitle class="text-sm font-medium text-muted-foreground">{{ stat.title }}</CardTitle>
          <component :is="stat.icon" class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ stat.value }}</div>
          <div class="mt-1 flex items-center gap-1">
            <ArrowUpRight v-if="stat.up" class="h-3 w-3 text-emerald-600" />
            <ArrowDownRight v-else class="h-3 w-3 text-red-500" />
            <span class="text-xs" :class="stat.up ? 'text-emerald-600' : 'text-red-500'">
              {{ stat.change }}
            </span>
            <span class="text-xs text-muted-foreground">{{ stat.desc }}</span>
          </div>
        </CardContent>
      </Card>
    </div>

    <Card class="shadow-sm">
      <CardHeader class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <CardTitle class="text-base">业绩概览</CardTitle>
          <CardDescription>近 12 个月合同金额与合同数量趋势</CardDescription>
        </div>
        <div class="flex flex-wrap items-center gap-3 text-xs">
          <div class="rounded-full border bg-muted/30 px-3 py-1.5 text-muted-foreground">
            合同总金额 {{ amountSummary }}
          </div>
          <div class="rounded-full border bg-muted/30 px-3 py-1.5 text-muted-foreground">
            合同总数量 {{ countSummary }}
          </div>
        </div>
      </CardHeader>
      <CardContent>
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

        <div v-else class="space-y-4">
          <div class="flex items-center gap-5 text-sm">
            <div class="flex items-center gap-2">
              <span class="h-2.5 w-2.5 rounded-full bg-primary" />
              <span class="text-muted-foreground">合同金额</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="h-2.5 w-2.5 rounded-full bg-amber-400" />
              <span class="text-muted-foreground">合同数量</span>
            </div>
          </div>

          <div
            class="relative rounded-xl border bg-muted/20 p-4"
            @mouseleave="setActiveTrendIndex(null)"
          >
            <div class="pointer-events-none absolute inset-x-4 top-4 bottom-10">
              <div
                v-for="index in 5"
                :key="index"
                class="absolute inset-x-0 border-t border-dashed border-border/70"
                :style="{ top: `${(index - 1) * 25}%` }"
              />
            </div>

            <svg
              :viewBox="`0 0 ${chartWidth} ${chartHeight}`"
              class="h-[320px] w-full"
              fill="none"
              preserveAspectRatio="none"
            >
              <path
                :d="amountPath"
                stroke="hsl(var(--primary))"
                stroke-width="3.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                :d="countPath"
                stroke="#FBBF24"
                stroke-width="3"
                stroke-dasharray="6 5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />

              <g v-for="point in countRenderPoints" :key="`count-${point.label}`">
                <circle
                  :cx="point.x"
                  :cy="point.y"
                  :r="activeTrendData?.item.label === point.label ? 5.2 : 4"
                  fill="white"
                  stroke="#FBBF24"
                  stroke-width="2"
                />
              </g>

              <g v-for="point in amountPoints" :key="`amount-${point.label}`">
                <circle
                  :cx="point.x"
                  :cy="point.y"
                  :r="activeTrendData?.item.label === point.label ? 5.8 : 4.5"
                  fill="white"
                  stroke="hsl(var(--primary))"
                  stroke-width="2.4"
                />
              </g>

              <g v-for="(point, index) in amountPoints" :key="`label-${point.label}`">
                <text
                  :x="point.x"
                  :y="chartHeight - 10"
                  text-anchor="middle"
                  class="fill-muted-foreground text-[11px]"
                >
                  {{ point.label }}
                </text>
                <rect
                  :x="point.x - 32"
                  y="0"
                  width="64"
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
                  stroke="rgba(148, 163, 184, 0.55)"
                  stroke-width="1"
                  stroke-dasharray="4 4"
                />
              </g>
            </svg>

            <div
              v-if="activeTrendData"
              class="pointer-events-none absolute z-10 w-52 rounded-xl border border-border/80 bg-background/95 p-3 shadow-xl backdrop-blur"
              :style="trendTooltipStyle"
            >
              <p class="mb-2 text-xs font-semibold text-muted-foreground">
                {{ activeTrendData.item.label }}
              </p>
              <div class="space-y-2">
                <div class="flex items-center justify-between gap-3">
                  <span class="flex items-center gap-2 text-xs text-muted-foreground">
                    <span class="h-2 w-2 rounded-full bg-primary" />
                    合同金额
                  </span>
                  <span class="text-sm font-semibold text-foreground">
                    {{ formatCurrency(activeTrendData.item.amount) }}
                  </span>
                </div>
                <div class="flex items-center justify-between gap-3">
                  <span class="flex items-center gap-2 text-xs text-muted-foreground">
                    <span class="h-2 w-2 rounded-full bg-amber-400" />
                    合同数量
                  </span>
                  <span class="text-sm font-semibold text-foreground">
                    {{ formatCount(activeTrendData.item.count) }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <div class="grid gap-4 lg:grid-cols-2">
      <Card class="shadow-sm">
        <CardHeader>
          <CardTitle class="text-base">近期成交</CardTitle>
          <CardDescription>
            {{ loading ? "正在加载..." : `最近 ${recentDeals.length} 笔合同记录` }}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div class="space-y-5">
            <div
              v-for="sale in recentDeals"
              :key="sale.id"
              class="flex items-center gap-3"
            >
              <Avatar class="h-9 w-9">
                <AvatarFallback class="bg-primary/10 text-primary text-xs font-semibold">
                  {{ (sale.customerName || "客").charAt(0) }}
                </AvatarFallback>
              </Avatar>
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium leading-none">
                  {{ sale.customerName || "未知客户" }}
                </p>
                <p class="mt-0.5 truncate text-xs text-muted-foreground">
                  {{ sale.contractName || sale.customerEmail || "-" }}
                </p>
              </div>
              <div class="text-right">
                <p class="text-sm font-semibold tabular-nums">
                  {{ formatCurrency(sale.amount) }}
                </p>
                <p class="text-xs text-muted-foreground">
                  {{ formatRelativeTime(sale.createdAt) }}
                </p>
              </div>
            </div>
            <p
              v-if="!loading && recentDeals.length === 0"
              class="text-sm text-muted-foreground"
            >
              暂无成交记录
            </p>
          </div>
        </CardContent>
      </Card>

      <Card class="shadow-sm">
        <CardHeader>
          <CardTitle class="text-base">近期动态</CardTitle>
          <CardDescription>销售与运营团队的最新跟进记录</CardDescription>
        </CardHeader>
        <CardContent>
          <div class="space-y-4">
            <div
              v-for="activity in recentActivities"
              :key="`${activity.type}-${activity.id}`"
              class="flex items-center gap-3"
            >
              <Avatar class="h-8 w-8">
                <AvatarFallback class="bg-muted text-muted-foreground text-xs font-medium">
                  {{ (activity.userName || "匿").charAt(0) }}
                </AvatarFallback>
              </Avatar>
              <div class="min-w-0 flex-1">
                <p class="text-sm">
                  <span class="font-medium">{{ activity.userName || "未知用户" }}</span>
                  <span class="text-muted-foreground"> {{ activity.action }} </span>
                  <span class="font-medium">{{ activity.target || "-" }}</span>
                </p>
                <p
                  v-if="activity.content"
                  class="mt-1 truncate text-xs text-muted-foreground"
                >
                  {{ activity.content }}
                </p>
              </div>
              <span class="shrink-0 text-xs text-muted-foreground">
                {{ formatRelativeTime(activity.createdAt) }}
              </span>
            </div>
            <p
              v-if="!loading && recentActivities.length === 0"
              class="text-sm text-muted-foreground"
            >
              暂无动态记录
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
