<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <div
    data-slot="input-group"
    :class="
      cn(
        'group/input-group flex h-10 w-full items-stretch overflow-hidden rounded-md border border-input bg-background text-sm ring-offset-background transition-[color,box-shadow] focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
