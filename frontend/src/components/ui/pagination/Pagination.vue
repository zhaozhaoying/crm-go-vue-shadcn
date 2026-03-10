<script setup lang="ts">
import { computed } from "vue"
import type { AcceptableValue } from "reka-ui"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from "lucide-vue-next"

interface Props {
  currentPage: number
  totalPages: number
  pageSize?: number
  pageSizeOptions?: number[]
  selectedCount?: number
  totalCount?: number
  showPageSize?: boolean
  showSelection?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  pageSize: 10,
  pageSizeOptions: () => [10, 20, 30, 50],
  showPageSize: true,
  showSelection: true,
})

const emit = defineEmits<{
  (e: "update:currentPage", value: number): void
  (e: "update:pageSize", value: number): void
}>()

const isFirstPage = computed(() => props.currentPage === 0)
const isLastPage = computed(() => props.currentPage >= props.totalPages - 1)
const hasSelectionSummary = computed(
  () => props.showSelection && props.selectedCount !== undefined && props.totalCount !== undefined
)

const goToFirstPage = () => {
  emit("update:currentPage", 0)
}

const goToPreviousPage = () => {
  if (!isFirstPage.value) {
    emit("update:currentPage", props.currentPage - 1)
  }
}

const goToNextPage = () => {
  if (!isLastPage.value) {
    emit("update:currentPage", props.currentPage + 1)
  }
}

const goToLastPage = () => {
  emit("update:currentPage", props.totalPages - 1)
}

const updatePageSize = (value: AcceptableValue) => {
  if (value === null || value === undefined) return
  emit("update:pageSize", Number(value))
  emit("update:currentPage", 0)
}
</script>

<template>
  <div class="flex items-center px-2" :class="hasSelectionSummary ? 'justify-between' : 'justify-end'">
    <div v-if="hasSelectionSummary" class="hidden flex-1 text-sm text-muted-foreground lg:flex">
      已选 {{ selectedCount }} / {{ totalCount }} 行
    </div>
    <div
      class="flex items-center gap-6"
      :class="hasSelectionSummary ? 'w-full lg:w-fit' : 'w-full justify-end'"
    >
      <div v-if="showPageSize && pageSize" class="hidden items-center gap-2 lg:flex">
        <Label class="text-sm font-medium whitespace-nowrap">每页行数</Label>
        <Select :model-value="String(pageSize)" @update:model-value="updatePageSize">
          <SelectTrigger class="w-18 h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent side="top">
            <SelectItem v-for="s in pageSizeOptions" :key="s" :value="String(s)">{{ s }}</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="flex w-fit items-center justify-center text-sm font-medium whitespace-nowrap">
        第 {{ currentPage + 1 }} / {{ totalPages }} 页
      </div>
      <div class="flex items-center gap-2" :class="hasSelectionSummary ? 'ml-auto lg:ml-0' : ''">
        <Button variant="outline" size="icon" class="hidden h-8 w-8 lg:flex" :disabled="isFirstPage" @click="goToFirstPage">
          <ChevronsLeft class="h-4 w-4" />
        </Button>
        <Button variant="outline" size="icon" class="h-8 w-8" :disabled="isFirstPage" @click="goToPreviousPage">
          <ChevronLeft class="h-4 w-4" />
        </Button>
        <Button variant="outline" size="icon" class="h-8 w-8" :disabled="isLastPage" @click="goToNextPage">
          <ChevronRight class="h-4 w-4" />
        </Button>
        <Button variant="outline" size="icon" class="hidden h-8 w-8 lg:flex" :disabled="isLastPage" @click="goToLastPage">
          <ChevronsRight class="h-4 w-4" />
        </Button>
      </div>
    </div>
  </div>
</template>
