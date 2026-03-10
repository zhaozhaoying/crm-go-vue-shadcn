<script setup lang="ts">
import {
  computed,
  onActivated,
  onMounted,
  ref,
  watch,
} from "vue";
import {
  Globe2,
  Loader2,
  SquareArrowOutUpRight,
  Search,
  Copy,
} from "lucide-vue-next";
import { toast } from "vue-sonner";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import ErrorFeedback from "@/components/custom/ErrorFeedback.vue";
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue";
import { Badge } from "@/components/ui/badge";
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
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useExternalCompanySearchTask } from "@/composables/useExternalCompanySearchTask";
import { useExternalCompanySearchTaskList } from "@/composables/useExternalCompanySearchTaskList";
import { useExternalCompanySearchResourceList } from "@/composables/useExternalCompanySearchResourceList";
import { getRequestErrorMessage } from "@/lib/http-error";
import type { ExternalCompanySearchResultItem } from "@/types/externalCompanySearch";
import {
  EXTERNAL_COMPANY_SEARCH_PLATFORM,
  EXTERNAL_COMPANY_SEARCH_TASK_STATUS,
} from "@/types/externalCompanySearch";

import CreateTaskDialog from "./components/CreateTaskDialog.vue";
import TaskManagementSheet from "./components/TaskManagementSheet.vue";

const supportedPlatforms = [
  {
    value: EXTERNAL_COMPANY_SEARCH_PLATFORM.ALIBABA,
    label: "Alibaba",
    description: "JSON 接口，实时性最好",
  },
  {
    value: EXTERNAL_COMPANY_SEARCH_PLATFORM.MADE_IN_CHINA,
    label: "Made-in-China",
    description: "HTML 翻页获取，已接入",
  },
  {
    value: EXTERNAL_COMPANY_SEARCH_PLATFORM.GOOGLE,
    label: "Google",
    description: "全球搜索引擎，覆盖面广",
  },
];

const searchWorkspace = useExternalCompanySearchTask();
const activeTask = searchWorkspace.task;
const events = searchWorkspace.events;
const creating = searchWorkspace.creating;
const loadingTask = searchWorkspace.loadingTask;
const canceling = searchWorkspace.canceling;
const streaming = searchWorkspace.streaming;
const isTerminalTask = searchWorkspace.isTerminalTask;
const actionError = searchWorkspace.actionError;
const streamError = searchWorkspace.streamError;
const createAndWatch = searchWorkspace.createAndWatch;
const openTask = searchWorkspace.openTask;
const syncTaskState = searchWorkspace.syncTaskState;
const connectStream = searchWorkspace.connectStream;
const cancelTask = searchWorkspace.cancelTask;

const taskListWorkspace = useExternalCompanySearchTaskList(10);
const taskLoading = taskListWorkspace.taskLoading;
const taskListError = taskListWorkspace.taskListError;
const taskItems = taskListWorkspace.taskItems;
const taskTotal = taskListWorkspace.taskTotal;
const taskPageIndex = taskListWorkspace.taskPageIndex;
const taskPageSize = taskListWorkspace.taskPageSize;
const taskKeyword = taskListWorkspace.taskKeyword;
const taskTotalPages = taskListWorkspace.taskTotalPages;
const loadTaskList = taskListWorkspace.loadTaskList;
const scheduleTaskListRefresh = taskListWorkspace.scheduleTaskListRefresh;
const handleTaskSearch = taskListWorkspace.handleTaskSearch;
const handleTaskPageChange = taskListWorkspace.handleTaskPageChange;
const handleTaskPageSizeChange = taskListWorkspace.handleTaskPageSizeChange;

