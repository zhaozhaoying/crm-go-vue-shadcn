<script setup lang="ts">
import { ref, watch, computed } from "vue";
import { Camera, Clock, Loader2 } from "lucide-vue-next";
import { toast } from "vue-sonner";

import { createOperationFollowRecord, getOperationFollowRecords, getFollowMethods } from "@/api/modules/followRecords";
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
import type { FollowMethod, OperationFollowRecord } from "@/api/modules/followRecords";

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
  followMethodId: string;
  appointmentTime: string;
  nextFollowTime: string;
  shootingTime: string;
}

const formError = ref("");
const loadingRecords = ref(false);
const form = ref<FormState>({
  content: "",
  followMethodId: "",
  appointmentTime: "",
  nextFollowTime: "",
  shootingTime: "",
});

const contentError = computed(() => {
  const content = form.value.content.trim();

  if (!content) {
    return "";
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

const followMethods = ref<FollowMethod[]>([]);
const records = ref<OperationFollowRecord[]>([]);
const loadingMethods = ref(false);

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
    const data = await getOperationFollowRecords(props.customerId);
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
      formError.value = "";
      form.value = {
        content: "",
        followMethodId: "",
        appointmentTime: "",
        nextFollowTime: "",
        shootingTime: "",
      };
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

const handleSubmit = async () => {
  if (!props.customerId) {
    formError.value = "客户ID不能为空";
    return;
  }

  const content = form.value.content.trim();

  if (!content) {
    formError.value = "请输入跟进内容";
    return;
  }

  if (contentError.value) {
    formError.value = contentError.value;
    return;
  }

  if (!form.value.followMethodId) {
    formError.value = "请选择跟进方式";
    return;
  }

  formError.value = "";

  try {
    await createOperationFollowRecord({
      customerId: props.customerId,
      content: form.value.content.trim(),
      followMethodId: Number(form.value.followMethodId),
      appointmentTime: formatDateTimeForApi(form.value.appointmentTime),
      nextFollowTime: formatDateTimeForApi(form.value.nextFollowTime),
      shootingTime: formatDateTimeForApi(form.value.shootingTime),
    });

    // 重置表单
    form.value = {
      content: "",
      followMethodId: "",
      appointmentTime: "",
      nextFollowTime: "",
      shootingTime: "",
    };

    // 刷新记录列表
    await loadRecords();

    // 通知父组件刷新客户列表
    emit("submit");

    toast.success("运营跟进记录添加成功");
  } catch (err: any) {
    formError.value = err?.response?.data?.message || err?.message || "添加失败";
    toast.error(formError.value);
  }
};
</script>

<template>
  <Dialog :open="open" @update:open="(val) => emit('update:open', val)">
    <DialogContent class="max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader>
        <DialogTitle>运营跟进</DialogTitle>
        <DialogDescription>记录运营跟进信息，包括约见、拍摄等安排</DialogDescription>
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
                <Label for="follow-method">
                  <span class="mr-1 text-destructive">*</span>跟进方式
                </Label>
                <Select v-model="form.followMethodId" :disabled="submitting || loadingMethods">
                  <SelectTrigger id="follow-method">
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
              </div>

              <div class="space-y-2">
                <Label for="appointment-time">
                  <Clock class="inline h-4 w-4 mr-1" />
                  约见时间
                </Label>
                <DatetimePicker
                  id="appointment-time"
                  v-model="form.appointmentTime"
                  placeholder="请选择约见时间"
                  :disabled="submitting"
                />
              </div>

              <div class="space-y-2">
                <Label for="shooting-time">
                  <Camera class="inline h-4 w-4 mr-1" />
                  拍摄时间
                </Label>
                <DatetimePicker
                  id="shooting-time"
                  v-model="form.shootingTime"
                  placeholder="请选择拍摄时间"
                  :disabled="submitting"
                />
              </div>

              <div class="space-y-2">
                <Label for="next-follow-time">
                  <Clock class="inline h-4 w-4 mr-1" />
                  下次跟进时间
                </Label>
                <DatetimePicker
                  id="next-follow-time"
                  v-model="form.nextFollowTime"
                  placeholder="请选择下次跟进时间"
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
                placeholder="请输入跟进内容"
                :rows="3"
                :disabled="submitting"
                :class="contentError ? 'border-destructive' : ''"
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
                    <TableHead class="w-24">跟进方式</TableHead>
                    <TableHead>跟进内容</TableHead>
                    <TableHead class="w-36">约见时间</TableHead>
                    <TableHead class="w-36">拍摄时间</TableHead>
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
                        {{ record.followMethodName || '-' }}
                      </Badge>
                    </TableCell>
                    <TableCell class="max-w-md">
                      <div class="truncate text-sm" :title="record.content">
                        {{ record.content }}
                      </div>
                    </TableCell>
                    <TableCell class="text-xs text-muted-foreground">
                      {{ formatDateTime(record.appointmentTime) }}
                    </TableCell>
                    <TableCell class="text-xs text-muted-foreground">
                      {{ formatDateTime(record.shootingTime) }}
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
