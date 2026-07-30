<script setup lang="ts">
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { getAssessmentAnnouncement } from "@/api/modules/systemSettings";
import { Loader2 } from "lucide-vue-next";
import { ref, watch } from "vue";

const props = defineProps<{ open: boolean }>();
const emit = defineEmits<{ "update:open": [value: boolean] }>();

const content = ref("");
const loading = ref(false);

const loadContent = async () => {
  loading.value = true;
  try {
    const data = await getAssessmentAnnouncement();
    content.value = data.assessmentAnnouncement || "";
  } catch {
    content.value = "";
  } finally {
    loading.value = false;
  }
};

watch(
  () => props.open,
  (val) => {
    if (val) {
      loadContent();
    }
  },
);

const onOpenChange = (val: boolean) => {
  emit("update:open", val);
};
</script>

<template>
  <Dialog :open="open" @update:open="onOpenChange">
    <DialogContent class="max-w-[75vw] max-h-[85vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle class="text-lg">考核公告</DialogTitle>
      </DialogHeader>
      <div class="py-4">
        <div v-if="loading" class="flex justify-center py-8">
          <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
        <div
          v-else-if="content"
          class="prose prose-sm max-w-none assessment-announcement-content"
          v-html="content"
        />
        <p v-else class="text-muted-foreground text-center py-8">
          暂无考核公告
        </p>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.assessment-announcement-content :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 1em 0;
}
.assessment-announcement-content :deep(th),
.assessment-announcement-content :deep(td) {
  border: 1px solid hsl(var(--border));
  padding: 0.5rem 0.75rem;
  text-align: left;
}
.assessment-announcement-content :deep(th) {
  background: hsl(var(--muted));
  font-weight: 600;
}
.assessment-announcement-content :deep(p[align="right"]) {
  text-align: right;
}
</style>
