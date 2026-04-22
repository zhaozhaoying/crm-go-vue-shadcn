<script setup lang="ts">
import { computed, onActivated, onMounted, ref, watch } from "vue";
import { Loader2, MapPin, RefreshCw, Search, Sparkles } from "lucide-vue-next";
import { toast } from "vue-sonner";

import {
  listResourcePool,
  searchAndStoreResourcePool,
} from "@/api/modules/resourcePool";
import { chinaPcaCode } from "@/data/china-pca-code";
import { getRequestErrorMessage } from "@/lib/http-error";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Pagination } from "@/components/ui/pagination";
import {
  Select,
  SelectContent,
  SelectGroup,
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
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue";
import CopyButton from "@/components/custom/CopyButton.vue";
import type { ResourcePoolItem } from "@/types/resourcePool";

interface SearchForm {
  province: string;
  city: string;
  area: string;
  address: string;
  radius: string;
  keyword: string;
}

const createEmptySearchForm = (): SearchForm => ({
  province: "",
  city: "",
  area: "",
  address: "",
  radius: "3000",
  keyword: "公司",
});

const BAIDU_MAP_AK =
  (import.meta.env.VITE_BAIDU_MAP_AK as string | undefined)?.trim() || "";

const BAIDU_MAP_REVERSE_GEO_URL =
  (
    import.meta.env.VITE_BAIDU_MAP_REVERSE_GEO_URL as string | undefined
  )?.trim() || "https://api.map.baidu.com/reverse_geocoding/v3/";

const loading = ref(false);
const searching = ref(false);
const locating = ref(false);
const error = ref("");
const items = ref<ResourcePoolItem[]>([]);
const totalCount = ref(0);
const pageIndex = ref(0);
const pageSize = ref(10);

const listKeyword = ref("");
const activeListKeyword = ref("");
const hasPhoneFilter = ref<"all" | "1" | "0">("all");
const activeHasPhoneFilter = ref<"all" | "1" | "0">("all");

const searchForm = ref<SearchForm>(createEmptySearchForm());

const radiusOptions = [
  { value: "500", label: "500m" },
  { value: "1000", label: "1km" },
  { value: "2000", label: "2km" },
  { value: "3000", label: "3km" },
  { value: "5000", label: "5km" },
  { value: "10000", label: "10km" },
  { value: "20000", label: "20km" },
  { value: "50000", label: "50km" },
];

const provinceOptions = chinaPcaCode;
const cityOptions = computed(() => {
  if (!searchForm.value.province || searchForm.value.province === "all")
    return [];
  const province = provinceOptions.find(
    (item) => item.code === searchForm.value.province,
  );
  return province?.children ?? [];
});
const areaOptions = computed(() => {
  if (!searchForm.value.city || searchForm.value.city === "all") return [];
  const city = cityOptions.value.find(
    (item) => item.code === searchForm.value.city,
  );
  return city?.children ?? [];
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(totalCount.value / pageSize.value)),
);
const withPhoneCount = computed(
  () =>
    items.value.filter((item) => Boolean(String(item.phone || "").trim()))
      .length,
);

const resolveRegionName = () => {
  const provinceName =
    provinceOptions.find((item) => item.code === searchForm.value.province)
      ?.name || "";
  const cityName =
    cityOptions.value.find((item) => item.code === searchForm.value.city)
      ?.name || "";
  const areaName =
    areaOptions.value.find((item) => item.code === searchForm.value.area)
      ?.name || "";
  return [provinceName, cityName, areaName].filter(Boolean).join("");
};

const formatDateTime = (raw: string) => {
  if (!raw) return "-";
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return raw;
  return date.toLocaleString("zh-CN", { hour12: false });
};

const clampRadius = (value: number) =>
  Math.max(1, Math.min(50000, Math.round(value || 3000)));

const normalizeRegionName = (value: string) =>
  String(value || "")
    .trim()
    .replace(/特别行政区|壮族自治区|回族自治区|维吾尔自治区|自治区/g, "")
    .replace(/省|市|区|县|旗|地区|盟/g, "");

const isRegionNameMatch = (optionName: string, targetName: string) => {
  const option = String(optionName || "").trim();
  const target = String(targetName || "").trim();
  if (!option || !target) return false;
  if (option === target) return true;
  if (option.includes(target) || target.includes(option)) return true;
  const normalizedOption = normalizeRegionName(option);
  const normalizedTarget = normalizeRegionName(target);
  if (!normalizedOption || !normalizedTarget) return false;
  return (
    normalizedOption === normalizedTarget ||
    normalizedOption.includes(normalizedTarget) ||
    normalizedTarget.includes(normalizedOption)
  );
};

const applyRegionByName = (provinceName = "", cityName = "", areaName = "") => {
  const matchedProvince = provinceOptions.find((province) =>
    isRegionNameMatch(province.name, provinceName),
  );
  if (!matchedProvince) return;

  searchForm.value.province = matchedProvince.code;

  const matchedCity = (matchedProvince.children || []).find((city) =>
    isRegionNameMatch(city.name, cityName),
  );
  searchForm.value.city = matchedCity?.code || "";

  const matchedArea = (matchedCity?.children || []).find((area) =>
    isRegionNameMatch(area.name, areaName),
  );
  searchForm.value.area = matchedArea?.code || "";
};

const resolveGeoErrorMessage = (err: unknown) => {
  if (typeof err === "object" && err && "code" in err) {
    const code = Number((err as { code: number }).code);
    if (code === 1) return "定位权限被拒绝，请先允许浏览器定位";
    if (code === 2) return "定位失败，请检查设备定位服务";
    if (code === 3) return "定位超时，请稍后重试";
  }
  return getRequestErrorMessage(err, "获取当前位置失败");
};

const fillAddressByCurrentLocation = async () => {
  if (locating.value || searching.value) return;
  if (!navigator.geolocation) {
    toast.error("当前浏览器不支持定位");
    return;
  }

  locating.value = true;
  try {
    const position = await new Promise<GeolocationPosition>(
      (resolve, reject) => {
        navigator.geolocation.getCurrentPosition(resolve, reject, {
          enableHighAccuracy: true,
          timeout: 12000,
          maximumAge: 0,
        });
      },
    );

    const latitude = Number(position.coords.latitude);
    const longitude = Number(position.coords.longitude);
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
      throw new Error("定位坐标无效");
    }

    let resolvedAddress = "";
    let province = "";
    let city = "";
    let district = "";

    if (BAIDU_MAP_AK) {
      try {
        const params = new URLSearchParams({
          ak: BAIDU_MAP_AK,
          output: "json",
          coordtype: "wgs84ll",
          location: `${latitude},${longitude}`,
        });
        const controller = new AbortController();
        const timeout = window.setTimeout(() => controller.abort(), 8000);
        const resp = await fetch(
          `${BAIDU_MAP_REVERSE_GEO_URL}?${params.toString()}`,
          {
            signal: controller.signal,
          },
        );
        window.clearTimeout(timeout);
        if (resp.ok) {
          const data = await resp.json();
          if (Number(data?.status) === 0 && data?.result) {
            resolvedAddress = String(
              data.result.formatted_address || "",
            ).trim();
            const component = data.result.addressComponent || {};
            province = String(component.province || "");
            city = String(component.city || "");
            district = String(component.district || "");
          }
        }
      } catch {
        // ignore and fallback to coordinates
      }
    }

    if (!resolvedAddress) {
      resolvedAddress = `${latitude.toFixed(6)},${longitude.toFixed(6)}`;
    }

    searchForm.value.address = resolvedAddress;
    if (province || city || district) {
      applyRegionByName(province, city, district);
    }
    toast.success("已根据当前位置填充地址");
  } catch (err) {
    toast.error(resolveGeoErrorMessage(err));
  } finally {
    locating.value = false;
  }
};

