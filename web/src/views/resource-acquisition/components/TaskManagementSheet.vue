<script setup lang="ts">
import {
  Ban,
  Clock3,
  Globe2,
  Loader2,
  RefreshCw,
  Search,
  Settings2,
  Wifi,
  WifiOff,
} from "lucide-vue-next";

import ErrorFeedback from "@/components/custom/ErrorFeedback.vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Pagination } from "@/components/ui/pagination";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import type { ExternalCompanySearchTask } from "@/types/externalCompanySearch";
import {
  EXTERNAL_COMPANY_SEARCH_PLATFORM,
  EXTERNAL_COMPANY_SEARCH_TASK_STATUS,
} from "@/types/externalCompanySearch";

defineProps<{
  activeTask: ExternalCompanySearchTask | null;
  taskItems: ExternalCompanySearchTask[];
  taskLoading: boolean;
  taskKeyword: string;
  taskPageIndex: number;
  taskTotalPages: number;
  taskPageSize: number;
  taskTotal: number;
  currentProgressPercent: number;
  canceling: boolean;
  canCancelTask: boolean;
  streaming: boolean;
  streamBadgeLabel: string;
  hasActiveNonTerminalTask: boolean;
  refreshing: boolean;
  streamError: string;
}>();

const emit = defineEmits<{
  "update:taskKeyword": [value: string];
  "open-task": [taskId: number];
  "cancel-task": [];
  refresh: [];
  search: [];
  "page-change": [page: number];
  "page-size-change": [size: number];
  "open-change": [open: boolean];
}>();

const platformLabel = (platform: number) => {
  switch (platform) {
    case EXTERNAL_COMPANY_SEARCH_PLATFORM.ALIBABA:
      return "Alibaba";
    case EXTERNAL_COMPANY_SEARCH_PLATFORM.MADE_IN_CHINA:
      return "Made-in-China";
    case EXTERNAL_COMPANY_SEARCH_PLATFORM.GOOGLE:
      return "Google";
    default:
      return `平台 ${platform}`;
  }
};

const taskStatusMeta = (status: number) => {
  switch (status) {
    case EXTERNAL_COMPANY_SEARCH_TASK_STATUS.PENDING:
      return {
        label: "待执行",
        class: "border-amber-200 bg-amber-50 text-amber-700",
      };
    case EXTERNAL_COMPANY_SEARCH_TASK_STATUS.RUNNING:
      return {
        label: "执行中",
        class: "border-sky-200 bg-sky-50 text-sky-700",
      };
    case EXTERNAL_COMPANY_SEARCH_TASK_STATUS.COMPLETED:
      return {
        label: "已完成",
        class: "border-emerald-200 bg-emerald-50 text-emerald-700",
      };
    case EXTERNAL_COMPANY_SEARCH_TASK_STATUS.FAILED:
      return { label: "失败", class: "border-red-200 bg-red-50 text-red-700" };
    case EXTERNAL_COMPANY_SEARCH_TASK_STATUS.CANCELED:
      return {
        label: "已取消",
        class: "border-zinc-200 bg-zinc-100 text-zinc-700",
      };
    default:
      return {
        label: `状态 ${status}`,
        class: "border-border bg-muted text-muted-foreground",
      };
  }
};

const formatDateTime = (raw?: string | null) => {
  if (!raw) return "-";
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return raw;
  return date.toLocaleString("zh-CN", { hour12: false });
};

const formatPageLimit = (pageLimit: number) => {
  if (pageLimit <= 0) return "全部页";
  return String(pageLimit);
};
</script>

