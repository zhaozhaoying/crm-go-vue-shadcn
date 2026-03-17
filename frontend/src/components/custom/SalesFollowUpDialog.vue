<script setup lang="ts">
import { ref, watch, computed } from "vue";
import { Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";

import {
  createSalesFollowRecord,
  getSalesFollowRecords,
  getCustomerLevels,
  getCustomerSources,
  getFollowMethods,
} from "@/api/modules/followRecords";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DatetimePicker } from "@/components/ui/datetime-picker";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type {
  CustomerLevel,
  CustomerSource,
  FollowMethod,
  SalesFollowRecord,
} from "@/api/modules/followRecords";

interface Props {
  open: boolean;
  customerId?: number | null;
  submitting?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  customerId: null,
  submitting: false,
});

const emit = defineEmits<{
  (e: "update:open", value: boolean): void;
  (e: "submit"): void;
}>();

interface FormState {
  content: string;
  nextFollowTime: string;
  customerLevelId: string;
  customerSourceId: string;
  followMethodId: string;
}

const serverError = ref("");
const submitAttempted = ref(false);
const loadingRecords = ref(false);
const form = ref<FormState>({
  content: "",
  nextFollowTime: "",
  customerLevelId: "",
  customerSourceId: "",
  followMethodId: "",
});

const contentError = computed(() => {
  const content = form.value.content.trim();

  if (!content) {
    return submitAttempted.value ? "请输入跟进内容" : "";
  }

  if (content.includes("跟进")) {
    return '跟进内容不能包含"跟进"两个字';
  }

  if (/[a-zA-Z]+/.test(content)) {
    return "跟进内容不能包含英文单词";
  }

  if (content.length < 10) {
    return "跟进内容必须至少10个字";
  }

  return "";
});

const followMethodError = computed(() => {
  if (form.value.followMethodId) {
    return "";
  }
  return submitAttempted.value ? "请选择跟进方式" : "";
});

const validationError = computed(() => contentError.value || followMethodError.value);
const formError = computed(() =>
  (submitAttempted.value ? validationError.value : "") || serverError.value,
);

const customerLevels = ref<CustomerLevel[]>([]);
const customerSources = ref<CustomerSource[]>([]);
const followMethods = ref<FollowMethod[]>([]);
const records = ref<SalesFollowRecord[]>([]);
const loadingLevels = ref(false);
const loadingSources = ref(false);
const loadingMethods = ref(false);

const loadCustomerLevels = async () => {
  loadingLevels.value = true;
  try {
    const data = await getCustomerLevels();
    customerLevels.value = data || [];
  } catch (err) {
    console.error("加载客户级别失败", err);
  } finally {
    loadingLevels.value = false;
  }
};

const loadCustomerSources = async () => {
  loadingSources.value = true;
  try {
    const data = await getCustomerSources();
    customerSources.value = data || [];
  } catch (err) {
    console.error("加载客户来源失败", err);
  } finally {
    loadingSources.value = false;
  }
};

const loadFollowMethods = async () => {
  loadingMethods.value = true;
  try {
    const data = await getFollowMethods();
    followMethods.value = data || [];
  } catch (err) {
    console.error("加载跟进方式失败", err);
  } finally {
    loadingMethods.value = false;
  }
};

const loadRecords = async () => {
  if (!props.customerId) return;

  loadingRecords.value = true;
  try {
    const data = await getSalesFollowRecords(props.customerId);
    records.value = data.items || [];
  } catch (err) {
    console.error("加载跟进记录失败", err);
  } finally {
    loadingRecords.value = false;
  }
};

watch(
  () => props.open,
  (open) => {
    if (open) {
      serverError.value = "";
      submitAttempted.value = false;
      form.value = {
        content: "",
        nextFollowTime: "",
        customerLevelId: "",
        customerSourceId: "",
        followMethodId: "",
      };
      loadCustomerLevels();
      loadCustomerSources();
      loadFollowMethods();
      loadRecords();
    }
  },
  { immediate: true },
);

