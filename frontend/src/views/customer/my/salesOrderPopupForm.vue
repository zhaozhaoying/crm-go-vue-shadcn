<script setup lang="ts">
import axios from "axios";
import { computed, ref, watch } from "vue";
import { Loader2 } from "lucide-vue-next";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
  InputGroupText,
} from "@/components/ui/input-group";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import ImageUploadCard from "@/components/custom/ImageUploadCard.vue";
import { getSystemSettings } from "@/api/modules/systemSettings";
import {
  checkContractNumberAvailable,
  uploadContractImage,
} from "@/api/modules/contracts";
import type { Contract, ContractFormPayload } from "@/types/contract";

interface Props {
  open: boolean;
  mode: "create" | "edit";
  contract?: Contract | null;
  submitting?: boolean;
  readonly?: boolean;
  customerId?: number | null;
}

const props = withDefaults(defineProps<Props>(), {
  contract: null,
  submitting: false,
  readonly: false,
  customerId: null,
});

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (e: "submit", payload: ContractFormPayload): void;
}>();

interface FormState {
  contractImage: string;
  paymentImage: string;
  cooperationType: string;
  contractNumberSuffix: string;
  contractName: string;
  contractAmount: string;
  paymentAmount: string;
  cooperationYears: string;
  nodeCount: string;
  remark: string;
}

const createEmptyForm = (): FormState => {
  return {
    contractImage: "",
    paymentImage: "",
    cooperationType: "domestic",
    contractNumberSuffix: "",
    contractName: "",
    contractAmount: "0",
    paymentAmount: "0",
    cooperationYears: "0",
    nodeCount: "0",
    remark: "",
  };
};

const form = ref<FormState>(createEmptyForm());
const formError = ref("");
const contractNumberPrefix = ref("zzy_");
const formReadonly = computed(() => props.readonly);
const remarkOnlyEditMode = computed(() => props.mode === "edit" && !formReadonly.value);
const baseFieldsReadonly = computed(() => formReadonly.value || remarkOnlyEditMode.value);
const contractNumberChecking = ref(false);
const contractNumberError = ref("");
const contractNumberCheckSupported = ref(true);
let contractNumberCheckSeq = 0;

const dialogTitle = computed(() => {
  if (formReadonly.value) return "查看销售提单";
  return props.mode === "create" ? "新增销售提单" : "编辑销售提单";
});

const extractSuffix = (prefix: string, fullNumber?: string) => {
  const normalizedPrefix = prefix.trim();
  const value = (fullNumber ?? "").trim();
  if (!normalizedPrefix) return value;
  if (value.startsWith(normalizedPrefix)) {
    return value.slice(normalizedPrefix.length).trim();
  }
  return value;
};

const parseNumber = (raw: string, fallback = 0) => {
  const value = Number(raw);
  if (!Number.isFinite(value)) return fallback;
  return value;
};

const parseUnixFromISOString = (value?: string) => {
  if (!value) return null;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;
  return Math.floor(date.getTime() / 1000);
};

const normalizeText = (value: unknown) => String(value ?? "").trim();

const isContractNumberCheckUnsupported = (error: unknown) => {
  if (!axios.isAxiosError(error)) return false;
  const status = error.response?.status;
  const data = error.response?.data as { message?: string } | undefined;
  const message = String(data?.message ?? error.message ?? "").toLowerCase();
  return status === 404 || message.includes("invalid contract id");
};

const loadContractNumberPrefix = async () => {
  try {
    const settings = await getSystemSettings();
    const prefix = settings.contractNumberPrefix?.trim();
    contractNumberPrefix.value = prefix || "zzy_";
  } catch {
    contractNumberPrefix.value = "zzy_";
  }
};

watch(
  () => [props.open, props.mode, props.contract, props.customerId, contractNumberPrefix.value],
  ([open]) => {
    if (!open) return;
    formError.value = "";
    contractNumberError.value = "";
    form.value = {
      contractImage: props.contract?.contractImage ?? "",
      paymentImage: props.contract?.paymentImage ?? "",
      cooperationType: props.contract?.cooperationType ?? "domestic",
      contractNumberSuffix: extractSuffix(
        contractNumberPrefix.value,
        props.contract?.contractNumber,
      ),
      contractName: props.contract?.contractName ?? "",
      contractAmount: String(props.contract?.contractAmount ?? 0),
      paymentAmount: String(props.contract?.paymentAmount ?? 0),
      cooperationYears: String(props.contract?.cooperationYears ?? 0),
      nodeCount: String(props.contract?.nodeCount ?? 0),
      remark: props.contract?.remark ?? "",
    };
  },
  { immediate: true },
);

