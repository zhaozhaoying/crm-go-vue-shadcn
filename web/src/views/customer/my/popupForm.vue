<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { Loader2, Plus, Trash2 } from "lucide-vue-next";
import { toTypedSchema } from "@vee-validate/zod";
import { useForm, useFieldArray, useField } from "vee-validate";
import { toast } from "vue-sonner";
import * as z from "zod";

import { validateCustomerUnique } from "@/api/modules/customers";
import { Button } from "@/components/ui/button";
import { DatetimePicker } from "@/components/ui/datetime-picker";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import { chinaPcaCode, type ChinaPcaNode } from "@/data/china-pca-code";
import {
  isValidCustomerPhone,
  normalizeCustomerPhoneInput,
  requiredString,
  requiredStringish,
} from "@/lib/form-validation";
import type {
  Customer,
  CustomerFormPayload,
  CustomerFormPhone,
} from "@/types/customer";

interface Props {
  open: boolean;
  mode: "create" | "edit";
  customer?: Customer | null;
  submitting?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  customer: null,
  submitting: false,
});

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (e: "submit", payload: CustomerFormPayload): void;
}>();

const phoneSchema = z.object({
  id: z.number().optional(),
  phone: requiredStringish("联系电话")
    .transform((value) => normalizeCustomerPhoneInput(value))
    .refine((value) => isValidCustomerPhone(value), {
      message: "请输入有效的手机号或座机号，例如 13800138000 或 01088886666",
    }),
  phoneLabel: z.string().default("手机"),
  isPrimary: z.boolean().default(false),
});

const isEditMode = computed(() => props.mode === "edit");

const formSchema = computed(() =>
  toTypedSchema(
    z.object({
      name: requiredString("客户名称"),
      legalName: requiredString("法人").min(2, "法人至少需要2个字"),
      contactName: requiredString("联系人").min(2, "联系人至少需要2个字"),
      email: z
        .string()
        .email({ message: "请输入有效的邮箱地址" })
        .optional()
        .or(z.literal("")),
      weixin: z.string().optional(),
      province: z.string().optional(),
      city: z.string().optional(),
      area: z.string().optional(),
      detailAddress: z.string().optional(),
      nextTime: z.string().optional(),
      remark: z.string().optional(),
      phones: z
        .array(phoneSchema)
        .min(1, "请至少填写一个联系电话")
        .superRefine((phones, ctx) => {
          const phoneNumbers = phones.map((p) => p.phone);
          const uniquePhoneNumbers = new Set(phoneNumbers);
          if (uniquePhoneNumbers.size !== phoneNumbers.length) {
            const seen = new Set();
            phones.forEach((phone, index) => {
              if (seen.has(phone.phone)) {
                ctx.addIssue({
                  code: z.ZodIssueCode.custom,
                  message: "手机号在当前表单中重复",
                  path: [index, "phone"],
                });
              }
              seen.add(phone.phone);
            });
          }
        }),
    })
  )
);

const { handleSubmit, errors, setValues, setFieldError, resetForm, values } =
  useForm<FormState>({
    validationSchema: formSchema,
  });

const { value: name } = useField<string>("name");
const { value: legalName } = useField<string>("legalName");
const { value: contactName } = useField<string>("contactName");
const { value: email } = useField<string>("email");
const { value: weixin } = useField<string>("weixin");
const { value: province } = useField<string>("province");
const { value: city } = useField<string>("city");
const { value: area } = useField<string>("area");
const { value: detailAddress } = useField<string>("detailAddress");
const { value: nextTime } = useField<string>("nextTime");
const { value: remark } = useField<string>("remark");

interface FormState {
  name: string;
  legalName?: string;
  contactName?: string;
  email?: string;
  weixin?: string;
  province?: string;
  city?: string;
  area?: string;
  detailAddress?: string;
  nextTime?: string;
  remark?: string;
  phones: CustomerFormPhone[];
}

const { fields: phoneFields, remove: removePhone, push: addPhone } = useFieldArray<CustomerFormPhone>("phones");

const createEmptyPhone = (isPrimary = false): CustomerFormPhone => ({
  phone: "",
  phoneLabel: "手机",
  isPrimary,
});

const toCodeNumber = (value?: string): number | undefined => {
  if (!value) return undefined;
  const num = Number(value);
  return Number.isFinite(num) ? num : undefined;
};

const checkingUnique = ref(false);
const formError = ref("");
const backendDuplicatePhones = ref<string[]>([]);
let uniqueCheckTimer: ReturnType<typeof setTimeout> | null = null;
let uniqueCheckSeq = 0;

