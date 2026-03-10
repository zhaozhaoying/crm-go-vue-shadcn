<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from "vue"
import { Download, Loader2, Plus, Trash2, ZoomIn } from "lucide-vue-next"

import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"

interface Props {
  modelValue: string
  disabled?: boolean
  accept?: string
  maxSizeMB?: number
  placeholder?: string
  uploadText?: string
  onUpload: (file: File) => Promise<{ url: string } | string>
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  accept: "image/png,image/jpeg,image/webp",
  maxSizeMB: 2,
  placeholder: "暂无图片",
  uploadText: "上传图片",
})

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void
  (e: "error", message: string): void
}>()

const fileInput = ref<HTMLInputElement | null>(null)
const previewUrl = ref("")
const localPreviewObjectUrl = ref("")
const dialogVisible = ref(false)
const uploading = ref(false)

const clearLocalPreviewObjectUrl = () => {
  if (localPreviewObjectUrl.value) {
    URL.revokeObjectURL(localPreviewObjectUrl.value)
    localPreviewObjectUrl.value = ""
  }
}

watch(
  () => props.modelValue,
  (value) => {
    if (!localPreviewObjectUrl.value) {
      previewUrl.value = value || ""
    }
  },
  { immediate: true },
)

const triggerUpload = () => {
  if (props.disabled || uploading.value) return
  fileInput.value?.click()
}

const handleRemove = () => {
  if (props.disabled || uploading.value) return
  clearLocalPreviewObjectUrl()
  previewUrl.value = ""
  emit("update:modelValue", "")
}

const handleDownload = () => {
  if (!previewUrl.value) return
  window.open(previewUrl.value, "_blank", "noopener,noreferrer")
}

const handlePreview = () => {
  if (!previewUrl.value) return
  dialogVisible.value = true
}

const onFileSelected = async (event: Event) => {
  const target = event.target as HTMLInputElement
  if (!target.files?.length) return
  const file = target.files[0]
  target.value = ""

  if (!file.type.startsWith("image/")) {
    emit("error", "请上传图片格式的文件")
    return
  }
  if (file.size > props.maxSizeMB * 1024 * 1024) {
    emit("error", `图片大小不能超过 ${props.maxSizeMB}MB`)
    return
  }

  clearLocalPreviewObjectUrl()
  localPreviewObjectUrl.value = URL.createObjectURL(file)
  previewUrl.value = localPreviewObjectUrl.value

  uploading.value = true
  try {
    const uploadResult = await props.onUpload(file)
    const url = typeof uploadResult === "string" ? uploadResult : uploadResult.url
    clearLocalPreviewObjectUrl()
    previewUrl.value = url || ""
    emit("update:modelValue", url || "")
  } catch (error) {
    clearLocalPreviewObjectUrl()
    previewUrl.value = props.modelValue || ""
    emit("error", error instanceof Error ? error.message : "图片上传失败")
  } finally {
    uploading.value = false
  }
}

onBeforeUnmount(() => {
  clearLocalPreviewObjectUrl()
})
</script>

<template>
  <div class="space-y-2 w-[146px]">
    <input ref="fileInput" class="hidden" type="file" :accept="accept" @change="onFileSelected" />

    <div class="h-[146px] w-[146px] rounded-md border">
      <div class="group relative flex h-full w-full items-center justify-center overflow-hidden rounded-md"
        :class="!disabled ? 'cursor-pointer' : ''" @click="triggerUpload">
        <img v-if="previewUrl" :src="previewUrl" alt="图片预览" class="h-full w-full object-cover" />

        <div v-else class="flex flex-col items-center gap-2 text-muted-foreground">
          <Loader2 v-if="uploading" class="h-4 w-4 animate-spin" />
          <Plus v-else class="h-4 w-4" />
        </div>

        <div v-if="previewUrl"
          class="absolute inset-0 flex items-center justify-center gap-2 bg-black/45 opacity-0 transition-opacity group-hover:opacity-100">
          <ZoomIn class="h-6 w-6 text-white" @click.stop="handlePreview" />
          <Download class="h-6 w-6 text-white" @click.stop="handleDownload" />
          <Trash2 class="h-6 w-6 text-white" @click.stop="handleRemove" />
        </div>
      </div>
    </div>

    <Dialog v-model:open="dialogVisible">
      <DialogContent class="sm:max-w-[720px]">
        <DialogHeader>
          <DialogTitle>图片预览</DialogTitle>
        </DialogHeader>
        <div class="flex items-center justify-center">
          <img :src="previewUrl" alt="预览" class="max-h-[70vh] w-auto max-w-full object-contain" />
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
