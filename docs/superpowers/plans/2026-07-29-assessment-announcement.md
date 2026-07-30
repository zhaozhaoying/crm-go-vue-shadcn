# 考核公告功能 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 CRM 系统中新增考核公告功能，所有用户可通过顶部栏图标查看，管理员可在系统设置中编辑富文本内容。

**Architecture:** 考核公告内容以 HTML 字符串存储在 `system_settings` 表（key: `assessment_announcement`）。后端复用现有 settings API 读写，新增独立公开 GET 接口。前端使用 TipTap 富文本编辑器供管理员编辑，弹窗组件供所有用户查看。

**Tech Stack:** Go/Gin + Vue 3 + TipTap + shadcn Dialog

## Global Constraints

- 所有 API 使用 envelope 格式: `{ code: 0, message: "ok", data: ... }`
- 前端 API 调用使用 `request<T>()` 封装
- 考核公告内容为 HTML 字符串
- 仅管理员可编辑，所有登录用户可查看
- 顶部栏公告按钮紧邻通知铃铛

---

### Task 1: Backend — 模型 & 设置读写

**Files:**
- Modify: `internal/model/system_setting.go`
- Modify: `internal/service/system_setting_service.go`

**Interfaces:**
- Consumes: (none — first task)
- Produces: `SystemSettingsResponse.AssessmentAnnouncement string`, `UpdateSystemSettingsRequest.AssessmentAnnouncement *string`

- [ ] **Step 1: 在 SystemSettingsResponse 中新增字段**

```go
// internal/model/system_setting.go
// 在 SystemSettingsResponse struct 末尾新增：

type SystemSettingsResponse struct {
	// ... 已有字段保持不变 ...
	CustomerLevels          []CustomerLevel  `json:"customerLevels" gorm:"-"`
	CustomerSources         []CustomerSource `json:"customerSources" gorm:"-"`
	AssessmentAnnouncement  string           `json:"assessmentAnnouncement" gorm:"-"`  // 新增
}
```

- [ ] **Step 2: 在 UpdateSystemSettingsRequest 中新增字段**

```go
// internal/model/system_setting.go
// 在 UpdateSystemSettingsRequest struct 末尾新增：

type UpdateSystemSettingsRequest struct {
	// ... 已有字段保持不变 ...
	VisitPurposes           []string `json:"visitPurposes" gorm:"-"`
	AssessmentAnnouncement  *string  `json:"assessmentAnnouncement" gorm:"-"`  // 新增
}
```

- [ ] **Step 3: 在 GetAllSettings 中读取考核公告**

```go
// internal/service/system_setting_service.go
// 在 GetAllSettings() 方法中，sources 读取之后、return 之前新增一行：

assessmentAnnouncement := s.getStringSetting("assessment_announcement", "")
```

- [ ] **Step 4: 在 GetAllSettings 的 return 中包含新字段**

```go
// 在 return 语句的 &model.SystemSettingsResponse{...} 中新增：
return &model.SystemSettingsResponse{
	// ... 已有字段保持不变 ...
	CustomerSources:         sources,
	AssessmentAnnouncement:  assessmentAnnouncement,  // 新增
}, nil
```

- [ ] **Step 5: 在 UpdateSettings 中处理考核公告写入**

```go
// internal/service/system_setting_service.go
// 在 UpdateSettings() 方法末尾，return nil 之前新增：

if req.AssessmentAnnouncement != nil {
	if err := s.repo.UpsertSetting(
		"assessment_announcement",
		*req.AssessmentAnnouncement,
		"考核公告内容",
	); err != nil {
		return err
	}
}
```

- [ ] **Step 6: 编译验证**

Run: `cd /Users/yanmengdi/web/other/crm-go-vue-shadcn && go build ./...`
Expected: 编译成功，无报错

---

### Task 2: Backend — 公开接口与新路由

**Files:**
- Modify: `internal/handler/system_setting_handler.go`
- Modify: `internal/router/router.go`

**Interfaces:**
- Consumes: `SystemSettingsResponse.AssessmentAnnouncement` (from Task 1)
- Produces: `GET /api/v1/announcement` → `{ assessmentAnnouncement: string }`

- [ ] **Step 1: 在 handler 中新增公开获取方法**