const resourceListWorkspace = useExternalCompanySearchResourceList(10);
const resourceLoading = resourceListWorkspace.resourceLoading;
const resourceListError = resourceListWorkspace.resourceListError;
const resourceItems = resourceListWorkspace.resourceItems;
const resourceTotal = resourceListWorkspace.resourceTotal;
const resourcePage = resourceListWorkspace.resourcePage;
const resourcePageSize = resourceListWorkspace.resourcePageSize;
const resourceSearch = resourceListWorkspace.resourceSearch;
const resourcePlatformFilter = resourceListWorkspace.resourcePlatformFilter;
const resourceNewOnly = resourceListWorkspace.resourceNewOnly;
const resourceTotalPages = resourceListWorkspace.resourceTotalPages;
const loadResourceList = resourceListWorkspace.loadResourceList;
const scheduleResourceListRefresh =
  resourceListWorkspace.scheduleResourceListRefresh;
const handleResourceSearch = resourceListWorkspace.handleResourceSearch;
const handleResultPageChange =
  resourceListWorkspace.handleResultPageChange;
const handleResultPageSizeChange =
  resourceListWorkspace.handleResultPageSizeChange;

const createTaskDialogRef = ref<InstanceType<typeof CreateTaskDialog> | null>(
  null,
);

let hasAutoOpenedLatestTask = false;
let hasInitializedTaskList = false;
let skipNextActivatedRefresh = false;
const currentProgressPercent = computed(() =>
  Math.min(100, Math.max(0, activeTask.value?.progressPercent ?? 0)),
);
const streamBadgeLabel = computed(() => {
  if (activeTask.value && isTerminalTask.value) return "任务已结束";
  if (streaming.value) return "实时连接中";
  if (activeTask.value) return "等待重连";
  return "未连接";
});
const canCancelTask = computed(() => {
  const status = activeTask.value?.status;
  return (
    status === EXTERNAL_COMPANY_SEARCH_TASK_STATUS.PENDING ||
    status === EXTERNAL_COMPANY_SEARCH_TASK_STATUS.RUNNING
  );
});
const hasActiveNonTerminalTask = computed(
  () => !!activeTask.value && !isTerminalTask.value,
);

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

const formatProducts = (item: ExternalCompanySearchResultItem) => {
  const raw = String(item.mainProducts || "").trim();
  if (!raw) return "-";
  try {
    const parsed = JSON.parse(raw) as Array<{ name?: string } | string>;
    if (Array.isArray(parsed)) {
      const names = parsed
        .map((entry) => (typeof entry === "string" ? entry : entry?.name || ""))
        .map((entry) => entry.trim())
        .filter(Boolean);
      return names.length ? names.slice(0, 3).join(" / ") : raw;
    }
  } catch {
    // fall back to raw string
  }
  return raw;
};

const handleCopy = async (text: string) => {
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
    toast.success("复制成功");
  } catch (error) {
    toast.error("复制失败");
  }
};

const handleCreateTask = async (payload: {
  keyword: string;
  platform: number;
}) => {
  try {
    await createAndWatch({
      platforms: [payload.platform],
      keyword: payload.keyword,
      regionKeyword: payload.keyword,
      pageLimit: 0,
      targetCount: 0,
      priority: 10,
    });
    await loadTaskList();
    hasInitializedTaskList = true;
    toast.success("任务已创建，后台开始获取");
    createTaskDialogRef.value?.handleCreated();
  } catch {
    // actionError is rendered inline
  }
};

const handleRefreshAll = async () => {
  await Promise.all([loadTaskList(), loadResourceList()]);
  if (activeTask.value) {
    await syncTaskState(activeTask.value.id);
  }
};

const handleOpenTask = async (taskId: number) => {
  try {
    await openTask(taskId);
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "打开任务失败"));
  }
};

const handleCancelTask = async () => {
  if (!activeTask.value) return;
  try {
    await cancelTask(activeTask.value.id);
    await loadTaskList();
    hasInitializedTaskList = true;
    toast.success("任务已取消");
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "取消任务失败"));
  }
};

const initializeTaskList = async (autoOpenLatest = false) => {
  await loadTaskList();
  hasInitializedTaskList = true;
  if (
    autoOpenLatest &&
    !hasAutoOpenedLatestTask &&
    !activeTask.value &&
    taskItems.value.length > 0
  ) {
    hasAutoOpenedLatestTask = true;
    await openTask(taskItems.value[0].id);
  }
};

