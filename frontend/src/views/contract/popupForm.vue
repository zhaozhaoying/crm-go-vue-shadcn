<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { CalendarRange, FileText, Globe2, Image, Loader2, ShieldCheck, Wallet } from "lucide-vue-next"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
  InputGroupText,
} from "@/components/ui/input-group"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import ImageUploadCard from "@/components/custom/ImageUploadCard.vue"
import { useAuthStore } from "@/stores/auth"
import { hasAnyRole, isAdminUser, normalizeRole } from "@/lib/auth-role"
import { getSystemSettings } from "@/api/modules/systemSettings"
import { listCustomersPage, listMyCustomers } from "@/api/modules/customers"
import { uploadContractImage } from "@/api/modules/contracts"
import { listUsers } from "@/api/modules/users"
import type { AuditContractRequest, Contract, ContractFormPayload } from "@/types/contract"
import type { Customer } from "@/types/customer"
import type { UserWithRole } from "@/types/user"

interface Props {
  open: boolean
  mode: "create" | "edit"
  contract?: Contract | null
  submitting?: boolean
  readonly?: boolean
  fixedCustomerId?: number | null
  auditMode?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  contract: null,
  submitting: false,
  readonly: false,
  fixedCustomerId: null,
  auditMode: false,
})

const emit = defineEmits<{
  (e: "update:open", value: boolean): void
  (e: "submit", payload: ContractFormPayload): void
  (e: "audit", payload: AuditContractRequest): void
}>()

interface FormState {
  contractImage: string
  paymentImage: string
  paymentStatus: string
  remark: string
  customerId: string
  cooperationType: string
  contractNumberSuffix: string
  contractName: string
  contractAmount: string
  paymentAmount: string
  cooperationYears: string
  nodeCount: string
  serviceUserId: string
  websiteName: string
  websiteUrl: string
  websiteUsername: string
  isOnline: boolean
  startDate: string
  endDate: string
  auditStatus: string
  auditComment: string
  expiryHandlingStatus: string
}

const authStore = useAuthStore()
const isAdmin = computed(() => isAdminUser(authStore.user))
const canViewAllCustomers = computed(
  () =>
    isAdmin.value ||
    hasAnyRole(authStore.user, ["finance_manager", "finance", "财务经理", "财务"]),
)
const SALES_ROLE_NAMES = [
  "sales_director",
  "sales_manager",
  "sales_staff",
  "销售总监",
  "销售经理",
  "销售员工",
  "销售",
]
const OPERATION_ROLE_CANDIDATES = [
  "ops_manager",
  "operation_manager",
  "ops_staff",
  "operation_staff",
  "运营经理",
  "运营员工",
  "运营",
]
const contractNumberPrefix = ref("zzy_")
const formReadonly = computed(() => props.readonly)
const isAuditMode = computed(() => props.auditMode && !formReadonly.value)
const isSalesOrderMode = computed(
  () => Number(props.fixedCustomerId || 0) > 0,
)
const isEditMode = computed(() => props.mode === "edit")
const isSalesEditRestricted = computed(
  () =>
    isEditMode.value &&
    !formReadonly.value &&
    !isAuditMode.value &&
    hasAnyRole(authStore.user, SALES_ROLE_NAMES),
)
const isOperationEditRestricted = computed(
  () =>
    isEditMode.value &&
    !formReadonly.value &&
    !isAuditMode.value &&
    hasAnyRole(authStore.user, OPERATION_ROLE_CANDIDATES),
)
const showBaseInfoSection = computed(() => !isOperationEditRestricted.value)
const showBusinessSection = computed(() => !isOperationEditRestricted.value)
const showAttachmentSection = computed(() => !isOperationEditRestricted.value)
const showSiteServiceSection = computed(
  () => !isSalesOrderMode.value && !isSalesEditRestricted.value,
)
const baseFieldsReadonly = computed(
  () => formReadonly.value || isOperationEditRestricted.value,
)
const businessFieldsReadonly = computed(
  () => formReadonly.value || isOperationEditRestricted.value,
)
const siteServiceReadonly = computed(
  () => formReadonly.value || isSalesEditRestricted.value,
)
const attachmentFieldsReadonly = computed(
  () => formReadonly.value || isOperationEditRestricted.value,
)
const canEditContractNumber = computed(
  () => !isAuditMode.value || isAdmin.value,
)
const SERVICE_USER_NONE = "none"
const OPERATION_ROLE_NAMES = new Set(OPERATION_ROLE_CANDIDATES.map((role) => normalizeRole(role)))

