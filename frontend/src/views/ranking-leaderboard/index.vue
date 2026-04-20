<script setup lang="ts">
import { onMounted, ref, watch } from "vue"
import { useRoute, useRouter } from "vue-router"
import {
  Crown,
  Download,
  Loader2,
  Medal,
} from "lucide-vue-next"
import { toast } from "vue-sonner"

import {
  getRankingLeaderboard,
  type RankingLeaderboardItem,
} from "@/api/modules/rankingLeaderboard"
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Tabs,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"
import { getRequestErrorMessage } from "@/lib/http-error"

type Period = "all" | "month" | "week"

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const errorMessage = ref("")
const activePeriod = ref<Period>("all")
const items = ref<RankingLeaderboardItem[]>([])

const validPeriods: Period[] = ["all", "month", "week"]

const resolvePeriodFromRoute = (): Period => {
  const q = String(route.query.period || "").trim() as Period
  if (validPeriods.includes(q)) return q
  return "all"
}

const displayEmptyText = "暂无电销排名数据"

const formatDuration = (seconds: number) => {
  const safe = Math.max(0, Math.floor(Number(seconds) || 0))
  const hours = Math.floor(safe / 3600)
  const minutes = Math.floor((safe % 3600) / 60)
  const remain = safe % 60
  if (hours > 0) return `${hours}小时${minutes}分${remain}秒`
  if (minutes > 0) return `${minutes}分${remain}秒`
  return `${remain}秒`
}

const formatAnswerRate = (value?: number) => {
  const safe = Number(value || 0)
  if (!Number.isFinite(safe) || safe <= 0) return "0%"
  return `${safe.toFixed(1)}%`
}

const getTelemarketingDisplayName = (item: RankingLeaderboardItem) => {
  return item.matchedUserName || item.seatName || item.seatWorkNumber || "-"
}

const downloadCsv = (rows: string[][], filename: string) => {
  const csv = rows
    .map((row) =>
      row
        .map((cell) => `"${String(cell).replace(/"/g, '""')}"`)
        .join(","),
    )
    .join("\n")

  const blob = new Blob(["\uFEFF" + csv], {
    type: "text/csv;charset=utf-8;",
  })
  const url = window.URL.createObjectURL(blob)
  const anchor = document.createElement("a")
  anchor.href = url
  anchor.download = filename
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  window.URL.revokeObjectURL(url)
}

const exportRankings = () => {
  if (items.value.length === 0) {
    toast.error("暂无可导出数据")
    return
  }

  const rows = [
    [
      "排名",
      "工号",
      "电销",
      "所属组",
      "总积分",
      "电话积分",
      "邀约积分",
      "新增客户积分",
      "拨打数",
      "接通数",
      "接通率",
      "通话时长",
      "新增客户数",
      "邀约数",
      "计分天数",
    ],
    ...items.value.map((item) => [
      String(item.rank),
      item.seatWorkNumber || "-",
      getTelemarketingDisplayName(item),
      item.groupName || "-",
      String(item.totalScore),
      String(item.callScore),
      String(item.invitationScore),
      String(item.newCustomerScore),
      String(item.callNum),
      String(item.answeredCallCount),
      formatAnswerRate(item.answerRate),
      formatDuration(item.callDurationSecond),
      String(item.newCustomerCount),
      String(item.invitationCount),
      String(item.scoreDays),
    ]),
  ]

  downloadCsv(rows, `ranking-leaderboard-${activePeriod.value}.csv`)
}

const fetchRankings = async (period: Period) => {
  loading.value = true
  errorMessage.value = ""
  try {
    const result = await getRankingLeaderboard({ period })
    items.value = result.items || []
    return true
  } catch (error) {
    items.value = []
    errorMessage.value = getRequestErrorMessage(error, "加载排名榜单失败")
    return false
  } finally {
    loading.value = false
  }
}

const handleTabChange = (value: string | number) => {
  const period = String(value) as Period
  activePeriod.value = period
  void router.replace({ query: { ...route.query, period } })
  void fetchRankings(period)
}

watch(
  () => route.query.period,
  (newPeriod) => {
    const period = String(newPeriod || "").trim() as Period
    if (validPeriods.includes(period) && period !== activePeriod.value) {
      activePeriod.value = period
      void fetchRankings(period)
    }
  },
  { immediate: false }
)

onMounted(() => {
  const period = resolvePeriodFromRoute()
  activePeriod.value = period
  void fetchRankings(period)
})
</script>

