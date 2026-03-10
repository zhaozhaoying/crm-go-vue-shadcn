<script setup lang="ts">
import { computed, onMounted, onActivated, ref } from "vue";
import { Loader2, RefreshCw, Search } from "lucide-vue-next";

import { getAllOperationFollowRecords, type OperationFollowRecord } from "@/api/modules/followRecords";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
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
import { getRequestErrorMessage } from "@/lib/http-error";
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue";

const loading = ref(false);
const error = ref("");
const records = ref<OperationFollowRecord[]>([]);
const totalCount = ref(0);
const showSearch = ref(false);
const pageIndex = ref(0);
const pageSize = ref(10);

interface SearchForm {
  customerName: string;
  operatorUserName: string;
  content: string;
}

const createEmptySearchForm = (): SearchForm => {
  return {
    customerName: "",
    operatorUserName: "",
    content: "",
  };
};

const searchForm = ref<SearchForm>(createEmptySearchForm());
const activeSearchForm = ref<SearchForm>(createEmptySearchForm());

const totalPages = computed(() =>
  Math.max(1, Math.ceil(totalCount.value / pageSize.value)),
);

const formatDate = (dateStr?: string) => {
  if (!dateStr) return "-";
  return new Date(dateStr).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
};

const buildListParams = () => {
  return {
    page: pageIndex.value + 1,
    pageSize: pageSize.value,
  };
};

const fetchRecords = async () => {
  loading.value = true;
  error.value = "";
  try {
    const result = await getAllOperationFollowRecords(
      buildListParams().page,
      buildListParams().pageSize,
    );
    records.value = result.items;
    totalCount.value = result.total;

    // 客户端搜索过滤
    if (activeSearchForm.value.customerName || activeSearchForm.value.operatorUserName || activeSearchForm.value.content) {
      records.value = records.value.filter(record => {
        const customerKeyword = activeSearchForm.value.customerName.trim();
        const customer = record.customer;
        const matchCustomer = !customerKeyword ||
          (customer?.id ?? record.customerId).toString().includes(customerKeyword) ||
          (customer?.name || "").includes(customerKeyword) ||
          (customer?.contactName || "").includes(customerKeyword);
        const matchOperator = !activeSearchForm.value.operatorUserName ||
          (record.operatorUserName || "").includes(activeSearchForm.value.operatorUserName);
        const matchContent = !activeSearchForm.value.content ||
          (record.content || "").includes(activeSearchForm.value.content);
        return matchCustomer && matchOperator && matchContent;
      });
    }
  } catch (err) {
    records.value = [];
    totalCount.value = 0;
    error.value = getRequestErrorMessage(err, "加载跟进记录失败");
  } finally {
    loading.value = false;
  }
};

const refreshList = () => {
  fetchRecords();
};

const handleSearchClick = () => {
  showSearch.value = !showSearch.value;
};

const handleSearch = () => {
  activeSearchForm.value = { ...searchForm.value };
  pageIndex.value = 0;
  fetchRecords();
};

const clearSearch = () => {
  searchForm.value = createEmptySearchForm();
  activeSearchForm.value = createEmptySearchForm();
  pageIndex.value = 0;
  fetchRecords();
};

const handlePageChange = (nextPage: number) => {
  if (nextPage === pageIndex.value) return;
  pageIndex.value = nextPage;
  fetchRecords();
};

const handlePageSizeChange = (nextPageSize: number) => {
  const changed = nextPageSize !== pageSize.value;
  pageSize.value = nextPageSize;
  pageIndex.value = 0;
  if (changed) {
    fetchRecords();
  }
};

onMounted(() => {
  fetchRecords();
});
onActivated(fetchRecords);
</script>

