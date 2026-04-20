<script setup lang="ts">
import { computed, ref, watch } from "vue"
import {
  CalendarDays,
  Loader2,
  Trophy,
} from "lucide-vue-next"

import {
  getTelemarketingDailyScoreDetail,
  type TelemarketingDailyScoreDetail,
} from "@/api/modules/telemarketingDailyScore"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { getRequestErrorMessage } from "@/lib/http-error"

const props = defineProps<{
  open: boolean
  seatWorkNumber?: string
  seatName?: string
  matchedUserName?: string
  scoreDate?: string
}>()

const emits = defineEmits<{
  (e: "update:open", value: boolean): void
}>()

const loading = ref(false)
const errorMessage = ref("")
const detail = ref<TelemarketingDailyScoreDetail | null>(null)

const safeSeatWorkNumber = computed(() => String(props.seatWorkNumber || "").trim())
const safeScoreDate = computed(() => String(props.scoreDate || "").trim())

const displayUserName = computed(() => {
  const matchedUserName = detail.value?.score.matchedUserName || props.matchedUserName || ""
  if (matchedUserName) {
    return matchedUserName
  }
  const seatWorkNumber = detail.value?.score.seatWorkNumber || safeSeatWorkNumber.value
  const seatName = detail.value?.score.seatName || props.seatName || ""
  if (!seatWorkNumber) return "-"
  if (!seatName) return seatWorkNumber
  return `${seatWorkNumber} · ${seatName}`
})

const displayScoreDate = computed(() => detail.value?.scoreDate || safeScoreDate.value || "-")
const totalUsers = computed(() => detail.value?.totalUsers || 0)
const totalScore = computed(() => detail.value?.score.totalScore || 0)
const callScore = computed(() => detail.value?.score.callScore || 0)
const callScoreByCount = computed(() => detail.value?.score.callScoreByCount || 0)
const callScoreByDuration = computed(() => detail.value?.score.callScoreByDuration || 0)
const callScoreType = computed(() => detail.value?.score.callScoreType || "")
const invitationScore = computed(() => detail.value?.score.invitationScore || 0)
const newCustomerScore = computed(() => detail.value?.score.newCustomerScore || 0)
const callNum = computed(() => detail.value?.score.callNum || 0)
const callDurationSecond = computed(() => detail.value?.score.callDurationSecond || 0)
const answeredCallCount = computed(() => detail.value?.score.answeredCallCount || 0)
const missedCallCount = computed(() => detail.value?.score.missedCallCount || 0)
const answerRate = computed(() => detail.value?.score.answerRate || 0)
const serviceNumber = computed(() => detail.value?.score.serviceNumber || "-")
const groupName = computed(() => detail.value?.score.groupName || "-")
const newCustomerCount = computed(() => detail.value?.score.newCustomerCount || 0)
const invitationCount = computed(() => detail.value?.score.invitationCount || 0)
const telemarketingSeatName = computed(() => detail.value?.score.seatName || props.seatName || "-")
const telemarketingMatchedUserName = computed(() => detail.value?.score.matchedUserName || props.matchedUserName || "-")

const callScoreBasisText = computed(() => {
  const type = callScoreType.value
  if (type === "call_num") return "按接通数量计分"
  if (type === "call_duration") return "按通话时长计分"
  return "今日电话未达积分标准"
})

const rankHeroLabel = computed(() => {
  if (!detail.value?.rank) return "--"
  return `#${detail.value.rank}`
})

const rankText = computed(() => {
  if (!detail.value?.rank) return "未生成排名"
  return `第 ${detail.value.rank} 名`
})

