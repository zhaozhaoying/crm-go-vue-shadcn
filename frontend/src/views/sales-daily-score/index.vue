<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue"
import {
  CalendarDays,
  Crown,
  Download,
  Eye,
  Loader2,
  Medal,
  RefreshCw,
  Trophy,
} from "lucide-vue-next"
import { toast } from "vue-sonner"

import {
  getSalesDailyScoreRankings,
  refreshTodaySalesDailyScoreRankings,
  type SalesDailyScoreRankingItem,
} from "@/api/modules/salesDailyScore"
import { DatetimePicker } from "@/components/ui/datetime-picker";
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { getRequestErrorMessage } from "@/lib/http-error"
import { useAuthStore } from "@/stores/auth"
import SalesDailyScoreDetail from "@/views/sales-daily-score/detail.vue"

const authStore = useAuthStore()
const salesDailyScoresRefreshEvent = "sales-daily-scores:refresh"

const loading = ref(false)
const refreshing = ref(false)
const errorMessage = ref("")
const scoreDate = ref(getTodayDate())
const activeScoreDate = ref(getTodayDate())
const items = ref<SalesDailyScoreRankingItem[]>([])
const detailOpen = ref(false)
const selectedItem = ref<SalesDailyScoreRankingItem | null>(null)

const currentUserId = computed(() => Number(authStore.user?.id || 0))

function getTodayDate() {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, "0")
  const day = String(now.getDate()).padStart(2, "0")
  return `${year}-${month}-${day}`
}

const formatDuration = (seconds: number) => {
  const safe = Math.max(0, Math.floor(Number(seconds) || 0))
  const hours = Math.floor(safe / 3600)
  const minutes = Math.floor((safe % 3600) / 60)
  const remain = safe % 60
  if (hours > 0) return `${hours}小时${minutes}分${remain}秒`
  if (minutes > 0) return `${minutes}分${remain}秒`
  return `${remain}秒`
}

const formatDateTime = (value?: string | null) => {
  if (!value) return "-"
  try {
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) {
      return "-"
    }
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, "0")
    const day = String(date.getDate()).padStart(2, "0")
    const hours = String(date.getHours()).padStart(2, "0")
    const minutes = String(date.getMinutes()).padStart(2, "0")
    const seconds = String(date.getSeconds()).padStart(2, "0")
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
  } catch {
    return "-"
  }
}

const exportRankings = () => {
  if (items.value.length === 0) {
    toast.error("暂无可导出数据")
    return
  }

  const rows = [
    [
      "积分日期",
      "排名",
      "销售",
      "角色",
      "总积分",
      "电话积分",
      "拜访积分",
      "新增客户积分",
      "拨打数",
      "通话时长",
      "拜访数",
      "新增客户数",
      "更新时间",
    ],
    ...items.value.map((item) => [
      activeScoreDate.value,
      String(item.rank),
      item.userName || "",
      item.roleName || "",
      String(item.totalScore),
      String(item.callScore),
      String(item.visitScore),
      String(item.newCustomerScore),
      String(item.callNum),
      formatDuration(item.callDurationSecond),
      String(item.visitCount),
      String(item.newCustomerCount),
      formatDateTime(item.updatedAt),
    ]),
  ]

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
  anchor.download = `sales-daily-scores-${activeScoreDate.value || getTodayDate()}.csv`
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  window.URL.revokeObjectURL(url)
}

const fetchRankings = async () => {
  loading.value = true
  errorMessage.value = ""
  try {
    const result = await getSalesDailyScoreRankings({ scoreDate: scoreDate.value })
    items.value = result.items || []
    activeScoreDate.value = result.scoreDate || scoreDate.value
  } catch (error) {
    items.value = []
    errorMessage.value = getRequestErrorMessage(error, "加载每日排名失败")
  } finally {
    loading.value = false
  }
}

const refreshTodayRankings = async () => {
  refreshing.value = true
  loading.value = true
  errorMessage.value = ""
  try {
    await refreshTodaySalesDailyScoreRankings()
    scoreDate.value = getTodayDate()
    toast.success("今日排名已刷新")
    await fetchRankings()
  } catch (error) {
    loading.value = false
    toast.error(getRequestErrorMessage(error, "刷新今日排名失败"))
  } finally {
    refreshing.value = false
  }
}

const handleScoreDateChange = () => {
  void fetchRankings()
}

const openDetail = (item: SalesDailyScoreRankingItem) => {
  selectedItem.value = item
  detailOpen.value = true
}

const handleDetailOpenChange = (open: boolean) => {
  detailOpen.value = open
  if (open) {
    return
  }
  selectedItem.value = null
}

onMounted(() => {
  void refreshTodayRankings()
  window.addEventListener(salesDailyScoresRefreshEvent, fetchRankings)
})