watch(
  () => props.open,
  (open) => {
    if (!open) return;
    loadContractNumberPrefix();
  },
  { immediate: true },
);

watch(
  () => form.value.contractNumberSuffix,
  (value) => {
    contractNumberError.value = "";
    const normalizedValue = String(value ?? "");
    const digitsOnly = normalizedValue.replace(/\D+/g, "");
    if (digitsOnly !== normalizedValue) {
      form.value.contractNumberSuffix = digitsOnly;
    }
  },
);

const close = () => {
  if (props.submitting) return;
  emit("update:open", false);
};

const uploadImage = async (file: File) => {
  return uploadContractImage(file);
};

const validateContractNumberUniqueness = async (showMessage: boolean) => {
  if (!contractNumberCheckSupported.value) {
    return true;
  }
  const suffix = normalizeText(form.value.contractNumberSuffix);
  if (!suffix) {
    if (showMessage) {
      contractNumberError.value = "合同编号后缀不能为空";
    }
    return false;
  }
  const prefix = contractNumberPrefix.value.trim() || "zzy_";
  const contractNumber = `${prefix}${suffix}`;
  const seq = ++contractNumberCheckSeq;
  contractNumberChecking.value = true;
  try {
    const result = await checkContractNumberAvailable(contractNumber, props.contract?.id);
    if (seq !== contractNumberCheckSeq) return false;
    if (!result.available) {
      contractNumberError.value = "合同编号已存在，请更换";
      return false;
    }
    contractNumberError.value = "";
    return true;
  } catch (error) {
    if (isContractNumberCheckUnsupported(error)) {
      contractNumberCheckSupported.value = false;
      contractNumberError.value = "";
      return true;
    }
    if (seq !== contractNumberCheckSeq) return false;
    if (showMessage) {
      contractNumberError.value = "合同编号校验失败，请重试";
    }
    return false;
  } finally {
    if (seq === contractNumberCheckSeq) {
      contractNumberChecking.value = false;
    }
  }
};

const handleContractNumberBlur = async () => {
  if (baseFieldsReadonly.value) return;
  await validateContractNumberUniqueness(false);
};

const submit = async () => {
  if (formReadonly.value) {
    close();
    return;
  }
  formError.value = "";

  const suffix = normalizeText(form.value.contractNumberSuffix);
  if (!suffix) {
    contractNumberError.value = "合同编号后缀不能为空";
    return;
  }
  if (!normalizeText(form.value.contractName)) {
    formError.value = "合同名称不能为空";
    return;
  }

  const resolvedCustomerId = Number(props.customerId || props.contract?.customerId || 0);
  if (resolvedCustomerId <= 0) {
    formError.value = "缺少客户ID";
    return;
  }

  const prefix = contractNumberPrefix.value.trim() || "zzy_";
  if (!remarkOnlyEditMode.value) {
    const numberAvailable = await validateContractNumberUniqueness(true);
    if (!numberAvailable) {
      return;
    }
  }

  const payload: ContractFormPayload = {
    contractImage: normalizeText(form.value.contractImage),
    paymentImage: normalizeText(form.value.paymentImage),
    paymentStatus: props.contract?.paymentStatus || "pending",
    remark: normalizeText(form.value.remark),
    customerId: resolvedCustomerId,
    cooperationType: form.value.cooperationType,
    contractNumber: `${prefix}${suffix}`,
    contractNumberSuffix: suffix,
    contractName: normalizeText(form.value.contractName),
    contractAmount: parseNumber(form.value.contractAmount, 0),
    paymentAmount: parseNumber(form.value.paymentAmount, 0),
    cooperationYears: parseNumber(form.value.cooperationYears, 0),
    nodeCount: parseNumber(form.value.nodeCount, 0),
    serviceUserId: props.contract?.serviceUserId ?? null,
    websiteName: props.contract?.websiteName || "",
    websiteUrl: props.contract?.websiteUrl || "",
    websiteUsername: props.contract?.websiteUsername || "",
    isOnline: props.contract?.isOnline ?? false,
    startDate: parseUnixFromISOString(props.contract?.startDate),
    endDate: parseUnixFromISOString(props.contract?.endDate),
    auditStatus: props.contract?.auditStatus || "pending",
    expiryHandlingStatus: props.contract?.expiryHandlingStatus || "pending",
  };

  emit("submit", payload);
};
</script>

