<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { Loader2, RefreshCw } from "lucide-vue-next"

import { listCustomerAssignments } from "@/api/modules/customers"
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
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
import type { CustomerAssignmentItem } from "@/types/customer"

const loading = ref(false)
const errorMessage = ref("")
const items = ref<CustomerAssignmentItem[]>([])
const totalCount = ref(0)
const pageIndex = ref(0)
const pageSize = ref(20)

const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / pageSize.value)))

const formatDate = (value?: string) => {
  if (!value) return "-"
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return "-"
  return parsed.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  })
}

const formatText = (value?: string) => {
  const trimmed = (value || "").trim()
  return trimmed || "-"
}

const fetchItems = async () => {
  loading.value = true
  errorMessage.value = ""
  try {
    const result = await listCustomerAssignments({
      page: pageIndex.value + 1,
      pageSize: pageSize.value,
    })
    items.value = result.items || []
    totalCount.value = result.total || 0
  } catch (error) {
    items.value = []
    totalCount.value = 0
    errorMessage.value = getRequestErrorMessage(error, "加载客户分配列表失败")
  } finally {
    loading.value = false
  }
}

const refreshList = () => {
  void fetchItems()
}

const handlePageChange = (nextPage: number) => {
  if (nextPage === pageIndex.value) return
  pageIndex.value = nextPage
  void fetchItems()
}

const handlePageSizeChange = (nextPageSize: number) => {
  const changed = nextPageSize !== pageSize.value
  pageSize.value = nextPageSize
  pageIndex.value = 0
  if (!changed) return
  void fetchItems()
}

onMounted(() => {
  void fetchItems()
})
</script>

<template>
  <div class="space-y-6">
    <Card class="border-border/60 shadow-sm">
      <CardHeader>
        <div class="flex items-center justify-between gap-3">
          <Button size="sm" variant="outline" class="bg-background" :disabled="loading" @click="refreshList">
            <Loader2 v-if="loading" class="h-3.5 w-3.5 animate-spin" />
            <RefreshCw v-else class="h-3.5 w-3.5" />
            刷新
          </Button>
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
                  <TableHead class="min-w-[160px] whitespace-nowrap">日期</TableHead>
                  <TableHead class="min-w-[100px] whitespace-nowrap">电销</TableHead>
                  <TableHead class="min-w-[100px] whitespace-nowrap">销售</TableHead>
                  <TableHead class="min-w-[160px] whitespace-nowrap">客户名称</TableHead>
                  <TableHead class="min-w-[160px] whitespace-nowrap">法人名称</TableHead>
                  <TableHead class="min-w-[140px] whitespace-nowrap">联系人名称</TableHead>
                  <TableHead class="min-w-[140px] whitespace-nowrap">手机号</TableHead>
                  <TableHead class="min-w-[220px] whitespace-nowrap">地址</TableHead>
                  <TableHead class="min-w-[240px] whitespace-nowrap">备注</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in items" :key="item.id">
                  <TableCell class="align-top">{{ formatDate(item.date) }}</TableCell>
                  <TableCell class="align-top">{{ formatText(item.insideSalesName) }}</TableCell>
                  <TableCell class="align-top">{{ formatText(item.salesName) }}</TableCell>
                  <TableCell class="align-top">{{ formatText(item.customerName) }}</TableCell>
                  <TableCell class="align-top">{{ formatText(item.legalName) }}</TableCell>
                  <TableCell class="align-top">{{ formatText(item.contactName) }}</TableCell>
                  <TableCell class="align-top">{{ formatText(item.mobile) }}</TableCell>
                  <TableCell class="max-w-[220px] whitespace-pre-wrap break-words align-top">
                    {{ formatText(item.address) }}
                  </TableCell>
                  <TableCell class="max-w-[240px] whitespace-pre-wrap break-words align-top">
                    {{ formatText(item.remark) }}
                  </TableCell>
                </TableRow>
                <EmptyTablePlaceholder
                  v-if="items.length === 0"
                  :colspan="9"
                  text="当前没有可显示的电销分配给销售的客户记录。"
                  height-class="h-40"
                />
              </TableBody>
            </Table>
          </div>
        </div>

        <div class="mt-4 flex justify-end">
          <Pagination
            :current-page="pageIndex"
            :total-pages="totalPages"
            :page-size="pageSize"
            @update:current-page="handlePageChange"
            @update:page-size="handlePageSizeChange"
          />
        </div>
      </CardContent>
    </Card>
  </div>
</template>
