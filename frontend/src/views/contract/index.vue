<script setup lang="ts">
import { computed, onActivated, onMounted, ref } from "vue"
import {
  Loader2,
  Plus,
  RefreshCw,
  Search,
  ShieldCheck,
  SquarePen,
} from "lucide-vue-next"
import { toast } from "vue-sonner"

import { auditContract, createContract, listContracts, updateContract } from "@/api/modules/contracts"
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Pagination } from "@/components/ui/pagination"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { getRequestErrorMessage } from "@/lib/http-error"
import { hasAnyRole, isAdminUser } from "@/lib/auth-role"
import { useAuthStore } from "@/stores/auth"
import type {
  AuditContractRequest,
  Contract,
  ContractFormPayload,
  ContractListParams,
} from "@/types/contract"

import PopupForm from "./popupForm.vue"

const authStore = useAuthStore()

const loading = ref(false)
const submitting = ref(false)
const auditSubmitting = ref(false)
const error = ref("")
const contracts = ref<Contract[]>([])
const totalCount = ref(0)
const pageIndex = ref(0)
const pageSize = ref(10)

const keyword = ref("")
const activeKeyword = ref("")
const paymentStatus = ref("all")
const cooperationType = ref("all")
const auditStatus = ref("all")
const expiryHandlingStatus = ref("all")

const dialogOpen = ref(false)
const dialogMode = ref<"create" | "edit">("create")
const editingContract = ref<Contract | null>(null)
const dialogReadonly = ref(false)
const auditDialogOpen = ref(false)
const auditingContract = ref<Contract | null>(null)
const imagePreviewOpen = ref(false)
const imagePreviewSrc = ref("")
const imagePreviewTitle = ref("")

const totalPages = computed(() =>
  Math.max(1, Math.ceil(totalCount.value / pageSize.value)),
)
const canAuditContract = computed(
  () =>
    isAdminUser(authStore.user) ||
    hasAnyRole(authStore.user, [
      "finance_manager",
      "财务经理",
    ]),
)
const canOperateSiteService = computed(() =>
  hasAnyRole(authStore.user, [
    "ops_manager",
    "operation_manager",
    "ops_staff",
    "operation_staff",
    "运营经理",
    "运营员工",
    "运营",
  ]),
)

const hasSalesRole = computed(() =>
  hasAnyRole(authStore.user, [
    "sales_director",
    "sales_manager",
    "sales_staff",
    "sales_inside",
    "sales_outside",
    "销售总监",
    "销售经理",
    "销售员工",
    "销售",
    "Inside销售",
    "Outside销售",
  ]),
)

const isPendingAudit = (status?: string) =>
  String(status || "").trim() === "pending"

const canEditContract = (contract: Contract) => {
  // 管理员：任何时候都能改
  if (isAdminUser(authStore.user)) return true
  // 运营：只有审核通过/驳回后才能改
  if (canOperateSiteService.value) return !isPendingAudit(contract.auditStatus)
  // 销售：任何时候都能打开编辑（审核后字段会被后端/表单限制为仅备注）
  if (hasSalesRole.value) return true
  // 其他角色：仅待审核时可改
  return isPendingAudit(contract.auditStatus)
}

const buildListParams = (): ContractListParams => ({
  page: pageIndex.value + 1,
  pageSize: pageSize.value,
  keyword: activeKeyword.value || undefined,
  paymentStatus: paymentStatus.value === "all" ? undefined : paymentStatus.value,
  cooperationType:
    cooperationType.value === "all" ? undefined : cooperationType.value,
  auditStatus: auditStatus.value === "all" ? undefined : auditStatus.value,
  expiryHandlingStatus:
    expiryHandlingStatus.value === "all"
      ? undefined
      : expiryHandlingStatus.value,
})

