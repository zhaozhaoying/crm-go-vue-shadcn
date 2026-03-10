<script setup lang="ts">
import { ref, watch } from "vue"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Loader2, X } from "lucide-vue-next"
import { createRole, updateRole } from "@/api/modules/users"
import type { Role } from "@/types/user"

const props = defineProps<{
  open: boolean
  mode: "create" | "edit"
  roleData?: Role | null
}>()

const emit = defineEmits<{
  (e: "update:open", value: boolean): void
  (e: "success"): void
}>()

const form = ref({
  name: "",
  label: "",
  sort: 0,
})
const formError = ref("")
const formSubmitting = ref(false)

watch(() => props.open, (val) => {
  if (!val) return
  formError.value = ""
  if (props.mode === "create") {
    form.value = { name: "", label: "", sort: 0 }
  } else if (props.roleData) {
    form.value = {
      name: props.roleData.name,
      label: props.roleData.label,
      sort: props.roleData.sort,
    }
  }
})

const close = () => { emit("update:open", false) }

const handleSubmit = async () => {
  formError.value = ""
  formSubmitting.value = true
  try {
    if (!form.value.name || !form.value.label) {
      formError.value = "请填写角色标识和角色名称"
      return
    }
    if (props.mode === "create") {
      await createRole({
        name: form.value.name,
        label: form.value.label,
        sort: form.value.sort,
      })
    } else if (props.roleData) {
      await updateRole(props.roleData.id, {
        name: form.value.name,
        label: form.value.label,
        sort: form.value.sort,
      })
    }
    emit("success")
    close()
  } catch (e) {
    formError.value = e instanceof Error ? e.message : "操作失败"
  } finally {
    formSubmitting.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center">
        <!-- 遮罩层 -->
        <div class="absolute inset-0 bg-black/60" @click="close" />

        <!-- 弹窗内容 -->
        <div class="relative z-10 w-full max-w-[500px] mx-4 rounded-xl bg-white shadow-2xl">
          <!-- 关闭按钮 -->
          <button
            type="button"
            @click="close"
            class="absolute right-4 top-4 z-20 flex h-8 w-8 items-center justify-center rounded-full bg-slate-100 text-slate-500 transition hover:bg-slate-200 hover:text-slate-700"
          >
            <X class="h-4 w-4" />
          </button>

          <!-- 头部 -->
          <div class="px-6 py-5 border-b border-slate-100 bg-slate-50/50 rounded-t-xl pr-14">
            <h2 class="text-lg font-semibold text-slate-900">
              {{ mode === "create" ? "添加角色" : "编辑角色" }}
            </h2>
          </div>

          <form @submit.prevent="handleSubmit">
            <div class="p-6 space-y-5">
              <div v-if="formError" class="rounded-lg bg-red-50 p-3 text-sm text-red-600 border border-red-100">
                {{ formError }}
              </div>

              <div class="space-y-1.5">
                <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">角色标识 <span class="text-red-500">*</span></Label>
                <Input v-model="form.name" required placeholder="如: manager" class="h-10 font-mono" />
                <p class="text-xs text-slate-400">用于系统内部识别，建议使用英文</p>
              </div>

              <div class="space-y-1.5">
                <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">角色名称 <span class="text-red-500">*</span></Label>
                <Input v-model="form.label" required placeholder="如: 经理" class="h-10" />
                <p class="text-xs text-slate-400">显示给用户的名称</p>
              </div>

              <div class="space-y-1.5">
                <Label class="text-slate-700 text-xs font-semibold uppercase tracking-wider">排序</Label>
                <Input v-model.number="form.sort" type="number" placeholder="0" class="h-10" />
                <p class="text-xs text-slate-400">数字越小越靠前</p>
              </div>
            </div>

            <!-- 底部按钮 -->
            <div class="px-6 py-4 border-t border-slate-100 bg-slate-50/50 flex justify-end gap-3 rounded-b-xl">
              <Button type="button" variant="outline" @click="close" class="h-10 px-5">取消</Button>
              <Button type="submit" :disabled="formSubmitting" class="h-10 px-5 bg-indigo-600 hover:bg-indigo-700 text-white">
                <Loader2 v-if="formSubmitting" class="mr-2 h-4 w-4 animate-spin" />
                {{ formSubmitting ? "保存中..." : "保存" }}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active, .modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from, .modal-leave-to {
  opacity: 0;
}
</style>