const buildListParams = () => {
  return {
    page: pageIndex.value + 1,
    pageSize: pageSize.value,
    keyword: activeListKeyword.value || undefined,
    hasPhone:
      activeHasPhoneFilter.value === "all"
        ? undefined
        : activeHasPhoneFilter.value,
  };
};

const fetchItems = async () => {
  loading.value = true;
  error.value = "";
  try {
    const result = await listResourcePool(buildListParams());
    items.value = result.items;
    totalCount.value = result.total;
  } catch (err) {
    items.value = [];
    totalCount.value = 0;
    error.value = getRequestErrorMessage(err, "加载地图资源失败");
  } finally {
    loading.value = false;
  }
};

const handleSearchAndStore = async () => {
  const region = resolveRegionName();
  const address = searchForm.value.address.trim();
  if (!region && !address) {
    toast.error("请至少选择省市区或输入地址");
    return;
  }

  const radius = Number(searchForm.value.radius) || 3000;
  const normalizedRadius = clampRadius(radius);
  searchForm.value.radius = String(normalizedRadius);

  searching.value = true;
  try {
    await searchAndStoreResourcePool({
      region: region || undefined,
      address: address || undefined,
      radius: normalizedRadius,
      keyword: searchForm.value.keyword.trim() || undefined,
    });
    toast.success("检索完成");

    pageIndex.value = 0;
    await fetchItems();
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "地图检索失败"));
  } finally {
    searching.value = false;
  }
};

