<script setup lang="ts">
import { computed, nextTick, ref } from "vue";
import { CalendarDate, getLocalTimeZone } from "@internationalized/date";
import { CalendarIcon, X } from "lucide-vue-next";
import { toDate } from "reka-ui/date";
import { Calendar } from "@/components/ui/calendar";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";

interface Props {
  modelValue?: string;
  placeholder?: string;
  disabled?: boolean;
  id?: string;
  contentAlign?: "start" | "center" | "end";
  dateOnly?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: "请选择日期时间",
  disabled: false,
  id: undefined,
  contentAlign: "start",
  dateOnly: false,
});

const emit = defineEmits<{
  (e: "update:modelValue", value: string | undefined): void;
  (e: "change", value: string | undefined): void;
}>();

const open = ref(false);
const tempDate = ref<any>(undefined);
const tempHours = ref(0);
const tempMinutes = ref(0);
const tempSeconds = ref(0);

const hourColumnRef = ref<HTMLElement | null>(null);
const minuteColumnRef = ref<HTMLElement | null>(null);
const secondColumnRef = ref<HTMLElement | null>(null);

const hourOptions = Array.from({ length: 24 }, (_, i) => i);
const minuteSecondOptions = Array.from({ length: 60 }, (_, i) => i);

const pad2 = (value: number) => String(value).padStart(2, "0");

const parseDateTime = (dateStr?: string) => {
  if (!dateStr) return undefined;
  const normalized = dateStr.includes("T")
    ? dateStr
    : dateStr.replace(" ", "T");
  const date = new Date(normalized);
  if (Number.isNaN(date.getTime())) return undefined;
  return date;
};

const formatDateTime = (date: Date): string => {
  const year = date.getFullYear();
  const month = pad2(date.getMonth() + 1);
  const day = pad2(date.getDate());
  const hours = pad2(date.getHours());
  const minutes = pad2(date.getMinutes());
  const seconds = pad2(date.getSeconds());
  return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}`;
};

const formatDateOnly = (date: Date): string => {
  const year = date.getFullYear();
  const month = pad2(date.getMonth() + 1);
  const day = pad2(date.getDate());
  return `${year}-${month}-${day}`;
};

const toCalendarDate = (date: Date) => {
  return new CalendarDate(date.getFullYear(), date.getMonth() + 1, date.getDate());
};

const selectedDate = computed(() => parseDateTime(props.modelValue));

const syncTempFromModel = () => {
  const base = selectedDate.value ?? new Date();
  tempDate.value = toCalendarDate(base);
  tempHours.value = base.getHours();
  tempMinutes.value = base.getMinutes();
  tempSeconds.value = base.getSeconds();
};

const scrollColumnTo = (
  column: HTMLElement | null,
  value: number,
  behavior: ScrollBehavior,
) => {
  if (!column) return;
  const target = column.querySelector<HTMLElement>(`button[data-value="${value}"]`);
  if (!target) return;
  const top = target.offsetTop - column.clientHeight / 2 + target.offsetHeight / 2;
  column.scrollTo({ top: Math.max(0, top), behavior });
};

const scrollToSelected = (behavior: ScrollBehavior = "auto") => {
  scrollColumnTo(hourColumnRef.value, tempHours.value, behavior);
  scrollColumnTo(minuteColumnRef.value, tempMinutes.value, behavior);
  scrollColumnTo(secondColumnRef.value, tempSeconds.value, behavior);
};

const handleOpenChange = (isOpen: boolean) => {
  open.value = isOpen;
  if (isOpen) {
    syncTempFromModel();
    nextTick(() => scrollToSelected("auto"));
  }
};

const handleDateSelect = (date: any) => {
  tempDate.value = date;
};

const handleNow = () => {
  const now = new Date();
  tempDate.value = toCalendarDate(now);
  tempHours.value = now.getHours();
  tempMinutes.value = now.getMinutes();
  tempSeconds.value = now.getSeconds();
  nextTick(() => scrollToSelected("smooth"));
};

const handleConfirm = () => {
  if (!tempDate.value) return;
  const pickedDate = toDate(tempDate.value as any, getLocalTimeZone());
  const next = new Date(
    pickedDate.getFullYear(),
    pickedDate.getMonth(),
    pickedDate.getDate(),
    tempHours.value,
    tempMinutes.value,
    tempSeconds.value,
  );
  const value = props.dateOnly ? formatDateOnly(next) : formatDateTime(next);
  emit("update:modelValue", value);
  emit("change", value);
  open.value = false;
};

const handleClear = (e: MouseEvent) => {
  e.stopPropagation();
  emit("update:modelValue", undefined);
  emit("change", undefined);
};

const displayText = computed(() => {
  const date = parseDateTime(props.modelValue);
  if (!date) return props.placeholder;
  if (props.dateOnly) {
    return formatDateOnly(date);
  }
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())} ${pad2(date.getHours())}:${pad2(date.getMinutes())}:${pad2(date.getSeconds())}`;
});
</script>

