<script lang="ts">
import { cva, type VariantProps } from "class-variance-authority";

export const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-lg border border-transparent bg-clip-padding text-sm font-medium transition-all focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 outline-none select-none shrink-0",
  {
    variants: {
      variant: {
        default: "bg-brand text-brand-foreground hover:bg-brand/80",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        outline:
          "border-brand bg-background text-brand hover:bg-brand hover:text-brand-foreground",
        ghost: "hover:bg-accent hover:text-accent-foreground",
      },
      size: {
        xs: "h-7 gap-1 px-2 text-xs",
        sm: "h-8 gap-1.5 px-2.5 text-xs",
        default: "h-9 gap-2 px-3",
        lg: "h-10 gap-2 px-4",
        icon: "h-9 w-9",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
);

export type ButtonVariants = VariantProps<typeof buttonVariants>;
</script>

<script setup lang="ts">
import { computed } from "vue";

import { cn } from "@/lib/utils";

const props = withDefaults(
  defineProps<{
    variant?: ButtonVariants["variant"];
    size?: ButtonVariants["size"];
    className?: string;
    type?: "button" | "submit" | "reset";
  }>(),
  {
    variant: "default",
    size: "default",
    type: "button",
  },
);

const buttonClass = computed(() =>
  cn(
    buttonVariants({
      variant: props.variant,
      size: props.size,
    }),
    props.className,
  ),
);
</script>

<template>
  <button :type="props.type" :class="buttonClass">
    <slot />
  </button>
</template>
