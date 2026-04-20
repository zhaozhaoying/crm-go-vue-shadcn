<script setup lang="ts">
import { computed, onMounted, onActivated, ref, watch } from "vue";
import { BaggageClaim, Loader2, Plus, RefreshCw, Search, SquarePen } from "lucide-vue-next";
import { toast } from "vue-sonner";

import {
  listPotentialCustomers,
  claimCustomer,
  createCustomer,
  updateCustomer,
} from "@/api/modules/customers";
import { listUsers } from "@/api/modules/users";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import OperationFollowUpDialog from "@/components/custom/OperationFollowUpDialog.vue";
import SalesFollowUpDialog from "@/components/custom/SalesFollowUpDialog.vue";
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
import { formatSevenDayCountdown } from "@/lib/customer-display";
import { isInsideSalesUser, isOperationUser } from "@/lib/auth-role";
import { getRequestErrorMessage } from "@/lib/http-error";
import { chinaPcaCode } from "@/data/china-pca-code";
import { useAuthStore } from "@/stores/auth";
import PopupForm from "../my/popupForm.vue";
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue";
import type { Customer, CustomerFormPayload } from "@/types/customer";
import type { UserWithRole } from "@/types/user";

const authStore = useAuthStore();

const loading = ref(false);
const submitting = ref(false);
const claimingId = ref<number | null>(null);
const error = ref("");
const customers = ref<Customer[]>([]);
const totalCount = ref(0);
const showSearch = ref(false);
const pageIndex = ref(0);
const pageSize = ref(10);
const allUsers = ref<UserWithRole[]>([]);

interface SearchForm {
  name: string;
  contactName: string;
  phone: string;
  weixin: string;
  ownerUserId: string;
  province: string;
  city: string;
  area: string;
}

const createEmptySearchForm = (): SearchForm => {
  return {
    name: "",
    contactName: "",
    phone: "",
    weixin: "",
    ownerUserId: "all",
    province: "",
    city: "",
    area: "",
  };
}

const searchForm = ref<SearchForm>(createEmptySearchForm());
const activeSearchForm = ref<SearchForm>(createEmptySearchForm());

const dialogOpen = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const editingCustomer = ref<Customer | null>(null);
const followUpDialogOpen = ref(false);
const followUpCustomerId = ref<number | null>(null);
const operationFollowUpDialogOpen = ref(false);
const operationFollowUpCustomerId = ref<number | null>(null);

const provinceOptions = chinaPcaCode;
const cityOptions = computed(() => {
  if (!searchForm.value.province || searchForm.value.province === "all") {
    return [];
  }
  const province = provinceOptions.find(
    (item) => item.code === searchForm.value.province,
  );
  return province?.children ?? [];
});
const areaOptions = computed(() => {
  if (!searchForm.value.city || searchForm.value.city === "all") {
    return [];
  }
  const city = cityOptions.value.find(
    (item) => item.code === searchForm.value.city,
  );
  return city?.children ?? [];
});

const totalPages = computed(() =>
  Math.max(1, Math.ceil(totalCount.value / pageSize.value)),
);
const isInsideSales = computed(() => isInsideSalesUser(authStore.user));
const isOperation = computed(() => isOperationUser(authStore.user));

const isPoolCustomer = (customer: Customer) => {
  if (customer.isInPool === true) return true;
  if (customer.ownerUserId === null || customer.ownerUserId === undefined)
    return true;
  return customer.status === "pool" || customer.status === "公海";
}

const renderOwner = (customer: Customer) => {
  if (isPoolCustomer(customer)) return "公海";
  return (
    customer.ownerUserName ||
    (customer.ownerUserId ? `用户 #${customer.ownerUserId}` : "未分配")
  );
}

const getPrimaryPhone = (customer: Customer) => {
  if (!customer.phones?.length) return "-";
  const primary = customer.phones.find((phone) => phone.isPrimary);
  return primary?.phone || customer.phones[0].phone;
}

const getUserDisplayName = (user: Pick<UserWithRole, "id" | "nickname" | "username">) => {
  return (user.nickname || user.username || "").trim() || `用户 #${user.id}`;
}

const ownerFilterOptions = computed(() =>
  allUsers.value.map((user) => ({
    value: String(user.id),
    label: getUserDisplayName(user),
  })),
);

const loadUserOptions = async () => {
  try {
    allUsers.value = await listUsers();
  } catch {
    allUsers.value = [];
  }
}

const regionNameCache = new Map<
  string,
  { provinceName: string; cityName: string; areaName: string }
>();