<template>
  <Popover :open="open" @update:open="handleOpenChange">
    <PopoverTrigger as-child>
      <Button
        :id="id"
        variant="outline"
        :class="cn('w-full justify-start text-left font-normal border-input hover:bg-accent hover:text-accent-foreground', !modelValue && 'text-muted-foreground')"
        :disabled="disabled"
      >
        <CalendarIcon class="mr-2 h-4 w-4" />
        <span class="flex-1">{{ displayText }}</span>
        <X
          v-if="modelValue && !disabled"
          class="h-4 w-4 opacity-50 hover:opacity-100"
          @click="handleClear"
        />
      </Button>
    </PopoverTrigger>

    <PopoverContent
      class="w-auto max-w-[calc(100vw-2rem)] overflow-hidden p-0"
      :align="contentAlign"
      :collision-padding="16"
      :side-offset="6"
    >
      <div class="bg-background">
        <div class="flex flex-col md:flex-row">
          <div class="border-b p-3 md:border-b-0 md:border-r">
            <Calendar
              mode="single"
              :model-value="tempDate"
              @update:model-value="handleDateSelect"
              initial-focus
            />
          </div>

          <div v-if="!dateOnly" class="w-full md:w-[248px]">
            <div class="flex h-11 items-center justify-center border-b text-base font-semibold">
              选择时间
            </div>

            <div class="p-3">
              <div class="mb-2 grid grid-cols-3 gap-2 text-center text-xs text-muted-foreground">
                <span>时</span>
                <span>分</span>
                <span>秒</span>
              </div>

              <div class="grid grid-cols-3 gap-2">
                <div
                  ref="hourColumnRef"
                  class="h-64 overflow-y-auto rounded-md border bg-background py-1"
                >
                  <button
                    v-for="hour in hourOptions"
                    :key="hour"
                    type="button"
                    :data-value="hour"
                    class="flex h-8 w-full items-center justify-center rounded-sm text-sm tabular-nums transition-colors"
                    :class="tempHours === hour ? 'bg-brand text-brand-foreground font-medium' : 'hover:bg-muted'"
                    @click="tempHours = hour"
                  >
                    {{ pad2(hour) }}
                  </button>
                </div>

                <div
                  ref="minuteColumnRef"
                  class="h-64 overflow-y-auto rounded-md border bg-background py-1"
                >
                  <button
                    v-for="minute in minuteSecondOptions"
                    :key="minute"
                    type="button"
                    :data-value="minute"
                    class="flex h-8 w-full items-center justify-center rounded-sm text-sm tabular-nums transition-colors"
                    :class="tempMinutes === minute ? 'bg-brand text-brand-foreground font-medium' : 'hover:bg-muted'"
                    @click="tempMinutes = minute"
                  >
                    {{ pad2(minute) }}
                  </button>
                </div>

                <div
                  ref="secondColumnRef"
                  class="h-64 overflow-y-auto rounded-md border bg-background py-1"
                >
                  <button
                    v-for="second in minuteSecondOptions"
                    :key="second"
                    type="button"
                    :data-value="second"
                    class="flex h-8 w-full items-center justify-center rounded-sm text-sm tabular-nums transition-colors"
                    :class="tempSeconds === second ? 'bg-brand text-brand-foreground font-medium' : 'hover:bg-muted'"
                    @click="tempSeconds = second"
                  >
                    {{ pad2(second) }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-between border-t px-3 py-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            :disabled="disabled"
            @click="handleNow"
          >
            此刻
          </Button>
          <Button
            type="button"
            size="sm"
            :disabled="disabled || !tempDate"
            @click="handleConfirm"
          >
            确认
          </Button>
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>
