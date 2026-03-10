<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { computed } from "vue"
import { Check } from "lucide-vue-next"
import { CheckboxIndicator, CheckboxRoot } from "reka-ui"
import { cn } from "@/lib/utils"

type CheckedState = boolean | 'indeterminate'

const props = defineProps<{
  defaultValue?: CheckedState
  checked?: CheckedState
  disabled?: boolean
  required?: boolean
  name?: string
  id?: string
  value?: string
  class?: HTMLAttributes["class"]
}>()

const emits = defineEmits<{
  'update:checked': [value: CheckedState]
}>()

const modelValue = computed({
  get() {
    return props.checked
  },
  set(val: CheckedState) {
    emits('update:checked', val)
  }
})
</script>

<template>
  <CheckboxRoot
    v-model="modelValue"
    :default-value="defaultValue"
    :disabled="disabled"
    :required="required"
    :name="name"
    :id="id"
    :value="value"
    :class="
      cn('grid place-content-center peer h-4 w-4 shrink-0 rounded-sm border border-primary ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground',
         props.class)"
  >
    <CheckboxIndicator class="grid place-content-center text-current">
      <slot>
        <Check class="h-4 w-4" />
      </slot>
    </CheckboxIndicator>
  </CheckboxRoot>
</template>