const close = () => {
  if (props.submitting) return;
  emit("update:open", false);
};

const formatDateTime = (dateStr?: string): string => {
  if (!dateStr) return "-";
  return new Date(dateStr).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
};

/**
 * 将日期时间字符串转换为 RFC3339 格式
 * 例如: 2026-03-02T16:39:13Z
 */
const formatDateTimeForApi = (dateTimeStr: string | undefined): string | undefined => {
  if (!dateTimeStr) return undefined;
  // DatetimePicker 返回 "YYYY-MM-DDTHH:mm:ss"
  // 需要转换为 RFC3339 格式: "YYYY-MM-DDTHH:mm:ssZ"
  try {
    const date = new Date(dateTimeStr);
    // 使用 toISOString() 返回 UTC 时间的 ISO 8601 格式 (RFC3339)
    // 例如: "2026-03-02T16:39:13.000Z"
    // 去掉毫秒部分: "2026-03-02T16:39:13Z"
    return date.toISOString().replace('.000Z', 'Z');
  } catch (err) {
    console.error('日期格式转换失败:', err);
    return undefined;
  }
};

const handleSubmit = async () => {
  submitAttempted.value = true;
  serverError.value = "";

  if (!props.customerId) {
    serverError.value = "客户ID不能为空";
    toast.error(serverError.value);
    return;
  }

  if (validationError.value) {
    toast.error(validationError.value);
    return;
  }

  try {
    await createSalesFollowRecord({
      customerId: props.customerId,
      content: form.value.content.trim(),
      nextFollowTime: formatDateTimeForApi(form.value.nextFollowTime),
      customerLevelId: form.value.customerLevelId
        ? Number(form.value.customerLevelId)
        : undefined,
      customerSourceId: form.value.customerSourceId
        ? Number(form.value.customerSourceId)
        : undefined,
      followMethodId: Number(form.value.followMethodId),
    });

    // 重置表单
    form.value = {
      content: "",
      nextFollowTime: "",
      customerLevelId: "",
      customerSourceId: "",
      followMethodId: "",
    };
    submitAttempted.value = false;

    // 刷新记录列表
    await loadRecords();

    // 通知父组件刷新客户列表
    emit("submit");

    toast.success("销售跟进记录添加成功");
  } catch (err: any) {
    serverError.value = err?.response?.data?.message || err?.message || "添加失败";
    toast.error(serverError.value);
  }
};
</script>