const handleListFilterSearch = () => {
  activeListKeyword.value = listKeyword.value.trim();
  activeHasPhoneFilter.value = hasPhoneFilter.value;
  pageIndex.value = 0;
  fetchItems();
};

const clearListFilter = () => {
  listKeyword.value = "";
  activeListKeyword.value = "";
  hasPhoneFilter.value = "all";
  activeHasPhoneFilter.value = "all";
  pageIndex.value = 0;
  fetchItems();
};

const clearMapQuery = () => {
  searchForm.value = createEmptySearchForm();
};

const handlePageChange = (nextPage: number) => {
  if (nextPage === pageIndex.value) return;
  pageIndex.value = nextPage;
  fetchItems();
};

const handlePageSizeChange = (nextPageSize: number) => {
  const changed = nextPageSize !== pageSize.value;
  pageSize.value = nextPageSize;
  pageIndex.value = 0;
  if (changed) fetchItems();
};

watch(
  () => searchForm.value.province,
  (provinceCode) => {
    if (!provinceCode || provinceCode === "all") {
      searchForm.value.city = "";
      searchForm.value.area = "";
      return;
    }
    if (
      !cityOptions.value.some((item) => item.code === searchForm.value.city)
    ) {
      searchForm.value.city = "";
      searchForm.value.area = "";
    }
  },
);

watch(
  () => searchForm.value.city,
  (cityCode) => {
    if (!cityCode || cityCode === "all") {
      searchForm.value.area = "";
      return;
    }
    if (
      !areaOptions.value.some((item) => item.code === searchForm.value.area)
    ) {
      searchForm.value.area = "";
    }
  },
);

onMounted(fetchItems);
onActivated(fetchItems);
</script>