const provinceOptions = chinaPcaCode;
const cityOptions = computed<ChinaPcaNode[]>(() => {
  const provinceCode = province.value;
  const provinceItem = provinceOptions.find((item) => item.code === provinceCode);
  return provinceItem?.children ?? [];
});
const areaOptions = computed<ChinaPcaNode[]>(() => {
  const cityCode = city.value;
  const cityItem = cityOptions.value.find((item) => item.code === cityCode);
  return cityItem?.children ?? [];
});

const dialogTitle = computed(() =>
  isEditMode.value ? "编辑客户" : "添加客户"
);
const submitText = computed(() => (isEditMode.value ? "保存" : "添加"));

const getExcludeCustomerId = () => {
  return props.mode === "edit" && props.customer?.id
    ? props.customer.id
    : undefined;
};

const clearUniqueState = () => {
  formError.value = "";
  backendDuplicatePhones.value = [];
};

const getFirstSubmitErrorMessage = () => {
  if (backendDuplicatePhones.value.length > 0) {
    return `系统中已存在联系电话：${backendDuplicatePhones.value.join("、")}`;
  }

  const firstPhoneError = phoneFields.value
    .map((_, index) => errors.value[`phones.${index}.phone`])
    .find((message): message is string => Boolean(message));
  if (firstPhoneError) return firstPhoneError;

  return (
    errors.value.name ||
    errors.value.legalName ||
    errors.value.contactName ||
    errors.value.weixin ||
    errors.value.email ||
    errors.value.phones ||
    "请先修正表单中的错误后再提交"
  );
};

const runUniqueCheck = async () => {
  const seq = ++uniqueCheckSeq;
  checkingUnique.value = true;

  const phonesToCheck = phoneFields.value
    ?.map((p) => normalizeCustomerPhoneInput(p.value.phone))
    .filter((p): p is string => !!p && isValidCustomerPhone(p));

  try {
    const result = await validateCustomerUnique({
      excludeCustomerId: getExcludeCustomerId(),
      name: name.value,
      legalName: "",
      contactName: "",
      weixin: weixin.value,
      phones: phonesToCheck,
    });

    if (seq !== uniqueCheckSeq) return;

    setFieldError("name", result.nameExists ? "公司名称已存在" : undefined);
    setFieldError("weixin", result.weixinExists ? "微信号已存在" : undefined);
    backendDuplicatePhones.value = result.duplicatePhones ?? [];

    const phoneMap = new Map(
      phoneFields.value.map((field, index) => [normalizeCustomerPhoneInput(field.value.phone), index])
    );
    phoneFields.value.forEach((field, index) => {
      const normalizedPhone = normalizeCustomerPhoneInput(field.value.phone);
      if (!backendDuplicatePhones.value.includes(normalizedPhone)) {
        if (errors.value[`phones.${index}.phone`] === "系统中已存在该联系电话") {
          setFieldError(`phones.${index}.phone`, undefined);
        }
      }
    });
    result.duplicatePhones?.forEach((dupPhone) => {
      const index = phoneMap.get(dupPhone);
      if (index !== undefined) {
        setFieldError(`phones.${index}.phone`, "系统中已存在该联系电话");
      }
    });
  } catch {
    // ignore
  } finally {
    if (seq === uniqueCheckSeq) {
      checkingUnique.value = false;
    }
  }
};

const scheduleUniqueCheck = (delay = 320) => {
  if (uniqueCheckTimer) {
    clearTimeout(uniqueCheckTimer);
  }
  uniqueCheckTimer = setTimeout(() => {
    void runUniqueCheck();
  }, delay);
};

watch(
  () => ({ open: props.open, customer: props.customer }),
  ({ open, customer }) => {
    if (!open) return;

    resetForm();
    clearUniqueState();
    if (customer) {
      setValues({
        name: customer.name ?? "",
        legalName: customer.legalName ?? "",
        contactName: customer.contactName ?? "",
        email: customer.email ?? "",
        weixin: customer.weixin ?? "",
        province: customer.province ? String(customer.province) : "",
        city: customer.city ? String(customer.city) : "",
        area: customer.area ? String(customer.area) : "",
        detailAddress: customer.detailAddress ?? "",
        nextTime: customer.nextTime ?? "",
        remark: customer.remark ?? "",
        phones: customer.phones?.length
          ? customer.phones.map((p: any) => ({ ...p }))
          : [createEmptyPhone(true)],
      });
    } else {
      setValues({ phones: [createEmptyPhone(true)] });
    }
    scheduleUniqueCheck(80);
  },
  { immediate: true, deep: true }
);