<template>
  <Dialog :open="open" @update:open="(val) => emit('update:open', val)">
    <DialogContent class="max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader>
        <DialogTitle>销售跟进</DialogTitle>
        <DialogDescription>记录客户跟进信息，更新客户级别和来源</DialogDescription>
      </DialogHeader>

      <div class="flex-1 overflow-y-auto -mx-6 px-6">
        <div class="space-y-6 py-4">
          <!-- 添加跟进记录表单 -->
          <div class="border rounded-lg p-4 space-y-4 bg-muted/30">
            <h3 class="font-medium text-sm">添加跟进记录</h3>

            <div v-if="formError" class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {{ formError }}
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-2">
                <Label for="customer-level">客户级别</Label>
                <Select v-model="form.customerLevelId" :disabled="submitting || loadingLevels">
                  <SelectTrigger id="customer-level">
                    <SelectValue placeholder="选择客户级别" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem
                        v-for="level in customerLevels"
                        :key="level.id"
                        :value="String(level.id)"
                      >
                        {{ level.name }}
                      </SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>

              <div class="space-y-2">
                <Label for="customer-source">客户来源</Label>
                <Select v-model="form.customerSourceId" :disabled="submitting || loadingSources">
                  <SelectTrigger id="customer-source">
                    <SelectValue placeholder="选择客户来源" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem
                        v-for="source in customerSources"
                        :key="source.id"
                        :value="String(source.id)"
                      >
                        {{ source.name }}
                      </SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>

              <div class="space-y-2">
                <Label for="follow-method">
                  <span class="mr-1 text-destructive">*</span>跟进方式
                </Label>
                <Select v-model="form.followMethodId" :disabled="submitting || loadingMethods">
                  <SelectTrigger
                    id="follow-method"
                    :class="followMethodError ? 'border-destructive focus:ring-destructive/20' : ''"
                  >
                    <SelectValue placeholder="选择跟进方式" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem
                        v-for="method in followMethods"
                        :key="method.id"
                        :value="String(method.id)"
                      >
                        {{ method.name }}
                      </SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
                <p v-if="followMethodError" class="text-sm text-destructive">
                  {{ followMethodError }}
                </p>
              </div>

              <div class="space-y-2">
                <Label for="next-follow-time">下次跟进时间</Label>
                <DatetimePicker
                  id="next-follow-time"
                  v-model="form.nextFollowTime"
                  placeholder="请选择下次跟进时间"
                  content-align="end"
                  :disabled="submitting"
                />
              </div>
            </div>

            <div class="space-y-2">
              <Label for="follow-content">
                <span class="mr-1 text-destructive">*</span>跟进内容
              </Label>
              <Textarea
                id="follow-content"
                v-model="form.content"
                placeholder="请输入跟进内容，根据内容文字字数大于10，内容不能出现跟进两个字，内容不能包含英文单词"
                :rows="3"
                :disabled="submitting"
                :class="contentError ? 'border-destructive focus-visible:ring-destructive/20' : ''"
              />
              <p v-if="contentError" class="text-sm text-destructive">
                {{ contentError }}
              </p>
            </div>

            <div class="flex justify-end">
              <Button
                type="button"
                :disabled="submitting"
                @click="handleSubmit"
              >
                <Loader2 v-if="submitting" class="mr-2 h-4 w-4 animate-spin" />
                添加跟进记录
              </Button>
            </div>
          </div>

          <!-- 历史记录表格 -->
          <div class="space-y-3">
            <h3 class="font-medium text-sm">历史跟进记录</h3>

            <div v-if="loadingRecords" class="flex items-center justify-center py-8">
              <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
            </div>

            <div v-else-if="records.length === 0" class="text-center py-8 text-muted-foreground text-sm border rounded-lg">
              暂无跟进记录
            </div>

            <div v-else class="border rounded-lg overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow class="bg-muted/50">
                    <TableHead class="w-24">客户级别</TableHead>
                    <TableHead class="w-24">客户来源</TableHead>
                    <TableHead class="w-24">跟进方式</TableHead>
                    <TableHead>跟进内容</TableHead>
                    <TableHead class="w-36">下次跟进</TableHead>
                    <TableHead class="w-32">操作人</TableHead>
                    <TableHead class="w-36">创建时间</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow
                    v-for="record in records"
                    :key="record.id"
                    class="hover:bg-muted/30"
                  >
                    <TableCell>
                      <Badge variant="outline" class="bg-background">
                        {{ record.customerLevelName || '-' }}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" class="bg-background">
                        {{ record.customerSourceName || '-' }}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" class="bg-background">
                        {{ record.followMethodName || '-' }}
                      </Badge>
                    </TableCell>
                    <TableCell class="max-w-md">
                      <div class="truncate text-sm" :title="record.content">
                        {{ record.content }}
                      </div>
                    </TableCell>
                    <TableCell class="text-xs text-muted-foreground">
                      {{ formatDateTime(record.nextFollowTime) }}
                    </TableCell>
                    <TableCell class="text-xs">
                      {{ record.operatorUserName || '-' }}
                    </TableCell>
                    <TableCell class="text-xs text-muted-foreground">
                      {{ formatDateTime(record.createdAt) }}
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
          </div>
        </div>
      </div>

      <div class="flex justify-end border-t pt-4">
        <Button
          type="button"
          variant="outline"
          :disabled="submitting"
          @click="close"
        >
          关闭
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