const fetchContracts = async () => {
  loading.value = true
  error.value = ""
  try {
    const result = await listContracts(buildListParams())
    contracts.value = result.items
    totalCount.value = result.total
  } catch (err) {
    contracts.value = []
    totalCount.value = 0
    error.value = getRequestErrorMessage(err, "加载合同失败")
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  dialogMode.value = "create"
  editingContract.value = null
  dialogReadonly.value = false
  dialogOpen.value = true
}

const openEdit = (contract: Contract) => {
  dialogMode.value = "edit"
  editingContract.value = contract
  dialogReadonly.value = !canEditContract(contract)
  dialogOpen.value = true
}

const openAudit = (contract: Contract) => {
  if (!canAuditContract.value || !isPendingAudit(contract.auditStatus)) return
  auditingContract.value = contract
  auditDialogOpen.value = true
}

const openImagePreview = (src?: string | null, title = "图片预览") => {
  const url = String(src || "").trim()
  if (!url) return
  imagePreviewSrc.value = url
  imagePreviewTitle.value = title
  imagePreviewOpen.value = true
}

const handleAuditSubmit = async (payload: AuditContractRequest) => {
  if (!auditingContract.value) return
  auditSubmitting.value = true
  try {
    const actionText =
      payload.auditStatus === "failed" ? "审核驳回" : "审核通过"
    await auditContract(auditingContract.value.id, payload)
    toast.success(`${actionText}成功`)
    auditDialogOpen.value = false
    auditingContract.value = null
    await fetchContracts()
  } catch (err) {
    const actionText =
      payload.auditStatus === "failed" ? "审核驳回" : "审核通过"
    toast.error(getRequestErrorMessage(err, `${actionText}失败`))
  } finally {
    auditSubmitting.value = false
  }
}

const handleSubmit = async (payload: ContractFormPayload) => {
  submitting.value = true
  try {
    if (dialogMode.value === "create") {
      await createContract(payload)
      toast.success("合同新增成功")
    } else if (editingContract.value) {
      await updateContract(editingContract.value.id, payload)
      toast.success("合同更新成功")
    }
    dialogOpen.value = false
    await fetchContracts()
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "保存失败"))
  } finally {
    submitting.value = false
  }
}

const handleSearch = () => {
  activeKeyword.value = keyword.value.trim()
  pageIndex.value = 0
  void fetchContracts()
}

const clearSearch = () => {
  keyword.value = ""
  activeKeyword.value = ""
  paymentStatus.value = "all"
  cooperationType.value = "all"
  auditStatus.value = "all"
  expiryHandlingStatus.value = "all"
  pageIndex.value = 0
  void fetchContracts()
}

const refresh = () => {
  void fetchContracts()
}

const handlePageChange = (nextPage: number) => {
  if (nextPage === pageIndex.value) return
  pageIndex.value = nextPage
  void fetchContracts()
}

const handlePageSizeChange = (nextPageSize: number) => {
  const changed = nextPageSize !== pageSize.value
  pageSize.value = nextPageSize
  pageIndex.value = 0
  if (changed) void fetchContracts()
}

const formatDate = (value?: string | null) => {
  if (!value) return "-"
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return "-"
  return date.toLocaleString("zh-CN", { hour12: false })
}

