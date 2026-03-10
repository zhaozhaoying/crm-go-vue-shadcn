<script setup lang="ts">
import { ref } from "vue";
import { Check, Copy } from "lucide-vue-next";
import { toast } from "vue-sonner";

const props = withDefaults(
  defineProps<{
    text: string;
    size?: "sm" | "xs";
    successMessage?: string;
  }>(),
  {
    size: "xs",
    successMessage: "已复制",
  },
);

const copied = ref(false);

const handleCopy = async (e: MouseEvent) => {
  e.stopPropagation();
  if (!props.text || copied.value) return;
  try {
    await navigator.clipboard.writeText(props.text);
    copied.value = true;
    toast.success(props.successMessage);
    setTimeout(() => {
      copied.value = false;
    }, 1500);
  } catch {
    toast.error("复制失败，请手动复制");
  }
};
</script>

<template>
  <button
    type="button"
    :title="copied ? '已复制' : '点击复制'"
    :class="[
      'inline-flex items-center justify-center rounded transition-all',
      'text-muted-foreground opacity-0 group-hover:opacity-100',
      'hover:bg-muted hover:text-foreground focus-visible:opacity-100',
      size === 'xs' ? 'h-5 w-5' : 'h-6 w-6',
    ]"
    @click="handleCopy"
  >
    <Check
      v-if="copied"
      :class="['text-emerald-500', size === 'xs' ? 'h-3 w-3' : 'h-3.5 w-3.5']"
    />
    <Copy v-else :class="[size === 'xs' ? 'h-3 w-3' : 'h-3.5 w-3.5']" />
  </button>
</template>