<template>
  <div class="space-y-6">
    <Card class="border-border/60 shadow-sm">
      <CardContent class="pt-6">
        <Tabs :model-value="activePeriod" @update:model-value="handleTabChange" class="w-full">
          <div class="flex flex-wrap items-center justify-between gap-3 mb-4">
            <TabsList class="w-full max-w-md">
              <TabsTrigger class="flex-1" value="all">总排名</TabsTrigger>
              <TabsTrigger class="flex-1" value="month">月排名</TabsTrigger>
              <TabsTrigger class="flex-1" value="week">周排名</TabsTrigger>
            </TabsList>
            <Button size="sm" variant="outline" class="bg-background" @click="exportRankings" :disabled="loading || items.length === 0">
              <Download class="h-3.5 w-3.5" />
              导出
            </Button>
          </div>

          <div class="overflow-hidden rounded-lg border border-border/60 bg-background">
            <div v-if="loading" class="flex items-center justify-center py-24">
              <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
            </div>

            <div v-else-if="errorMessage" class="py-20 text-center text-destructive">
              {{ errorMessage }}
            </div>

            <div v-else class="overflow-x-auto">
              <Table class="w-max min-w-full">
                <TableHeader class="sticky top-0 z-20 bg-muted/40">
                  <TableRow>
                    <TableHead class="w-20 whitespace-nowrap">排名</TableHead>
                    <TableHead class="w-24 whitespace-nowrap">工号</TableHead>
                    <TableHead class="w-28 whitespace-nowrap">电销</TableHead>
                    <TableHead class="w-24 whitespace-nowrap">总积分</TableHead>
                    <TableHead class="w-24 whitespace-nowrap">电话积分</TableHead>
                    <TableHead class="w-24 whitespace-nowrap">邀约积分</TableHead>
                    <TableHead class="w-28 whitespace-nowrap">新增客户积分</TableHead>
                    <TableHead class="w-20 whitespace-nowrap">拨打数</TableHead>
                    <TableHead class="w-20 whitespace-nowrap">接通数</TableHead>
                    <TableHead class="w-20 whitespace-nowrap">接通率</TableHead>
                    <TableHead class="w-32 whitespace-nowrap">通话时长</TableHead>
                    <TableHead class="w-24 whitespace-nowrap">新增客户数</TableHead>
                    <TableHead class="w-20 whitespace-nowrap">邀约数</TableHead>
                    <TableHead class="w-20 whitespace-nowrap">计分天数</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow
                    v-for="item in items"
                    :key="`${item.seatWorkNumber || item.matchedUserId || item.rank}-${item.rank}`"
                    class="group transition-colors hover:bg-muted/20"
                  >
                    <TableCell>
                      <Badge :variant="item.rank <= 3 ? 'default' : 'outline'" class="gap-1.5">
                        <Crown v-if="item.rank === 1" class="h-3.5 w-3.5 text-amber-200" />
                        <Medal
                          v-else-if="item.rank === 2 || item.rank === 3"
                          class="h-3.5 w-3.5"
                          :class="item.rank === 2 ? 'text-slate-100' : 'text-orange-100'"
                        />
                        第{{ item.rank }}名
                      </Badge>
                    </TableCell>
                    <TableCell class="font-medium">{{ item.seatWorkNumber || "-" }}</TableCell>
                    <TableCell class="font-medium">{{ getTelemarketingDisplayName(item) }}</TableCell>
                    <TableCell>
                      <span class="text-base font-semibold text-primary">{{ item.totalScore }}</span>
                    </TableCell>
                    <TableCell>{{ item.callScore }}</TableCell>
                    <TableCell>{{ item.invitationScore }}</TableCell>
                    <TableCell>{{ item.newCustomerScore }}</TableCell>
                    <TableCell>{{ item.callNum }}</TableCell>
                    <TableCell>{{ item.answeredCallCount }}</TableCell>
                    <TableCell>{{ formatAnswerRate(item.answerRate) }}</TableCell>
                    <TableCell>{{ formatDuration(item.callDurationSecond) }}</TableCell>
                    <TableCell>{{ item.newCustomerCount }}</TableCell>
                    <TableCell>{{ item.invitationCount }}</TableCell>
                    <TableCell>{{ item.scoreDays }}</TableCell>
                  </TableRow>
                  <EmptyTablePlaceholder v-if="items.length === 0" :colspan="15" :text="displayEmptyText" />
                </TableBody>
              </Table>
            </div>
          </div>
        </Tabs>
      </CardContent>
    </Card>
  </div>
</template>
