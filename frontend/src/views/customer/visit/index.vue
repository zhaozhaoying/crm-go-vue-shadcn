<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  Loader2,
  RefreshCw,
  Search,
  MapPin,
  Camera,
  Eye,
  X,
  Image as ImageIcon,
  Navigation,
  Upload,
} from "lucide-vue-next";

import {
  getCustomerVisits,
  createCustomerVisit,
  uploadVisitImage,
  type CustomerVisit,
} from "@/api/modules/customerVisit";
import { getSystemSettings } from "@/api/modules/systemSettings";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Pagination } from "@/components/ui/pagination";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { getVisitPurposeOptions } from "@/constants/customerVisit";
import { getRequestErrorMessage } from "@/lib/http-error";
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue";

const goCheckIn = () => {
  window.open("https://checkin.zhaozhaoying.cn", "_blank", "noopener,noreferrer");
};

// === List State ===
const loading = ref(false);
const error = ref("");
const records = ref<CustomerVisit[]>([]);
const totalCount = ref(0);
const showSearch = ref(false);
const pageIndex = ref(0);
const pageSize = ref(10);
const searchKeyword = ref("");
const activeKeyword = ref("");

const totalPages = computed(() =>
  Math.max(1, Math.ceil(totalCount.value / pageSize.value))
);

// === Create Dialog State ===
const showCreateDialog = ref(false);
const creating = ref(false);
const createError = ref("");
const visitPurposeOptions = ref<string[]>(getVisitPurposeOptions());

const createForm = ref({
  customerName: "",
  checkInLat: 0,
  checkInLng: 0,
  province: "",
  city: "",
  area: "",
  detailAddress: "",
  images: [] as string[],
  visitPurpose: "",
  remark: "",
});

const uploadingImage = ref(false);

// === Image Preview State ===
const previewDialogOpen = ref(false);
const previewImages = ref<string[]>([]);
const previewIndex = ref(0);

// === Format Helpers ===
const formatDate = (dateStr?: string) => {
  if (!dateStr) return "-";
  return new Date(dateStr).toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
};

const parseImages = (imagesStr: string): string[] => {
  try {
    const parsed = JSON.parse(imagesStr);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
};

const formatRegion = (record: CustomerVisit) => {
  const parts = [record.province, record.city, record.area]
    .map((value) => (value || "").trim())
    .filter(Boolean);
  return parts.length > 0 ? parts.join(" / ") : "-";
};

const getPreviewImages = (record: CustomerVisit) => parseImages(record.images);

// === List Operations ===
const fetchRecords = async () => {
  loading.value = true;
  error.value = "";
  try {
    const result = await getCustomerVisits({
      page: pageIndex.value + 1,
      pageSize: pageSize.value,
      keyword: activeKeyword.value || undefined,
    });
    records.value = result.items || [];
    totalCount.value = result.total;
  } catch (err) {
    records.value = [];
    totalCount.value = 0;
    error.value = getRequestErrorMessage(err, "加载上门拜访记录失败");
  } finally {
    loading.value = false;
  }
};

const refreshList = () => {
  fetchRecords();
};

const handleSearchClick = () => {
  showSearch.value = !showSearch.value;
};

const handleSearch = () => {
  activeKeyword.value = searchKeyword.value;
  pageIndex.value = 0;
  fetchRecords();
};

const clearSearch = () => {
  searchKeyword.value = "";
  activeKeyword.value = "";
  pageIndex.value = 0;
  fetchRecords();
};

const handlePageChange = (nextPage: number) => {
  if (nextPage === pageIndex.value) return;
  pageIndex.value = nextPage;
  fetchRecords();
};

const handlePageSizeChange = (nextPageSize: number) => {
  const changed = nextPageSize !== pageSize.value;
  pageSize.value = nextPageSize;
  pageIndex.value = 0;
  if (changed) fetchRecords();
};

// === Create Operations ===
const resetCreateForm = () => {
  createForm.value = {
    customerName: "",
    checkInLat: 0,
    checkInLng: 0,
    province: "",
    city: "",
    area: "",
    detailAddress: "",
    images: [],
    visitPurpose: "",
    remark: "",
  };
  createError.value = "";
};

const openCreateDialog = () => {
  resetCreateForm();
  showCreateDialog.value = true;
};

const loadVisitPurposeOptions = async () => {
  try {
    const data = await getSystemSettings();
    visitPurposeOptions.value = getVisitPurposeOptions(data.visitPurposes);
  } catch {
    visitPurposeOptions.value = getVisitPurposeOptions();
  }
};

const handleImageUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (!target.files?.length) return;

  const files = Array.from(target.files);
  target.value = "";

  uploadingImage.value = true;
  try {
    for (const file of files) {
      if (!file.type.startsWith("image/")) continue;
      if (file.size > 20 * 1024 * 1024) {
        createError.value = `图片 ${file.name} 超过20MB限制`;
        continue;
      }
      const result = await uploadVisitImage(file);
      createForm.value.images.push(result.url);
    }
  } catch (err) {
    createError.value = getRequestErrorMessage(err, "图片上传失败");
  } finally {
    uploadingImage.value = false;
  }
};

