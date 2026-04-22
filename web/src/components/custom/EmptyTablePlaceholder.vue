<script setup lang="ts">
import { Database } from "lucide-vue-next";

import { TableCell, TableRow } from "@/components/ui/table";

interface Props {
  colspan: number;
  text?: string;
  heightClass?: string;
}

withDefaults(defineProps<Props>(), {
  text: "暂无数据",
  heightClass: "h-28",
});
</script>

<template>
  <TableRow>
    <TableCell :colspan="colspan" :class="[heightClass, 'text-center text-muted-foreground']">
      <div class="flex items-center justify-center gap-2">
        <Database class="h-8 w-8 text-muted-foreground/70" />
        <span class="text-sm">{{ text }}</span>
      </div>
    </TableCell>
  </TableRow>
</template>
