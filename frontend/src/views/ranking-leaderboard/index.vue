<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  CalendarDays,
  Crown,
  Download,
  Loader2,
  Medal,
  RefreshCw,
  Trophy,
  Activity,
  Phone,
  PhoneCall,
  Percent,
  Clock,
  UserPlus,
  CalendarCheck,
} from "lucide-vue-next";
import { toast } from "vue-sonner";

import {
  getRankingLeaderboard,
  getRankingLeaderboardDetail,
  type RankingLeaderboardItem,
  type RankingLeaderboardPeriod,
} from "@/api/modules/rankingLeaderboard";
import {
  getTelemarketingDailyScoreDetail,
  getTelemarketingDailyScoreRankings,
  type TelemarketingDailyScore,
  type TelemarketingDailyScoreDetail,
  type TelemarketingDailyScoreRankingItem,
} from "@/api/modules/telemarketingDailyScore";
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { getRequestErrorMessage } from "@/lib/http-error";

type Period = Exclude<RankingLeaderboardPeriod, "all">;

interface RankingFilters {
  period: Period;
  startDate: string;
  endDate: string;
}

interface LeaderboardDetailState {
  period: string;
  startDate: string;
  endDate: string;
  rank: number;
  totalUsers: number;
  hasData: boolean;
  score: RankingLeaderboardItem;
}

const rankingLeaderboardRefreshEvent = "ranking-leaderboard:refresh";
const validPeriods: Period[] = ["month", "week", "day"];

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const refreshing = ref(false);
const errorMessage = ref("");
const activePeriod = ref<Period>("day");
const startDate = ref("");
const endDate = ref("");
const activeStartDate = ref("");
const activeEndDate = ref("");
const items = ref<RankingLeaderboardItem[]>([]);
const detailOpen = ref(false);
const detailLoading = ref(false);
const detailErrorMessage = ref("");
const detail = ref<LeaderboardDetailState | null>(null);
const selectedItem = ref<RankingLeaderboardItem | null>(null);

const displayEmptyText = "暂无电销排名数据";

function formatDate(date: Date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function parseDate(value?: string | null) {
  const normalized = String(value || "")
    .trim()
    .slice(0, 10);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(normalized)) {
    return null;
  }
  const [year, month, day] = normalized.split("-").map((item) => Number(item));
  const date = new Date(year, month - 1, day);
  if (Number.isNaN(date.getTime())) {
    return null;
  }
  return date;
}

function getTodayDate() {
  return formatDate(new Date());
}

function getDateOffset(date: Date, offsetDays: number) {
  const cloned = new Date(date);
  cloned.setDate(cloned.getDate() + offsetDays);
  return cloned;
}

function getWeekStartDate(date: Date) {
  const weekday = date.getDay() === 0 ? 7 : date.getDay();
  return formatDate(getDateOffset(date, -(weekday - 1)));
}

function getMonthStartDate(date: Date) {
  return formatDate(new Date(date.getFullYear(), date.getMonth(), 1));
}

function normalizeDateValue(value?: string | null) {
  return String(value || "")
    .trim()
    .slice(0, 10);
}

function buildPresetRange(period: Period) {
  const today = parseDate(getTodayDate()) || new Date();
  switch (period) {
    case "day":
      return {
        startDate: formatDate(today),
        endDate: formatDate(today),
      };
    case "week":
      return {
        startDate: getWeekStartDate(today),
        endDate: formatDate(today),
      };
    case "month":
    default:
      return {
        startDate: getMonthStartDate(today),
        endDate: formatDate(today),
      };
  }
}

function normalizeFilters(period: Period) {
  const normalized = buildPresetRange(period);
  return {
    period,
    startDate: normalized.startDate,
    endDate: normalized.endDate,
  };
}

function resolveFiltersFromRoute(): RankingFilters {
  const routePeriod = String(route.query.period || "").trim() as Period;
  const period = validPeriods.includes(routePeriod) ? routePeriod : "month";
  return normalizeFilters(period);
}

function routeMatchesFilters(filters: RankingFilters) {
  return (
    String(route.query.period || "").trim() === filters.period &&
    !normalizeDateValue(route.query.startDate as string | undefined) &&
    !normalizeDateValue(route.query.endDate as string | undefined)
  );
}