```go
// internal/handler/system_setting_handler.go
// 在文件末尾，最后一个函数之后新增：

// GetAssessmentAnnouncement godoc
// @Summary 获取考核公告
// @Tags 考核公告
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/announcement [get]
func (h *SystemSettingHandler) GetAssessmentAnnouncement(c *gin.Context) {
	settings, err := h.service.GetAllSettings()
	if err != nil {
		ErrorWithDetail(c, 500, 10120, "加载考核公告失败", err)
		return
	}
	Success(c, gin.H{"assessmentAnnouncement": settings.AssessmentAnnouncement})
}
```

需要在文件顶部 import 中添加 `"github.com/gin-gonic/gin"`（应该已有）。

- [ ] **Step 2: 在 router 中注册新路由**

```go
// internal/router/router.go
// 在 protected 路由组内，合适位置（如 /auth/me 附近）新增：

protected.GET("/announcement", systemSettingHandler.GetAssessmentAnnouncement)
```

建议放在 `protected.GET("/auth/me", ...)` 这行附近。

- [ ] **Step 3: 编译验证**

Run: `cd /Users/yanmengdi/web/other/crm-go-vue-shadcn && go build ./...`
Expected: 编译成功，无报错

---

### Task 3: Frontend — 安装 TipTap 依赖

**Files:**
- Modify: `web/package.json` (通过 pnpm add)

- [ ] **Step 1: 安装 TipTap**

```bash
cd /Users/yanmengdi/web/other/crm-go-vue-shadcn/web && pnpm add @tiptap/vue-3 @tiptap/starter-kit
```

Expected: 安装成功，package.json 和 pnpm-lock.yaml 更新

---

### Task 4: Frontend — API 层更新

**Files:**
- Modify: `web/src/api/modules/systemSettings.ts`

- [ ] **Step 1: 在 SystemSettings 接口中新增字段**

```typescript
// 在 SystemSettings 接口末尾新增：
export interface SystemSettings {
  // ... 已有字段保持不变 ...
  customerSources: CustomerSource[];
  assessmentAnnouncement: string;  // 新增
}
```

- [ ] **Step 2: 在 UpdateSystemSettingsRequest 接口中新增字段**

```typescript
// 在 UpdateSystemSettingsRequest 接口末尾新增：
export interface UpdateSystemSettingsRequest {
  // ... 已有字段保持不变 ...
  visitPurposes?: string[];
  assessmentAnnouncement?: string;  // 新增
}
```

- [ ] **Step 3: 新增 getAssessmentAnnouncement API 函数**

```typescript
// 在文件末尾新增：
export const getAssessmentAnnouncement = () => {
  return request<{ assessmentAnnouncement: string }>({
    method: "GET",
    url: "/v1/announcement",
  });
};
```

---

### Task 5: Frontend — 考核公告弹窗组件

**Files:**
- Create: `web/src/components/custom/AssessmentAnnouncementModal.vue`

- [ ] **Step 1: 创建弹窗组件**

```vue
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
    <DialogContent class="max-w-2xl max-h-[80vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle class="text-lg">考核公告</DialogTitle>
      </DialogHeader>
      <div class="py-4">
        <div v-if="loading" class="flex justify-center py-8">
          <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
        <div
          v-else-if="content"
          class="prose prose-sm max-w-none"
          v-html="content"
        />
        <p v-else class="text-muted-foreground text-center py-8">
          暂无考核公告
        </p>
      </div>
    </DialogContent>
  </Dialog>
</template>
```

---

### Task 6: Frontend — 顶部栏新增公告图标按钮

**Files:**
- Modify: `web/src/layouts/AuthenticatedLayout.vue`

- [ ] **Step 1: 修改 AuthenticatedLayout.vue**

需要在顶部栏 header 区域新增公告图标按钮。当前 header 结构是：

```vue
<header class="flex h-16 shrink-0 items-center gap-2 ...">
  <div class="flex items-center gap-2 px-4 flex-1">
    <SidebarTrigger class="-ml-1 h-4 w-4" />
    <Separator orientation="vertical" class="mr-2 h-4" />
    <AppBreadcrumb />
  </div>
</header>
```

修改为在 header 右侧新增按钮：