const resolveRegionName = (
  provinceCode?: number,
  cityCode?: number,
  areaCode?: number,
) => {
  const cacheKey = `${provinceCode ?? ""}-${cityCode ?? ""}-${areaCode ?? ""}`;
  const cached = regionNameCache.get(cacheKey);
  if (cached) return cached;

  let provinceName = "";
  let cityName = "";
  let areaName = "";

  if (!provinceCode) {
    const emptyRegion = { provinceName, cityName, areaName };
    regionNameCache.set(cacheKey, emptyRegion);
    return emptyRegion;
  }

  const pCode = String(provinceCode);
  const province = chinaPcaCode.find((p) => p.code === pCode);
  provinceName = province?.name ?? pCode;
  if (!cityCode || !province?.children) {
    const provinceRegion = { provinceName, cityName, areaName };
    regionNameCache.set(cacheKey, provinceRegion);
    return provinceRegion;
  }

  const cCode = String(cityCode);
  const city = province.children.find((c) => c.code === cCode);
  cityName = city?.name ?? cCode;
  if (!areaCode || !city?.children) {
    const cityRegion = { provinceName, cityName, areaName };
    regionNameCache.set(cacheKey, cityRegion);
    return cityRegion;
  }

  const aCode = String(areaCode);
  const area = city.children.find((a) => a.code === aCode);
  areaName = area?.name ?? aCode;

  const fullRegion = { provinceName, cityName, areaName };
  regionNameCache.set(cacheKey, fullRegion);
  return fullRegion;
}

const normalizeRegionCode = (value: string): string | undefined => {
  if (!value || value === "all") return undefined;
  return value;
}

const normalizeOwnerUserId = (value: string): number | undefined => {
  if (!value || value === "all") return undefined;
  const userId = Number(value);
  return Number.isFinite(userId) && userId > 0 ? userId : undefined;
}

const buildListParams = () => {
  return {
    page: pageIndex.value + 1,
    pageSize: pageSize.value,
    name: activeSearchForm.value.name || undefined,
    contactName: activeSearchForm.value.contactName || undefined,
    phone: activeSearchForm.value.phone || undefined,
    weixin: activeSearchForm.value.weixin || undefined,
    ownerUserId: normalizeOwnerUserId(activeSearchForm.value.ownerUserId),
    province: normalizeRegionCode(activeSearchForm.value.province),
    city: normalizeRegionCode(activeSearchForm.value.city),
    area: normalizeRegionCode(activeSearchForm.value.area),
  };
}

const fetchCustomers = async () => {
  loading.value = true;
  error.value = "";
  try {
    const result = await listPotentialCustomers(buildListParams());
    customers.value = result.items;
    totalCount.value = result.total;
  } catch (err) {
    customers.value = [];
    totalCount.value = 0;
    error.value = getRequestErrorMessage(err, "加载客户失败");
  } finally {
    loading.value = false;
  }
}

const openCreate = () => {
  dialogMode.value = "create";
  editingCustomer.value = null;
  dialogOpen.value = true;
}

const openEdit = (customer: Customer) => {
  dialogMode.value = "edit";
  editingCustomer.value = customer;
  dialogOpen.value = true;
}

const handleClaim = async (customer: Customer) => {
  claimingId.value = customer.id;
  try {
    await claimCustomer(customer.id);
    toast.success("领取成功，客户已归属到我的客户");
    await fetchCustomers();
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "领取失败"));
  } finally {
    claimingId.value = null;
  }
}

const openFollowUp = (customer: Customer) => {
  followUpCustomerId.value = customer.id;
  followUpDialogOpen.value = true;
}

const handleFollowUpSubmit = () => {
  fetchCustomers();
}

const openOperationFollowUp = (customer: Customer) => {
  operationFollowUpCustomerId.value = customer.id;
  operationFollowUpDialogOpen.value = true;
}

const handleOperationFollowUpSubmit = () => {
  fetchCustomers();
}

const refreshList = () => {
  fetchCustomers();
}

const handleSearchClick = () => {
  showSearch.value = !showSearch.value;
}

const handleSearch = () => {
  activeSearchForm.value = { ...searchForm.value };
  pageIndex.value = 0;
  fetchCustomers();
}

const clearSearch = () => {
  searchForm.value = createEmptySearchForm();
  activeSearchForm.value = createEmptySearchForm();
  pageIndex.value = 0;
  fetchCustomers();
}

const handlePageChange = (nextPage: number) => {
  if (nextPage === pageIndex.value) return;
  pageIndex.value = nextPage;
  fetchCustomers();
}

