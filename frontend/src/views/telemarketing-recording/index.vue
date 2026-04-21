<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from "vue"
import {
  CalendarDays,
  Clock3,
  Download,
  Loader2,
  Play,
  RefreshCw,
  Search,
  UserRound,
  Waves,
} from "lucide-vue-next"
import { toast } from "vue-sonner"

import {
  getTelemarketingRecordingDetail,
  getTelemarketingRecordings,
  syncTelemarketingRecordings,
  type TelemarketingRecording,
  type TelemarketingRecordingDetailResponse,
} from "@/api/modules/telemarketingRecording"
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Pagination } from "@/components/ui/pagination"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { hasAnyRole, isAdminUser } from "@/lib/auth-role"
import { getRequestErrorMessage } from "@/lib/http-error"
import { useAuthStore } from "@/stores/auth"

const telemarketingRecordingsRefreshEvent = "telemarketing-recordings:refresh"

const authStore = useAuthStore()

const loading = ref(false)
const syncing = ref(false)
const errorMessage = ref("")
const items = ref<TelemarketingRecording[]>([])
const totalCount = ref(0)
const pageIndex = ref(0)
const pageSize = ref(10)

const keyword = ref("")
const activeKeyword = ref("")
const startDate = ref("")
const activeStartDate = ref("")
const endDate = ref("")
const activeEndDate = ref("")
const minDuration = ref<string | number>("")
const maxDuration = ref<string | number>("")
const activeMinDuration = ref<number | undefined>(undefined)
const activeMaxDuration = ref<number | undefined>(undefined)

const detailOpen = ref(false)
const selectedRecord = ref<TelemarketingRecording | null>(null)
const detailLoading = ref(false)
const detailErrorMessage = ref("")
const detail = ref<TelemarketingRecordingDetailResponse | null>(null)
const detailAudioRef = ref<HTMLAudioElement | null>(null)

const playbackDialogOpen = ref(false)
const playbackDialogRecord = ref<TelemarketingRecording | null>(null)
const playbackDialogLoading = ref(false)
const playbackDialogErrorMessage = ref("")
const playbackDialogDetail = ref<TelemarketingRecordingDetailResponse | null>(null)
const playbackDialogAudioRef = ref<HTMLAudioElement | null>(null)

const playbackLoadingId = ref("")

let listRequestId = 0
let detailRequestId = 0

const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize.value)))
const selectedRecordId = computed(() => String(selectedRecord.value?.id || "").trim())
const selectedRecording = computed(() => detail.value?.recording || selectedRecord.value || null)
const canSync = computed(
  () =>
    isAdminUser(authStore.user) ||
    hasAnyRole(authStore.user, ["finance_manager", "finance", "财务经理", "财务"]),
)
const playbackDialogTitle = computed(() =>
  pickDisplayText(
    playbackDialogDetail.value?.playbackFilename,
    playbackDialogRecord.value?.recordFilename,
    playbackDialogRecord.value?.outlineNumber,
    "录音文件",
  ),
)

const normalizeDisplayText = (value?: string | number | null) => {
  const trimmed = String(value ?? "").trim()
  if (!trimmed) return "-"
  const lowered = trimmed.toLowerCase()
  if (lowered === "null" || lowered === "undefined") return "-"
  return trimmed
}

