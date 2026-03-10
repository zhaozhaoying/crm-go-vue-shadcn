<script setup lang="ts">
import { ref } from "vue";
import { Loader2, Plus, Check } from "lucide-vue-next";

import ErrorFeedback from "@/components/custom/ErrorFeedback.vue";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

interface Platform {
  value: number;
  label: string;
  description: string;
}

const props = defineProps<{
  creating: boolean;
  actionError: string;
  platforms: Platform[];
  defaultPlatform: string;
}>();

const emit = defineEmits<{
  create: [payload: { keyword: string; platform: number }];
}>();

const open = ref(false);
const keyword = ref("");
const selectedPlatform = ref(props.defaultPlatform);

const handleSubmit = () => {
  const trimmed = keyword.value.trim();
  if (!trimmed) return;
  emit("create", {
    keyword: trimmed,
    platform: Number(selectedPlatform.value),
  });
};

const handleCreated = () => {
  open.value = false;
  keyword.value = "";
};

defineExpose({ handleCreated });
</script>

<template>
  <Dialog v-model:open="open">
    <DialogTrigger as-child>
      <Button size="sm"> <Plus class="h-4 w-4 mr-1" />新建查询任务 </Button>
    </DialogTrigger>
    <DialogContent class="sm:max-w-[600px]">
      <DialogHeader>
        <DialogTitle>新建查询任务</DialogTitle>
        <DialogDescription>
          输入关键词并选择平台后，将持续获取直到平台没有下一页。
        </DialogDescription>
      </DialogHeader>
      <div class="space-y-6 py-4">
        <!-- 平台选择 - 卡片式 -->
        <div class="space-y-3">
          <label class="text-sm font-medium">选择获取平台</label>
          <div class="grid grid-cols-3 gap-2">
            <button
              v-for="p in platforms"
              :key="p.value"
              type="button"
              @click="selectedPlatform = String(p.value)"
              :class="cn(
                'flex items-center justify-center rounded-md px-3 py-2 text-sm font-medium transition-all',
                selectedPlatform === String(p.value)
                  ? 'bg-black text-white'
                  : 'bg-white text-black border border-gray-300 hover:border-gray-400'
              )"
            >
              {{ p.label }}
            </button>
          </div>
        </div>

        <!-- 关键词输入 -->
        <div class="space-y-2">
          <label class="text-sm font-medium">搜索关键词</label>
          <Input
            v-model="keyword"
            placeholder="例如：led light, solar panel, etc."
            class="w-full"
            @keyup.enter="handleSubmit"
          />
          <p class="text-xs text-muted-foreground">
            输入产品关键词，系统将自动获取所有相关供应商信息
          </p>
        </div>

        <!-- 错误提示 -->
        <div v-if="actionError">
          <ErrorFeedback :message="actionError" />
        </div>
      </div>
      <DialogFooter>
        <Button variant="outline" @click="open = false">取消</Button>
        <Button :disabled="creating || !keyword.trim()" @click="handleSubmit">
          <Loader2 v-if="creating" class="h-4 w-4 mr-1 animate-spin" />
          开始获取
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