<template>
  <div class="w-full flex flex-col gap-4 lg:gap-6">
    <Card class="shadow-sm border-border/60 overflow-hidden">
      <CardContent class="space-y-4 pt-4">
        <div
          class="grid grid-cols-1 gap-x-4 gap-y-3 md:grid-cols-2 lg:grid-cols-3"
        >
          <div class="space-y-1.5">
            <div class="flex h-5 items-center justify-between">
              <label class="text-xs text-muted-foreground">省份</label>
            </div>
            <Select v-model="searchForm.province">
              <SelectTrigger class="h-9 w-full">
                <SelectValue placeholder="选择省份" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem
                    v-for="province in provinceOptions"
                    :key="province.code"
                    :value="province.code"
                  >
                    {{ province.name }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-1.5">
            <div class="flex h-5 items-center justify-between">
              <label class="text-xs text-muted-foreground">城市</label>
            </div>
            <Select
              v-model="searchForm.city"
              :disabled="!searchForm.province || searchForm.province === 'all'"
            >
              <SelectTrigger class="h-9 w-full">
                <SelectValue placeholder="选择城市" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem
                    v-for="city in cityOptions"
                    :key="city.code"
                    :value="city.code"
                  >
                    {{ city.name }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-1.5">
            <div class="flex h-5 items-center justify-between">
              <label class="text-xs text-muted-foreground">区县</label>
              <span class="text-[10px] text-muted-foreground/70"
                >（必填项之一）</span
              >
            </div>
            <Select
              v-model="searchForm.area"
              :disabled="!searchForm.city || searchForm.city === 'all'"
            >
              <SelectTrigger class="h-9 w-full">
                <SelectValue placeholder="选择区县" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部</SelectItem>
                  <SelectItem
                    v-for="area in areaOptions"
                    :key="area.code"
                    :value="area.code"
                  >
                    {{ area.name }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-1.5">
            <div class="flex h-5 items-center justify-between">
              <label class="text-xs text-muted-foreground"
                >地址（可替代行政区）</label
              >
              <span class="text-[10px] text-muted-foreground/70"
                >（必填项之一）</span
              >
            </div>
            <Input
              v-model="searchForm.address"
              placeholder="输入详细地址，例如：人民广场"
              class="h-9 w-full"
            />
          </div>

          <div class="space-y-1.5">
            <div class="flex h-5 items-center justify-between">
              <label class="text-xs text-muted-foreground">关键词</label>
            </div>
            <Input
              v-model="searchForm.keyword"
              placeholder="默认：公司"
              class="h-9 w-full"
            />
          </div>

          <div class="space-y-1.5">
            <div class="flex h-5 items-center justify-between">
              <label class="text-xs text-muted-foreground">搜索半径</label>
            </div>
            <Select v-model="searchForm.radius">
              <SelectTrigger class="h-9 w-full">
                <SelectValue placeholder="选择半径范围" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem
                    v-for="opt in radiusOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.label }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div
          class="flex flex-wrap items-center gap-2 rounded-md border border-border/60 bg-muted/20 p-2"
        >
          <div class="flex items-center gap-2 px-1">
            <Badge
              variant="outline"
              class="h-6 bg-background text-[11px] font-normal text-muted-foreground shadow-sm"
            >
              当前位置：{{ resolveRegionName() || "等待定位/未选择" }}
            </Badge>
          </div>
          <div class="ml-auto flex flex-wrap items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              :disabled="locating || searching"
              @click="fillAddressByCurrentLocation"
            >
              <Loader2 v-if="locating" class="h-4 w-4 animate-spin" />
              <MapPin v-else class="h-4 w-4" />
              <span>{{ locating ? "定位中" : "当前位置填充地址" }}</span>
            </Button>
            <Button
              size="sm"
              :disabled="searching"
              @click="handleSearchAndStore"
            >
              <Loader2 v-if="searching" class="h-4 w-4 animate-spin" />
              <Sparkles v-else class="h-4 w-4" />
              <span>{{ searching ? "检索中" : "检索并入库" }}</span>
            </Button>
            <Button
              size="sm"
              variant="outline"
              :disabled="searching"
              @click="clearMapQuery"
            >
              <RefreshCw class="h-4 w-4" />
              <span>清空条件</span>
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card class="shadow-sm border-border/60 overflow-hidden">
      <CardHeader class="border-b space-y-3 bg-muted/10">
        <div class="flex flex-wrap items-center gap-2 rounded-md">
          <Input
            v-model="listKeyword"
            placeholder="名称 / 电话 / 地址"
            class="h-9 w-[280px] min-w-[220px]"
          />
          <Select v-model="hasPhoneFilter">
            <SelectTrigger class="h-9 w-36">
              <SelectValue placeholder="电话筛选" />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="all">全部</SelectItem>
                <SelectItem value="1">有电话</SelectItem>
                <SelectItem value="0">无电话</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>

          <Button size="sm" @click="handleListFilterSearch">
            <Search class="h-4 w-4" />
            <span>筛选</span>
          </Button>
          <Button size="sm" variant="outline" @click="clearListFilter">
            <RefreshCw class="h-4 w-4" />
            <span>重置</span>
          </Button>
        </div>
      </CardHeader>

      <CardContent class="pt-4">
        <div
          class="overflow-hidden rounded-lg border border-border/60 bg-background"
        >
          <div v-if="loading" class="flex items-center justify-center py-24">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
          </div>

          <Table v-else class="min-w-full">
            <TableHeader class="sticky top-0 z-20 bg-muted/40">
              <TableRow>
                <TableHead class="min-w-[240px]">企业信息</TableHead>
                <TableHead class="min-w-[140px]">联系电话</TableHead>
                <TableHead class="min-w-[260px]">地址信息</TableHead>
                <TableHead class="min-w-[180px]">坐标</TableHead>
                <TableHead class="min-w-[140px]">检索条件</TableHead>
                <TableHead class="w-[170px]">更新时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="error">
                <TableCell
                  :colspan="6"
                  class="h-24 text-center text-destructive"
                >
                  {{ error }}
                </TableCell>
              </TableRow>

              <template v-else>
                <TableRow
                  v-for="item in items"
                  :key="item.id"
                  class="group hover:bg-muted/30 transition-colors"
                >
                  <TableCell>
                    <div class="flex flex-col gap-1">
                      <p class="text-xs text-muted-foreground">
                        资源ID：{{ item.id }}
                      </p>
                      <div class="flex items-center gap-1">
                        <p class="font-medium leading-5">
                          {{ item.name || "-" }}
                        </p>
                        <CopyButton
                          v-if="item.name"
                          :text="item.name"
                          success-message="企业名称已复制"
                        />
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant="outline"
                      :class="
                        item.phone
                          ? 'border-emerald-200 bg-emerald-50 text-emerald-700'
                          : 'text-muted-foreground'
                      "
                    >
                      {{ item.phone || "无电话" }}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div class="space-y-1">
                      <p class="text-xs text-muted-foreground">
                        {{
                          [item.province, item.city, item.area]
                            .filter(Boolean)
                            .join("/") || "-"
                        }}
                      </p>
                      <p
                        class="max-w-[320px] truncate"
                        :title="item.address || undefined"
                      >
                        {{ item.address || "-" }}
                      </p>
                    </div>
                  </TableCell>
                  <TableCell class="text-xs font-mono text-muted-foreground">
                    {{ item.latitude?.toFixed(6) || "-" }},
                    {{ item.longitude?.toFixed(6) || "-" }}
                  </TableCell>
                  <TableCell>
                    <div class="space-y-1">
                      <p class="text-xs text-muted-foreground">
                        关键词：{{ item.searchKeyword || "-" }}
                      </p>
                      <p class="text-xs text-muted-foreground">
                        半径：{{ item.searchRadius || "-" }}m
                      </p>
                    </div>
                  </TableCell>
                  <TableCell class="text-xs">{{
                    formatDateTime(item.updatedAt)
                  }}</TableCell>
                </TableRow>

                <EmptyTablePlaceholder v-if="items.length === 0" :colspan="6" />
              </template>
            </TableBody>
          </Table>
        </div>

        <div class="mt-4">
          <Pagination
            :current-page="pageIndex"
            :total-pages="totalPages"
            :page-size="pageSize"
            :total-count="totalCount"
            @update:current-page="handlePageChange"
            @update:page-size="handlePageSizeChange"
          />
        </div>
      </CardContent>
    </Card>
  </div>
</template>