<template>
  <div class="w-full flex flex-col gap-4 lg:gap-6">
    <Card class="shadow-sm border-border/60">
      <CardHeader v-if="showSearch" class="border-b space-y-3">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >客户</label
            >
            <Input
              v-model="searchForm.customerName"
              placeholder="客户ID/名称/联系人"
              class="h-9 w-40"
            />
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >操作人</label
            >
            <Input
              v-model="searchForm.operatorUserName"
              placeholder="操作人"
              class="h-9 w-40"
            />
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >跟进内容</label
            >
            <Input
              v-model="searchForm.content"
              placeholder="跟进内容"
              class="h-9 w-40"
            />
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <div class="flex items-center gap-2 ml-auto">
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
        <div class="mb-4 flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <Button size="sm" @click="refreshList">
              <RefreshCw class="h-4 w-4" />
            </Button>
          </div>
          <div class="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              class="h-9 w-9"
              @click="handleSearchClick"
            >
              <Search class="h-4 w-4" />
            </Button>
          </div>
        </div>

        <div
          class="overflow-hidden rounded-lg border border-border/60 bg-background"
        >
          <div v-if="loading" class="flex items-center justify-center py-24">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
          </div>

          <div v-else-if="error" class="py-20 text-center text-destructive">
            {{ error }}
          </div>

          <Table v-else class="w-max min-w-full">
            <TableHeader class="sticky top-0 z-20 bg-muted/40">
              <TableRow>
                <TableHead class="w-16 whitespace-nowrap">编号</TableHead>
                <TableHead class="w-48 whitespace-nowrap">客户信息</TableHead>
                <TableHead class="whitespace-nowrap">跟进内容</TableHead>
                <TableHead class="w-28 whitespace-nowrap">跟进方式</TableHead>
                <TableHead class="w-28 whitespace-nowrap">客户级别</TableHead>
                <TableHead class="w-28 whitespace-nowrap">操作人</TableHead>
                <TableHead class="w-40 whitespace-nowrap">创建时间</TableHead>
                <TableHead class="w-40 whitespace-nowrap">约见时间</TableHead>
                <TableHead class="w-40 whitespace-nowrap">拍摄时间</TableHead>
                <TableHead class="w-40 whitespace-nowrap">下次跟进时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow
                v-for="record in records"
                :key="record.id"
                class="group hover:bg-muted/30 transition-colors"
              >
                <TableCell class="text-muted-foreground">{{
                  record.id
                }}</TableCell>
                <TableCell class="align-top">
                  <div class="space-y-0.5">
                    <div class="font-medium">
                      {{ record.customer?.name || "-" }}
                    </div>
                    <div class="text-xs text-muted-foreground">
                      ID: {{ record.customer?.id ?? record.customerId }}
                    </div>
                    <div class="text-xs text-muted-foreground">
                      联系人: {{ record.customer?.contactName || "-" }}
                    </div>
                  </div>
                </TableCell>
                <TableCell
                  class="max-w-[300px] truncate"
                  :title="record.content"
                >
                  {{ record.content || "-" }}
                </TableCell>
                <TableCell>
                  <Badge variant="outline" class="bg-background">
                    {{ record.followMethodName || '-' }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Badge variant="secondary">
                    {{ record.customerLevelName || '-' }}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Badge
                    variant="outline"
                    class="bg-background text-muted-foreground"
                  >
                    {{ record.operatorUserName || "未知" }}
                  </Badge>
                </TableCell>
                <TableCell class="text-xs text-muted-foreground">
                  {{ formatDate(record.createdAt) }}
                </TableCell>
                <TableCell class="text-xs text-muted-foreground">
                  <span>
                    {{ formatDate(record.appointmentTime) }}
                  </span>
                </TableCell>
                <TableCell class="text-xs text-muted-foreground">
                  <span>
                    {{ formatDate(record.shootingTime) }}
                  </span>
                </TableCell>
                <TableCell class="text-xs text-muted-foreground">
                  <span>
                    {{ formatDate(record.nextFollowTime) }}
                  </span>
                </TableCell>
              </TableRow>

              <EmptyTablePlaceholder
                v-if="records.length === 0"
                :colspan="10"
                text="暂无跟进记录"
              />
            </TableBody>
          </Table>
        </div>

        <div class="mt-4">
          <Pagination
            :current-page="pageIndex"
            :total-pages="totalPages"
            :page-size="pageSize"
            :show-selection="false"
            :total-count="totalCount"
            @update:current-page="handlePageChange"
            @update:page-size="handlePageSizeChange"
          />
        </div>
      </CardContent>
    </Card>
  </div>
</template>
