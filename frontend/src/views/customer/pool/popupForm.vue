<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue"
import { Loader2, Plus, Star, Trash2 } from "lucide-vue-next"

import { validateCustomerUnique } from "@/api/modules/customers"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { chinaPcaCode, type ChinaPcaNode } from "@/data/china-pca-code"
import type { Customer, CustomerFormPayload, CustomerFormPhone } from "@/types/customer"

interface Props {
  open: boolean
  mode: "create" | "edit"
  customer?: Customer | null
  submitting?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  customer: null,
  submitting: false,
})

const emit = defineEmits<{
  (e: "update:open", value: boolean): void
  (e: "submit", payload: CustomerFormPayload): void
}>()

interface FormState {
  name: string
  legalName: string
  contactName: string
  email: string
  weixin: string
  province: string
  city: string
  area: string
  detailAddress: string
  remark: string
  phones: CustomerFormPhone[]
}

const createEmptyPhone = (isPrimary = false): CustomerFormPhone => {
  return {
    phone: "",
    phoneLabel: "手机",
    isPrimary,
  }
}

const toCodeNumber = (value: string): number | undefined => {
  if (!value) return undefined
  const num = Number(value)
  return Number.isFinite(num) ? num : undefined
}

const formError = ref("")
const form = ref<FormState>({
  name: "",
  legalName: "",
  contactName: "",
  email: "",
  weixin: "",
  province: "",
  city: "",
  area: "",
  detailAddress: "",
  remark: "",
  phones: [createEmptyPhone(true)],
})

const uniqueNameError = ref("")
const uniqueLegalNameError = ref("")
const uniqueWeixinError = ref("")
const duplicatePhones = ref<string[]>([])
const checkingUnique = ref(false)
let uniqueCheckTimer: ReturnType<typeof setTimeout> | null = null
let uniqueCheckSeq = 0

const CN_MOBILE_PHONE_REGEX = /^1[3-9]\d{9}$/

const normalizePhoneInput = (value: string): string => {
  let digits = String(value ?? "").replace(/\D/g, "")
  if (digits.length === 13 && digits.startsWith("86")) {
    digits = digits.slice(2)
  }
  return digits.slice(0, 11)
}

const isValidCnMobile = (phone: string): boolean => {
  return CN_MOBILE_PHONE_REGEX.test(phone)
}

const duplicatePhoneSet = computed(() => new Set(duplicatePhones.value))
const localDuplicatePhones = computed(() => getLocalDuplicatePhones())
const localDuplicatePhoneSet = computed(() => new Set(localDuplicatePhones.value))
const phoneFieldErrors = computed(() => form.value.phones.map((item) => getPhoneFieldError(item.phone)))
const hasPhoneFieldError = computed(() => phoneFieldErrors.value.some((item) => Boolean(item)))

const provinceOptions = chinaPcaCode
const cityOptions = computed<ChinaPcaNode[]>(() => {
  const province = provinceOptions.find((item) => item.code === form.value.province)
  return province?.children ?? []
})
const areaOptions = computed<ChinaPcaNode[]>(() => {
  const city = cityOptions.value.find((item) => item.code === form.value.city)
  return city?.children ?? []
})

const dialogTitle = computed(() => props.mode === "create" ? "添加客户" : "编辑客户")
const submitText = computed(() => props.mode === "create" ? "添加" : "保存")

const clearUniqueErrors = () => {
  uniqueNameError.value = ""
  uniqueLegalNameError.value = ""
  uniqueWeixinError.value = ""
  duplicatePhones.value = []
}

const hasUniqueConflict = () => {
  return Boolean(
    uniqueNameError.value
    || uniqueLegalNameError.value
    || uniqueWeixinError.value
    || duplicatePhones.value.length > 0
    || localDuplicatePhones.value.length > 0
  )
}