watch(
  () => province.value,
  (provinceCode) => {
    if (!provinceCode) {
      city.value = "";
      area.value = "";
      return;
    }
    if (!cityOptions.value.some((item) => item.code === city.value)) {
      city.value = "";
      area.value = "";
    }
  }
);

watch(
  () => city.value,
  (cityCode) => {
    if (!cityCode) {
      area.value = "";
      return;
    }
    if (!areaOptions.value.some((item) => item.code === area.value)) {
      area.value = "";
    }
  }
);

watch(
  () => [
    name.value,
    legalName.value,
    contactName.value,
    weixin.value,
    ...(phoneFields.value?.map((p) => p.value.phone) ?? []),
  ],
  () => {
    if (!props.open) return;
    formError.value = "";
    scheduleUniqueCheck();
  },
  { deep: true }
);

onBeforeUnmount(() => {
  if (uniqueCheckTimer) {
    clearTimeout(uniqueCheckTimer);
  }
  uniqueCheckSeq += 1;
});

const close = () => {
  if (props.submitting) return;
  emit("update:open", false);
};

const handleAddPhone = () => {
  formError.value = "";
  addPhone(createEmptyPhone(phoneFields.value.length === 0));
};

const handleRemovePhone = (index: number) => {
  formError.value = "";
  if (phoneFields.value.length <= 1) return;
  const isPrimary = phoneFields.value[index].value.isPrimary;
  removePhone(index);
  if (isPrimary && phoneFields.value.length > 0) {
    phoneFields.value[0].value.isPrimary = true;
  }
};

const setPrimaryPhone = (index: number) => {
  formError.value = "";
  phoneFields.value.forEach((field, idx) => {
    field.value.isPrimary = idx === index;
  });
};

const onSubmit = handleSubmit(async (formValues) => {
  formError.value = "";
  await runUniqueCheck();
  if (Object.keys(errors.value).length > 0) {
    formError.value = getFirstSubmitErrorMessage();
    toast.error(formError.value);
    return;
  }

  const payload: CustomerFormPayload = {
    ...formValues,
    province: toCodeNumber(formValues.province),
    city: toCodeNumber(formValues.city),
    area: toCodeNumber(formValues.area),
    nextTime: formValues.nextTime || undefined,
  };

  emit("submit", payload);
});
</script>