```vue
<script setup lang="ts">
// 在现有 imports 基础上新增：
import { ref } from "vue";
import { Megaphone } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import AssessmentAnnouncementModal from "@/components/custom/AssessmentAnnouncementModal.vue";

// 新增状态：
const announcementOpen = ref(false);
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    <SidebarInset class="min-w-0 overflow-x-hidden">
      <header
        class="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12"
      >
        <div class="flex items-center gap-2 px-4 flex-1">
          <SidebarTrigger class="-ml-1 h-4 w-4" />
          <Separator orientation="vertical" class="mr-2 h-4" />
          <AppBreadcrumb />
        </div>
        <!-- 新增：公告图标按钮 -->
        <div class="flex items-center gap-1 px-4">
          <Button
            variant="ghost"
            size="icon"
            class="h-8 w-8"
            @click="announcementOpen = true"
          >
            <Megaphone class="h-4 w-4" />
          </Button>
        </div>
      </header>
      <div class="flex min-w-0 flex-1 flex-col gap-4 p-4 pt-0">
        <RouterView v-slot="{ Component, route }">
          <component :is="Component" :key="route.fullPath" />
        </RouterView>
      </div>
    </SidebarInset>
    <AppCommand />
    <!-- 新增：公告弹窗 -->
    <AssessmentAnnouncementModal v-model:open="announcementOpen" />
  </SidebarProvider>
</template>
```

---

### Task 7: Frontend — 系统设置页新增富文本编辑器

**Files:**
- Modify: `web/src/views/settings/index.vue`

- [ ] **Step 1: 在 script 中新增 TipTap 编辑器相关导入和逻辑**

在 `<script setup>` 顶部新增导入：

```typescript
import { useEditor, EditorContent } from "@tiptap/vue-3";
import StarterKit from "@tiptap/starter-kit";
import {
  Bold,
  Italic,
  Strikethrough,
  List,
  ListOrdered,
  Heading2,
  Undo,
  Redo,
} from "lucide-vue-next";
```

新增编辑器状态和保存逻辑（在现有 ref 声明附近）：

```typescript
const savingAnnouncement = ref(false);

const announcementEditor = useEditor({
  content: "",
  extensions: [StarterKit],
  editorProps: {
    attributes: {
      class:
        "prose prose-sm max-w-none min-h-[200px] px-4 py-3 border rounded-md focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
    },
  },
});

const loadAnnouncementToEditor = () => {
  if (announcementEditor.value) {
    announcementEditor.value.commands.setContent(
      settings.value.assessmentAnnouncement || "",
    );
  }
};
```

在 `loadSettings` 函数中，`settings.value = { ... }` 之后新增加载编辑器内容：

```typescript
// 在 loadSettings 函数内，settings.value = { ... } 之后新增：
loadAnnouncementToEditor();
```

新增保存考核公告的方法：

```typescript
const saveAnnouncement = async () => {
  if (!announcementEditor.value) return;
  const html = announcementEditor.value.getHTML();

  savingAnnouncement.value = true;
  try {
    await updateSystemSettings({
      assessmentAnnouncement: html,
    });
    settings.value.assessmentAnnouncement = html;
    toast.success("考核公告已保存");
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "保存考核公告失败"));
  } finally {
    savingAnnouncement.value = false;
  }
};
```

- [ ] **Step 2: 在 template 中新增考核公告编辑卡片**

在设置页 template 中，建议放在"客户来源"卡片和"跟进方式"卡片之间（或作为新的独立卡片在右侧），新增：