const normalizePhonesForCheck = (phones: CustomerFormPhone[] = form.value.phones) => {
  const deduplicated: string[] = []
  const seen = new Set<string>()

  phones.forEach((item) => {
    const phone = normalizePhoneInput(item.phone)
    if (!isValidCnMobile(phone) || seen.has(phone)) {
      return
    }
    seen.add(phone)
    deduplicated.push(phone)
  })

  return deduplicated
}

const getExcludeCustomerId = () => {
  return props.mode === "edit" && props.customer?.id ? props.customer.id : undefined
}

const getLocalDuplicatePhones = (phones: CustomerFormPhone[] = form.value.phones): string[] => {
  const counts = new Map<string, number>()

  phones.forEach((item) => {
    const phone = normalizePhoneInput(item.phone)
    if (!isValidCnMobile(phone)) return
    counts.set(phone, (counts.get(phone) ?? 0) + 1)
  })

  return Array.from(counts.entries())
    .filter(([, count]) => count > 1)
    .map(([phone]) => phone)
}

const syncBackendDuplicatePhones = () => {
  const currentValidPhones = new Set(normalizePhonesForCheck())
  duplicatePhones.value = duplicatePhones.value.filter((phone) => currentValidPhones.has(phone))
}

const getPhoneFieldError = (value: string): string => {
  const phone = normalizePhoneInput(value)
  if (!phone) return ""
  if (phone.length !== 11) return "手机号必须为11位数字"
  if (!isValidCnMobile(phone)) return "请输入有效的中国大陆手机号"
  if (localDuplicatePhoneSet.value.has(phone)) return "手机号在当前表单中重复"
  if (duplicatePhoneSet.value.has(phone)) return "系统中已存在该手机号"
  return ""
}

const runUniqueCheck = async (phonesForCheck?: string[]) => {
  const payload = {
    excludeCustomerId: getExcludeCustomerId(),
    name: form.value.name.trim(),
    legalName: form.value.legalName.trim(),
    weixin: form.value.weixin.trim(),
    phones: phonesForCheck ?? normalizePhonesForCheck(),
  }

  if (!payload.name && !payload.legalName && !payload.weixin && payload.phones.length === 0) {
    clearUniqueErrors()
    return false
  }

  const seq = ++uniqueCheckSeq
  checkingUnique.value = true
  try {
    const result = await validateCustomerUnique(payload)
    if (seq !== uniqueCheckSeq) {
      return hasUniqueConflict()
    }
    uniqueNameError.value = result.nameExists ? "公司名称已存在" : ""
    uniqueLegalNameError.value = result.legalNameExists ? "公司法人已存在" : ""
    uniqueWeixinError.value = result.weixinExists ? "微信号已存在" : ""
    duplicatePhones.value = result.duplicatePhones ?? []
  } catch {
    // ignore network fluctuation, keep current editing flow
  } finally {
    if (seq === uniqueCheckSeq) {
      checkingUnique.value = false
    }
  }

  return hasUniqueConflict()
}

const scheduleUniqueCheck = (delay = 320) => {
  if (uniqueCheckTimer) {
    clearTimeout(uniqueCheckTimer)
  }
  uniqueCheckTimer = setTimeout(() => {
    void runUniqueCheck()
  }, delay)
}

const isDuplicatePhone = (phone: string) => {
  const normalized = normalizePhoneInput(phone)
  return Boolean(
    normalized
    && (duplicatePhoneSet.value.has(normalized) || localDuplicatePhoneSet.value.has(normalized))
  )
}