const createEmptyForm = (): FormState => ({
  contractImage: "",
  paymentImage: "",
  paymentStatus: "pending",
  remark: "",
  customerId: "",
  cooperationType: "domestic",
  contractNumberSuffix: "",
  contractName: "",
  contractAmount: "0",
  paymentAmount: "0",
  cooperationYears: "0",
  nodeCount: "0",
  serviceUserId: SERVICE_USER_NONE,
  websiteName: "",
  websiteUrl: "",
  websiteUsername: "",
  isOnline: false,
  startDate: "",
  endDate: "",
  auditStatus: "pending",
  auditComment: "",
  expiryHandlingStatus: "pending",
})

const form = ref<FormState>(createEmptyForm())
const formError = ref("")
const customerOptions = ref<Customer[]>([])
const customerLoading = ref(false)
const allUserOptions = ref<UserWithRole[]>([])
const serviceUserLoading = ref(false)

const dialogTitle = computed(() => {
  if (isAuditMode.value) return "审核合同"
  if (formReadonly.value) return "查看合同"
  return props.mode === "create" ? "新增合同" : "编辑合同"
})
const submitText = computed(() => (props.mode === "create" ? "新增" : "保存"))
const dialogDescription = computed(() => {
  if (isAuditMode.value) {
    return "审核时会一并保存本页调整内容，并写入审核备注与审核人信息。"
  }
  if (formReadonly.value) {
    return "查看合同详情、商务信息与履约状态。"
  }
  if (isOperationEditRestricted.value) {
    return "运营编辑时仅可维护站点与服务区域，上线时间会在保存时自动生成。"
  }
  if (isSalesEditRestricted.value) {
    return "销售编辑时不展示站点与服务区域，其余合同信息可继续维护。"
  }
  return "按业务信息分区填写，减少跨区来回滚动和遗漏。"
})

const statusLabelMap: Record<string, string> = {
  pending: "待审核",
  success: "审核通过",
  failed: "审核驳回",
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

const auditStatusTone = computed(() => {
  switch (form.value.auditStatus) {
    case "success":
      return "border-emerald-200 bg-emerald-50 text-emerald-700"
    case "failed":
      return "border-red-200 bg-red-50 text-red-700"
    default:
      return "border-amber-200 bg-amber-50 text-amber-700"
  }
})

const paymentStatusTone = computed(() => {
  switch (form.value.paymentStatus) {
    case "paid":
      return "border-emerald-200 bg-emerald-50 text-emerald-700"
    case "partial":
      return "border-sky-200 bg-sky-50 text-sky-700"
    default:
      return "border-zinc-200 bg-zinc-50 text-zinc-700"
  }
})

const currentCustomerLabel = computed(() => {
  if (props.contract?.customerName) return props.contract.customerName
  const currentId = Number(form.value.customerId || 0)
  if (currentId > 0) {
    const customer = customerOptions.value.find((item) => item.id === currentId)
    if (customer?.name) return customer.name
    return `客户 #${currentId}`
  }
  return "未选择客户"
})

const isOperationUser = (user: UserWithRole) =>
  [user.roleName, user.roleLabel].some((role) =>
    OPERATION_ROLE_NAMES.has(normalizeRole(role)),
  )

const serviceUserOptions = computed(() => {
  const currentId = Number(form.value.serviceUserId || props.contract?.serviceUserId || 0)
  const options = allUserOptions.value.filter((user) => {
    const enabled = String(user.status || "").trim().toLowerCase() !== "disabled"
    return isOperationUser(user) && (enabled || user.id === currentId)
  })

  if (
    currentId > 0 &&
    !options.some((user) => user.id === currentId)
  ) {
    options.unshift({
      id: currentId,
      username: "",
      nickname: props.contract?.serviceUserName || "",
      email: "",
      mobile: "",
      avatar: "",
      roleId: 0,
      parentId: null,
      status: "enabled",
      createdAt: "",
      updatedAt: "",
      roleName: "",
      roleLabel: "运营组",
    })
  }

  return options
})

const auditHistoryText = computed(() => {
  const parts: string[] = []
  if (props.contract?.auditedByName) parts.push(`审核人 ${props.contract.auditedByName}`)
  if (props.contract?.auditedAt) {
    const date = new Date(props.contract.auditedAt)
    if (!Number.isNaN(date.getTime())) {
      parts.push(date.toLocaleString("zh-CN", { hour12: false }))
    }
  }
  return parts.join(" · ")
})

const toDatetimeLocal = (value?: string) => {
  if (!value) return ""
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ""
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, "0")
  const d = String(date.getDate()).padStart(2, "0")
  const hh = String(date.getHours()).padStart(2, "0")
  const mm = String(date.getMinutes()).padStart(2, "0")
  return `${y}-${m}-${d}T${hh}:${mm}`
}

