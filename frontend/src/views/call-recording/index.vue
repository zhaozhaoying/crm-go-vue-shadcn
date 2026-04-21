<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue"
import { Loader2, RefreshCw, Search } from "lucide-vue-next"
import { toast } from "vue-sonner"

import {
  getCallRecordingAudio,
  getCallRecordings,
  syncCallRecordings,
  type CallRecording,
} from "@/api/modules/callRecording"
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Pagination } from "@/components/ui/pagination"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { getRequestErrorMessage } from "@/lib/http-error"

const loading = ref(false)
const syncing = ref(false)
const error = ref("")
const items = ref<CallRecording[]>([])
const totalCount = ref(0)
const pageIndex = ref(0)
const pageSize = ref(20)
const keyword = ref("")
const activeKeyword = ref("")
const minDuration = ref<string | number>("")
const maxDuration = ref<string | number>("")
const activeMinDuration = ref<number | undefined>(undefined)
const activeMaxDuration = ref<number | undefined>(undefined)

const audioUrls = ref<Record<string, string>>({})
const audioLoadingId = ref("")

const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize.value)))

const formatDateTime = (value?: number) => {
  const safe = Number(value || 0)
  if (!safe) return "-"
  const date = new Date(safe)
  if (Number.isNaN(date.getTime())) return "-"
  return date.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  })
}

const formatDateTimeParts = (value?: number) => {
  const formatted = formatDateTime(value)
  if (formatted === "-") {
    return { date: "-", time: "" }
  }
  const [date = "-", time = ""] = formatted.split(" ")
  return { date, time }
}

const formatDuration = (seconds?: number) => {
  const safe = Math.max(0, Math.floor(Number(seconds) || 0))
  const minutes = Math.floor(safe / 60)
  const remain = safe % 60
  if (minutes > 0) return `${minutes}分${remain}秒`
  return `${remain}秒`
}

const normalizeDisplayText = (value?: string | null) => {
  const trimmed = String(value ?? "").trim()
  if (!trimmed) return "-"
  const lowered = trimmed.toLowerCase()
  if (lowered === "null" || lowered === "undefined") return "-"
  return trimmed
}

const pickDisplayText = (...values: Array<string | null | undefined>) => {
  for (const value of values) {
    const normalized = normalizeDisplayText(value)
    if (normalized !== "-") return normalized
  }
  return "-"
}

const normalizeDuration = (value: string | number) => {
  const trimmed = String(value ?? "").trim()
  if (!trimmed) return undefined
  const parsed = Number(trimmed)
  if (!Number.isFinite(parsed)) return undefined
  return Math.max(0, Math.floor(parsed))
}

const fetchRecords = async () => {
  loading.value = true
  error.value = ""
  try {
    const result = await getCallRecordings({
      page: pageIndex.value + 1,
      pageSize: pageSize.value,
      keyword: activeKeyword.value || undefined,
      minDuration: activeMinDuration.value,
      maxDuration: activeMaxDuration.value,
    })
    items.value = result.items || []
    totalCount.value = result.total || 0
  } catch (err) {
    items.value = []
    totalCount.value = 0
    error.value = getRequestErrorMessage(err, "加载通话录音失败")
  } finally {
    loading.value = false
  }
}

const refreshList = async () => {
  syncing.value = true
  try {
    const result = await syncCallRecordings({
      minTime: "60",
      limit: 40,
    })
    toast.success(`已同步 ${result.totalSaved} 条通话录音`)
    pageIndex.value = 0
    await fetchRecords()
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "同步通话录音失败"))
    await fetchRecords()
  } finally {
    syncing.value = false
  }
}

const handleSearch = () => {
  activeKeyword.value = keyword.value.trim()
  activeMinDuration.value = normalizeDuration(minDuration.value)
  activeMaxDuration.value = normalizeDuration(maxDuration.value)
  if (
    activeMinDuration.value !== undefined &&
    activeMaxDuration.value !== undefined &&
    activeMinDuration.value > activeMaxDuration.value
  ) {
    toast.error("最小时长不能大于最大时长")
    return
  }
  pageIndex.value = 0
  void fetchRecords()
}