watch(
  () => [props.open, props.mode, props.customer],
  ([open]) => {
    if (!open) return
    formError.value = ""
    clearUniqueErrors()
    form.value = {
      name: props.customer?.name ?? "",
      legalName: props.customer?.legalName ?? "",
      contactName: props.customer?.contactName ?? "",
      email: props.customer?.email ?? "",
      weixin: props.customer?.weixin ?? "",
      province: props.customer?.province ? String(props.customer.province) : "",
      city: props.customer?.city ? String(props.customer.city) : "",
      area: props.customer?.area ? String(props.customer.area) : "",
      detailAddress: props.customer?.detailAddress ?? "",
      remark: props.customer?.remark ?? "",
      phones: props.customer?.phones?.length
        ? props.customer.phones.map((item) => ({
            id: item.id,
            phone: normalizePhoneInput(item.phone ?? ""),
            phoneLabel: item.phoneLabel ?? "手机",
            isPrimary: item.isPrimary,
          }))
        : [createEmptyPhone(true)],
    }
    scheduleUniqueCheck(80)
  },
  { immediate: true }
)

watch(
  () => form.value.province,
  (provinceCode) => {
    if (!provinceCode) {
      form.value.city = ""
      form.value.area = ""
      return
    }
    if (!cityOptions.value.some((item) => item.code === form.value.city)) {
      form.value.city = ""
      form.value.area = ""
    }
  }
)

watch(
  () => form.value.city,
  (cityCode) => {
    if (!cityCode) {
      form.value.area = ""
      return
    }
    if (!areaOptions.value.some((item) => item.code === form.value.area)) {
      form.value.area = ""
    }
  }
)

watch(
  () => [
    props.open,
    form.value.name,
    form.value.legalName,
    form.value.weixin,
    form.value.phones.map((item) => normalizePhoneInput(item.phone)).join("|"),
  ],
  ([open]) => {
    if (!open) return
    syncBackendDuplicatePhones()
    const hasPartialPhone = form.value.phones.some((item) => {
      const phone = normalizePhoneInput(item.phone)
      return phone.length > 0 && !isValidCnMobile(phone)
    })
    scheduleUniqueCheck(hasPartialPhone ? 420 : 280)
  }
)

onBeforeUnmount(() => {
  if (uniqueCheckTimer) {
    clearTimeout(uniqueCheckTimer)
  }
  uniqueCheckSeq += 1
})

const close = () => {
  if (props.submitting) return
  emit("update:open", false)
}

const addPhone = () => {
  form.value.phones.push(createEmptyPhone(form.value.phones.length === 0))
}

const handlePhoneInput = (index: number, value: string | number) => {
  const phone = normalizePhoneInput(String(value ?? ""))
  if (!form.value.phones[index]) return
  form.value.phones[index].phone = phone
  syncBackendDuplicatePhones()
  if (phone.length === 11) {
    scheduleUniqueCheck(120)
  }
}

const handlePhoneBlur = () => {
  if (!props.open) return
  scheduleUniqueCheck(120)
}

const removePhone = (index: number) => {
  if (form.value.phones.length <= 1) return
  const deletingPrimary = form.value.phones[index].isPrimary
  form.value.phones.splice(index, 1)
  if (deletingPrimary && form.value.phones.length > 0) {
    form.value.phones[0].isPrimary = true
  }
  syncBackendDuplicatePhones()
  scheduleUniqueCheck(120)
}

const setPrimaryPhone = (index: number) => {
  form.value.phones = form.value.phones.map((item, idx) => ({
    ...item,
    isPrimary: idx === index,
  }))
}

const normalizePhoneList = (): CustomerFormPhone[] => {
  const cleaned = form.value.phones
    .map((item) => ({
      ...item,
      phone: normalizePhoneInput(item.phone),
      phoneLabel: item.phoneLabel?.trim() || "手机",
    }))
    .filter((item) => item.phone.length > 0)

  if (cleaned.length > 0 && !cleaned.some((item) => item.isPrimary)) {
    cleaned[0].isPrimary = true
  }

  return cleaned
}