const formatDisplayDatetime = (value?: string) => {
  if (!value) return "-"
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return "-"
  return date.toLocaleString("zh-CN", { hour12: false })
}

const startDateDisplayText = computed(() => {
  if (form.value.startDate) return formatDisplayDatetime(form.value.startDate)
  if (form.value.isOnline) return "保存后按后端提交时间生成"
  return "开启上线后自动生成"
})

const endDateDisplayText = computed(() => {
  if (form.value.endDate) return formatDisplayDatetime(form.value.endDate)
  if (form.value.isOnline) return "保存后按后端时间 + 合作年限生成"
  return "开启上线后自动生成"
})

const showExpiryHandlingStatus = computed(() => {
  if (isSalesOrderMode.value) return false
  if (!form.value.endDate) return false
  const endDate = new Date(form.value.endDate)
  if (Number.isNaN(endDate.getTime())) return false
  return endDate.getTime() <= Date.now()
})

const extractSuffix = (prefix: string, fullNumber?: string) => {
  const normalizedPrefix = prefix.trim()
  const value = (fullNumber ?? "").trim()
  if (!normalizedPrefix) return value
  if (value.startsWith(normalizedPrefix)) {
    return value.slice(normalizedPrefix.length).trim()
  }
  return value
}

const loadContractNumberPrefix = async () => {
  try {
    const settings = await getSystemSettings()
    const prefix = settings.contractNumberPrefix?.trim()
    contractNumberPrefix.value = prefix || "zzy_"
  } catch {
    contractNumberPrefix.value = "zzy_"
  }
}

watch(
  () => [props.open, props.mode, props.contract, contractNumberPrefix.value],
  ([open]) => {
    if (!open) return
    formError.value = ""
    const resolvedCustomerId = props.contract?.customerId
      ? String(props.contract.customerId)
      : props.fixedCustomerId
        ? String(props.fixedCustomerId)
        : ""
    form.value = {
      contractImage: props.contract?.contractImage ?? "",
      paymentImage: props.contract?.paymentImage ?? "",
      paymentStatus: props.contract?.paymentStatus ?? "pending",
      remark: props.contract?.remark ?? "",
      customerId: resolvedCustomerId,
      cooperationType: props.contract?.cooperationType ?? "domestic",
      contractNumberSuffix: extractSuffix(contractNumberPrefix.value, props.contract?.contractNumber),
      contractName: props.contract?.contractName ?? "",
      contractAmount: String(props.contract?.contractAmount ?? 0),
      paymentAmount: String(props.contract?.paymentAmount ?? 0),
      cooperationYears: String(props.contract?.cooperationYears ?? 0),
      nodeCount: String(props.contract?.nodeCount ?? 0),
      serviceUserId: props.contract?.serviceUserId ? String(props.contract.serviceUserId) : SERVICE_USER_NONE,
      websiteName: props.contract?.websiteName ?? "",
      websiteUrl: props.contract?.websiteUrl ?? "",
      websiteUsername: props.contract?.websiteUsername ?? "",
      isOnline: Boolean(props.contract?.isOnline),
      startDate: toDatetimeLocal(props.contract?.startDate),
      endDate: toDatetimeLocal(props.contract?.endDate),
      auditStatus: props.contract?.auditStatus ?? "pending",
      auditComment: props.contract?.auditComment ?? "",
      expiryHandlingStatus: props.contract?.expiryHandlingStatus ?? "pending",
    }
  },
  { immediate: true },
)