const formatDuration = (seconds: number) => {
  const safe = Math.max(0, Math.floor(Number(seconds) || 0));
  const hours = Math.floor(safe / 3600);
  const minutes = Math.floor((safe % 3600) / 60);
  const remain = safe % 60;
  if (hours > 0) return `${hours}小时${minutes}分${remain}秒`;
  if (minutes > 0) return `${minutes}分${remain}秒`;
  return `${remain}秒`;
};

const formatAnswerRate = (value?: number) => {
  const safe = Number(value || 0);
  if (!Number.isFinite(safe) || safe <= 0) return "0%";
  return `${safe.toFixed(1)}%`;
};

const firstNonEmptyText = (
  ...values: Array<string | number | undefined | null>
) => {
  for (const value of values) {
    const normalized = String(value ?? "").trim();
    if (normalized) {
      return normalized;
    }
  }
  return "";
};

const resolveLeaderboardIdentityKey = (
  item?: Partial<RankingLeaderboardItem> | null,
  fallbackRank?: number,
) => {
  if (!item) {
    return "";
  }
  const seatWorkNumber = String(item.seatWorkNumber || "").trim();
  if (seatWorkNumber) {
    return `w:${seatWorkNumber}`;
  }
  const matchedUserId = Number(item.matchedUserId || 0);
  if (Number.isFinite(matchedUserId) && matchedUserId > 0) {
    return `u:${matchedUserId}`;
  }
  const currentIdentityKey = String(item.identityKey || "").trim();
  if (currentIdentityKey) {
    return currentIdentityKey;
  }
  const fallbackText = firstNonEmptyText(
    item.matchedUserName,
    item.seatName,
    item.groupName,
    item.totalScore,
    fallbackRank,
  );
  return fallbackText ? `n:${fallbackText}` : "";
};

const normalizeRankingLeaderboardItem = (
  item: RankingLeaderboardItem,
): RankingLeaderboardItem => ({
  ...item,
  identityKey: resolveLeaderboardIdentityKey(item, item.rank),
});

const mapTelemarketingDailyScoreToLeaderboardItem = (
  item: Pick<
    TelemarketingDailyScore,
    | "scoreDate"
    | "seatWorkNumber"
    | "seatName"
    | "matchedUserId"
    | "matchedUserName"
    | "groupName"
    | "roleName"
    | "callNum"
    | "answeredCallCount"
    | "answerRate"
    | "callDurationSecond"
    | "newCustomerCount"
    | "invitationCount"
    | "callScore"
    | "invitationScore"
    | "newCustomerScore"
    | "totalScore"
  >,
  rank: number,
): RankingLeaderboardItem => {
  const scoreDate = normalizeDateValue(item.scoreDate) || getTodayDate();
  const seatWorkNumber = String(item.seatWorkNumber || "").trim();
  const identitySeed =
    seatWorkNumber || String(item.matchedUserId || "") || String(rank);
  return {
    identityKey: `day:${scoreDate}:${identitySeed}`,
    rank,
    seatWorkNumber,
    seatName: item.seatName || "",
    matchedUserId: item.matchedUserId,
    matchedUserName: item.matchedUserName || "",
    groupName: item.groupName || "",
    roleName: item.roleName || "",
    callNum: item.callNum || 0,
    answeredCallCount: item.answeredCallCount || 0,
    answerRate: item.answerRate || 0,
    callDurationSecond: item.callDurationSecond || 0,
    newCustomerCount: item.newCustomerCount || 0,
    invitationCount: item.invitationCount || 0,
    callScore: item.callScore || 0,
    invitationScore: item.invitationScore || 0,
    newCustomerScore: item.newCustomerScore || 0,
    totalScore: item.totalScore || 0,
    scoreDays: 1,
  };
};

const mapTelemarketingDailyRankingItem = (
  item: TelemarketingDailyScoreRankingItem,
): RankingLeaderboardItem =>
  mapTelemarketingDailyScoreToLeaderboardItem(item, item.rank);