<template>
  <Dialog :open="open" @update:open="(val) => emit('update:open', val)">
    <DialogContent
      class="flex max-h-[85vh] flex-col overflow-hidden p-0 sm:max-w-[760px]"
    >
      <DialogHeader class="shrink-0 px-6 pt-6 pb-2">
        <DialogTitle>{{ dialogTitle }}</DialogTitle>
      </DialogHeader>

      <form class="flex min-h-0 flex-1 flex-col" @submit.prevent="onSubmit">
        <div class="min-h-0 flex-1 overflow-y-auto px-6 pb-4">
          <div
            v-if="formError"
            class="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          >
            {{ formError }}
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-1.5 sm:col-span-2">
              <Label for="customer-name">
                <span class="mr-1 text-destructive">*</span>客户名称
              </Label>
              <Input
                id="customer-name"
                v-model="name"
                placeholder="请输入客户名称"
                :disabled="submitting"
              />
              <p v-if="errors.name" class="text-xs text-destructive">
                {{ errors.name }}
              </p>
            </div>

            <div class="space-y-1.5">
              <Label for="customer-legal-name">
                <span class="mr-1 text-destructive">*</span>法人
              </Label>
              <Input
                id="customer-legal-name"
                v-model="legalName"
                placeholder="请输入法人姓名"
                :disabled="submitting"
              />
              <p v-if="errors.legalName" class="text-xs text-destructive">
                {{ errors.legalName }}
              </p>
            </div>

            <div class="space-y-1.5">
              <Label for="customer-contact">
                <span class="mr-1 text-destructive">*</span>联系人
              </Label>
              <Input
                id="customer-contact"
                v-model="contactName"
                placeholder="请输入联系人"
                :disabled="submitting"
              />
               <p v-if="errors.contactName" class="text-xs text-destructive">
                {{ errors.contactName }}
              </p>
            </div>
            <div class="space-y-1.5 sm:col-span-2">
              <div class="flex items-center justify-between">
                <Label>
                  <span class="mr-1 text-destructive">*</span>联系电话
                </Label>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  :disabled="submitting"
                  @click="handleAddPhone"
                >
                  <Plus class="h-4 w-4" />
                  <span>新增电话</span>
                </Button>
              </div>

              <div class="space-y-2 rounded-md border border-border/70 p-3">
                <div
                  v-for="(field, idx) in phoneFields"
                  :key="field.key"
                  class="space-y-1"
                >
                  <div class="flex gap-2 items-center">
                    <Input
                      v-model="field.value.phone"
                      placeholder="请输入手机号或座机号"
                      :disabled="submitting"
                      class="flex-1"
                    />
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      class="shrink-0 border-input text-muted-foreground hover:text-foreground"
                      :disabled="submitting || phoneFields.length <= 1"
                      @click="handleRemovePhone(idx)"
                    >
                      <Trash2 class="h-4 w-4" />
                    </Button>
                  </div>
                  <p
                    v-if="errors[`phones.${idx}.phone`]"
                    class="text-xs text-destructive"
                  >
                    {{ errors[`phones.${idx}.phone`] }}
                  </p>
                </div>
              </div>
              <p v-if="errors.phones" class="text-xs text-destructive">
                {{ errors.phones }}
              </p>
              <p class="text-xs text-muted-foreground">
                支持手机号或座机号，座机示例：01088886666
              </p>
              <p
                v-if="backendDuplicatePhones.length > 0"
                class="text-xs text-destructive"
              >
                系统中已存在联系电话：{{ backendDuplicatePhones.join("、") }}
              </p>
            </div>

            <div class="space-y-1.5">
              <Label for="customer-weixin">微信</Label>
              <Input
                id="customer-weixin"
                v-model="weixin"
                placeholder="请输入微信号"
                :disabled="submitting"
              />
              <p v-if="errors.weixin" class="text-xs text-destructive">
                {{ errors.weixin }}
              </p>
            </div>

            <div class="space-y-1.5">
              <Label for="customer-email">邮箱</Label>
              <Input
                id="customer-email"
                v-model="email"
                type="email"
                placeholder="请输入邮箱"
                :disabled="submitting"
              />
              <p v-if="errors.email" class="text-xs text-destructive">
                {{ errors.email }}
              </p>
            </div>

            <div class="space-y-1.5">
              <Label for="customer-province">省份</Label>
              <Select v-model="province" :disabled="submitting">
                <SelectTrigger id="customer-province" class="h-10">
                  <SelectValue placeholder="请选择省份" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem
                      v-for="prov in provinceOptions"
                      :key="prov.code"
                      :value="prov.code"
                    >
                      {{ prov.name }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-1.5">
              <Label for="customer-city">城市</Label>
              <Select
                v-model="city"
                :disabled="submitting || !province"
              >
                <SelectTrigger id="customer-city" class="h-10">
                  <SelectValue placeholder="请选择城市" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem
                      v-for="cityOption in cityOptions"
                      :key="cityOption.code"
                      :value="cityOption.code"
                    >
                      {{ cityOption.name }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-1.5">
              <Label for="customer-area">区县</Label>
              <Select v-model="area" :disabled="submitting || !city">
                <SelectTrigger id="customer-area" class="h-10">
                  <SelectValue placeholder="请选择区县" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem
                      v-for="areaOption in areaOptions"
                      :key="areaOption.code"
                      :value="areaOption.code"
                    >
                      {{ areaOption.name }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-1.5">
              <Label for="customer-next-time">下次跟进时间</Label>
              <DatetimePicker
                id="customer-next-time"
                v-model="nextTime"
                placeholder="请选择下次跟进时间"
                :disabled="submitting"
              />
            </div>

            <div class="space-y-1.5 sm:col-span-2">
              <Label for="customer-detail-address">详细地址</Label>
              <textarea
                id="customer-detail-address"
                v-model="detailAddress"
                class="flex min-h-12 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                placeholder="请输入详细地址"
                :disabled="submitting"
              />
            </div>

            <div class="space-y-1.5 sm:col-span-2">
              <Label for="customer-remark">备注</Label>
              <textarea
                id="customer-remark"
                v-model="remark"
                class="flex min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                placeholder="请输入备注"
                :disabled="submitting"
              />
            </div>
          </div>
        </div>
      </form>

      <DialogFooter class="shrink-0 border-t px-6 py-4">
        <Button
          type="button"
          variant="outline"
          :disabled="submitting"
          @click="close"
        >
          取消
        </Button>
        <Button
          type="button"
          :disabled="submitting || checkingUnique"
          @click="onSubmit"
        >
          <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />
          {{ submitText }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