const handlePageSizeChange = (nextPageSize: number) => {
  const changed = nextPageSize !== pageSize.value;
  pageSize.value = nextPageSize;
  pageIndex.value = 0;
  if (changed) {
    fetchCustomers();
  }
}

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

const handleSubmit = async (payload: CustomerFormPayload) => {
  submitting.value = true;
  try {
    if (dialogMode.value === "create") {
      await createCustomer({
        ...payload,
        status: "owned",
      });
    } else if (editingCustomer.value) {
      await updateCustomer(editingCustomer.value.id, payload);
    }
    dialogOpen.value = false;
    await fetchCustomers();
    toast.success(
      dialogMode.value === "create"
        ? isInsideSales.value
          ? "客户添加成功，已自动分配负责人"
          : "客户添加成功"
        : "客户更新成功",
    );
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "保存失败"));
  } finally {
    submitting.value = false;
  }
}

onMounted(async () => {
  await Promise.all([loadUserOptions(), fetchCustomers()]);
});
onActivated(async () => {
  if (allUsers.value.length === 0) {
    await loadUserOptions();
  }
  await fetchCustomers();
});
</script>

<template>
  <div class="w-full flex flex-col gap-4 lg:gap-6">
    <Card class="shadow-sm border-border/60">
      <CardHeader v-if="showSearch" class="border-b space-y-3">
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >客户名称</label
            >
            <Input
              v-model="searchForm.name"
              placeholder="客户名称"
              class="h-9 w-40"
            />
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >联系人</label
            >
            <Input
              v-model="searchForm.contactName"
              placeholder="联系人"
              class="h-9 w-40"
            />
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >电话</label
            >
            <Input
              v-model="searchForm.phone"
              placeholder="电话"
              class="h-9 w-40"
            />
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >微信</label
            >
            <Input
              v-model="searchForm.weixin"
              placeholder="微信"
              class="h-9 w-40"
            />
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >负责人</label
            >
            <Select v-model="searchForm.ownerUserId">
              <SelectTrigger class="h-9 w-40">
                <SelectValue placeholder="全部负责人" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="all">全部负责人</SelectItem>
                  <SelectItem
                    v-for="user in ownerFilterOptions"
                    :key="user.value"
                    :value="user.value"
                  >
                    {{ user.label }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >省份</label
            >
            <Select v-model="searchForm.province">
              <SelectTrigger class="h-9 w-40">
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
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >城市</label
            >
            <Select
              v-model="searchForm.city"
              :disabled="!searchForm.province || searchForm.province === 'all'"
            >
              <SelectTrigger class="h-9 w-40">
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
          <div class="flex items-center gap-2">
            <label class="text-sm text-muted-foreground whitespace-nowrap"
              >区县</label
            >
            <Select
              v-model="searchForm.area"
              :disabled="!searchForm.city || searchForm.city === 'all'"
            >
              <SelectTrigger class="h-9 w-40">
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
          <div
            class="flex w-full justify-end gap-2 pt-1 lg:ml-auto lg:w-auto lg:pt-0"
          >
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
            <Button size="sm" @click="refreshList">
              <RefreshCw class="h-4 w-4" />
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

          <Table v-else class="w-max min-w-full">
            <TableHeader class="sticky top-0 z-20 bg-muted/40">
              <TableRow>
                <TableHead class="w-16">编号</TableHead>
                <TableHead>客户名称</TableHead>
                <TableHead>法人</TableHead>
                <TableHead>联系人</TableHead>
                <TableHead>联系电话</TableHead>
                <TableHead>微信</TableHead>
                <TableHead>邮箱</TableHead>
                <TableHead>客户级别</TableHead>
                <TableHead>客户来源</TableHead>
                <TableHead>负责人</TableHead>
                <TableHead>省份</TableHead>
                <TableHead>城市</TableHead>
                <TableHead>区县</TableHead>
                <TableHead>下次跟进时间</TableHead>
                <TableHead>7天倒计时</TableHead>
                <TableHead>备注</TableHead>
                <TableHead
                  class="sticky right-0 z-30 w-[180px] min-w-[180px] bg-muted/95 text-center before:absolute before:left-0 before:top-0 before:h-full before:w-px before:bg-border"
                  >操作</TableHead
                >
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="error">
                <TableCell
                  :colspan="17"
                  class="h-24 text-center text-destructive"
                >
                  {{ error }}
                </TableCell>
              </TableRow>

              <template v-else>
                <TableRow
                  v-for="customer in customers"
                  :key="customer.id"
                  class="group hover:bg-muted/30 transition-colors"
                >
                  <TableCell class="text-muted-foreground">{{
                    customer.id
                  }}</TableCell>
                  <TableCell class="font-medium">
                    <div class="flex flex-col gap-0.5">
                      <span>{{ customer.name }}</span>
                    </div>
                  </TableCell>
                  <TableCell>{{ customer.legalName || "-" }}</TableCell>
                  <TableCell>{{ customer.contactName || "-" }}</TableCell>
                  <TableCell>{{ getPrimaryPhone(customer) }}</TableCell>
                  <TableCell>{{ customer.weixin || "-" }}</TableCell>
                  <TableCell>{{ customer.email || "-" }}</TableCell>
                  <TableCell>{{ customer.customerLevelName || "-" }}</TableCell>
                  <TableCell>{{ customer.customerSourceName || "-" }}</TableCell>
                  <TableCell>
                    <Badge
                      variant="outline"
                      class="bg-background text-muted-foreground"
                    >
                      {{ renderOwner(customer) }}
                    </Badge>
                  </TableCell>
                  <TableCell>{{
                    resolveRegionName(
                      customer.province,
                      customer.city,
                      customer.area,
                    ).provinceName || "-"
                  }}</TableCell>
                  <TableCell>{{
                    resolveRegionName(
                      customer.province,
                      customer.city,
                      customer.area,
                    ).cityName || "-"
                  }}</TableCell>
                  <TableCell>{{
                    resolveRegionName(
                      customer.province,
                      customer.city,
                      customer.area,
                    ).areaName || "-"
                  }}</TableCell>
                  <TableCell class="text-xs">{{
                    customer.nextTime || "-"
                  }}</TableCell>
                  <TableCell class="text-xs whitespace-nowrap">{{
                    formatSevenDayCountdown(customer)
                  }}</TableCell>
                  <TableCell class="text-xs text-muted-foreground">
                    <p
                      class="max-w-[220px] truncate"
                      :title="typeof customer.remark === 'string' && customer.remark.trim() ? customer.remark : undefined"
                    >
                      {{ typeof customer.remark === "string" && customer.remark.trim() ? customer.remark : "-" }}
                    </p>
                  </TableCell>
                  <TableCell
                    class="sticky right-0 z-10 w-[180px] min-w-[180px] bg-background text-center border-l before:absolute before:left-0 before:top-0 before:h-full before:w-px before:bg-border"
                  >
                    <div class="grid grid-cols-2 gap-1.5">
                      <Button
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2"
                        :disabled="claimingId === customer.id || submitting"
                        @click="openFollowUp(customer)"
                      >
                        <span>销售跟进</span>
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2"
                        :disabled="claimingId === customer.id || submitting"
                        @click="openOperationFollowUp(customer)"
                      >
                        <span>运营跟进</span>
                      </Button>
                      <Button
                        v-if="!isOperation"
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2"
                        :disabled="claimingId === customer.id || submitting"
                        @click="handleClaim(customer)"
                      >
                        <Loader2
                          v-if="claimingId === customer.id"
                          class="h-4 w-4 flex-shrink-0 animate-spin"
                        />
                        <BaggageClaim
                          v-else
                          class="h-4 w-4 flex-shrink-0 text-emerald-600"
                        />
                        <span :class="claimingId === customer.id ? '' : 'text-emerald-600'">
                          {{ claimingId === customer.id ? "领取中" : "领取" }}
                        </span>
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2"
                        :disabled="claimingId === customer.id || submitting"
                        @click="openEdit(customer)"
                      >
                        <SquarePen class="h-4 w-4 flex-shrink-0" />
                        <span>编辑</span>
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>

                <EmptyTablePlaceholder v-if="customers.length === 0" :colspan="17" />
              </template>
            </TableBody>
          </Table>
        </div>

        <div class="mt-4">
          <Pagination
            :current-page="pageIndex"
            :total-pages="totalPages"
            :page-size="pageSize"
            :selected-count="0"
            :total-count="totalCount"
            @update:current-page="handlePageChange"
            @update:page-size="handlePageSizeChange"
          />
        </div>
      </CardContent>
    </Card>

    <PopupForm
      v-model:open="dialogOpen"
      :mode="dialogMode"
      :customer="editingCustomer"
      :submitting="submitting"
      @submit="handleSubmit"
    />
    <SalesFollowUpDialog
      v-model:open="followUpDialogOpen"
      :customer-id="followUpCustomerId"
      @submit="handleFollowUpSubmit"
    />
    <OperationFollowUpDialog
      v-model:open="operationFollowUpDialogOpen"
      :customer-id="operationFollowUpCustomerId"
      @submit="handleOperationFollowUpSubmit"
    />
  </div>
</template>