const removeImage = (index: number) => {
  createForm.value.images.splice(index, 1);
};

const handleCreate = async () => {
  createError.value = "";

  if (!createForm.value.customerName.trim()) {
    createError.value = "请输入客户公司名称";
    return;
  }
  creating.value = true;
  try {
    await createCustomerVisit({
      customerName: createForm.value.customerName.trim(),
      checkInLat: createForm.value.checkInLat,
      checkInLng: createForm.value.checkInLng,
      province: createForm.value.province,
      city: createForm.value.city,
      area: createForm.value.area,
      detailAddress: createForm.value.detailAddress.trim(),
      images: JSON.stringify(createForm.value.images),
      visitPurpose: createForm.value.visitPurpose.trim(),
      remark: createForm.value.remark.trim(),
    });
    showCreateDialog.value = false;
    pageIndex.value = 0;
    fetchRecords();
  } catch (err) {
    createError.value = getRequestErrorMessage(err, "创建上门拜访记录失败");
  } finally {
    creating.value = false;
  }
};

// === Image Preview ===
const openImagePreview = (images: string[], index: number) => {
  previewImages.value = images;
  previewIndex.value = index;
  previewDialogOpen.value = true;
};

onMounted(() => {
  void loadVisitPurposeOptions();
  fetchRecords();
});
</script>

