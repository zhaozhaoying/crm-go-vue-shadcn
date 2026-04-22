<script setup lang="ts">
import { Ripple } from '@/components/custom/ripple'
import { Sparkles } from 'lucide-vue-next'
</script>

<template>
  <div class="flex h-svh flex-col bg-background">
    <!-- Top nav -->
    <header class="bg-transparent">
      <div class="mx-auto flex max-w-7xl items-center justify-between p-6 md:p-10">
        <RouterLink to="/" class="flex items-center gap-2 font-medium">
          <img src="@/assets/logo-header.png" alt="">
        </RouterLink>
      </div>
    </header>

    <!-- Main content with ripple background -->
    <main class="relative mx-auto h-full w-full max-w-7xl grow content-center overflow-hidden">
      <div class="relative flex h-full w-full flex-col items-center justify-center">
        <div class="relative z-10 w-full max-w-sm px-4">
          <slot />
        </div>
        <Ripple
          circle-class="border-[hsl(var(--primary))] bg-[#0000]/15 dark:bg-[#fff]/20 rounded-full"
        />
      </div>
    </main>
  </div>
</template>