watch(
  () => [props.open, props.fixedCustomerId, props.mode, props.contract],
  ([open]) => {
    if (!open) return
    if (props.contract?.customerId) return
    if (props.mode === "create" && props.fixedCustomerId) {
      form.value.customerId = String(props.fixedCustomerId)
    }
  },
)

const loadCustomerOptions = async () => {
  customerLoading.value = true
  try {
    const result = canViewAllCustomers.value
      ? await listCustomersPage({
        page: 1,
        pageSize: 500,
        ownershipScope: "all",
        excludePool: "1",
      })
      : await listMyCustomers({
        page: 1,
        pageSize: 500,
        ownershipScope: "mine",
        excludePool: "1",
      })
    customerOptions.value = result.items
  } catch {
    customerOptions.value = []
    formError.value = "加载客户列表失败"
  } finally {
    customerLoading.value = false
  }
}

const loadServiceUserOptions = async () => {
  serviceUserLoading.value = true
  try {
    allUserOptions.value = (await listUsers()) || []
  } catch {
    allUserOptions.value = []
    formError.value = "加载运营组用户失败"
  } finally {
    serviceUserLoading.value = false
  }
}

watch(
  () => [props.open, canViewAllCustomers.value, isSalesOrderMode.value],
  ([open]) => {
    if (!open) return
    if (isSalesOrderMode.value) return
    loadCustomerOptions()
  },
  { immediate: true },
)

watch(
  () => [props.open, isSalesOrderMode.value],
  ([open]) => {
    if (!open) return
    if (isSalesOrderMode.value) return
    loadServiceUserOptions()
  },
  { immediate: true },
)

watch(
  () => props.open,
  (open) => {
    if (!open) return
    loadContractNumberPrefix()
  },
  { immediate: true },
)

watch(
  () => form.value.contractNumberSuffix,
  (value) => {
    const digitsOnly = value.replace(/\D+/g, "")
    if (digitsOnly !== value) {
      form.value.contractNumberSuffix = digitsOnly
    }
  },
)

const close = () => {
  if (props.submitting) return
  emit("update:open", false)
}

const parseNumber = (raw: string, fallback = 0) => {
  const value = Number(raw)
  if (!Number.isFinite(value)) return fallback
  return value
}

const parseUnix = (raw: string): number | null => {
  if (!raw) return null
  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) return null
  return Math.floor(date.getTime() / 1000)
}

const uploadImage = async (file: File) => {
  return uploadContractImage(file)
}

const buildPayload = (): ContractFormPayload => {
  const startUnix = parseUnix(form.value.startDate)
  const endUnix = parseUnix(form.value.endDate)
  const suffix = form.value.contractNumberSuffix.trim()
  const prefix = contractNumberPrefix.value.trim() || "zzy_"

  return {
    contractImage: form.value.contractImage.trim(),
    paymentImage: form.value.paymentImage.trim(),
    paymentStatus: form.value.paymentStatus,
    remark: form.value.remark.trim(),
    customerId: Number(form.value.customerId),
    cooperationType: form.value.cooperationType,
    contractNumber: `${prefix}${suffix}`,
    contractNumberSuffix: suffix,
    contractName: form.value.contractName.trim(),
    contractAmount: parseNumber(form.value.contractAmount, 0),
    paymentAmount: parseNumber(form.value.paymentAmount, 0),
    cooperationYears: parseNumber(form.value.cooperationYears, 0),
    nodeCount: parseNumber(form.value.nodeCount, 0),
    serviceUserId:
      form.value.serviceUserId && form.value.serviceUserId !== SERVICE_USER_NONE
        ? parseNumber(form.value.serviceUserId, 0)
        : null,
    websiteName: form.value.websiteName.trim(),
    websiteUrl: form.value.websiteUrl.trim(),
    websiteUsername: form.value.websiteUsername.trim(),
    isOnline: form.value.isOnline,
    startDate: startUnix,
    endDate: endUnix,
    auditStatus: form.value.auditStatus,
    expiryHandlingStatus: form.value.expiryHandlingStatus,
  }
}

