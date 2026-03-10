<script setup lang="ts">
import { ref } from "vue"
import { Button } from "@/components/ui/button"
import { CircleAlert, Trash2, ShieldBan, Info } from "lucide-vue-next"

export interface ConfirmOptions {
  title: string
  description?: string
  confirmText?: string
  cancelText?: string
  variant?: "danger" | "warning" | "info"
}

const visible = ref(false)
const options = ref<ConfirmOptions>({ title: "" })
let resolveFn: ((val: boolean) => void) | null = null

const open = (opts: ConfirmOptions): Promise<boolean> => {
  options.value = opts
  visible.value = true
  return new Promise((resolve) => {
    resolveFn = resolve
  })
}

const close = (result: boolean) => {
  visible.value = false
  resolveFn?.(result)
  resolveFn = null
}

const iconMap = {
  danger: Trash2,
  warning: ShieldBan,
  info: Info,
}

const iconBgMap = {
  danger: "bg-red-50 dark:bg-red-950/40",
  warning: "bg-orange-50 dark:bg-orange-950/40",
  info: "bg-blue-50 dark:bg-blue-950/40",
}

const iconColorMap = {
  danger: "text-red-500",
  warning: "text-orange-500",
  info: "text-blue-500",
}

const btnClassMap = {
  danger: "bg-red-600 text-white hover:bg-red-700",
  warning: "",
  info: "",
}

defineExpose({ open })
</script>

<template>
  <Teleport to="body">
    <Transition name="confirm-fade">
      <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-[2px]" @click="close(false)" />
        <Transition name="confirm-scale" appear>
          <div
            v-if="visible"
            class="relative z-10 w-full max-w-sm mx-4 rounded-xl bg-background shadow-2xl border"
          >
            <div class="p-6 flex flex-col items-center text-center gap-3">
              <div
                class="flex h-12 w-12 items-center justify-center rounded-full"
                :class="iconBgMap[options.variant || 'danger']"
              >
                <component
                  :is="iconMap[options.variant || 'danger']"
                  class="h-6 w-6"
                  :class="iconColorMap[options.variant || 'danger']"
                />
              </div>
              <h3 class="text-base font-semibold">{{ options.title }}</h3>
              <p v-if="options.description" class="text-sm text-muted-foreground leading-relaxed">
                {{ options.description }}
              </p>
            </div>
            <div class="flex gap-3 px-6 pb-6">
              <Button
                variant="outline"
                class="flex-1"
                @click="close(false)"
              >
                {{ options.cancelText || '取消' }}
              </Button>
              <Button
                class="flex-1"
                :class="btnClassMap[options.variant || 'danger']"
                @click="close(true)"
              >
                {{ options.confirmText || '确认' }}
              </Button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.confirm-fade-enter-active,
.confirm-fade-leave-active {
  transition: opacity 0.2s ease;
}
.confirm-fade-enter-from,
.confirm-fade-leave-to {
  opacity: 0;
}
.confirm-scale-enter-active {
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}
.confirm-scale-leave-active {
  transition: all 0.15s ease-in;
}
.confirm-scale-enter-from {
  opacity: 0;
  transform: scale(0.95);
}
.confirm-scale-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>
