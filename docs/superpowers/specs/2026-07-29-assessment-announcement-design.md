# 考核公告功能 — 设计文档

## 概述

在 CRM 系统中新增"考核公告"功能：所有用户可通过顶部栏图标查看公告内容（弹窗展示），管理员可在系统设置中编辑公告内容（富文本）。

## 需求

- 所有登录用户可查看考核公告
- 点击顶部栏的公告图标按钮，弹窗展示公告内容
- 公告内容支持富文本格式（加粗、列表、表格等）
- 公告内容可自定义编辑
- 内容存储在 `system_settings` 数据库表中
- 仅管理员可编辑

## 架构

### 数据存储

使用现有的 `system_settings` 表，新增一条记录：

| key                        | value          | description  |
| -------------------------- | -------------- | ------------ |
| `assessment_announcement`  | `<html 字符串>` | 考核公告内容 |

### API

| 方法 | 路径                     | 权限     | 说明              |
| ---- | ------------------------ | -------- | ----------------- |
| GET  | `/api/v1/announcement`   | 登录用户 | 获取考核公告内容  |
| GET  | `/api/v1/settings`       | 登录用户 | 已有，新增返回字段 |
| PUT  | `/api/v1/settings`       | 管理员   | 已有，新增更新字段 |

### 前端组件树

```
AuthenticatedLayout.vue (顶部栏新增图标按钮)
  └── AssessmentAnnouncementModal.vue (弹窗展示公告内容)

views/settings/index.vue (新增富文本编辑器区域，仅管理员)
```

## 实现清单

### 后端

1. `internal/model/system_setting.go`
   - `SystemSettingsResponse` 新增 `AssessmentAnnouncement string`
   - `UpdateSystemSettingsRequest` 新增 `AssessmentAnnouncement *string`

2. `internal/service/system_setting_service.go`
   - `GetAllSettings()` 读取 `assessment_announcement` 设置
   - `UpdateSettings()` 写入 `assessment_announcement` 设置

3. `internal/handler/system_setting_handler.go`
   - 新增 `GetAssessmentAnnouncement` 公开接口

4. `internal/router/router.go`
   - 注册新路由 `GET /api/v1/announcement`

### 前端

1. 安装 TipTap 富文本编辑器依赖
2. `web/src/api/modules/settings.ts` — 新增获取公告 API
3. `web/src/layouts/AuthenticatedLayout.vue` — 顶部栏加公告图标按钮
4. `web/src/components/custom/AssessmentAnnouncementModal.vue` — 新建弹窗组件
5. `web/src/views/settings/index.vue` — 加富文本编辑区