const submit = async () => {
  if (formReadonly.value) {
    close()
    return
  }
  formError.value = ""
  if (!form.value.contractNumberSuffix.trim()) {
    formError.value = "合同编号后缀不能为空"
    return
  }
  if (!form.value.contractName.trim()) {
    formError.value = "合同名称不能为空"
    return
  }
  if (
    isSalesOrderMode.value &&
    props.fixedCustomerId &&
    Number(form.value.customerId || 0) <= 0
  ) {
    form.value.customerId = String(props.fixedCustomerId)
  }
  if (!form.value.customerId || Number(form.value.customerId) <= 0) {
    formError.value = "请选择客户"
    return
  }

  try {
    emit("submit", buildPayload())
  } catch (error) {
    formError.value = error instanceof Error ? error.message : "图片上传失败"
  }
}

const submitAudit = (nextStatus: "success" | "failed") => {
  if (!isAuditMode.value) return
  formError.value = ""
  if (!form.value.contractNumberSuffix.trim()) {
    formError.value = "合同编号后缀不能为空"
    return
  }
  if (!form.value.contractName.trim()) {
    formError.value = "合同名称不能为空"
    return
  }
  if (!form.value.customerId || Number(form.value.customerId) <= 0) {
    formError.value = "请选择客户"
    return
  }
  try {
    form.value.auditStatus = nextStatus
    emit("audit", {
      ...buildPayload(),
      auditStatus: nextStatus,
      auditComment: form.value.auditComment.trim(),
    })
  } catch (error) {
    formError.value = error instanceof Error ? error.message : "合同审核失败"
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="(v) => emit('update:open', v)">
    <DialogContent class="flex max-h-[88vh] flex-col overflow-hidden border-border/70 bg-gradient-to-b from-background to-muted/10 p-0 sm:max-w-[980px]">
      <DialogHeader class="border-b bg-muted/20 px-6 pb-5 pt-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-2">
            <DialogTitle class="text-xl font-semibold tracking-tight">
              {{ dialogTitle }}
            </DialogTitle>
            <p class="max-w-2xl text-sm leading-6 text-muted-foreground">
              {{ dialogDescription }}
            </p>
          </div>

          <div class="grid gap-2 sm:grid-cols-2 lg:min-w-[300px]">
            <div class="rounded-xl border bg-background px-3 py-2 shadow-sm">
              <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-muted-foreground">
                客户
              </p>
              <p class="mt-1 text-sm font-medium text-foreground">
                {{ currentCustomerLabel }}
              </p>
            </div>
            <div class="rounded-xl border bg-background px-3 py-2 shadow-sm">
              <p class="text-[11px] font-medium uppercase tracking-[0.18em] text-muted-foreground">
                编号前缀
              </p>
              <p class="mt-1 font-mono text-sm text-foreground">
                {{ contractNumberPrefix }}
              </p>
            </div>
          </div>
        </div>
      </DialogHeader>

      <form class="flex min-h-0 flex-1 flex-col" @submit.prevent="submit">
        <div class="min-h-0 flex-1 overflow-y-auto px-6 py-5">
          <div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_300px]">
            <div class="space-y-5">
              <section
                v-if="showBaseInfoSection"
                class="rounded-2xl border bg-background/95 p-5 shadow-sm"
              >
                <div class="mb-4 flex items-center gap-2">
                  <div class="flex h-9 w-9 items-center justify-center rounded-xl border bg-muted/40 text-foreground">
                    <FileText class="h-4 w-4" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-foreground">基础信息</h3>
                    <p class="text-xs text-muted-foreground">合同身份、客户归属和基础合作关系。</p>
                  </div>
                </div>

                <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <div class="space-y-2 md:col-span-2">
                    <Label>合同编号</Label>
                    <InputGroup>
                      <InputGroupAddon class="px-2">
                        <InputGroupText class="text-xs">
                          {{ contractNumberPrefix }}
                        </InputGroupText>
                      </InputGroupAddon>
                      <InputGroupInput
                        v-model="form.contractNumberSuffix"
                        type="number"
                        inputmode="numeric"
                        pattern="[0-9]*"
                        placeholder="请输入合同编号后缀"
                        :disabled="baseFieldsReadonly || !canEditContractNumber"
                      />
                    </InputGroup>
                    <p
                      v-if="isAuditMode && !canEditContractNumber"
                      class="text-xs text-muted-foreground"
                    >
                      审核时仅管理员可修改合同编号。
                    </p>
                  </div>

                  <div class="space-y-1.5">
                    <Label>合同名称</Label>
                    <Input v-model="form.contractName" placeholder="请输入合同名称" :disabled="baseFieldsReadonly" />
                  </div>

                  <div class="space-y-1.5">
                    <Label>合作类型</Label>
                    <Select v-model="form.cooperationType" :disabled="baseFieldsReadonly">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="domestic">内贸</SelectItem>
                          <SelectItem value="foreign">外贸</SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </div>

                  <div v-if="!isSalesOrderMode" class="space-y-1.5">
                    <Label>客户</Label>
                    <Select v-model="form.customerId" :disabled="customerLoading || baseFieldsReadonly || !!props.fixedCustomerId">
                      <SelectTrigger>
                        <SelectValue :placeholder="customerLoading ? '加载客户中...' : '请选择客户'" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem v-for="customer in customerOptions" :key="customer.id" :value="String(customer.id)">
                            {{ customer.name || `#${customer.id}` }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </div>

                  <div v-if="!isSalesOrderMode" class="space-y-1.5">
                    <Label>客服对接人</Label>
                    <Select
                      v-model="form.serviceUserId"
                      :disabled="serviceUserLoading || baseFieldsReadonly"
                    >
                      <SelectTrigger>
                        <SelectValue :placeholder="serviceUserLoading ? '加载运营组用户中...' : '请选择运营组用户'" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem :value="SERVICE_USER_NONE">未分配</SelectItem>
                          <SelectItem
                            v-for="user in serviceUserOptions"
                            :key="user.id"
                            :value="String(user.id)"
                          >
                            {{ user.nickname || user.username || `用户 #${user.id}` }}
                            {{ user.roleLabel ? `（${user.roleLabel}）` : "" }}
                          </SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              </section>

              <section
                v-if="showBusinessSection"
                class="rounded-2xl border bg-background/95 p-5 shadow-sm"
              >
                <div class="mb-4 flex items-center gap-2">
                  <div class="flex h-9 w-9 items-center justify-center rounded-xl border bg-muted/40 text-foreground">
                    <Wallet class="h-4 w-4" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-foreground">商务与履约</h3>
                    <p class="text-xs text-muted-foreground">金额、回款、周期和合同生命周期状态。</p>
                  </div>
                </div>

                <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
                  <div class="space-y-1.5">
                    <Label>合同金额</Label>
                    <InputGroup>
                      <InputGroupInput
                        v-model="form.contractAmount"
                        type="number"
                        min="0"
                        step="0.01"
                        :disabled="businessFieldsReadonly"
                      />
                      <InputGroupAddon class="px-2">
                        <InputGroupText class="text-xs">元</InputGroupText>
                      </InputGroupAddon>
                    </InputGroup>
                  </div>

                  <div class="space-y-1.5">
                    <Label>回款金额</Label>
                    <InputGroup>
                      <InputGroupInput
                        v-model="form.paymentAmount"
                        type="number"
                        min="0"
                        step="0.01"
                        :disabled="businessFieldsReadonly"
                      />
                      <InputGroupAddon class="px-2">
                        <InputGroupText class="text-xs">元</InputGroupText>
                      </InputGroupAddon>
                    </InputGroup>
                  </div>

                  <div v-if="!isSalesOrderMode" class="space-y-1.5">
                    <Label>回款状态</Label>
                    <Select v-model="form.paymentStatus" :disabled="businessFieldsReadonly">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="pending">未回款</SelectItem>
                          <SelectItem value="partial">部分回款</SelectItem>
                          <SelectItem value="paid">已回款</SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </div>

                  <div class="space-y-1.5">
                    <Label>合作年限</Label>
                    <InputGroup>
                      <InputGroupInput
                        v-model="form.cooperationYears"
                        type="number"
                        min="0"
                        :disabled="businessFieldsReadonly"
                      />
                      <InputGroupAddon class="px-2">
                        <InputGroupText class="text-xs">年</InputGroupText>
                      </InputGroupAddon>
                    </InputGroup>
                  </div>

                  <div class="space-y-1.5">
                    <Label>节点个数</Label>
                    <InputGroup>
                      <InputGroupInput
                        v-model="form.nodeCount"
                        type="number"
                        min="0"
                        :disabled="businessFieldsReadonly"
                      />
                      <InputGroupAddon class="px-2">
                        <InputGroupText class="text-xs">个</InputGroupText>
                      </InputGroupAddon>
                    </InputGroup>
                  </div>

                  <div v-if="showExpiryHandlingStatus" class="space-y-1.5">
                    <Label>过期处理状态</Label>
                    <Select v-model="form.expiryHandlingStatus" :disabled="businessFieldsReadonly">
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectGroup>
                          <SelectItem value="pending">未处理</SelectItem>
                          <SelectItem value="renewed">已续签</SelectItem>
                          <SelectItem value="ended">不再合作</SelectItem>
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              </section>

              <section
                v-if="showSiteServiceSection"
                class="rounded-2xl border bg-background/95 p-5 shadow-sm"
              >
                <div class="mb-4 flex items-center gap-2">
                  <div class="flex h-9 w-9 items-center justify-center rounded-xl border bg-muted/40 text-foreground">
                    <Globe2 class="h-4 w-4" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-foreground">站点与服务</h3>
                    <p class="text-xs text-muted-foreground">站点投放信息、上线状态与履约时间。</p>
                  </div>
                </div>

                <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <div class="space-y-1.5">
                    <Label>网站名称</Label>
                    <Input v-model="form.websiteName" placeholder="例如 官网、专题页" :disabled="siteServiceReadonly" />
                  </div>
                  <div class="space-y-1.5">
                    <Label>网站地址</Label>
                    <Input v-model="form.websiteUrl" placeholder="https://..." :disabled="siteServiceReadonly" />
                  </div>
                  <div class="space-y-1.5">
                    <Label>网站账号</Label>
                    <Input v-model="form.websiteUsername" placeholder="录入账号或标识" :disabled="siteServiceReadonly" />
                  </div>
                  <div class="flex items-center justify-between rounded-xl border bg-muted/20 px-4 py-3">
                    <div class="space-y-1">
                      <p class="text-sm font-medium text-foreground">是否上线</p>
                      <p class="text-xs text-muted-foreground">开启后，开始时间按保存提交时间写入，结束时间自动按合作年限推算。</p>
                    </div>
                    <Switch id="is-online" v-model:checked="form.isOnline" :disabled="siteServiceReadonly" />
                  </div>
                </div>
              </section>

              <section
                v-if="showAttachmentSection"
                class="rounded-2xl border bg-background/95 p-5 shadow-sm"
              >
                <div class="mb-4 flex items-center gap-2">
                  <div class="flex h-9 w-9 items-center justify-center rounded-xl border bg-muted/40 text-foreground">
                    <Image class="h-4 w-4" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-foreground">附件与备注</h3>
                    <p class="text-xs text-muted-foreground">保留合同原件、回款凭证和业务说明。</p>
                  </div>
                </div>

                <div class="grid gap-5 lg:grid-cols-[320px_minmax(0,1fr)]">
                  <div class="flex flex-wrap gap-4">
                    <div class="space-y-2">
                      <Label>合同图片</Label>
                      <ImageUploadCard
                        v-model="form.contractImage"
                        placeholder="暂无合同图片"
                        :disabled="attachmentFieldsReadonly"
                        :on-upload="uploadImage"
                        @error="(msg) => (formError = msg)"
                      />
                    </div>
                    <div class="space-y-2">
                      <Label>回款图片</Label>
                      <ImageUploadCard
                        v-model="form.paymentImage"
                        placeholder="暂无回款图片"
                        :disabled="attachmentFieldsReadonly"
                        :on-upload="uploadImage"
                        @error="(msg) => (formError = msg)"
                      />
                    </div>
                  </div>

                  <div class="space-y-1.5">
                    <Label>备注</Label>
                    <Textarea
                      v-model="form.remark"
                      :rows="8"
                      placeholder="补充合同背景、商务说明、注意事项等"
                      :disabled="attachmentFieldsReadonly"
                    />
                  </div>
                </div>
              </section>
            </div>

            <div class="space-y-5">
              <section class="rounded-2xl border bg-background/95 p-5 shadow-sm">
                <div class="mb-4 flex items-center gap-2">
                  <div class="flex h-9 w-9 items-center justify-center rounded-xl border bg-muted/40 text-foreground">
                    <ShieldCheck class="h-4 w-4" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-foreground">审核面板</h3>
                    <p class="text-xs text-muted-foreground">
                      {{ isAuditMode ? "审核时可调整内容后再执行通过或驳回。" : "展示审核轨迹和当前审核状态。" }}
                    </p>
                  </div>
                </div>

                <div class="space-y-3">
                  <div class="rounded-xl border p-3" :class="auditStatusTone">
                    <p class="text-[11px] uppercase tracking-[0.18em]">当前审核状态</p>
                    <p class="mt-1 text-sm font-medium">{{ renderStatus(form.auditStatus) }}</p>
                  </div>

                  <div v-if="auditHistoryText" class="rounded-xl border bg-muted/20 p-3">
                    <p class="text-[11px] uppercase tracking-[0.18em] text-muted-foreground">最近审核</p>
                    <p class="mt-1 text-sm text-foreground">{{ auditHistoryText }}</p>
                  </div>

                  <div
                    v-if="props.contract?.auditComment && !isAuditMode"
                    class="rounded-xl border bg-muted/20 p-3"
                  >
                    <p class="text-[11px] uppercase tracking-[0.18em] text-muted-foreground">审核备注</p>
                    <p class="mt-1 whitespace-pre-wrap text-sm leading-6 text-foreground">
                      {{ props.contract.auditComment }}
                    </p>
                  </div>

                  <div v-if="isAuditMode" class="space-y-1.5">
                    <Label>审核备注</Label>
                    <Textarea
                      v-model="form.auditComment"
                      :rows="6"
                      placeholder="填写审核说明、调整依据或驳回原因"
                      :disabled="formReadonly"
                    />
                    <p class="text-xs leading-5 text-muted-foreground">
                      审核将同步保存当前表单修改内容，并写入审核人和审核时间。
                    </p>
                  </div>
                </div>
              </section>
            </div>
          </div>

          <p v-if="formError" class="mt-4 text-sm text-destructive">{{ formError }}</p>
        </div>

        <DialogFooter class="border-t bg-background/90 px-6 py-4">
          <Button type="button" variant="outline" @click="close">{{ formReadonly ? "关闭" : "取消" }}</Button>
          <template v-if="isAuditMode">
            <Button
              type="button"
              variant="outline"
              class="border-destructive/40 text-destructive hover:bg-destructive/5"
              :disabled="submitting"
              @click="submitAudit('failed')"
            >
              <Loader2 v-if="submitting && form.auditStatus === 'failed'" class="mr-2 h-4 w-4 animate-spin" />
              驳回
            </Button>
            <Button
              type="button"
              :disabled="submitting"
              @click="submitAudit('success')"
            >
              <Loader2 v-if="submitting && form.auditStatus === 'success'" class="mr-2 h-4 w-4 animate-spin" />
              审核通过
            </Button>
          </template>
          <Button v-else-if="!formReadonly" type="submit" :disabled="submitting">
            <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />
            {{ submitText }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