const pickDisplayText = (...values: Array<string | number | null | undefined>) => {
  for (const value of values) {
    const normalized = normalizeDisplayText(value)
    if (normalized !== "-") {
      return normalized
    }
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

const toTimestampMs = (value?: number | string | null) => {
  const safe = Number(value || 0)
  if (!Number.isFinite(safe) || safe <= 0) return 0
  if (safe < 1_000_000_000_000) return safe * 1000
  return safe
}

const formatDateTime = (value?: number | string | null) => {
  const timestamp = toTimestampMs(value)
  if (!timestamp) return "-"
  const date = new Date(timestamp)
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

const formatDateTimeParts = (value?: number | string | null) => {
  const formatted = formatDateTime(value)
  if (formatted === "-") {
    return { date: "-", time: "" }
  }
  const [date = "-", time = ""] = formatted.split(" ")
  return { date, time }
}

const formatDateOnly = (value?: string | null) => {
  const trimmed = String(value || "").trim()
  if (!trimmed) return "-"
  const date = new Date(trimmed)
  if (Number.isNaN(date.getTime())) return trimmed
  return date.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  })
}

const formatDuration = (seconds?: number) => {
  const safe = Math.max(0, Math.floor(Number(seconds) || 0))
  const hours = Math.floor(safe / 3600)
  const minutes = Math.floor((safe % 3600) / 60)
  const remain = safe % 60
  if (hours > 0) return `${hours}小时${minutes}分${remain}秒`
  if (minutes > 0) return `${minutes}分${remain}秒`
  return `${remain}秒`
}

const getRecordDurationText = (record?: TelemarketingRecording | null) => {
  if (!record) return "-"
  const durationText = normalizeDisplayText(record.durationText)
  if (durationText !== "-") return durationText
  const validDurationText = normalizeDisplayText(record.validDurationText)
  if (validDurationText !== "-") return validDurationText
  return formatDuration(record.durationSecond)
}

const getCallTypeLabel = (value?: number) => {
  if (value === 1) return "呼出"
  if (value === 2) return "呼入"
  if (Number(value) > 0) return `类型 ${value}`
  return "-"
}

const getAnswerResultLabel = (value?: number) => {
  if (value === 1) return "已接通"
  if (Number(value) > 0) return `结果 ${value}`
  return "-"
}

const getHangupPartyLabel = (value?: number) => {
  if (value === 1) return "坐席侧挂断"
  if (value === 2) return "客户侧挂断"
  if (Number(value) > 0) return `挂断方 ${value}`
  return "-"
}

const getRecordStatusLabel = (value?: number) => {
  if (value === 5) return "录音可用"
  if (value === 1) return "录音处理中"
  if (value === 0) return "无录音"
  if (Number(value) > 0) return `状态 ${value}`
  return "-"
}

const getAnswerBadgeVariant = (value?: number) => {
  if (value === 1) return "default"
  if (Number(value) > 0) return "secondary"
  return "outline"
}

const getRecordBadgeVariant = (value?: number) => {
  if (value === 5) return "default"
  if (value === 1) return "secondary"
  return "outline"
}

const resultSummaryText = computed(() => {
  if (loading.value) return "正在加载录音列表..."
  return `共 ${totalCount.value} 条录音`
})

const displayEmptyText = computed(() => {
  if (activeKeyword.value || activeStartDate.value || activeEndDate.value || activeMinDuration.value || activeMaxDuration.value) {
    return "当前筛选条件下暂无录音"
  }
  return "暂无电销录音数据"
})

const detailInfoRows = computed(() => {
  const record = selectedRecording.value
  if (!record) return []
  return [
    { label: "通话类型", value: getCallTypeLabel(record.callType) },
    { label: "客户号码", value: pickDisplayText(record.outlineNumber) },
    { label: "归属地", value: pickDisplayText(record.attribution, record.districtName) },
    { label: "客户姓名", value: pickDisplayText(record.customerName) },
    { label: "客户公司", value: pickDisplayText(record.customerCompany, record.enterpriseName) },
    { label: "结束原因", value: record.stopReason > 0 ? `原因 ${record.stopReason}` : "-" },
    { label: "质检评价", value: pickDisplayText(record.evaluateValue, record.cmResult, record.cmDescription) },
    { label: "通话流水", value: pickDisplayText(record.ccNumber) },
  ]
})

const detailSeatRows = computed(() => {
  const record = selectedRecording.value
  if (!record) return []
  return [
    { label: "本地用户", value: pickDisplayText(record.matchedUserName) },
    { label: "本地角色", value: pickDisplayText(record.roleName) },
    { label: "坐席工号", value: pickDisplayText(record.serviceSeatWorkNumber) },
    { label: "米话坐席", value: pickDisplayText(record.serviceSeatName) },
    { label: "坐席号码", value: pickDisplayText(record.serviceNumber) },
    { label: "所属分组", value: pickDisplayText(record.serviceGroupName, record.groupNames) },
    { label: "设备号码", value: pickDisplayText(record.serviceDeviceNumber, record.switchNumber) },
    { label: "企业名称", value: pickDisplayText(record.enterpriseName) },
  ]
})

const detailTimelineRows = computed(() => {
  const record = selectedRecording.value
  if (!record) return []
  return [
    { label: "发起时间", value: formatDateTime(record.initiateTime) },
    { label: "振铃时间", value: formatDateTime(record.ringTime) },
    { label: "接通时间", value: formatDateTime(record.confirmTime || record.conversationTime) },
    { label: "结束时间", value: formatDateTime(record.disconnectTime) },
    { label: "通话时长", value: getRecordDurationText(record) },
    { label: "客户振铃", value: formatDuration(record.customerRingDuration) },
    { label: "坐席振铃", value: formatDuration(record.seatRingDuration) },
    { label: "播放链接过期", value: detail.value?.playbackExpiresAt ? formatDateTime(detail.value.playbackExpiresAt) : "-" },
    { label: "上游创建时间", value: formatDateOnly(record.remoteCreatedAt) },
    { label: "上游更新时间", value: formatDateOnly(record.remoteUpdatedAt) },
  ]
})

const fetchRecords = async () => {
  const requestId = ++listRequestId
  loading.value = true
  errorMessage.value = ""

  try {
    const result = await getTelemarketingRecordings({
      page: pageIndex.value + 1,
      pageSize: pageSize.value,
      keyword: activeKeyword.value || undefined,
      startDate: activeStartDate.value || undefined,
      endDate: activeEndDate.value || undefined,
      minDuration: activeMinDuration.value,
      maxDuration: activeMaxDuration.value,
    })
    if (requestId !== listRequestId) return
    items.value = result.items || []
    totalCount.value = Number(result.total || 0)
  } catch (error) {
    if (requestId !== listRequestId) return
    items.value = []
    totalCount.value = 0
    errorMessage.value = getRequestErrorMessage(error, "加载电销录音库失败")
  } finally {
    if (requestId === listRequestId) {
      loading.value = false
    }
  }
}

const refreshList = async () => {
  await fetchRecords()
}

const syncList = async () => {
  if (!canSync.value) {
    await fetchRecords()
    return
  }

  syncing.value = true
  try {
    const result = await syncTelemarketingRecordings({
      pageSize: 100,
      timePeriod: "30d",
    })
    toast.success(`已同步 ${result.totalSaved} 条电销录音`)
    pageIndex.value = 0
    await fetchRecords()
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "同步电销录音失败"))
  } finally {
    syncing.value = false
  }
}