```vue
<!-- 考核公告编辑 -->
<Card>
  <CardHeader class="pb-4">
    <div class="flex items-center justify-between">
      <div>
        <CardTitle class="text-base">考核公告</CardTitle>
        <CardDescription class="text-xs mt-1">
          设置考核公告内容，所有用户可在顶部栏查看
        </CardDescription>
      </div>
      <Button
        @click="saveAnnouncement"
        :disabled="savingAnnouncement"
        size="sm"
        class="gap-1.5"
      >
        <Loader2
          v-if="savingAnnouncement"
          class="h-3.5 w-3.5 animate-spin"
        />
        <Save v-else class="h-3.5 w-3.5" />
        保存
      </Button>
    </div>
  </CardHeader>
  <CardContent class="space-y-3">
    <!-- 富文本工具栏 -->
    <div
      v-if="announcementEditor"
      class="flex flex-wrap items-center gap-1 rounded-md border p-1"
    >
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        :class="{
          'bg-accent text-accent-foreground':
            announcementEditor.isActive('bold'),
        }"
        @click="announcementEditor.chain().focus().toggleBold().run()"
      >
        <Bold class="h-3.5 w-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        :class="{
          'bg-accent text-accent-foreground':
            announcementEditor.isActive('italic'),
        }"
        @click="announcementEditor.chain().focus().toggleItalic().run()"
      >
        <Italic class="h-3.5 w-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        :class="{
          'bg-accent text-accent-foreground':
            announcementEditor.isActive('strike'),
        }"
        @click="announcementEditor.chain().focus().toggleStrike().run()"
      >
        <Strikethrough class="h-3.5 w-3.5" />
      </Button>
      <div class="w-px h-5 bg-border mx-0.5" />
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        :class="{
          'bg-accent text-accent-foreground': announcementEditor.isActive(
            'heading',
            { level: 2 },
          ),
        }"
        @click="
          announcementEditor.chain().focus().toggleHeading({ level: 2 }).run()
        "
      >
        <Heading2 class="h-3.5 w-3.5" />
      </Button>
      <div class="w-px h-5 bg-border mx-0.5" />
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        :class="{
          'bg-accent text-accent-foreground':
            announcementEditor.isActive('bulletList'),
        }"
        @click="announcementEditor.chain().focus().toggleBulletList().run()"
      >
        <List class="h-3.5 w-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        :class="{
          'bg-accent text-accent-foreground':
            announcementEditor.isActive('orderedList'),
        }"
        @click="announcementEditor.chain().focus().toggleOrderedList().run()"
      >
        <ListOrdered class="h-3.5 w-3.5" />
      </Button>
      <div class="w-px h-5 bg-border mx-0.5" />
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        @click="announcementEditor.chain().focus().undo().run()"
      >
        <Undo class="h-3.5 w-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        @click="announcementEditor.chain().focus().redo().run()"
      >
        <Redo class="h-3.5 w-3.5" />
      </Button>
    </div>
    <!-- 编辑器内容区 -->
    <EditorContent :editor="announcementEditor" />
  </CardContent>
</Card>
```

建议将此卡片放在右侧列的 `</div>` 之前（即跟进方式卡片之后）。

- [ ] **Step 3: 将 assessmentAnnouncement 加入默认值**

在 `settings` ref 的默认值中新增：

```typescript
const settings = ref<SystemSettings>({
  // ... 已有字段 ...
  customerSources: [],
  assessmentAnnouncement: "", // 新增
});
```

---

### Task 8: 验证

- [ ] **Step 1: 编译 Go 后端**

Run: `cd /Users/yanmengdi/web/other/crm-go-vue-shadcn && go build ./...`
Expected: 编译成功

- [ ] **Step 2: TypeCheck 前端**

Run: `cd /Users/yanmengdi/web/other/crm-go-vue-shadcn/web && pnpm typecheck`
Expected: 类型检查通过

- [ ] **Step 3: 启动后端验证 API**

Run: `cd /Users/yanmengdi/web/other/crm-go-vue-shadcn && go run main.go`

用 curl 测试：
```bash
# 需要先登录获取 token
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/announcement
```
Expected: 返回 `{"code":0,"message":"ok","data":{"assessmentAnnouncement":""}}`

- [ ] **Step 4: 启动前端验证 UI**

Run: `cd /Users/yanmengdi/web/other/crm-go-vue-shadcn/web && pnpm dev`

检查点：
- 顶部栏出现 📢 图标按钮
- 点击弹出考核公告弹窗
- 管理员进入系统设置页 → 能看到"考核公告"编辑卡片
- 富文本编辑器可正常使用（加粗、斜体、列表等）
- 保存后，其他用户点击顶部栏图标可看到更新后的内容

---

### Task 9: 提交

- [ ] **Step 1: 提交所有变更**

```bash
git add internal/model/system_setting.go \
        internal/service/system_setting_service.go \
        internal/handler/system_setting_handler.go \
        internal/router/router.go \
        web/src/api/modules/systemSettings.ts \
        web/src/components/custom/AssessmentAnnouncementModal.vue \
        web/src/layouts/AuthenticatedLayout.vue \
        web/src/views/settings/index.vue \
        web/package.json \
        web/pnpm-lock.yaml \
        docs/superpowers/plans/2026-07-29-assessment-announcement.md \
        docs/superpowers/specs/2026-07-29-assessment-announcement-design.md

git commit -m "feat: 新增考核公告功能

- 后端: system_settings 表存储考核公告内容(HTML)
- 后端: 新增 GET /api/v1/announcement 公开接口
- 前端: 顶部栏新增公告图标按钮(Megaphone)
- 前端: 新增 AssessmentAnnouncementModal 弹窗组件
- 前端: 系统设置页新增 TipTap 富文本编辑器
- 仅管理员可编辑，所有登录用户可查看"
```