<template>
  <Sheet @update:open="emit('open-change', $event)">
    <SheetTrigger as-child>
      <Button variant="outline" size="sm" class="bg-background">
        <Settings2 class="h-4 w-4 mr-1" />任务管理
      </Button>
    </SheetTrigger>
    <SheetContent
      class="w-[90vw] sm:w-[800px] sm:max-w-[800px] border-l flex flex-col p-6 overflow-y-auto"
    >
      <SheetHeader class="pb-4 border-b">
        <SheetTitle>获取任务管理</SheetTitle>
        <SheetDescription>
          查看所有新建的查询任务，并追踪其实时进度和状态。
        </SheetDescription>
      </SheetHeader>

      <div class="py-4 space-y-6 flex-1 flex flex-col">
        <div
          class="rounded-2xl border bg-muted/20 p-4 flex items-start justify-between gap-4 shrink-0"
        >
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <div
                class="h-9 w-9 rounded-xl flex items-center justify-center border"
                :class="
                  streaming
                    ? 'border-emerald-200 bg-emerald-50 text-emerald-600'
                    : 'border-zinc-200 bg-background text-zinc-500'
                "
              >
                <Wifi v-if="streaming" class="h-4 w-4" />
                <WifiOff v-else class="h-4 w-4" />
              </div>
              <div class="space-y-0.5">
                <p class="text-sm font-medium text-foreground">服务状态</p>
                <div class="flex items-center gap-2">
                  <Badge
                    variant="outline"
                    :class="
                      streaming
                        ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                        : 'border-zinc-200 bg-background text-muted-foreground'
                    "
                  >
                    {{ streamBadgeLabel }}
                  </Badge>
                  <span class="text-xs text-muted-foreground">
                    {{
                      hasActiveNonTerminalTask
                        ? '运行中的任务会实时推送结果'
                        : '当前无运行中任务，面板进入静态查看'
                    }}
                  </span>
                </div>
              </div>
            </div>

            <p
              v-if="!streaming && hasActiveNonTerminalTask"
              class="text-xs text-amber-700 bg-amber-50 border border-amber-200 rounded-lg px-3 py-2"
            >
              实时推送已断开，点击右侧同步会补拉事件并重新接入。
            </p>
          </div>

          <Button
            variant="outline"
            size="sm"
            class="bg-background shrink-0"
            :disabled="refreshing"
            @click="emit('refresh')"
          >
            <Loader2 v-if="refreshing" class="h-3.5 w-3.5 mr-1.5 animate-spin" />
            <RefreshCw v-else class="h-3.5 w-3.5 mr-1.5" />
            同步
          </Button>
        </div>

        <!-- 当前选中的任务进度和详情 -->
        <div
          v-if="activeTask"
          class="space-y-5 rounded-2xl border bg-gradient-to-br from-background to-muted/30 p-5 shadow-sm shrink-0 relative overflow-hidden"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="space-y-1.5 min-w-0">
              <div class="flex items-center gap-2 flex-wrap">
                <p
                  class="font-semibold text-foreground truncate max-w-[200px]"
                  :title="activeTask.keyword"
                >
                  {{ activeTask.keyword }}
                </p>
                <Badge
                  variant="secondary"
                  class="text-[10px] h-5 px-1.5 font-medium bg-muted"
                >
                  {{ platformLabel(activeTask.platform) }}
                </Badge>
              </div>
              <div
                class="flex items-center gap-2 text-[11px] text-muted-foreground"
              >
                <Clock3 class="h-3 w-3" />
                <span>{{ formatDateTime(activeTask.updatedAt) }}</span>
              </div>
            </div>
            <Badge
              variant="outline"
              :class="['shrink-0', taskStatusMeta(activeTask.status).class]"
            >
              {{ taskStatusMeta(activeTask.status).label }}
            </Badge>
          </div>

          <div
            class="grid grid-cols-4 gap-3 text-center text-sm bg-background/50 rounded-xl p-3 border border-border/40"
          >
            <div class="space-y-1">
              <p class="text-[11px] text-muted-foreground font-medium">进度</p>
              <p class="font-semibold tabular-nums text-foreground">
                {{ currentProgressPercent }}%
              </p>
            </div>
            <div class="space-y-1">
              <p class="text-[11px] text-muted-foreground font-medium">已查</p>
              <p class="font-semibold tabular-nums text-foreground">
                {{ activeTask.fetchedCount }}
              </p>
            </div>
            <div class="space-y-1">
              <p class="text-[11px] text-emerald-600/80 font-medium">已存</p>
              <p class="font-semibold tabular-nums text-emerald-600">
                {{ activeTask.savedCount }}
              </p>
            </div>
            <div class="space-y-1">
              <p class="text-[11px] text-red-600/80 font-medium">失败</p>
              <p class="font-semibold tabular-nums text-red-600">
                {{ activeTask.failedCount }}
              </p>
            </div>
          </div>

          <ErrorFeedback
            :message="activeTask.errorMessage || streamError"
            class="bg-background rounded-lg"
          />

          <div v-if="canCancelTask" class="pt-1 flex justify-end">
            <Button
              size="sm"
              variant="ghost"
              class="h-7 text-xs text-red-500 hover:text-red-600 hover:bg-red-50"
              :disabled="canceling"
              @click="emit('cancel-task')"
            >
              <Loader2 v-if="canceling" class="w-3.5 h-3.5 mr-1 animate-spin" />
              <Ban v-else class="w-3.5 h-3.5 mr-1" />
              中止任务
            </Button>
          </div>
        </div>

        <!-- 历史任务列表 -->
        <div
          class="flex flex-col gap-4 flex-1 overflow-hidden pointer-events-auto"
        >
          <!-- Search Input -->
          <div class="relative shrink-0">
            <Search
              class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/70"
            />
            <Input
              :model-value="taskKeyword"
              placeholder="检索历史任务关键词..."
              class="h-9 pl-9 pr-4 bg-muted/30 border-border/60 transition-colors focus-visible:bg-background focus-visible:ring-1 focus-visible:ring-primary/20 shadow-sm"
              @update:model-value="emit('update:taskKeyword', String($event))"
              @keyup.enter="emit('search')"
            />
          </div>

          <!-- Task List Items -->
          <div
            class="flex-1 overflow-y-auto min-h-[300px] border border-border/60 rounded-xl bg-background"
          >
            <div
              v-if="taskLoading"
              class="flex flex-col items-center justify-center py-24 gap-3 h-full"
            >
              <Loader2 class="h-6 w-6 animate-spin text-primary/60" />
              <p class="text-sm text-muted-foreground">加载中...</p>
            </div>
            <Table v-else>
              <TableHeader class="sticky top-0 bg-muted/50 z-10 shadow-sm">
                <TableRow>
                  <TableHead class="w-24">状态</TableHead>
                  <TableHead>关键词</TableHead>
                  <TableHead>平台</TableHead>
                  <TableHead class="text-center">页数</TableHead>
                  <TableHead class="text-center">总爬取</TableHead>
                  <TableHead class="text-center">成功数</TableHead>
                  <TableHead class="text-center">失败数</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow
                  v-for="taskItem in taskItems"
                  :key="taskItem.id"
                  class="cursor-pointer transition-colors hover:bg-muted/30"
                  :class="
                    taskItem.id === activeTask?.id
                      ? 'bg-primary/5 relative'
                      : ''
                  "
                  @click="emit('open-task', taskItem.id)"
                >
                  <TableCell class="relative">
                    <!-- Active Item Indicator Line -->
                    <div
                      v-if="taskItem.id === activeTask?.id"
                      class="absolute left-0 top-0 bottom-0 w-1 bg-primary"
                    ></div>
                    <Badge
                      variant="outline"
                      class="text-[10px] px-1.5 shrink-0 bg-background"
                      :class="taskStatusMeta(taskItem.status).class"
                    >
                      {{ taskStatusMeta(taskItem.status).label }}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div class="flex flex-col gap-0.5">
                      <span
                        class="font-medium text-sm truncate max-w-[160px]"
                        :title="taskItem.keyword"
                      >
                        {{ taskItem.keyword }}
                      </span>
                      <span
                        class="text-[10px] text-muted-foreground/70 font-mono"
                      >
                        {{ formatDateTime(taskItem.updatedAt) }}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div
                      class="flex items-center gap-1.5 text-xs text-muted-foreground"
                    >
                      <Globe2 class="h-3 w-3 opacity-70" />
                      <span>{{ platformLabel(taskItem.platform) }}</span>
                    </div>
                  </TableCell>
                  <TableCell class="text-center">
                    <span
                      class="text-xs tabular-nums text-muted-foreground font-mono"
                      >{{ taskItem.pageNo }}</span
                    >
                    <span class="text-[10px] text-muted-foreground/50 mx-0.5"
                      >/</span
                    >
                    <span
                      class="text-xs tabular-nums text-muted-foreground/70 font-mono"
                      >{{ formatPageLimit(taskItem.pageLimit) }}</span
                    >
                  </TableCell>
                  <TableCell class="text-center">
                    <span
                      class="text-xs font-semibold tabular-nums text-foreground"
                      >{{ taskItem.fetchedCount }}</span
                    >
                  </TableCell>
                  <TableCell class="text-center">
                    <span
                      class="text-xs font-semibold tabular-nums text-emerald-600"
                      >{{ taskItem.savedCount }}</span
                    >
                  </TableCell>
                  <TableCell class="text-center">
                    <span
                      class="text-xs font-semibold tabular-nums"
                      :class="
                        taskItem.failedCount > 0
                          ? 'text-red-500'
                          : 'text-muted-foreground'
                      "
                    >
                      {{ taskItem.failedCount }}
                    </span>
                  </TableCell>
                </TableRow>

                <TableRow v-if="taskItems.length === 0">
                  <TableCell
                    colspan="7"
                    class="h-48 text-center text-muted-foreground"
                  >
                    <div
                      class="flex flex-col items-center justify-center gap-2"
                    >
                      <Search class="h-6 w-6 text-muted-foreground/40" />
                      <p class="text-sm">暂无任务记录</p>
                      <p class="text-xs mt-1">请修改搜索词或创建新任务</p>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>

          <div class="shrink-0 pt-3 border-t bg-background mt-auto">
            <Pagination
              :current-page="taskPageIndex"
              :total-pages="taskTotalPages"
              :page-size="taskPageSize"
              :selected-count="0"
              :total-count="taskTotal"
              :show-selection="false"
              @update:current-page="emit('page-change', $event)"
              @update:page-size="emit('page-size-change', $event)"
            />
          </div>
        </div>
      </div>
    </SheetContent>
  </Sheet>
</template>
