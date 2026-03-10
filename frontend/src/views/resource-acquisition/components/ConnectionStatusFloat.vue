<script setup lang="ts">
import { ref, watch } from "vue";
import { RefreshCw, Wifi, WifiOff } from "lucide-vue-next";
import { useDraggable, useStorage, useWindowSize } from "@vueuse/core";

import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";

defineProps<{
  streaming: boolean;
  streamBadgeLabel: string;
  hasActiveNonTerminalTask: boolean;
  refreshing: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const floatingBtnRef = ref<HTMLElement | null>(null);
const isDragging = ref(false);

const { width: windowWidth, height: windowHeight } = useWindowSize();
const btnSize = 48;

const floatingBtnPos = useStorage("floatingBtnPos", {
  x: typeof window !== "undefined" ? window.innerWidth - btnSize - 24 : 0,
  y: typeof window !== "undefined" ? window.innerHeight - btnSize - 24 : 0,
});

const {
  x: dragX,
  y: dragY,
  style: floatingBtnStyle,
} = useDraggable(floatingBtnRef, {
  initialValue: { x: floatingBtnPos.value.x, y: floatingBtnPos.value.y },
  onStart: () => {
    isDragging.value = false;
  },
  onMove: () => {
    isDragging.value = true;
  },
});

const handleClickCapture = (e: MouseEvent) => {
  if (isDragging.value) {
    e.preventDefault();
    e.stopPropagation();
  }
};

watch(
  [dragX, dragY, windowWidth, windowHeight],
  () => {
    const clampedX = Math.max(
      0,
      Math.min(dragX.value, windowWidth.value - btnSize),
    );
    const clampedY = Math.max(
      0,
      Math.min(dragY.value, windowHeight.value - btnSize),
    );
    if (dragX.value !== clampedX) dragX.value = clampedX;
    if (dragY.value !== clampedY) dragY.value = clampedY;
  },
  { immediate: true },
);

watch([dragX, dragY], () => {
  floatingBtnPos.value = { x: dragX.value, y: dragY.value };
});
</script>

<template>
  <div
    ref="floatingBtnRef"
    class="fixed z-50 cursor-move touch-none"
    :style="floatingBtnStyle"
  >
    <Popover>
      <PopoverTrigger as-child>
        <Button
          size="icon"
          class="h-12 w-12 rounded-full shadow-lg border-2 transition-all hover:scale-105 pointer-events-auto"
          :class="
            streaming
              ? 'border-emerald-200 bg-emerald-50 text-emerald-700 hover:bg-emerald-100'
              : 'border-zinc-200 bg-zinc-50 text-zinc-700 hover:bg-zinc-100'
          "
          @click.capture="handleClickCapture"
        >
          <Wifi v-if="streaming" class="h-5 w-5" />
          <WifiOff v-else class="h-5 w-5" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        align="end"
        side="top"
        class="w-64 p-4 text-sm animate-in zoom-in-95 duration-200"
      >
        <div class="space-y-4">
          <div class="space-y-1">
            <h4 class="font-medium text-foreground">服务连接状态</h4>
            <div class="flex items-center gap-2 mt-1">
              <span class="relative flex h-2 w-2">
                <span
                  v-if="streaming"
                  class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"
                ></span>
                <span
                  class="relative inline-flex rounded-full h-2 w-2"
                  :class="streaming ? 'bg-emerald-500' : 'bg-zinc-400'"
                ></span>
              </span>
              <p class="text-xs text-muted-foreground">
                {{ streamBadgeLabel }}
              </p>
            </div>
          </div>
          <p
            v-if="!streaming && hasActiveNonTerminalTask"
            class="text-xs text-amber-600 bg-amber-50 p-2 rounded border border-amber-200"
          >
            数据推送可能已断开，请点下方刷新以重连。
          </p>
          <Button
            variant="outline"
            size="sm"
            class="w-full bg-background mt-2"
            :disabled="refreshing"
            @click="emit('refresh')"
          >
            <RefreshCw class="w-3.5 h-3.5 mr-2" />
            重新同步
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  </div>
</template>