const mapTelemarketingDailyDetail = (
  result: TelemarketingDailyScoreDetail,
): LeaderboardDetailState => {
  const scoreDate =
    normalizeDateValue(result.scoreDate || result.score.scoreDate) ||
    getTodayDate();
  return {
    period: "day",
    startDate: scoreDate,
    endDate: scoreDate,
    rank: result.rank,
    totalUsers: result.totalUsers,
    hasData: result.hasData,
    score: mapTelemarketingDailyScoreToLeaderboardItem(
      result.score,
      result.rank,
    ),
  };
};

const getTelemarketingDisplayName = (item: RankingLeaderboardItem) => {
  return item.matchedUserName || item.seatName || item.seatWorkNumber || "-";
};

const activeRangeText = computed(() => {
  if (!activeStartDate.value || !activeEndDate.value) {
    return "-";
  }
  if (activeStartDate.value === activeEndDate.value) {
    return activeStartDate.value;
  }
  return `${activeStartDate.value} 至 ${activeEndDate.value}`;
});

const periodLabel = computed(() => {
  switch (activePeriod.value) {
    case "day":
      return "日排名";
    case "week":
      return "周排名";
    case "month":
    default:
      return "月排名";
  }
});

const refreshButtonText = computed(() => {
  if (
    activePeriod.value === "day" &&
    activeStartDate.value === activeEndDate.value &&
    activeEndDate.value === getTodayDate()
  ) {
    return "刷新排名";
  }
  return "刷新列表";
});

const selectedIdentityKey = computed(() =>
  resolveLeaderboardIdentityKey(selectedItem.value, selectedItem.value?.rank),
);
const detailScore = computed(() => detail.value?.score || selectedItem.value);
const detailDisplayName = computed(() => {
  if (!detailScore.value) {
    return "-";
  }
  return getTelemarketingDisplayName(detailScore.value);
});
const detailRankText = computed(() => {
  if (!detail.value?.rank) {
    return "未生成排名";
  }
  return `第 ${detail.value.rank} 名`;
});
const detailHeroRank = computed(() => {
  if (!detail.value?.rank) {
    return "--";
  }
  return `#${detail.value.rank}`;
});
const detailScoreSegments = computed(() => {
  if (!detailScore.value) {
    return [];
  }

  const totalScore =
    detailScore.value.totalScore > 0 ? detailScore.value.totalScore : 1;
  return [
    {
      key: "call",
      label: "电话积分",
      value: detailScore.value.callScore,
      percent: (detailScore.value.callScore / totalScore) * 100,
      barClass: "bg-sky-500",
    },
    {
      key: "invitation",
      label: "邀约积分",
      value: detailScore.value.invitationScore,
      percent: (detailScore.value.invitationScore / totalScore) * 100,
      barClass: "bg-emerald-500",
    },
    {
      key: "customer",
      label: "新增客户积分",
      value: detailScore.value.newCustomerScore,
      percent: (detailScore.value.newCustomerScore / totalScore) * 100,
      barClass: "bg-amber-500",
    },
  ];
});

const downloadCsv = (rows: string[][], filename: string) => {
  const csv = rows
    .map((row) =>
      row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(","),
    )
    .join("\n");

  const blob = new Blob(["\uFEFF" + csv], {
    type: "text/csv;charset=utf-8;",
  });
  const url = window.URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  window.URL.revokeObjectURL(url);
};