const formatCurrency = (value?: number) => {
  const amount = Number(value || 0)
  return new Intl.NumberFormat("zh-CN", {
    style: "currency",
    currency: "CNY",
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amount)
}

const resolveWebsiteUrl = (value?: string | null) => {
  const url = String(value || "").trim()
  if (!url) return ""
  if (/^(https?:)?\/\//i.test(url)) return url
  return `https://${url}`
}

const statusLabelMap: Record<string, string> = {
  pending: "未审核",
  success: "审核成功",
  failed: "审核失败",
  paid: "已回款",
  partial: "部分回款",
  domestic: "内贸",
  foreign: "外贸",
  renewed: "已续签",
  ended: "不再合作",
}

const renderStatus = (value?: string) => {
  if (!value) return "-"
  return statusLabelMap[value] ?? value
}

const getContractDisplayName = (item: Contract) => item.contractName || "-"

const getCustomerDisplayName = (item: Contract) =>
  item.customerName || (item.customerId ? `#${item.customerId}` : "-")

const getCooperationTypeBadgeClass = (value?: string) => {
  switch (value) {
    case "domestic":
      return "border-sky-200 bg-sky-50 text-sky-700"
    case "foreign":
      return "border-amber-200 bg-amber-50 text-amber-700"
    default:
      return "border-zinc-200 bg-zinc-50 text-zinc-700"
  }
}

onMounted(() => {
  void fetchContracts()
})
onActivated(() => {
  void fetchContracts()
})
</script>

<template>
  <div class="w-full flex flex-col gap-4 lg:gap-6">
    <Card class="overflow-hidden border-border/60 shadow-sm">
      <CardHeader class="space-y-5 border-b bg-gradient-to-br from-background via-muted/10 to-background">
        <div class="flex flex-wrap items-center gap-3 bg-background/90">
          <div class="relative min-w-[220px] flex-1">
            <Search
              class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input v-model="keyword" placeholder="合同编号 / 合同名称 / 客户" class="h-10 pl-9" @keyup.enter="handleSearch" />
          </div>

          <Select v-model="cooperationType">
            <SelectTrigger class="h-10 w-[150px]">
              <SelectValue placeholder="合作类型" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="all">全部类型</SelectItem>
                <SelectItem value="domestic">内贸</SelectItem>
                <SelectItem value="foreign">外贸</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>

          <Select v-model="auditStatus">
            <SelectTrigger class="h-10 w-[150px]">
              <SelectValue placeholder="审核状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="all">全部审核</SelectItem>
                <SelectItem value="pending">未审核</SelectItem>
                <SelectItem value="success">审核成功</SelectItem>
                <SelectItem value="failed">审核失败</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>

          <div class="ml-auto flex items-center gap-2">
            <Button size="sm" @click="handleSearch">
              <Search class="h-4 w-4" />
              <span>搜索</span>
            </Button>
            <Button size="sm" variant="outline" @click="clearSearch">
              <RefreshCw class="h-4 w-4" />
              <span>重置</span>
            </Button>
          </div>
          <div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
            <div class="flex flex-wrap items-center gap-2">
              <Button size="sm" variant="outline" class="bg-background" @click="refresh">
                <RefreshCw class="h-4 w-4" />
                <span>刷新</span>
              </Button>
              <Button size="sm" @click="openCreate">
                <Plus class="h-4 w-4" />
                <span>新增合同</span>
              </Button>
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent class="pt-4">
        <div class="overflow-hidden rounded-lg border border-border/60 bg-background">
          <div v-if="loading" class="flex items-center justify-center py-24">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
          <div v-else-if="error" class="p-6 text-sm text-destructive">
            {{ error }}
          </div>
          <Table v-else class="min-w-[2100px]">
            <TableHeader class="sticky top-0 z-10 bg-muted/50">
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>合同编号</TableHead>
                <TableHead>合同名称</TableHead>
                <TableHead>客户</TableHead>
                <TableHead>负责销售</TableHead>
                <TableHead>负责运营</TableHead>
                <TableHead>合同图片</TableHead>
                <TableHead>回款图片</TableHead>
                <TableHead>合同金额</TableHead>
                <TableHead>合作类型</TableHead>
                <TableHead>合作年限</TableHead>
                <TableHead>合作节点</TableHead>
                <TableHead>网站名称</TableHead>
                <TableHead>网站地址</TableHead>
                <TableHead>网站账号</TableHead>
                <TableHead>是否上线</TableHead>
                <TableHead>开始时间</TableHead>
                <TableHead>结束时间</TableHead>

                <TableHead>备注</TableHead>
                <TableHead
                  class="sticky right-0 z-20 w-[80px] min-w-[80px] border-l border-border bg-muted/95 text-center">
                  操作
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="item in contracts" :key="item.id">
                <TableCell class="font-mono text-xs text-muted-foreground">
                  {{ item.id }}
                </TableCell>

                <TableCell class="font-mono text-sm">
                  {{ item.contractNumber || "-" }}
                </TableCell>

                <TableCell class="w-[160px] min-w-[160px] max-w-[160px]">
                  <template v-if="getContractDisplayName(item) !== '-'">
                    <TooltipProvider :delayDuration="200">
                      <Tooltip>
                        <TooltipTrigger as-child>
                          <div class="w-full cursor-help truncate text-left">
                            {{ getContractDisplayName(item) }}
                          </div>
                        </TooltipTrigger>
                        <TooltipContent class="max-w-sm whitespace-pre-wrap break-words text-left">
                          <p>{{ getContractDisplayName(item) }}</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </template>
                  <span v-else>-</span>
                </TableCell>

                <TableCell class="w-[160px] min-w-[160px] max-w-[160px]">
                  <template v-if="getCustomerDisplayName(item) !== '-'">
                    <TooltipProvider :delayDuration="200">
                      <Tooltip>
                        <TooltipTrigger as-child>
                          <div class="w-full cursor-help truncate text-left">
                            {{ getCustomerDisplayName(item) }}
                          </div>
                        </TooltipTrigger>
                        <TooltipContent class="max-w-sm whitespace-pre-wrap break-words text-left">
                          <p>{{ getCustomerDisplayName(item) }}</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </template>
                  <span v-else>-</span>
                </TableCell>


                <TableCell>
                  {{ item.userName || (item.userId ? `#${item.userId}` : "-") }}
                </TableCell>

                <TableCell>
                  {{
                    item.serviceUserName ||
                    (item.serviceUserId ? `#${item.serviceUserId}` : "-")
                  }}
                </TableCell>

                <TableCell class="text-center">
                  <button v-if="item.contractImage" type="button"
                    class="group inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-lg border border-border bg-muted/20 align-middle"
                    @click="openImagePreview(item.contractImage, `${item.contractName || '合同'} - 合同图片`)">
                    <img :src="item.contractImage" alt="合同图片"
                      class="h-full w-full object-cover transition-transform duration-200 group-hover:scale-105" />
                  </button>
                  <span v-else>-</span>
                </TableCell>

                <TableCell class="text-center">
                  <button v-if="item.paymentImage" type="button"
                    class="group inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-lg border border-border bg-muted/20 align-middle"
                    @click="openImagePreview(item.paymentImage, `${item.contractName || '合同'} - 回款图片`)">
                    <img :src="item.paymentImage" alt="回款图片"
                      class="h-full w-full object-cover transition-transform duration-200 group-hover:scale-105" />
                  </button>
                  <span v-else>-</span>
                </TableCell>

                <TableCell class="font-medium text-red-700">
                  {{ formatCurrency(item.contractAmount) }}
                </TableCell>


                <TableCell>
                  <Badge variant="outline" :class="getCooperationTypeBadgeClass(item.cooperationType)">
                    {{ renderStatus(item.cooperationType) }}
                  </Badge>
                </TableCell>

                <TableCell>
                  {{ item.cooperationYears || 0 }} 年
                </TableCell>

                <TableCell>
                  {{ item.nodeCount || 0 }} 个
                </TableCell>

                <TableCell class="max-w-[180px] whitespace-normal break-words">
                  {{ item.websiteName || "-" }}
                </TableCell>

                <TableCell class="max-w-[240px] whitespace-normal break-all">
                  <a v-if="item.websiteUrl" :href="resolveWebsiteUrl(item.websiteUrl)" target="_blank"
                    rel="noopener noreferrer" class="text-sky-700 underline-offset-4 hover:underline"
                    :title="item.websiteUrl">
                    {{ item.websiteUrl }}
                  </a>
                  <span v-else>-</span>
                </TableCell>

                <TableCell class="max-w-[180px] whitespace-normal break-all">
                  {{ item.websiteUsername || "-" }}
                </TableCell>

                <TableCell>
                  <Badge variant="outline" :class="item.isOnline
                    ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                    : 'border-zinc-200 bg-zinc-50 text-zinc-700'
                    ">
                    {{ item.isOnline ? "已上线" : "未上线" }}
                  </Badge>
                </TableCell>

                <TableCell>
                  {{ formatDate(item.startDate) }}
                </TableCell>

                <TableCell>
                  {{ formatDate(item.endDate) }}
                </TableCell>



                <TableCell class="max-w-[280px] text-sm text-muted-foreground">
                  <template v-if="item.remark">
                    <TooltipProvider :delayDuration="200">
                      <Tooltip>
                        <TooltipTrigger as-child>
                          <div class="line-clamp-3 cursor-help whitespace-normal break-words text-left leading-6">
                            {{ item.remark }}
                          </div>
                        </TooltipTrigger>
                        <TooltipContent class="max-w-sm whitespace-pre-wrap break-words text-left">
                          <p>{{ item.remark }}</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </template>
                  <span v-else>-</span>
                </TableCell>

                <TableCell
                  class="sticky right-0 z-10 w-[80px] min-w-[80px] border-l border-border bg-background text-center">
                  <div class="flex items-center justify-center gap-2 whitespace-nowrap">
                    <Button v-if="canAuditContract && isPendingAudit(item.auditStatus)" variant="ghost" size="sm"
                      class="h-8 shrink-0 gap-1 text-emerald-700" :disabled="auditSubmitting || submitting"
                      @click="openAudit(item)">
                      <ShieldCheck class="h-4 w-4" />
                      <span>去审核</span>
                    </Button>
                    <Button variant="ghost" size="sm" class="h-8 shrink-0 gap-1" @click="openEdit(item)">
                      <SquarePen class="h-4 w-4" />
                      <span>编辑</span>
                    </Button>
                  </div>
                </TableCell>
              </TableRow>

              <EmptyTablePlaceholder v-if="contracts.length === 0" :colspan="20" text="暂无合同数据" />
            </TableBody>
          </Table>
        </div>

        <div class="mt-4">
          <Pagination :current-page="pageIndex" :total-pages="totalPages" :page-size="pageSize" :selected-count="0"
            :total-count="totalCount" @update:current-page="handlePageChange"
            @update:page-size="handlePageSizeChange" />
        </div>
      </CardContent>
    </Card>

    <PopupForm v-model:open="dialogOpen" :mode="dialogMode" :contract="editingContract" :readonly="dialogReadonly"
      :submitting="submitting" @submit="handleSubmit" />
    <PopupForm v-model:open="auditDialogOpen" mode="edit" :contract="auditingContract" :submitting="auditSubmitting"
      :audit-mode="true" @audit="handleAuditSubmit" />
    <Dialog v-model:open="imagePreviewOpen">
      <DialogContent class="sm:max-w-[920px]">
        <DialogHeader>
          <DialogTitle>{{ imagePreviewTitle }}</DialogTitle>
        </DialogHeader>
        <div class="flex items-center justify-center rounded-lg bg-muted/20 p-2">
          <img v-if="imagePreviewSrc" :src="imagePreviewSrc" :alt="imagePreviewTitle"
            class="max-h-[75vh] w-auto max-w-full rounded-md object-contain" />
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
