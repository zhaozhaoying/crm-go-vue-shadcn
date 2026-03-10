<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<{
  class?: HTMLAttributes["class"]
}>()
</script>

<template>
  <div
    data-slot="input-group-addon"
    :class="
      cn(
        'flex shrink-0 items-center bg-muted text-muted-foreground first:border-r last:border-l border-input',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