const exportRankings = () => {
  if (items.value.length === 0) {
    toast.error("暂无可导出数据");
    return;
  }

  const rows = [
    [
      "榜单类型",
      "开始日期",
      "结束日期",
      "排名",
      "工号",
      "电销",
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
      periodLabel.value,
      activeStartDate.value,
      activeEndDate.value,
      String(item.rank),
      item.seatWorkNumber || "-",
      getTelemarketingDisplayName(item),
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
  ];

  downloadCsv(
    rows,
    `telemarketing-rankings-${activePeriod.value}-${activeStartDate.value}-${activeEndDate.value}.csv`,
  );
};

const fetchRankings = async (
  filters: RankingFilters,
  options?: {
    fromRefresh?: boolean;
  },
) => {
  if (options?.fromRefresh) {
    refreshing.value = true;
  } else {
    loading.value = true;
  }
  errorMessage.value = "";

  try {
    if (filters.period === "day") {
      const result = await getTelemarketingDailyScoreRankings({
        scoreDate: filters.endDate,
      });
      const scoreDate = normalizeDateValue(result.scoreDate) || filters.endDate;
      items.value = (result.items || []).map(mapTelemarketingDailyRankingItem);
      activePeriod.value = "day";
      activeStartDate.value = scoreDate;
      activeEndDate.value = scoreDate;
      startDate.value = scoreDate;
      endDate.value = scoreDate;
      return true;
    }

    const result = await getRankingLeaderboard(filters);
    items.value = (result.items || []).map(normalizeRankingLeaderboardItem);
    activePeriod.value = (
      validPeriods.includes(result.period as Period)
        ? result.period
        : filters.period
    ) as Period;
    activeStartDate.value = result.startDate || filters.startDate;
    activeEndDate.value = result.endDate || filters.endDate;
    startDate.value = activeStartDate.value;
    endDate.value = activeEndDate.value;
    return true;
  } catch (error) {
    items.value = [];
    errorMessage.value = getRequestErrorMessage(
      error,
      filters.period === "day" ? "加载电销每日排名失败" : "加载电销排名失败",
    );
    return false;
  } finally {
    if (options?.fromRefresh) {
      refreshing.value = false;
    } else {
      loading.value = false;
    }
  }
};

const fetchDetail = async () => {
  if (!detailOpen.value) {
    return;
  }

  detailLoading.value = true;
  detailErrorMessage.value = "";
  detail.value = null;

  try {
    if (activePeriod.value === "day") {
      const seatWorkNumber = String(
        selectedItem.value?.seatWorkNumber || "",
      ).trim();
      if (!seatWorkNumber) {
        throw new Error("无效的坐席工号");
      }
      const result = await getTelemarketingDailyScoreDetail(seatWorkNumber, {
        scoreDate: activeEndDate.value || endDate.value,
      });
      detail.value = mapTelemarketingDailyDetail(result);
      return;
    }

    if (!selectedIdentityKey.value) {
      throw new Error("无效的榜单标识");
    }
    const result = await getRankingLeaderboardDetail(
      selectedIdentityKey.value,
      {
        period: activePeriod.value,
        startDate: activeStartDate.value || startDate.value,
        endDate: activeEndDate.value || endDate.value,
      },
    );
    detail.value = {
      ...result,
      score: normalizeRankingLeaderboardItem(result.score),
    };
  } catch (error) {
    detail.value = null;
    detailErrorMessage.value = getRequestErrorMessage(
      error,
      activePeriod.value === "day"
        ? "加载电销积分详情失败"
        : "加载榜单详情失败",
    );
  } finally {
    detailLoading.value = false;
  }
};

const replaceRouteQuery = async (filters: RankingFilters) => {
  await router.replace({
    query: {
      period: filters.period,
    },
  });
};

const handlePeriodChange = (value: string | number) => {
  const period = String(value) as Period;
  if (!validPeriods.includes(period)) {
    return;
  }
  void replaceRouteQuery(normalizeFilters(period));
};

const refreshRankings = async () => {
  const success = await fetchRankings(
    {
      period: activePeriod.value,
      startDate: startDate.value,
      endDate: endDate.value,
    },
    { fromRefresh: true },
  );
  if (success) {
    toast.success(
      activePeriod.value === "day" &&
        activeStartDate.value === activeEndDate.value &&
        activeEndDate.value === getTodayDate()
        ? "日排名已刷新"
        : "榜单已更新",
    );
  }
};

const openDetail = (item: RankingLeaderboardItem) => {
  selectedItem.value = item;
  detailOpen.value = true;
};

const resetDetailState = () => {
  detailLoading.value = false;
  detailErrorMessage.value = "";
  detail.value = null;
};

const handleDetailOpenChange = (open: boolean) => {
  detailOpen.value = open;
  if (!open) {
    selectedItem.value = null;
    resetDetailState();
  }
};

const handleRankingRefreshEvent = () => {
  void refreshRankings();
};

watch(
  () => [route.query.period, route.query.startDate, route.query.endDate],
  () => {
    const filters = resolveFiltersFromRoute();
    activePeriod.value = filters.period;
    startDate.value = filters.startDate;
    endDate.value = filters.endDate;

    if (!routeMatchesFilters(filters)) {
      void replaceRouteQuery(filters);
      return;
    }

    void fetchRankings(filters);
  },
  { immediate: true },
);

watch(
  () => [
    detailOpen.value,
    selectedIdentityKey.value,
    activePeriod.value,
    activeStartDate.value,
    activeEndDate.value,
  ],
  ([open, identityKey]) => {
    if (!open || !identityKey) {
      if (!open) {
        resetDetailState();
      }
      return;
    }
    void fetchDetail();
  },
);

onMounted(() => {
  window.addEventListener(
    rankingLeaderboardRefreshEvent,
    handleRankingRefreshEvent,
  );
});
onBeforeUnmount(() => {
  window.removeEventListener(
    rankingLeaderboardRefreshEvent,
    handleRankingRefreshEvent,
  );
});
</script>

<template>
  <div class="space-y-6">
    <Card class="border-border/60 shadow-sm">
      <CardHeader class="space-y-4">
        <div class="overflow-hidden">
          <div class="flex flex-wrap items-center gap-4 py-2">
            <Button size="sm" variant="outline" class="h-10 rounded-md bg-background px-4" @click="refreshRankings"
              :disabled="loading || refreshing">
              <Loader2 v-if="refreshing" class="h-3.5 w-3.5 animate-spin" />
              <RefreshCw v-else class="h-3.5 w-3.5" />
              {{ refreshButtonText }}
            </Button>
            <Button size="sm" variant="outline" class="h-10 rounded-md bg-background px-4" @click="exportRankings"
              :disabled="loading || items.length === 0">
              <Download class="h-3.5 w-3.5" />
              导出
            </Button>

            <Tabs :model-value="activePeriod" @update:model-value="handlePeriodChange" class="min-w-[240px] flex-1">
              <TabsList class="grid h-10 w-full grid-cols-3 rounded-lg bg-muted/60">
                <TabsTrigger value="day">日排名</TabsTrigger>
                <TabsTrigger value="week">周排名</TabsTrigger>
                <TabsTrigger value="month">月排名</TabsTrigger>
              </TabsList>
            </Tabs>
          </div>
        </div>
      </CardHeader>

      <CardContent class="pt-0">
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
                  <TableHead
                    class="sticky right-0 z-30 w-[80px] min-w-[80px] border-l border-border bg-muted/95 text-center before:absolute before:left-0 before:top-0 before:h-full before:w-px before:bg-border">
                    操作
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in items" :key="item.identityKey"
                  class="group transition-colors hover:bg-muted/20">
                  <TableCell>
                    <Badge :variant="item.rank <= 3 ? 'default' : 'outline'" class="gap-1.5">
                      <Crown v-if="item.rank === 1" class="h-3.5 w-3.5 text-amber-200" />
                      <Medal v-else-if="item.rank === 2 || item.rank === 3" class="h-3.5 w-3.5" :class="item.rank === 2 ? 'text-slate-100' : 'text-orange-100'
                        " />
                      第{{ item.rank }}名
                    </Badge>
                  </TableCell>
                  <TableCell class="font-medium">{{
                    item.seatWorkNumber || "-"
                    }}</TableCell>
                  <TableCell class="font-medium">{{
                    getTelemarketingDisplayName(item)
                    }}</TableCell>
                  <TableCell>
                    <span class="text-base font-semibold text-primary">{{
                      item.totalScore
                      }}</span>
                  </TableCell>
                  <TableCell>{{ item.callScore }}</TableCell>
                  <TableCell>{{ item.invitationScore }}</TableCell>
                  <TableCell>{{ item.newCustomerScore }}</TableCell>
                  <TableCell>{{ item.callNum }}</TableCell>
                  <TableCell>{{ item.answeredCallCount }}</TableCell>
                  <TableCell>{{ formatAnswerRate(item.answerRate) }}</TableCell>
                  <TableCell>{{
                    formatDuration(item.callDurationSecond)
                    }}</TableCell>
                  <TableCell>{{ item.newCustomerCount }}</TableCell>
                  <TableCell>{{ item.invitationCount }}</TableCell>
                  <TableCell>{{ item.scoreDays }}</TableCell>
                  <TableCell
                    class="sticky right-0 z-10 w-[80px] min-w-[80px] border-l border-border bg-background text-center before:absolute before:left-0 before:top-0 before:h-full before:w-px before:bg-border">
                    <div class="flex justify-end">
                      <Button variant="ghost" size="sm" @click="openDetail(item)">详情</Button>
                    </div>
                  </TableCell>
                </TableRow>
                <EmptyTablePlaceholder v-if="items.length === 0" :colspan="15" :text="displayEmptyText" />
              </TableBody>
            </Table>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>

  <Sheet :open="detailOpen" @update:open="handleDetailOpenChange">
    <SheetContent side="right"
      class="w-[94vw] max-w-none overflow-y-auto border-l bg-background p-0 sm:w-[720px] sm:max-w-[720px]">
      <div class="flex min-h-full flex-col bg-muted/10">
        <SheetHeader class="border-b bg-background px-6 py-5 text-left">
          <div class="flex items-start justify-between gap-4 pr-8">
            <div class="space-y-1">
              <SheetTitle class="flex items-center gap-3 text-2xl font-semibold tracking-tight">
                {{ detailDisplayName }}
                <span class="text-lg font-semibold text-muted-foreground">{{
                  detailHeroRank
                  }}</span>
                <Badge variant="outline" class="ml-2 font-normal text-xs">{{
                  periodLabel
                  }}</Badge>
              </SheetTitle>
              <SheetDescription v-if="detailScore?.seatWorkNumber">
                工号：{{ detailScore.seatWorkNumber }}
              </SheetDescription>
            </div>
          </div>
        </SheetHeader>

        <div class="flex-1 space-y-6 px-6 py-6">
          <Card v-if="detailErrorMessage" class="border-red-200 bg-red-50/40">
            <CardContent class="pt-6">
              <p class="text-sm text-red-600">{{ detailErrorMessage }}</p>
            </CardContent>
          </Card>

          <div v-else-if="detailLoading" class="flex items-center justify-center py-24">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
          </div>

          <template v-else-if="detail && detailScore">
            <Card class="border-border/60 shadow-sm overflow-hidden">
              <div class="bg-primary/5 px-6 py-5 border-b border-border/60 flex items-center justify-between">
                <div>
                  <div class="font-medium text-primary flex items-center gap-2">
                    <Trophy class="h-5 w-5" />
                    总积分
                  </div>
                </div>
                <div class="h-16 w-16 flex items-center justify-center">
                  <div class="mt-2 text-4xl font-bold text-primary">
                    {{ detailScore.totalScore }}
                  </div>
                </div>
              </div>
              <CardContent class="px-6 py-5 space-y-5">
                <div class="space-y-3">
                  <div class="flex items-center justify-between text-sm font-medium">
                    <span>积分构成</span>
                    <span class="text-muted-foreground">占比</span>
                  </div>
                  <div class="flex h-3 overflow-hidden rounded-full bg-muted">
                    <div v-for="segment in detailScoreSegments" :key="segment.key" class="h-full transition-all"
                      :class="segment.barClass" :style="{ width: `${Math.max(segment.percent, 0)}%` }"
                      :title="`${segment.label}: ${segment.value}分`" />
                  </div>
                  <div class="grid gap-3 pt-2 md:grid-cols-3">
                    <div v-for="segment in detailScoreSegments" :key="segment.key"
                      class="rounded-lg border border-border/60 bg-background px-4 py-3 transition-colors hover:bg-muted/50">
                      <div class="flex items-center gap-2">
                        <div class="h-2.5 w-2.5 rounded-full" :class="segment.barClass"></div>
                        <div class="text-sm text-muted-foreground">
                          {{ segment.label }}
                        </div>
                      </div>
                      <div class="mt-2 text-2xl font-semibold">
                        {{ segment.value }}
                      </div>
                      <p class="mt-1 text-xs text-muted-foreground">
                        占比 {{ formatAnswerRate(segment.percent) }}
                      </p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card class="border-border/60 shadow-sm">
              <CardHeader class="pb-4">
                <CardTitle class="text-lg flex items-center gap-2">
                  <Activity class="h-5 w-5 text-primary" />
                  详细数据
                </CardTitle>
                <CardDescription>
                  该员工在当前排名周期内的详细业务产出及工作量统计
                </CardDescription>
              </CardHeader>
              <CardContent class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                <div
                  class="group rounded-lg border border-border/60 bg-background px-4 py-3 transition-all hover:border-primary/20 hover:bg-muted/20 hover:shadow-sm">
                  <div class="flex items-center justify-between text-sm font-medium text-muted-foreground">
                    拨打数
                    <Phone class="h-4 w-4 text-muted-foreground/50 transition-colors group-hover:text-primary/70" />
                  </div>
                  <div class="mt-1.5 text-2xl font-semibold">
                    {{ detailScore.callNum }}
                  </div>
                  <div class="mt-3 text-xs text-muted-foreground bg-muted/40 p-2 rounded-md">
                    通过系统呼出的总电话数量
                  </div>
                </div>

                <div
                  class="group rounded-lg border border-border/60 bg-background px-4 py-3 transition-all hover:border-primary/20 hover:bg-muted/20 hover:shadow-sm">
                  <div class="flex items-center justify-between text-sm font-medium text-muted-foreground">
                    接通数
                    <PhoneCall class="h-4 w-4 text-muted-foreground/50 transition-colors group-hover:text-primary/70" />
                  </div>
                  <div class="mt-1.5 text-2xl font-semibold">
                    {{ detailScore.answeredCallCount }}
                  </div>
                  <div class="mt-3 text-xs text-muted-foreground bg-muted/40 p-2 rounded-md">
                    客户实际接听的电话数量
                  </div>
                </div>

                <div
                  class="group rounded-lg border border-border/60 bg-background px-4 py-3 transition-all hover:border-primary/20 hover:bg-muted/20 hover:shadow-sm">
                  <div class="flex items-center justify-between text-sm font-medium text-muted-foreground">
                    接通率
                    <Percent class="h-4 w-4 text-muted-foreground/50 transition-colors group-hover:text-primary/70" />
                  </div>
                  <div class="mt-1.5 text-2xl font-semibold">
                    {{ formatAnswerRate(detailScore.answerRate) }}
                  </div>
                  <div class="mt-3 text-xs text-muted-foreground bg-muted/40 p-2 rounded-md">
                    接通数占比总拨打数的百分比
                  </div>
                </div>

                <div
                  class="group rounded-lg border border-border/60 bg-background px-4 py-3 transition-all hover:border-primary/20 hover:bg-muted/20 hover:shadow-sm">
                  <div class="flex items-center justify-between text-sm font-medium text-muted-foreground">
                    通话时长
                    <Clock class="h-4 w-4 text-muted-foreground/50 transition-colors group-hover:text-primary/70" />
                  </div>
                  <div class="mt-1.5 text-xl font-semibold">
                    {{ formatDuration(detailScore.callDurationSecond) }}
                  </div>
                  <div class="mt-3 text-xs text-muted-foreground bg-muted/40 p-2 rounded-md">
                    接通电话的累计主叫通话时长
                  </div>
                </div>

                <div
                  class="group rounded-lg border border-border/60 bg-background px-4 py-3 transition-all hover:border-primary/20 hover:bg-muted/20 hover:shadow-sm">
                  <div class="flex items-center justify-between text-sm font-medium text-muted-foreground">
                    新增客户数
                    <UserPlus class="h-4 w-4 text-muted-foreground/50 transition-colors group-hover:text-primary/70" />
                  </div>
                  <div class="mt-1.5 text-2xl font-semibold">
                    {{ detailScore.newCustomerCount }}
                  </div>
                  <div class="mt-3 text-xs text-muted-foreground bg-muted/40 p-2 rounded-md">
                    系统录入的有效新增客户数量
                  </div>
                </div>

                <div
                  class="group rounded-lg border border-border/60 bg-background px-4 py-3 transition-all hover:border-primary/20 hover:bg-muted/20 hover:shadow-sm">
                  <div class="flex items-center justify-between text-sm font-medium text-muted-foreground">
                    邀约数
                    <CalendarCheck
                      class="h-4 w-4 text-muted-foreground/50 transition-colors group-hover:text-primary/70" />
                  </div>
                  <div class="mt-1.5 text-2xl font-semibold">
                    {{ detailScore.invitationCount }}
                  </div>
                  <div class="mt-3 text-xs text-muted-foreground bg-muted/40 p-2 rounded-md">
                    成功录入系统的到店邀约记录数
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