const clearSearch = () => {
  keyword.value = ""
  activeKeyword.value = ""
  minDuration.value = ""
  maxDuration.value = ""
  activeMinDuration.value = undefined
  activeMaxDuration.value = undefined
  pageIndex.value = 0
  void fetchRecords()
}

const handlePageChange = (nextPage: number) => {
  if (nextPage === pageIndex.value) return
  pageIndex.value = nextPage
  void fetchRecords()
}

const handlePageSizeChange = (nextPageSize: number) => {
  const changed = nextPageSize !== pageSize.value
  pageSize.value = nextPageSize
  pageIndex.value = 0
  if (changed) {
    void fetchRecords()
  }
}

const loadAudio = async (record: CallRecording) => {
  if (!record.preRecordUrl) {
    toast.error("当前记录没有录音地址")
    return
  }
  if (audioUrls.value[record.id] || audioLoadingId.value === record.id) {
    return
  }

  audioLoadingId.value = record.id
  try {
    const blob = await getCallRecordingAudio(record.id)
    const objectUrl = URL.createObjectURL(blob)
    audioUrls.value = {
      ...audioUrls.value,
      [record.id]: objectUrl,
    }
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "加载录音文件失败"))
  } finally {
    audioLoadingId.value = ""
  }
}

onMounted(() => {
  void fetchRecords()
})

onBeforeUnmount(() => {
  Object.values(audioUrls.value).forEach((url) => URL.revokeObjectURL(url))
})
</script>