const handleSubmit = async () => {
  const phones = normalizePhoneList()

  if (phones.length === 0) {
    formError.value = "请至少填写一个联系电话"
    return
  }

  if (phones.some((item) => !isValidCnMobile(item.phone))) {
    formError.value = "请填写有效的中国大陆手机号"
    return
  }

  const duplicatedInForm = getLocalDuplicatePhones(phones)
  if (duplicatedInForm.length > 0) {
    formError.value = `手机号重复：${duplicatedInForm.join("、")}`
    return
  }

  if (uniqueCheckTimer) {
    clearTimeout(uniqueCheckTimer)
    uniqueCheckTimer = null
  }
  const hasConflict = await runUniqueCheck(phones.map((item) => item.phone))
  if (hasConflict) {
    formError.value = "公司名称、法人、手机号或微信存在重复，请修改后再保存"
    return
  }

  const payload: CustomerFormPayload = {
    name: form.value.name.trim(),
    legalName: form.value.legalName.trim() || "",
    contactName: form.value.contactName.trim() || "",
    email: form.value.email.trim() || "",
    weixin: form.value.weixin.trim() || "",
    province: toCodeNumber(form.value.province),
    city: toCodeNumber(form.value.city),
    area: toCodeNumber(form.value.area),
    detailAddress: form.value.detailAddress.trim() || "",
    remark: form.value.remark.trim() || "",
    phones,
  }

  if (!payload.name) {
    formError.value = "客户名称不能为空"
    return
  }

  formError.value = ""
  emit("submit", payload)
}
</script>