const scoreSegments = computed(() => {
  if (!detail.value) {
    return []
  }

  const divisor = totalScore.value > 0 ? totalScore.value : 1
  return [
    {
      key: "call",
      label: "电话积分",
      value: callScore.value,
      percent: (callScore.value / divisor) * 100,
      barClass: "bg-sky-500",
    },
    {
      key: "invitation",
      label: "邀约积分",
      value: invitationScore.value,
      percent: (invitationScore.value / divisor) * 100,
      barClass: "bg-emerald-500",
    },
    {
      key: "customer",
      label: "新增客户积分",
      value: newCustomerScore.value,
      percent: (newCustomerScore.value / divisor) * 100,
      barClass: "bg-amber-500",
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

const formatAnswerRate = (value?: number) => {
  const safe = Number(value || 0)
  if (!Number.isFinite(safe) || safe <= 0) return "0%"
  return `${safe.toFixed(1)}%`
}

const resetDetailState = () => {
  loading.value = false
  errorMessage.value = ""
  detail.value = null
}

const fetchDetail = async () => {
  if (!props.open) {
    return
  }

  loading.value = true
  errorMessage.value = ""
  detail.value = null
  try {
    if (!safeSeatWorkNumber.value) {
      errorMessage.value = "无效的坐席工号"
      return
    }
    detail.value = await getTelemarketingDailyScoreDetail(safeSeatWorkNumber.value, {
      scoreDate: safeScoreDate.value || undefined,
    })
  } catch (error) {
    detail.value = null
    errorMessage.value = getRequestErrorMessage(error, "加载电销积分详情失败")
  } finally {
    loading.value = false
  }
}

const handleOpenChange = (open: boolean) => {
  emits("update:open", open)
  if (!open) {
    resetDetailState()
  }
}

watch(
  () => [props.open, safeSeatWorkNumber.value, safeScoreDate.value],
  ([open]) => {
    if (!open) {
      resetDetailState()
      return
    }
    void fetchDetail()
  },
  { immediate: true },
)
</script>

<template>
  <Sheet :open="open" @update:open="handleOpenChange">
    <SheetContent
      side="right"
      class="w-[94vw] max-w-none overflow-y-auto border-l bg-background p-0 sm:w-[720px] sm:max-w-[720px]">
      <div class="flex min-h-full flex-col bg-muted/10">
        <SheetHeader class="border-b bg-background px-6 py-5 text-left">
          <div class="flex items-start justify-between gap-4 pr-8">
            <div class="space-y-3">
              <div class="flex flex-wrap items-center gap-2">
                <Badge variant="outline" class="px-3 py-1">
                  电销积分详情
                </Badge>
                <Badge variant="outline" class="px-3 py-1">
                  <CalendarDays class="h-3.5 w-3.5" />
                  {{ displayScoreDate }}
                </Badge>
              </div>
              <div class="space-y-1">
                <SheetTitle class="text-2xl font-semibold tracking-tight">
                  {{ displayUserName }}
                </SheetTitle>
                <SheetDescription class="text-sm text-muted-foreground">
                  查看当日积分构成、行为数据和排名位置
                </SheetDescription>
              </div>
            </div>
            <div class="hidden shrink-0 rounded-xl border bg-muted/30 px-4 py-3 text-right sm:block">
              <div class="text-xs text-muted-foreground">当前排名</div>
              <div class="mt-1 text-2xl font-semibold">{{ rankHeroLabel }}</div>
            </div>
          </div>
        </SheetHeader>

        <div class="flex-1 space-y-6 px-6 py-6">
          <Card v-if="errorMessage" class="border-red-200 bg-red-50/40">
            <CardContent class="pt-6">
              <p class="text-sm text-red-600">{{ errorMessage }}</p>
            </CardContent>
          </Card>

          <div v-else-if="loading" class="flex items-center justify-center py-24">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
          </div>

          <template v-else-if="detail">
            <div class="grid gap-4 lg:grid-cols-2">
              <div class="grid h-full gap-4 sm:grid-cols-2 lg:grid-cols-1">
                <Card class="h-full border-border/60 shadow-sm">
                  <CardContent class="flex h-full flex-col justify-between px-6 py-5">
                    <div class="text-sm text-muted-foreground">总积分</div>
                    <div class="mt-2 text-4xl font-semibold tracking-tight text-foreground">
                      {{ totalScore }}
                    </div>
                    <p class="mt-2 text-sm text-muted-foreground">电话、邀约、新增客户三项累计</p>
                  </CardContent>
                </Card>
              </div>
              <div class="grid h-full gap-4 sm:grid-cols-2 lg:grid-cols-1">
                <Card class="h-full border-border/60 shadow-sm">
                  <CardContent class="flex h-full flex-col justify-between px-6 py-5">
                    <div class="flex items-center justify-between gap-3">
                      <div>
                        <div class="text-sm text-muted-foreground">今日排名</div>
                        <div class="mt-2 text-3xl font-semibold tracking-tight text-foreground">{{ rankText }}</div>
                      </div>
                      <div class="rounded-lg bg-amber-50 p-3 text-amber-600">
                        <Trophy class="h-5 w-5" />
                      </div>
                    </div>
                    <p class="mt-2 text-sm text-muted-foreground">共 {{ totalUsers }} 人参与排名</p>
                  </CardContent>
                </Card>
              </div>
            </div>

            <Card class="border-border/60 shadow-sm">
              <CardHeader class="space-y-2 pb-4">
                <div class="flex items-center justify-between gap-3">
                  <div>
                    <CardTitle class="text-lg">积分构成</CardTitle>
                    <p class="mt-1 text-sm text-muted-foreground">总积分由电话、邀约和新增客户三部分组成。</p>
                  </div>
                  <Badge variant="outline">总分 {{ totalScore }} 分</Badge>
                </div>
              </CardHeader>
              <CardContent class="space-y-5">
                <div class="space-y-3">
                  <div class="flex h-2 overflow-hidden rounded-full bg-muted">
                    <div
                      v-for="segment in scoreSegments"
                      :key="segment.key"
                      class="h-full transition-all"
                      :class="segment.barClass"
                      :style="{ width: `${segment.percent}%` }"
                    />
                  </div>
                  <div class="grid gap-3 md:grid-cols-3">
                    <div v-for="segment in scoreSegments" :key="segment.key" class="rounded-xl border bg-muted/20 p-4">
                      <div class="text-sm text-muted-foreground">{{ segment.label }}</div>
                      <div class="mt-2 flex items-end justify-between gap-2">
                        <div class="text-2xl font-semibold">{{ segment.value }}</div>
                        <div class="text-xs text-muted-foreground">{{ segment.percent.toFixed(0) }}%</div>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="grid gap-4 lg:grid-cols-[1.15fr_0.85fr]">
                  <div class="rounded-xl border bg-muted/10 p-4">
                    <div class="flex items-center justify-between gap-3">
                      <div>
                        <h3 class="text-base font-semibold">电话积分</h3>
                        <p class="mt-1 text-sm text-muted-foreground">{{ callScoreBasisText }}</p>
                      </div>
                      <Badge>{{ callScore }} 分</Badge>
                    </div>
                    <div class="mt-4 grid gap-3 sm:grid-cols-2">
                      <div class="rounded-lg border bg-background px-4 py-3">
                        <div class="text-xs text-muted-foreground">按接通数可得</div>
                        <div class="mt-1 text-lg font-semibold">{{ callScoreByCount }} 分</div>
                        <div class="mt-1 text-xs text-muted-foreground">今日接通 {{ answeredCallCount }} 通</div>
                      </div>
                      <div class="rounded-lg border bg-background px-4 py-3">
                        <div class="text-xs text-muted-foreground">按通话时长可得</div>
                        <div class="mt-1 text-lg font-semibold">{{ callScoreByDuration }} 分</div>
                        <div class="mt-1 text-xs text-muted-foreground">今日通话 {{ formatDuration(callDurationSecond) }}</div>
                      </div>
                    </div>
                  </div>

                  <div class="grid gap-3">
                    <div class="rounded-xl border bg-muted/10 px-4 py-4">
                      <div class="flex items-center justify-between gap-3">
                        <div>
                          <h3 class="text-base font-semibold">邀约积分</h3>
                          <p class="mt-1 text-sm text-muted-foreground">按当日邀约数累计</p>
                        </div>
                        <Badge variant="secondary">{{ invitationScore }} 分</Badge>
                      </div>
                      <p class="mt-3 text-sm text-muted-foreground">今日邀约 {{ invitationCount }} 次</p>
                    </div>

                    <div class="rounded-xl border bg-muted/10 px-4 py-4">
                      <div class="flex items-center justify-between gap-3">
                        <div>
                          <h3 class="text-base font-semibold">新增客户积分</h3>
                          <p class="mt-1 text-sm text-muted-foreground">按当日新增客户数累计</p>
                        </div>
                        <Badge variant="secondary">{{ newCustomerScore }} 分</Badge>
                      </div>
                      <p class="mt-3 text-sm text-muted-foreground">今日新增客户 {{ newCustomerCount }} 个</p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card class="border-border/60 shadow-sm">
              <CardHeader class="space-y-2 pb-4">
                <div class="flex items-center justify-between gap-3">
                  <div>
                    <CardTitle class="text-lg">数据概览</CardTitle>
                    <p class="mt-1 text-sm text-muted-foreground">查看坐席归属、接通情况和业务指标。</p>
                  </div>
                </div>
              </CardHeader>
              <CardContent class="grid gap-4 lg:grid-cols-2">
                <div class="rounded-xl border bg-muted/10 p-4">
                  <div class="text-sm font-semibold">通话表现</div>
                  <div class="mt-3 space-y-2 text-sm text-muted-foreground">
                    <p>拨打数：{{ callNum }}</p>
                    <p>接通数：{{ answeredCallCount }}</p>
                    <p>未接通数：{{ missedCallCount }}</p>
                    <p>接通率：{{ formatAnswerRate(answerRate) }}</p>
                    <p>通话时长：{{ formatDuration(callDurationSecond) }}</p>
                  </div>
                </div>

                <div class="rounded-xl border bg-muted/10 p-4">
                  <div class="text-sm font-semibold">人员归属</div>
                  <div class="mt-3 space-y-2 text-sm text-muted-foreground">
                    <p>本地电销：{{ telemarketingMatchedUserName }}</p>
                    <p>米话工号：{{ safeSeatWorkNumber || "-" }}</p>
                    <p>米话坐席：{{ telemarketingSeatName }}</p>
                    <p>分机号：{{ serviceNumber }}</p>
                    <p>所属组：{{ groupName }}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </template>
        </div>
      </div>
    </SheetContent>
  </Sheet>
</template>