<template>
  <div class="w-full flex flex-col gap-4 lg:gap-6">
    <Card class="shadow-sm border-border/60">
      <CardHeader class="border-b space-y-3">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div class="flex items-center gap-3">
            <Button size="sm" variant="outline" :disabled="syncing" @click="refreshList">
              <Loader2 v-if="syncing" class="h-4 w-4 animate-spin" />
              <RefreshCw v-else class="h-4 w-4" />
              <span>{{ syncing ? "同步中" : "同步最新" }}</span>
            </Button>
          </div>
          <div class="flex flex-wrap items-center justify-end gap-3">
            <Input
              v-model="keyword"
              placeholder="姓名/电话"
              class="h-9 w-full sm:w-56"
              @keyup.enter="handleSearch"
            />
            <Input
              v-model="minDuration"
              type="number"
              min="0"
              step="1"
              placeholder="最小时长(秒)"
              class="h-9 w-full sm:w-36"
              @keyup.enter="handleSearch"
            />
            <Input
              v-model="maxDuration"
              type="number"
              min="0"
              step="1"
              placeholder="最大时长(秒)"
              class="h-9 w-full sm:w-36"
              @keyup.enter="handleSearch"
            />
            <Button size="sm" @click="handleSearch">
              <Search class="h-4 w-4" />
              <span>搜索</span>
            </Button>
            <Button size="sm" variant="outline" @click="clearSearch">
              <RefreshCw class="h-4 w-4" />
              <span>重置</span>
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent class="pt-4">
        <div class="overflow-hidden rounded-lg bg-background">
          <div v-if="loading" class="flex items-center justify-center py-24">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
          </div>

          <div v-else-if="error" class="py-20 text-center text-destructive">
            {{ error }}
          </div>

          <div v-else class="overflow-x-auto">
            <Table class="w-full min-w-[920px] table-fixed">
              <TableHeader class="sticky top-0 z-20 bg-muted/40">
                <TableRow>
                  <TableHead class="w-[190px] text-center">坐席信息</TableHead>
                  <TableHead class="w-[340px] text-center">主叫 / 被叫</TableHead>
                  <TableHead class="w-[300px] text-center">录音</TableHead>
                  <TableHead class="w-[190px] text-center">开始时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow
                  v-for="record in items"
                  :key="record.id"
                  class="group hover:bg-muted/30 transition-colors"
                >
                  <TableCell class="align-middle py-4">
                    <div class="mx-auto flex w-[146px] flex-col rounded-2xl bg-muted/[0.14] px-4 py-3 text-left">
                      <div class="truncate text-sm font-semibold text-foreground">
                        {{ pickDisplayText(record.realName, record.mobile) }}
                      </div>
                      <div class="mt-1 truncate text-xs text-muted-foreground tabular-nums">
                        {{ pickDisplayText(record.mobile) }}
                      </div>
                    </div>
                  </TableCell>

                  <TableCell class="align-middle py-4">
                    <div class="mx-auto flex w-[312px] flex-col gap-2 text-sm">
                      <div class="grid grid-cols-[42px_102px_minmax(0,1fr)_auto] items-center gap-2 rounded-2xl bg-muted/[0.12] px-3 py-2">
                        <span class="inline-flex items-center justify-center rounded-full bg-background/90 px-2 py-1 text-[11px] font-medium text-muted-foreground">
                          主叫
                        </span>
                        <span class="truncate font-medium tabular-nums text-foreground">
                          {{ pickDisplayText(record.telA, record.mobile) }}
                        </span>
                        <span class="truncate text-muted-foreground">
                          {{ pickDisplayText(record.callerAttr) }}
                        </span>
                        <Badge variant="secondary" class="justify-self-start whitespace-nowrap border-transparent px-2.5">
                          {{
                            pickDisplayText(record.finishStatusName) !== "-"
                              ? record.finishStatusName
                              : `结果 ${record.finishStatus}`
                          }}
                        </Badge>
                      </div>

                      <div class="grid grid-cols-[42px_102px_minmax(0,1fr)_auto] items-center gap-2 rounded-2xl bg-muted/[0.08] px-3 py-2">
                        <span class="inline-flex items-center justify-center rounded-full bg-background/90 px-2 py-1 text-[11px] font-medium text-muted-foreground">
                          被叫
                        </span>
                        <span class="truncate font-medium tabular-nums text-foreground">
                          {{ pickDisplayText(record.telB, record.phone) }}
                        </span>
                        <span class="truncate text-muted-foreground">
                          {{ pickDisplayText(record.calleeAttr) }}
                        </span>
                        <Badge variant="outline" class="justify-self-start whitespace-nowrap border-transparent bg-background/90 px-2.5">
                          {{
                            pickDisplayText(record.callStatusName) !== "-"
                              ? record.callStatusName
                              : `状态 ${record.callStatus}`
                          }}
                        </Badge>
                      </div>
                    </div>
                  </TableCell>

                  <TableCell class="align-middle py-4">
                    <div class="mx-auto flex w-[272px] flex-col items-center gap-2 rounded-2xl bg-muted/[0.14] px-4 py-3">
                      <div class="min-h-6 text-sm font-semibold text-foreground tabular-nums">
                        {{ formatDuration(record.duration) }}
                      </div>
                      <audio
                        v-if="record.preRecordUrl"
                        class="h-10 w-full shrink-0"
                        controls
                        preload="none"
                        :src="record.preRecordUrl"
                      />
                      <span v-else class="text-sm text-muted-foreground">暂无录音</span>
                    </div>
                  </TableCell>

                  <TableCell class="align-middle py-4">
                    <div class="mx-auto flex w-[156px] flex-col items-center rounded-2xl bg-muted/[0.14] px-3 py-3 text-center">
                      <div class="text-sm font-medium text-foreground tabular-nums whitespace-nowrap">
                        {{ formatDateTimeParts(record.startTime || record.createTime).date }}
                      </div>
                      <div class="mt-1 text-xs text-muted-foreground tabular-nums whitespace-nowrap">
                        {{ formatDateTimeParts(record.startTime || record.createTime).time || "--:--:--" }}
                      </div>
                    </div>
                  </TableCell>
                </TableRow>

                <EmptyTablePlaceholder v-if="items.length === 0" :colspan="4" text="暂无通话录音数据" />
              </TableBody>
            </Table>
          </div>
        </div>

        <div class="mt-4">
          <Pagination :current-page="pageIndex" :total-pages="totalPages" :page-size="pageSize"
            :page-size-options="[20, 40, 60, 100]" :show-selection="false" :total-count="totalCount"
            @update:current-page="handlePageChange" @update:page-size="handlePageSizeChange" />
        </div>
      </CardContent>
    </Card>
  </div>
</template>