<template>
  <Dialog :open="open" @update:open="(v) => emit('update:open', v)">
    <DialogContent class="flex max-h-[85vh] flex-col overflow-hidden p-0 sm:max-w-[820px]">
      <DialogHeader class="border-b px-6 pt-6 pb-4">
        <DialogTitle>{{ dialogTitle }}</DialogTitle>
      </DialogHeader>

      <form class="flex min-h-0 flex-1 flex-col" @submit.prevent="submit">
        <div class="min-h-0 flex-1 overflow-y-auto px-6 py-4">
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div class="space-y-2 md:col-span-2">
              <Label>合同编号</Label>
              <InputGroup>
                <InputGroupAddon class="px-2">
                  <InputGroupText class="text-xs">{{ contractNumberPrefix }}</InputGroupText>
                </InputGroupAddon>
                <InputGroupInput
                  v-model="form.contractNumberSuffix"
                  type="text"
                  inputmode="numeric"
                  pattern="[0-9]*"
                  placeholder="请输入合同编号"
                  :disabled="baseFieldsReadonly"
                  @blur="handleContractNumberBlur"
                />
              </InputGroup>
              <p v-if="contractNumberError" class="text-xs text-destructive">
                {{ contractNumberError }}
              </p>
              <p v-else-if="contractNumberChecking" class="text-xs text-muted-foreground">
                正在校验合同编号...
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

            <div class="space-y-1.5">
              <Label>合同金额</Label>
              <InputGroup>
                <InputGroupInput
                  v-model="form.contractAmount"
                  type="number"
                  min="0"
                  step="0.01"
                  :disabled="baseFieldsReadonly"
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
                  :disabled="baseFieldsReadonly"
                />
                <InputGroupAddon class="px-2">
                  <InputGroupText class="text-xs">元</InputGroupText>
                </InputGroupAddon>
              </InputGroup>
            </div>

            <div class="space-y-1.5">
              <Label>合作年限</Label>
              <InputGroup>
                <InputGroupInput
                  v-model="form.cooperationYears"
                  type="number"
                  min="0"
                  :disabled="baseFieldsReadonly"
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
                  :disabled="baseFieldsReadonly"
                />
                <InputGroupAddon class="px-2">
                  <InputGroupText class="text-xs">个</InputGroupText>
                </InputGroupAddon>
              </InputGroup>
            </div>

            <div class="space-y-2">
              <Label>合同图片</Label>
              <ImageUploadCard
                v-model="form.contractImage"
                placeholder="暂无合同图片"
                :disabled="baseFieldsReadonly"
                :on-upload="uploadImage"
                @error="(msg) => (formError = msg)"
              />
            </div>

            <div class="space-y-2">
              <Label>回款图片</Label>
              <ImageUploadCard
                v-model="form.paymentImage"
                placeholder="暂无回款图片"
                :disabled="baseFieldsReadonly"
                :on-upload="uploadImage"
                @error="(msg) => (formError = msg)"
              />
            </div>

            <div class="space-y-1.5 md:col-span-2">
              <Label>备注</Label>
              <Textarea
                v-model="form.remark"
                :rows="3"
                placeholder="请输入备注"
                :disabled="formReadonly"
              />
            </div>
          </div>
          <p v-if="formError" class="mt-4 text-sm text-destructive">{{ formError }}</p>
        </div>

        <DialogFooter class="border-t px-6 py-4">
          <Button type="button" variant="outline" @click="close">{{ formReadonly ? "关闭" : "取消" }}</Button>
          <Button v-if="!formReadonly" type="submit" :disabled="submitting">
            <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />
            保存
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