<template>
  <div class="w-full flex flex-col gap-4 lg:gap-6">
    <Card class="shadow-sm border-border/60">
      <CardHeader v-if="showSearch" class="border-b space-y-3">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >关键词</label
            >
            <Input
              v-model="searchKeyword"
              placeholder="客户名称/地址/拜访目的"
              class="h-9 w-56"
              @keyup.enter="handleSearch"
            />
          </div>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <div class="flex items-center gap-2 ml-auto">
            <Button size="sm" @click="handleSearch">
              <Search class="h-4 w-4" />
              <span>搜索</span>
            </Button>
            <Button size="sm" variant="outline" @click="clearSearch">
              <RefreshCw class="h-4 w-4" />
              <span>重置</span>
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent class="pt-4">
        <div class="mb-4 flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
             <Button size="sm" variant="outline" @click="refreshList">
              <RefreshCw class="h-4 w-4" />
            </Button>
            <Button size="sm" @click="goCheckIn">
              <Navigation class="h-4 w-4" />
              <span>去签到</span>
            </Button>
          </div>
          <div class="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              class="h-9 w-9"
              @click="handleSearchClick"
            >
              <Search class="h-4 w-4" />
            </Button>
          </div>
        </div>

        <div
          class="overflow-hidden rounded-lg border border-border/60 bg-background"
        >
          <div v-if="loading" class="flex items-center justify-center py-24">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
          </div>

          <div v-else-if="error" class="py-20 text-center text-destructive">
            {{ error }}
          </div>

          <div v-else class="overflow-x-auto">
            <Table class="w-max min-w-full">
              <TableHeader class="sticky top-0 z-20 bg-muted/40">
                <TableRow>
                  <TableHead class="w-16 whitespace-nowrap">编号</TableHead>
                  <TableHead class="w-40 whitespace-nowrap"
                    >客户公司名称</TableHead
                  >
                  <TableHead class="w-40 whitespace-nowrap">省市区</TableHead>
                  <TableHead class="w-40 whitespace-nowrap"
                    >详细地址</TableHead
                  >
                  <TableHead class="w-28 whitespace-nowrap"
                    >拜访目的</TableHead
                  >
                  <TableHead class="w-20 whitespace-nowrap"
                    >签到图片</TableHead
                  >
                  <TableHead class="w-28 whitespace-nowrap">备注</TableHead>
                  <TableHead class="w-24 whitespace-nowrap">签到人</TableHead>
                  <TableHead class="w-28 whitespace-nowrap"
                    >签到日期</TableHead
                  >
                  <TableHead class="w-40 whitespace-nowrap"
                    >创建时间</TableHead
                  >
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow
                  v-for="record in records"
                  :key="record.id"
                  class="group hover:bg-muted/30 transition-colors"
                >
                  <TableCell class="text-muted-foreground">{{
                    record.id
                  }}</TableCell>
                  <TableCell class="font-medium">
                    <template v-if="record.customerName">
                      <TooltipProvider :delayDuration="200">
                        <Tooltip>
                          <TooltipTrigger as-child>
                            <div class="max-w-[160px] cursor-help truncate">
                              {{ record.customerName }}
                            </div>
                          </TooltipTrigger>
                          <TooltipContent
                            class="max-w-sm whitespace-pre-wrap break-words text-left"
                          >
                            <p>{{ record.customerName }}</p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </template>
                    <span v-else>-</span>
                  </TableCell>
                  <TableCell class="max-w-[200px]">
                    <span
                      class="truncate block text-sm text-muted-foreground"
                      :title="formatRegion(record)"
                    >
                      {{ formatRegion(record) }}
                    </span>
                  </TableCell>
                  <TableCell class="max-w-[200px]">
                    <template v-if="record.detailAddress">
                      <TooltipProvider :delayDuration="200">
                        <Tooltip>
                          <TooltipTrigger as-child>
                            <div class="block cursor-help truncate text-sm">
                              {{ record.detailAddress }}
                            </div>
                          </TooltipTrigger>
                          <TooltipContent
                            class="max-w-sm whitespace-pre-wrap break-words text-left"
                          >
                            <p>{{ record.detailAddress }}</p>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    </template>
                    <span v-else class="text-muted-foreground">-</span>
                  </TableCell>
                  <TableCell>
                    <Badge
                      v-if="record.visitPurpose"
                      variant="secondary"
                      class="max-w-[120px] truncate"
                    >
                      {{ record.visitPurpose }}
                    </Badge>
                    <span v-else class="text-muted-foreground">-</span>
                  </TableCell>
                  <TableCell>
                    <div
                      v-if="getPreviewImages(record).length > 0"
                      class="flex items-center gap-2"
                    >
                      <button
                        v-for="(image, idx) in getPreviewImages(record).slice(0, 3)"
                        :key="`${record.id}-${idx}`"
                        class="relative h-10 w-10 overflow-hidden rounded-md border border-border/60 bg-muted cursor-pointer"
                        @click="openImagePreview(getPreviewImages(record), idx)"
                      >
                        <img
                          :src="image"
                          alt="签到图片"
                          class="h-full w-full object-cover"
                        />
                        <span
                          v-if="idx === 2 && getPreviewImages(record).length > 3"
                          class="absolute inset-0 flex items-center justify-center bg-black/55 text-[11px] font-medium text-white"
                        >
                          +{{ getPreviewImages(record).length - 3 }}
                        </span>
                      </button>
                    </div>
                    <span v-else class="text-muted-foreground">-</span>
                  </TableCell>
                  <TableCell
                    class="max-w-[150px] truncate text-sm"
                    :title="record.remark"
                  >
                    {{ record.remark || "-" }}
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant="outline"
                      class="bg-background text-muted-foreground"
                    >
                      {{ record.operatorUserName || "未知" }}
                    </Badge>
                  </TableCell>
                  <TableCell class="text-sm">
                    <Badge variant="outline" class="bg-background">
                      {{ record.visitDate || "-" }}
                    </Badge>
                  </TableCell>
                  <TableCell class="text-xs text-muted-foreground">
                    {{ formatDate(record.createdAt) }}
                  </TableCell>
                </TableRow>
                <EmptyTablePlaceholder
                  v-if="records.length === 0"
                  :colspan="10"
                  text="暂无上门拜访记录"
                />
              </TableBody>
            </Table>
          </div>
        </div>

        <div class="mt-4">
          <Pagination
            :current-page="pageIndex"
            :total-pages="totalPages"
            :page-size="pageSize"
            :show-selection="false"
            :total-count="totalCount"
            @update:current-page="handlePageChange"
            @update:page-size="handlePageSizeChange"
          />
        </div>
      </CardContent>
    </Card>

    <!-- Create Dialog -->
    <Dialog v-model:open="showCreateDialog">
      <DialogContent class="sm:max-w-[560px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2">
            <MapPin class="h-5 w-5 text-primary" />
            新增上门拜访签到
          </DialogTitle>
          <DialogDescription>
            记录上门拜访信息，签到日期自动为今天。
          </DialogDescription>
        </DialogHeader>
        <div class="space-y-4 py-2">
          <!-- Customer Name -->
          <div class="space-y-2">
            <Label class="text-sm font-medium">
              客户公司名称 <span class="text-destructive">*</span>
            </Label>
            <Input
              v-model="createForm.customerName"
              placeholder="请输入客户公司名称"
            />
          </div>

          <!-- Detail Address -->
          <div class="space-y-2">
            <Label class="text-sm font-medium">
              详细地址 <span class="text-destructive">*</span>
            </Label>
            <Input
              v-model="createForm.detailAddress"
              placeholder="请输入详细地址"
            />
          </div>

          <!-- Visit Purpose -->
          <div class="space-y-2">
            <Label class="text-sm font-medium">拜访目的</Label>
            <Select v-model="createForm.visitPurpose">
              <SelectTrigger>
                <SelectValue placeholder="请选择拜访目的" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="option in visitPurposeOptions" :key="option" :value="option">
                  {{ option }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Images Upload -->
          <div class="space-y-2">
            <Label class="text-sm font-medium">签到图片</Label>
            <div class="flex flex-wrap gap-3">
              <!-- Uploaded images -->
              <div
                v-for="(url, idx) in createForm.images"
                :key="idx"
                class="relative group h-[80px] w-[80px] rounded-lg border overflow-hidden"
              >
                <img
                  :src="url"
                  alt="签到图片"
                  class="h-full w-full object-cover"
                />
                <div
                  class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-1"
                >
                  <button
                    class="p-1 rounded-full hover:bg-white/20 cursor-pointer"
                    @click="openImagePreview(createForm.images, idx)"
                  >
                    <Eye class="h-4 w-4 text-white" />
                  </button>
                  <button
                    class="p-1 rounded-full hover:bg-white/20 cursor-pointer"
                    @click="removeImage(idx)"
                  >
                    <X class="h-4 w-4 text-white" />
                  </button>
                </div>
              </div>
              <!-- Upload button -->
              <label
                class="flex h-[80px] w-[80px] cursor-pointer items-center justify-center rounded-lg border-2 border-dashed border-muted-foreground/30 hover:border-primary/50 hover:bg-muted/30 transition-all"
              >
                <input
                  type="file"
                  accept="image/png,image/jpeg,image/webp"
                  multiple
                  class="hidden"
                  @change="handleImageUpload"
                  :disabled="uploadingImage"
                />
                <div class="flex flex-col items-center gap-1 text-muted-foreground">
                  <Loader2
                    v-if="uploadingImage"
                    class="h-5 w-5 animate-spin"
                  />
                  <template v-else>
                    <Upload class="h-5 w-5" />
                    <span class="text-[10px]">上传</span>
                  </template>
                </div>
              </label>
            </div>
            <p class="text-xs text-muted-foreground">
              支持 JPG、PNG、WEBP 格式，单张最大 20MB
            </p>
          </div>

          <!-- Remark -->
          <div class="space-y-2">
            <Label class="text-sm font-medium">备注</Label>
            <Textarea
              v-model="createForm.remark"
              placeholder="备注信息..."
              :rows="3"
            />
          </div>

          <!-- Error -->
          <div
            v-if="createError"
            class="text-sm text-destructive bg-destructive/10 px-3 py-2 rounded-md"
          >
            {{ createError }}
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            @click="showCreateDialog = false"
            :disabled="creating"
          >
            取消
          </Button>
          <Button @click="handleCreate" :disabled="creating || uploadingImage">
            <Loader2 v-if="creating" class="h-4 w-4 animate-spin" />
            <span>{{ creating ? "提交中..." : "提交签到" }}</span>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Image Preview Dialog -->
    <Dialog v-model:open="previewDialogOpen">
      <DialogContent class="sm:max-w-[800px]">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2">
            <ImageIcon class="h-5 w-5" />
            图片预览 ({{ previewIndex + 1 }}/{{ previewImages.length }})
          </DialogTitle>
          <DialogDescription>点击左右切换图片</DialogDescription>
        </DialogHeader>
        <div class="flex items-center justify-center min-h-[300px]">
          <img
            v-if="previewImages[previewIndex]"
            :src="previewImages[previewIndex]"
            alt="预览"
            class="max-h-[60vh] w-auto max-w-full object-contain rounded-md"
          />
        </div>
        <div
          v-if="previewImages.length > 1"
          class="flex items-center justify-center gap-3"
        >
          <Button
            variant="outline"
            size="sm"
            :disabled="previewIndex <= 0"
            @click="previewIndex--"
          >
            上一张
          </Button>
          <span class="text-sm text-muted-foreground">
            {{ previewIndex + 1 }} / {{ previewImages.length }}
          </span>
          <Button
            variant="outline"
            size="sm"
            :disabled="previewIndex >= previewImages.length - 1"
            @click="previewIndex++"
          >
            下一张
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