const handleSearch = () => {
  activeKeyword.value = keyword.value.trim()
  activeStartDate.value = startDate.value.trim()
  activeEndDate.value = endDate.value.trim()
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

  if (
    activeStartDate.value &&
    activeEndDate.value &&
    activeStartDate.value > activeEndDate.value
  ) {
    toast.error("开始日期不能晚于结束日期")
    return
  }

  pageIndex.value = 0
  void fetchRecords()
}

const clearSearch = () => {
  keyword.value = ""
  activeKeyword.value = ""
  startDate.value = ""
  activeStartDate.value = ""
  endDate.value = ""
  activeEndDate.value = ""
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

const resetDetailState = () => {
  detailRequestId += 1
  detailLoading.value = false
  detailErrorMessage.value = ""
  detail.value = null
}

const resetPlaybackDialogState = () => {
  playbackDialogLoading.value = false
  playbackDialogErrorMessage.value = ""
  playbackDialogDetail.value = null
}

const selectRecord = (record: TelemarketingRecording, options?: { resetPlayback?: boolean }) => {
  const nextId = String(record.id || "").trim()
  const shouldResetPlayback = Boolean(options?.resetPlayback) || nextId !== selectedRecordId.value
  if (shouldResetPlayback) {
    resetDetailState()
  }
  selectedRecord.value = record
  detailOpen.value = true
}

const requestRecordingDetail = async (
  record?: TelemarketingRecording | null,
): Promise<TelemarketingRecordingDetailResponse | null> => {
  const targetRecord = record || selectedRecord.value
  const id = String(targetRecord?.id || "").trim()
  if (!id) return null

  return getTelemarketingRecordingDetail(id)
}

const fetchDetail = async (
  record?: TelemarketingRecording | null,
): Promise<TelemarketingRecordingDetailResponse | null> => {
  const targetRecord = record || selectedRecord.value
  const id = String(targetRecord?.id || "").trim()
  if (!id) return null

  const requestId = ++detailRequestId
  detailLoading.value = true
  detailErrorMessage.value = ""
  detail.value = null

  try {
    const result = await requestRecordingDetail(targetRecord)
    if (requestId !== detailRequestId) return null
    detail.value = result
    return result
  } catch (error) {
    if (requestId !== detailRequestId) return null
    detail.value = null
    detailErrorMessage.value = getRequestErrorMessage(error, "加载录音详情失败")
    return null
  } finally {
    if (requestId === detailRequestId) {
      detailLoading.value = false
    }
  }
}

const openDetail = (record: TelemarketingRecording) => {
  selectRecord(record)
}

const playAudio = async (audioElement: HTMLAudioElement | null) => {
  if (!audioElement) return
  try {
    audioElement.currentTime = 0
    await audioElement.play()
  } catch {
    // 浏览器自动播放策略可能阻止播放，保留控件供用户手动点击
  }
}

const ensureDetailPlayback = async () => {
  const record = selectedRecording.value
  if (!record) return null

  if (detail.value?.playbackUrl && detail.value.recording.id === record.id) {
    return detail.value
  }
  return fetchDetail(record)
}

const playInDetail = async () => {
  const playbackDetail = await ensureDetailPlayback()
  if (!playbackDetail?.playbackUrl) {
    toast.error("当前录音暂时没有可用的播放地址")
    return
  }

  await nextTick()
  await playAudio(detailAudioRef.value)
}

const downloadFromDetail = async () => {
  const playbackDetail = await ensureDetailPlayback()
  const playbackUrl = playbackDetail?.playbackUrl
  if (!playbackUrl) {
    toast.error("当前录音暂时没有可用的播放地址")
    return
  }

  const anchor = document.createElement("a")
  anchor.href = playbackUrl
  anchor.target = "_blank"
  anchor.rel = "noopener noreferrer"
  anchor.download = playbackDetail.playbackFilename || `${selectedRecordId.value || "recording"}.mp3`
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
}

const openPlaybackDialog = async (record: TelemarketingRecording) => {
  const targetID = String(record.id || "").trim()
  playbackDialogRecord.value = record
  playbackDialogOpen.value = true
  resetPlaybackDialogState()
  playbackDialogLoading.value = true
  playbackLoadingId.value = targetID

  try {
    const result = await requestRecordingDetail(record)
    if (String(playbackDialogRecord.value?.id || "").trim() !== targetID) return
    playbackDialogDetail.value = result
    if (!result?.playbackUrl) {
      playbackDialogErrorMessage.value = "当前录音暂时没有可用的播放地址"
      return
    }
    await nextTick()
    await playAudio(playbackDialogAudioRef.value)
  } catch (error) {
    if (String(playbackDialogRecord.value?.id || "").trim() !== targetID) return
    playbackDialogErrorMessage.value = getRequestErrorMessage(error, "加载录音详情失败")
  } finally {
    if (String(playbackDialogRecord.value?.id || "").trim() === targetID) {
      playbackDialogLoading.value = false
      if (playbackLoadingId.value === targetID) {
        playbackLoadingId.value = ""
      }
    }
  }
}

const handleDetailOpenChange = (open: boolean) => {
  detailOpen.value = open
  if (!open) {
    selectedRecord.value = null
    resetDetailState()
  }
}

const downloadFromPlaybackDialog = () => {
  const playbackUrl = playbackDialogDetail.value?.playbackUrl
  if (!playbackUrl) {
    toast.error("当前录音暂时没有可用的播放地址")
    return
  }

  const anchor = document.createElement("a")
  anchor.href = playbackUrl
  anchor.target = "_blank"
  anchor.rel = "noopener noreferrer"
  anchor.download =
    playbackDialogDetail.value?.playbackFilename ||
    `${String(playbackDialogRecord.value?.id || "recording").trim() || "recording"}.mp3`
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
}

const handlePlaybackDialogOpenChange = (open: boolean) => {
  playbackDialogOpen.value = open
  if (!open) {
    playbackDialogRecord.value = null
    resetPlaybackDialogState()
    playbackLoadingId.value = ""
  }
}

const handleRefreshEvent = () => {
  void fetchRecords()
}

onMounted(() => {
  void fetchRecords()
  window.addEventListener(telemarketingRecordingsRefreshEvent, handleRefreshEvent)
})

onBeforeUnmount(() => {
  window.removeEventListener(telemarketingRecordingsRefreshEvent, handleRefreshEvent)
})
</script>

<template>
  <div class="w-full flex flex-col gap-3 lg:gap-4">
    <Card class="border-border/60 shadow-sm">
      <CardHeader class="space-y-2 border-b px-4 py-4 sm:px-5">
        <div class="flex flex-col gap-2 xl:flex-row xl:items-center xl:justify-between">
          <div class="grid flex-1 gap-2 sm:grid-cols-2 xl:grid-cols-6">
            <Button size="sm" variant="outline" class="h-9" :disabled="loading || syncing" @click="refreshList">
              <RefreshCw class="h-4 w-4" />
              <span>刷新列表</span>
            </Button>
            <Button
              v-if="canSync"
              size="sm"
              class="h-9"
              :disabled="syncing"
              @click="syncList"
            >
              <Loader2 v-if="syncing" class="h-4 w-4 animate-spin" />
              <RefreshCw v-else class="h-4 w-4" />
              <span>{{ syncing ? "同步中" : "同步最新" }}</span>
            </Button>
            <Input
              v-model="keyword"
              placeholder="工号 / 姓名 / 号码"
              class="h-9 xl:col-span-2"
              @keyup.enter="handleSearch"
            />
            <Input
              v-model="minDuration"
              type="number"
              min="0"
              step="1"
              placeholder="最小时长(秒)"
              class="h-9"
              @keyup.enter="handleSearch"
            />
            <Input
              v-model="maxDuration"
              type="number"
              min="0"
              step="1"
              placeholder="最大时长(秒)"
              class="h-9"
              @keyup.enter="handleSearch"
            />
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <Button size="sm" class="h-9" @click="handleSearch">
              <Search class="h-4 w-4" />
              <span>搜索</span>
            </Button>
            <Button size="sm" variant="outline" class="h-9" @click="clearSearch">
              <RefreshCw class="h-4 w-4" />
              <span>重置</span>
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent class="pt-3">
        <div class="overflow-hidden rounded-xl bg-background">
          <div v-if="loading" class="flex items-center justify-center py-24">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
          </div>

          <div v-else-if="errorMessage" class="py-20 text-center text-destructive">
            {{ errorMessage }}
          </div>

          <div v-else class="overflow-x-auto">
            <Table class="w-full min-w-[900px] table-fixed">
              <TableHeader class="sticky top-0 z-20 bg-muted/40">
                <TableRow>
                  <TableHead class="w-[80px] text-center">坐席信息</TableHead>
                  <TableHead class="w-[180px] text-center">客户号码</TableHead>
                  <TableHead class="w-[180px] text-center">录音信息</TableHead>
                  <TableHead class="w-[96px] text-center">操作</TableHead>
                </TableRow>
              </TableHeader>

              <TableBody>
                <TableRow
                  v-for="record in items"
                  :key="record.id"
                  class="group h-[88px] transition-colors hover:bg-muted/30"
                >
                  <TableCell class="align-middle py-2.5 text-center">
                    <div class="mx-auto flex w-[80px] flex-col items-center gap-1.5 rounded-xl bg-muted/[0.14] px-3 py-2.5 text-center">
                      <div class="flex items-center justify-center gap-2">
                        <span class="rounded-full bg-background px-2 py-0.5 text-[11px] font-medium text-muted-foreground">
                          工号 {{ pickDisplayText(record.serviceSeatWorkNumber) }} {{ pickDisplayText(record.matchedUserName, record.serviceSeatName) }}
                        </span>
                      </div>
                    </div>
                  </TableCell>

                  <TableCell class="align-middle py-2.5 text-center">
                    <div class="mx-auto flex w-[180px] flex-col items-center gap-1.5 rounded-xl bg-muted/[0.12] px-3 py-2.5 text-center">
                      <div class="text-sm font-semibold tabular-nums text-foreground">
                        {{ pickDisplayText(record.attribution, record.districtName) }}-{{ pickDisplayText(record.outlineNumber) }}
                      </div>
                      <div class="text-xs text-muted-foreground">
                        {{ formatDateTimeParts(record.initiateTime).date }}-{{ formatDateTimeParts(record.initiateTime).time || "--:--:--" }}
                      </div>
                      <div class="flex flex-wrap justify-center gap-1.5">
                        <Badge :variant="getAnswerBadgeVariant(record.callAnswerResult)" class="h-5 px-2 text-[11px]">
                          {{ getAnswerResultLabel(record.callAnswerResult) }}
                        </Badge>
                        <Badge variant="outline" class="h-5 px-2 text-[11px]">
                          {{ getHangupPartyLabel(record.callHangupParty) }}
                        </Badge>
                      </div>
                    </div>
                  </TableCell>

                  <TableCell class="align-middle py-2.5 text-center">
                    <div class="mx-auto flex w-[180px] flex-col items-center gap-1.5 rounded-xl bg-muted/[0.14] px-3 py-2.5 text-center">
                      <div class="flex items-center justify-center gap-2">
                        <div class="text-sm font-semibold text-foreground">
                          {{ getRecordDurationText(record) }}
                        </div>
                        <Button
                          size="sm"
                          variant="secondary"
                          class="h-7 rounded-lg px-2.5 text-[11px] font-medium"
                          :disabled="playbackLoadingId === record.id"
                          @click.stop="openPlaybackDialog(record)"
                        >
                          <Loader2 v-if="playbackLoadingId === record.id" class="h-3.5 w-3.5 animate-spin" />
                          <Play v-else class="h-3.5 w-3.5" />
                          <span>{{ playbackLoadingId === record.id ? "连接中" : "播放录音" }}</span>
                        </Button>
                      </div>
                      <div class="max-w-full truncate text-center text-xs text-muted-foreground">
                        {{ pickDisplayText(record.recordFilename) }}
                      </div>
                    </div>
                  </TableCell>

                  <TableCell class="align-middle py-2.5">
                    <div class="flex justify-center">
                      <Button size="sm" variant="outline" class="h-8 px-3 text-xs" @click="openDetail(record)">
                        详情
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>

                <EmptyTablePlaceholder
                  v-if="items.length === 0"
                  :colspan="4"
                  :text="displayEmptyText"
                />
              </TableBody>
            </Table>
          </div>
        </div>

        <div class="mt-3">
          <Pagination
            :current-page="pageIndex"
            :total-pages="totalPages"
            :page-size="pageSize"
            :page-size-options="[10, 20, 30, 50]"
            :show-selection="false"
            :total-count="totalCount"
            @update:current-page="handlePageChange"
            @update:page-size="handlePageSizeChange"
          />
        </div>
      </CardContent>
    </Card>

    <Sheet :open="detailOpen" @update:open="handleDetailOpenChange">
      <SheetContent
        side="right"
        class="w-[94vw] max-w-none overflow-y-auto border-l bg-background p-0 sm:w-[700px] sm:max-w-[700px]"
      >
        <div class="flex min-h-full flex-col bg-muted/10">
          <SheetHeader class="border-b bg-background px-5 py-4 text-left">
            <div class="flex items-start justify-between gap-3 pr-8">
              <div class="space-y-2">
                <div class="flex flex-wrap items-center gap-2">
                  <Badge variant="outline" class="px-2.5 py-0.5">
                    电销录音详情
                  </Badge>
                  <Badge variant="outline" class="px-2.5 py-0.5">
                    <CalendarDays class="h-3.5 w-3.5" />
                    {{ formatDateTime(selectedRecording?.initiateTime) }}
                  </Badge>
                </div>

                <div class="space-y-1">
                  <SheetTitle class="text-xl font-semibold tracking-tight">
                    {{ pickDisplayText(selectedRecording?.matchedUserName, selectedRecording?.serviceSeatName, selectedRecording?.serviceSeatWorkNumber) }}
                  </SheetTitle>
                </div>
              </div>

              <div class="hidden shrink-0 rounded-xl border bg-muted/30 px-3 py-2.5 text-right sm:block">
                <div class="text-xs text-muted-foreground">当前号码</div>
                <div class="mt-1 text-sm font-semibold tabular-nums text-foreground">
                  {{ pickDisplayText(selectedRecording?.outlineNumber) }}
                </div>
              </div>
            </div>
          </SheetHeader>

          <div class="flex-1 space-y-4 px-5 py-4">
            <template v-if="selectedRecording">
              <Card class="border-border/60 shadow-sm">
                <CardContent class="space-y-3 px-5 py-4">
                  <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                    <div class="space-y-1">
                      <div class="flex items-center gap-2 text-[13px] text-muted-foreground">
                        <Waves class="h-4 w-4" />
                        <span>点击播放或下载时才会实时换取录音链接</span>
                      </div>
                      <div class="text-base font-semibold text-foreground">
                        {{ pickDisplayText(detail?.playbackFilename, selectedRecording.recordFilename) }}
                      </div>
                      <div class="text-xs text-muted-foreground">
                        过期时间：{{ detail?.playbackExpiresAt ? formatDateTime(detail.playbackExpiresAt) : "未刷新" }}
                      </div>
                    </div>

                    <div class="flex flex-wrap gap-2">
                      <Button size="sm" variant="outline" class="h-8" :disabled="!selectedRecording || detailLoading" @click="playInDetail">
                        <Loader2 v-if="detailLoading" class="h-4 w-4 animate-spin" />
                        <Play v-else class="h-4 w-4" />
                        <span>播放录音</span>
                      </Button>
                      <Button size="sm" class="h-8" :disabled="!selectedRecording || detailLoading" @click="downloadFromDetail">
                        <Loader2 v-if="detailLoading" class="h-4 w-4 animate-spin" />
                        <Download v-else class="h-4 w-4" />
                        <span>下载录音</span>
                      </Button>
                    </div>
                  </div>

                  <div v-if="detailErrorMessage" class="text-sm text-destructive">
                    {{ detailErrorMessage }}
                  </div>

                  <div v-if="detail?.playbackUrl" class="rounded-xl border bg-background/80 p-3">
                    <audio
                      ref="detailAudioRef"
                      class="h-11 w-full"
                      controls
                      preload="none"
                      :src="detail.playbackUrl"
                    />
                  </div>
                </CardContent>
              </Card>

              <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
                <Card class="border-border/60 shadow-sm">
                  <CardContent class="px-4 py-4">
                    <div class="flex items-center gap-2 text-muted-foreground">
                      <Clock3 class="h-4 w-4" />
                      <span class="text-sm">通话时长</span>
                    </div>
                    <div class="mt-2 text-xl font-semibold text-foreground">
                      {{ getRecordDurationText(selectedRecording) }}
                    </div>
                  </CardContent>
                </Card>

                <Card class="border-border/60 shadow-sm">
                  <CardContent class="px-4 py-4">
                    <div class="text-sm text-muted-foreground">接通结果</div>
                    <div class="mt-2 text-xl font-semibold text-foreground">
                      {{ getAnswerResultLabel(selectedRecording.callAnswerResult) }}
                    </div>
                  </CardContent>
                </Card>

                <Card class="border-border/60 shadow-sm">
                  <CardContent class="px-4 py-4">
                    <div class="text-sm text-muted-foreground">挂断方</div>
                    <div class="mt-2 text-xl font-semibold text-foreground">
                      {{ getHangupPartyLabel(selectedRecording.callHangupParty) }}
                    </div>
                  </CardContent>
                </Card>

                <Card class="border-border/60 shadow-sm">
                  <CardContent class="px-4 py-4">
                    <div class="flex items-center gap-2 text-muted-foreground">
                      <UserRound class="h-4 w-4" />
                      <span class="text-sm">录音状态</span>
                    </div>
                    <div class="mt-2 text-xl font-semibold text-foreground">
                      {{ getRecordStatusLabel(selectedRecording.recordStatus) }}
                    </div>
                  </CardContent>
                </Card>
              </div>

              <div class="grid gap-3 lg:grid-cols-2">
                <Card class="border-border/60 shadow-sm">
                  <CardHeader class="px-5 py-4">
                    <CardTitle class="text-base">通话信息</CardTitle>
                  </CardHeader>
                  <CardContent class="space-y-2.5 px-5 pb-5 pt-0">
                    <div
                      v-for="item in detailInfoRows"
                      :key="item.label"
                      class="grid grid-cols-[92px_minmax(0,1fr)] gap-3 text-sm"
                    >
                      <div class="text-muted-foreground">{{ item.label }}</div>
                      <div class="break-all text-foreground">{{ item.value }}</div>
                    </div>
                  </CardContent>
                </Card>

                <Card class="border-border/60 shadow-sm">
                  <CardHeader class="px-5 py-4">
                    <CardTitle class="text-base">坐席映射</CardTitle>
                  </CardHeader>
                  <CardContent class="space-y-2.5 px-5 pb-5 pt-0">
                    <div
                      v-for="item in detailSeatRows"
                      :key="item.label"
                      class="grid grid-cols-[92px_minmax(0,1fr)] gap-3 text-sm"
                    >
                      <div class="text-muted-foreground">{{ item.label }}</div>
                      <div class="break-all text-foreground">{{ item.value }}</div>
                    </div>
                  </CardContent>
                </Card>
              </div>

              <Card class="border-border/60 shadow-sm">
                <CardHeader class="px-5 py-4">
                  <CardTitle class="text-base">时间线</CardTitle>
                </CardHeader>
                <CardContent class="grid gap-2.5 px-5 pb-5 pt-0 md:grid-cols-2">
                  <div
                    v-for="item in detailTimelineRows"
                    :key="item.label"
                    class="grid grid-cols-[92px_minmax(0,1fr)] gap-3 text-sm"
                  >
                    <div class="text-muted-foreground">{{ item.label }}</div>
                    <div class="break-all text-foreground">{{ item.value }}</div>
                  </div>
                </CardContent>
              </Card>
            </template>

            <Card v-else class="border-border/60 shadow-sm">
              <CardContent class="py-14 text-center text-sm text-muted-foreground">
                暂无可展示的录音详情
              </CardContent>
            </Card>
          </div>
        </div>
      </SheetContent>
    </Sheet>

    <Dialog :open="playbackDialogOpen" @update:open="handlePlaybackDialogOpenChange">
      <DialogContent class="overflow-hidden border-border/70 bg-background/95 p-0 sm:max-w-[460px]">
        <DialogHeader class="border-b bg-muted/[0.16] px-5 py-4">
          <DialogTitle class="text-base font-semibold tracking-tight">录音播放</DialogTitle>
        </DialogHeader>

        <div class="space-y-3 px-5 py-4">
          <div class="space-y-2">
            <div class="truncate text-sm font-semibold text-foreground">
              {{ playbackDialogTitle }}
            </div>
            <div class="flex flex-wrap gap-1.5">
              <Badge variant="outline" class="h-5 px-2 text-[11px]">
                {{ pickDisplayText(playbackDialogRecord?.outlineNumber, "未知号码") }}
              </Badge>
              <Badge variant="outline" class="h-5 px-2 text-[11px]">
                {{ getRecordDurationText(playbackDialogRecord) }}
              </Badge>
              <Badge variant="outline" class="h-5 px-2 text-[11px]">
                {{ pickDisplayText(playbackDialogRecord?.attribution, playbackDialogRecord?.districtName, "未知归属地") }}
              </Badge>
            </div>
          </div>

          <div v-if="playbackDialogLoading" class="flex items-center justify-center py-10">
            <Loader2 class="h-5 w-5 animate-spin text-muted-foreground" />
          </div>

          <div v-else-if="playbackDialogErrorMessage" class="rounded-xl border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive">
            {{ playbackDialogErrorMessage }}
          </div>

          <div v-else-if="playbackDialogDetail?.playbackUrl" class="space-y-3 rounded-xl border bg-background/90 p-3">
            <div class="text-xs text-muted-foreground">
              {{ pickDisplayText(playbackDialogDetail.playbackFilename, playbackDialogRecord?.recordFilename, "录音文件") }}
            </div>
            <audio
              ref="playbackDialogAudioRef"
              class="h-11 w-full"
              controls
              preload="none"
              :src="playbackDialogDetail.playbackUrl"
            />
            <div class="flex justify-end">
              <Button
                size="sm"
                class="h-8"
                :disabled="!playbackDialogDetail?.playbackUrl"
                @click="downloadFromPlaybackDialog"
              >
                <Download class="h-4 w-4" />
                <span>下载录音</span>
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