onBeforeUnmount(() => {
  window.removeEventListener(salesDailyScoresRefreshEvent, fetchRankings)
})
</script>

<template>
  <div class="space-y-6">
    <Card class="shadow-sm border-border/60">
      <CardHeader>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-2">
            <Button size="sm" variant="outline" class="bg-background"
              @click="refreshTodayRankings" :disabled="refreshing">
              <Loader2 v-if="refreshing" class="h-3.5 w-3.5 animate-spin" />
              <RefreshCw v-else class="h-3.5 w-3.5" />
              刷新排名
            </Button>
            <Button size="sm" variant="outline" class="bg-background" @click="exportRankings"
              :disabled="loading || items.length === 0">
              <Download class="h-3.5 w-3.5" />
              导出
            </Button>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <DatetimePicker id="shooting-time" v-model="scoreDate" placeholder="请选择拍摄时间" :date-only="true"
              @change="handleScoreDateChange" />
          </div>
        </div>
      </CardHeader>
      <CardContent class="pt-2">
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
                  <TableHead class="w-32 whitespace-nowrap">销售</TableHead>
                  <TableHead class="w-28 whitespace-nowrap">角色</TableHead>
                  <TableHead class="w-24 whitespace-nowrap">总积分</TableHead>
                  <TableHead class="w-24 whitespace-nowrap">电话积分</TableHead>
                  <TableHead class="w-24 whitespace-nowrap">拜访积分</TableHead>
                  <TableHead class="w-24 whitespace-nowrap">新增客户积分</TableHead>
                  <TableHead class="w-20 whitespace-nowrap">拨打数</TableHead>
                  <TableHead class="w-32 whitespace-nowrap">通话时长</TableHead>
                  <TableHead class="w-20 whitespace-nowrap">拜访数</TableHead>
                  <TableHead class="w-24 whitespace-nowrap">新增客户数</TableHead>
                  <TableHead class="w-44 whitespace-nowrap">更新时间</TableHead>
                  <TableHead
                    class="sticky right-0 z-30 w-[80px] min-w-[80px] bg-muted/95 text-center border-l border-border before:absolute before:left-0 before:top-0 before:h-full before:w-px before:bg-border">
                    操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in items" :key="`${item.scoreDate}-${item.userId}`"
                  class="group transition-colors hover:bg-muted/20"
                  :class="item.userId === currentUserId ? 'bg-primary/5 ring-1 ring-inset ring-primary/15' : ''">
                  <TableCell>
                    <Badge :variant="item.rank <= 3 ? 'default' : 'outline'" class="gap-1.5">
                      <Crown v-if="item.rank === 1" class="h-3.5 w-3.5 text-amber-200" />
                      <Medal v-else-if="item.rank === 2 || item.rank === 3" class="h-3.5 w-3.5"
                        :class="item.rank === 2 ? 'text-slate-100' : 'text-orange-100'" />
                      第{{ item.rank }}名
                    </Badge>
                  </TableCell>
                  <TableCell class="font-medium">{{ item.userName || "-" }}</TableCell>
                  <TableCell class="text-muted-foreground">{{ item.roleName || "-" }}</TableCell>
                  <TableCell>
                    <span class="text-base font-semibold text-primary">{{ item.totalScore }}</span>
                  </TableCell>
                  <TableCell>{{ item.callScore }}</TableCell>
                  <TableCell>{{ item.visitScore }}</TableCell>
                  <TableCell>{{ item.newCustomerScore }}</TableCell>
                  <TableCell>{{ item.callNum }}</TableCell>
                  <TableCell>{{ formatDuration(item.callDurationSecond) }}</TableCell>
                  <TableCell>{{ item.visitCount }}</TableCell>
                  <TableCell>{{ item.newCustomerCount }}</TableCell>
                  <TableCell class="tabular-nums whitespace-nowrap text-muted-foreground">
                    {{ formatDateTime(item.updatedAt) }}
                  </TableCell>
                  <TableCell
                    class="sticky right-0 z-10 w-[80px] min-w-[80px] border-l border-border bg-background text-center before:absolute before:left-0 before:top-0 before:h-full before:w-px before:bg-border">
                    <div class="flex justify-end">
                      <Button variant="ghost" size="sm" @click="openDetail(item)">
                        详情
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
                <EmptyTablePlaceholder v-if="items.length === 0" :colspan="13" text="暂无每日排名数据" />
              </TableBody>
            </Table>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>

  <SalesDailyScoreDetail :open="detailOpen" :user-id="selectedItem?.userId" :user-name="selectedItem?.userName"
    :score-date="activeScoreDate" @update:open="handleDetailOpenChange" />
</template>