<template>
  <Dialog :open="open" @update:open="(val) => emit('update:open', val)">
    <DialogContent class="flex max-h-[85vh] flex-col overflow-hidden p-0 sm:max-w-[760px]">
      <DialogHeader class="shrink-0 px-6 pt-6 pb-2">
        <DialogTitle>{{ dialogTitle }}</DialogTitle>
        <DialogDescription>填写客户完整信息后保存。</DialogDescription>
      </DialogHeader>

      <form class="flex min-h-0 flex-1 flex-col" @submit.prevent="handleSubmit">
        <div class="min-h-0 flex-1 overflow-y-auto px-6 pb-4">
          <div v-if="formError" class="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {{ formError }}
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2 sm:col-span-2">
              <Label for="customer-name"><span class="mr-1 text-destructive">*</span>客户名称</Label>
              <Input
                id="customer-name"
                v-model="form.name"
                placeholder="请输入客户名称"
                :disabled="submitting"
              />
              <p v-if="uniqueNameError" class="text-xs text-destructive">
                {{ uniqueNameError }}
              </p>
            </div>

            <div class="space-y-2">
              <Label for="customer-legal-name">法人</Label>
              <Input
                id="customer-legal-name"
                v-model="form.legalName"
                placeholder="请输入法人姓名"
                :disabled="submitting"
              />
              <p v-if="uniqueLegalNameError" class="text-xs text-destructive">
                {{ uniqueLegalNameError }}
              </p>
            </div>

            <div class="space-y-2">
              <Label for="customer-contact">联系人</Label>
              <Input
                id="customer-contact"
                v-model="form.contactName"
                placeholder="请输入联系人"
                :disabled="submitting"
              />
            </div>
            <div class="space-y-2 sm:col-span-2">
              <div class="flex items-center justify-between">
                <Label><span class="mr-1 text-destructive">*</span>联系电话</Label>
                <Button type="button" variant="outline" size="sm" :disabled="submitting" @click="addPhone">
                  <Plus class="h-4 w-4" />
                  <span>新增手机号</span>
                </Button>
              </div>

              <div class="space-y-2 rounded-md border border-border/70 p-3">
                <div
                  v-for="(phone, idx) in form.phones"
                  :key="phone.id ?? `new-${idx}`"
                  class="space-y-1"
                >
                  <div class="grid gap-2 sm:grid-cols-[1fr_120px_auto_auto]">
                    <Input
                      :model-value="phone.phone"
                      placeholder="请输入手机号"
                      :disabled="submitting"
                      :class="isDuplicatePhone(phone.phone) || phoneFieldErrors[idx] ? 'border-destructive focus-visible:ring-destructive' : ''"
                      @update:model-value="(value) => handlePhoneInput(idx, value)"
                      @blur="handlePhoneBlur"
                    />
                    <Input
                      v-model="phone.phoneLabel"
                      placeholder="标签"
                      :disabled="submitting"
                    />
                    <Button
                      type="button"
                      size="sm"
                      :variant="phone.isPrimary ? 'default' : 'outline'"
                      :disabled="submitting"
                      @click="setPrimaryPhone(idx)"
                    >
                      <Star class="h-4 w-4" />
                      <span>{{ phone.isPrimary ? "主号" : "设为主号" }}</span>
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      class="h-9 w-9"
                      :disabled="submitting || form.phones.length <= 1"
                      @click="removePhone(idx)"
                    >
                      <Trash2 class="h-4 w-4" />
                    </Button>
                  </div>
                  <p v-if="phoneFieldErrors[idx]" class="text-xs text-destructive">
                    {{ phoneFieldErrors[idx] }}
                  </p>
                </div>
              </div>
              <p v-if="localDuplicatePhones.length > 0" class="text-xs text-destructive">
                当前表单手机号重复：{{ localDuplicatePhones.join("、") }}
              </p>
              <p v-if="duplicatePhones.length > 0" class="text-xs text-destructive">
                系统中已存在手机号：{{ duplicatePhones.join("、") }}
              </p>
            </div>

            <div class="space-y-2">
              <Label for="customer-weixin">微信</Label>
              <Input
                id="customer-weixin"
                v-model="form.weixin"
                placeholder="请输入微信号"
                :disabled="submitting"
              />
              <p v-if="uniqueWeixinError" class="text-xs text-destructive">
                {{ uniqueWeixinError }}
              </p>
            </div>

            <div class="space-y-2">
              <Label for="customer-email">邮箱</Label>
              <Input
                id="customer-email"
                v-model="form.email"
                type="email"
                placeholder="请输入邮箱"
                :disabled="submitting"
              />
            </div>

            <div class="space-y-2">
              <Label for="customer-province">省(编码)</Label>
              <Select v-model="form.province" :disabled="submitting">
                <SelectTrigger id="customer-province" class="h-10">
                  <SelectValue placeholder="请选择省份" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem
                      v-for="province in provinceOptions"
                      :key="province.code"
                      :value="province.code"
                    >
                      {{ province.name }} ({{ province.code }})
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-2">
              <Label for="customer-city">市(编码)</Label>
              <Select v-model="form.city" :disabled="submitting || !form.province">
                <SelectTrigger id="customer-city" class="h-10">
                  <SelectValue placeholder="请选择城市" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem
                      v-for="city in cityOptions"
                      :key="city.code"
                      :value="city.code"
                    >
                      {{ city.name }} ({{ city.code }})
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-2">
              <Label for="customer-area">区(编码)</Label>
              <Select v-model="form.area" :disabled="submitting || !form.city">
                <SelectTrigger id="customer-area" class="h-10">
                  <SelectValue placeholder="请选择区县" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem
                      v-for="area in areaOptions"
                      :key="area.code"
                      :value="area.code"
                    >
                      {{ area.name }} ({{ area.code }})
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-2 sm:col-span-2">
              <Label for="customer-detail-address">详细地址</Label>
              <Input
                id="customer-detail-address"
                v-model="form.detailAddress"
                placeholder="请输入详细地址"
                :disabled="submitting"
              />
            </div>

            

            <div class="space-y-2 sm:col-span-2">
              <Label for="customer-remark">备注</Label>
              <textarea
                id="customer-remark"
                v-model="form.remark"
                class="flex min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                placeholder="请输入备注"
                :disabled="submitting"
              />
            </div>
          </div>
        </div>
      </form>

      <DialogFooter class="shrink-0 border-t px-6 py-4">
        <Button type="button" variant="outline" :disabled="submitting" @click="close">
          取消
        </Button>
        <Button type="button" :disabled="submitting || checkingUnique || hasPhoneFieldError" @click="handleSubmit">
          <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />
          {{ submitText }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
