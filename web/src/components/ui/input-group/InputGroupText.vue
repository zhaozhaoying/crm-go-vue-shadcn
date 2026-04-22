<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <span
    data-slot="input-group-text"
    :class="cn('text-sm text-muted-foreground whitespace-nowrap', props.class)"
  >
    <slot />
  </span>
</template>