const ensureTaskListLoaded = async () => {
  if (hasInitializedTaskList || taskLoading.value) return;
  await initializeTaskList(true);
};

const handleTaskSheetOpenChange = (open: boolean) => {
  if (!open) return;
  void ensureTaskListLoaded();
};

watch(
  () => events.value.length,
  (nextLength, previousLength) => {
    if (nextLength !== previousLength) {
      scheduleTaskListRefresh();
      scheduleResourceListRefresh();
    }
  },
);

watch(
  () => actionError.value,
  (message) => {
    if (message) {
      toast.error(message);
    }
  },
);

watch(
  () => taskListError.value,
  (message) => {
    if (message) {
      toast.error(message);
    }
  },
);

onMounted(() => {
  skipNextActivatedRefresh = true;
  void loadResourceList();
});

onActivated(() => {
  if (skipNextActivatedRefresh) {
    skipNextActivatedRefresh = false;
    if (activeTask.value && !streaming.value && !isTerminalTask.value) {
      void connectStream();
    }
    return;
  }
  void loadResourceList();
  if (activeTask.value && !streaming.value && !isTerminalTask.value) {
    void connectStream();
  }
});
</script>

<template>
  <div class="w-full flex flex-col gap-4 lg:gap-6">
    <!-- 顶部操作栏 -->
    <div
      class="flex flex-col gap-3 px-4 lg:px-6 shrink-0 lg:flex-row lg:items-center lg:justify-between"
    >
      <div class="flex items-center gap-2">
        <CreateTaskDialog
          ref="createTaskDialogRef"
          :creating="creating"
          :action-error="actionError"
          :platforms="supportedPlatforms"
          :default-platform="String(EXTERNAL_COMPANY_SEARCH_PLATFORM.ALIBABA)"
          @create="handleCreateTask"
        />

        <TaskManagementSheet
          :active-task="activeTask"
          :task-items="taskItems"
          :task-loading="taskLoading"
          :task-keyword="taskKeyword"
          :task-page-index="taskPageIndex"
          :task-total-pages="taskTotalPages"
          :task-page-size="taskPageSize"
          :task-total="taskTotal"
          :current-progress-percent="currentProgressPercent"
          :canceling="canceling"
          :can-cancel-task="canCancelTask"
          :streaming="streaming"
          :stream-badge-label="streamBadgeLabel"
          :has-active-non-terminal-task="hasActiveNonTerminalTask"
          :refreshing="taskLoading || loadingTask || resourceLoading"
          :stream-error="streamError"
          @update:task-keyword="taskKeyword = $event"
          @open-task="handleOpenTask"
          @cancel-task="handleCancelTask"
          @refresh="handleRefreshAll"
          @search="handleTaskSearch"
          @page-change="handleTaskPageChange"
          @page-size-change="handleTaskPageSizeChange"
          @open-change="handleTaskSheetOpenChange"
        />
      </div>

      <div class="flex flex-wrap items-center gap-2 lg:justify-end">
        <div class="relative w-full sm:w-[280px] lg:w-[320px]">
          <Search
            class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground"
          />
          <Input
            v-model="resourceSearch"
            type="search"
            placeholder="搜索企业 / 产品 / 关键词..."
            class="pl-8 h-9 bg-background"
            @keyup.enter="handleResourceSearch"
          />
        </div>

        <Select v-model="resourcePlatformFilter">
          <SelectTrigger class="h-9 w-[150px] bg-background">
            <SelectValue placeholder="平台筛选" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="all">全部平台</SelectItem>
              <SelectItem
                v-for="platform in supportedPlatforms"
                :key="platform.value"
                :value="String(platform.value)"
              >
                {{ platform.label }}
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>

        <label
          for="resource-new-only"
          class="flex h-9 items-center gap-2 rounded-md border bg-background px-3 text-sm text-foreground"
        >
          <Switch id="resource-new-only" v-model:checked="resourceNewOnly" />
          <span>只看新发掘</span>
        </label>
      </div>
    </div>

    <!-- 企业列表及分页 -->
    <div class="relative flex flex-col gap-4 px-4 lg:px-6">
      <ErrorFeedback
        v-if="resourceListError"
        :message="resourceListError"
        class="shrink-0"
      />

      <div class="overflow-hidden rounded-lg border bg-background">
        <div
          v-if="resourceLoading"
          class="flex items-center justify-center py-24"
        >
          <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
        </div>

        <Table v-else>
          <TableHeader class="bg-muted/50 sticky top-0 z-10">
            <TableRow>
              <TableHead class="w-16">ID</TableHead>
              <TableHead>企业</TableHead>
              <TableHead>平台</TableHead>
              <TableHead>城市</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>关键词</TableHead>
              <TableHead class="w-[30%]">主营产品</TableHead>
              <TableHead class="w-24">入库状态</TableHead>
              <TableHead class="w-20 text-center sticky right-0 z-20 bg-muted"
                >操作</TableHead
              >
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="item in resourceItems"
              :key="item.id"
              class="hover:bg-muted/20 group"
            >
              <TableCell>{{ item.id }}</TableCell>
              <TableCell>
                <div class="min-w-0 space-y-1">
                  <p class="truncate font-medium" :title="item.companyName">
                    {{ item.companyName }}
                  </p>
                </div>
              </TableCell>
              <TableCell>
                <Badge variant="outline" class="font-normal bg-background">
                  {{ platformLabel(item.platform) }}
                </Badge>
              </TableCell>
              <TableCell>{{
                item.city || item.province || item.country || "-"
              }}</TableCell>
              <TableCell>{{ item.businessType || "-" }}</TableCell>
              <TableCell>
                <div
                  class="max-w-[160px] truncate text-xs text-muted-foreground"
                  :title="item.keyword"
                >
                  {{ item.keyword || "-" }}
                </div>
              </TableCell>
              <TableCell>
                <div
                  class="max-w-[320px] truncate text-xs text-muted-foreground"
                  :title="formatProducts(item)"
                >
                  {{ formatProducts(item) }}
                </div>
              </TableCell>
              <TableCell>
                <Badge
                  :class="
                    item.isNewCompany
                      ? 'bg-emerald-500 hover:bg-emerald-600 border-none'
                      : 'border-border/60 bg-muted/40 text-muted-foreground'
                  "
                >
                  {{ item.isNewCompany ? "新发掘" : "已存在" }}
                </Badge>
              </TableCell>
              <TableCell
                class="text-center sticky right-0 bg-background z-10 border-l border-border"
              >
                <div class="flex items-center justify-center gap-1.5">
                  <TooltipProvider :delayDuration="200">
                    <Tooltip>
                      <TooltipTrigger as-child>
                        <Button
                          variant="ghost"
                          size="icon"
                          class="h-7 w-7 shrink-0"
                          @click="handleCopy(item.companyName)"
                        >
                          <Copy class="h-3.5 w-3.5 text-muted-foreground" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>复制企业名称</p>
                      </TooltipContent>
                    </Tooltip>

                    <Tooltip v-if="item.companyUrl">
                      <TooltipTrigger as-child>
                        <Button
                          variant="ghost"
                          size="icon"
                          as-child
                          class="h-7 w-7 shrink-0"
                        >
                          <a
                            :href="item.companyUrl"
                            target="_blank"
                            rel="noreferrer"
                          >
                            <SquareArrowOutUpRight
                              class="h-3.5 w-3.5 text-primary/80"
                            />
                          </a>
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>跳转企业主页</p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
              </TableCell>
            </TableRow>

            <EmptyTablePlaceholder
              v-if="resourceItems.length === 0"
              :colspan="9"
            />
          </TableBody>
        </Table>
      </div>

      <div class="shrink-0 pb-4">
        <Pagination
          :current-page="Math.max(0, resourcePage - 1)"
          :total-pages="resourceTotalPages"
          :page-size="resourcePageSize"
          :selected-count="0"
          :total-count="resourceTotal"
          :show-selection="false"
          @update:current-page="handleResultPageChange"
          @update:page-size="handleResultPageSizeChange"
        />
      </div>
    </div>
  </div>
</template>
