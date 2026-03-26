<script setup lang="ts">
import {
  computed,
  onActivated,
  onDeactivated,
  onMounted,
  onUnmounted,
  ref,
  watch,
} from "vue";
import {
  ClipboardList,
  FileText,
  Loader2,
  Plus,
  RefreshCw,
  Search,
  SquarePen,
  Trash2,
} from "lucide-vue-next";
import { toast } from "vue-sonner";

import {
  listMyCustomers,
  batchReassignCustomersByRanking,
  type BatchRankedReassignCustomersResponseItem,
  convertCustomer,
  createCustomer,
  releaseCustomer,
  updateCustomer,
} from "@/api/modules/customers";
import { listUsers } from "@/api/modules/users";
import {
  listContracts,
  createContract,
  updateContract,
} from "@/api/modules/contracts";
import { getSystemSettings } from "@/api/modules/systemSettings";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import {
  getDealDropCountdown,
  getFollowUpDropCountdown,
} from "@/lib/customer-display";
import { isAdminUser, isInsideSalesUser } from "@/lib/auth-role";
import { getRequestErrorMessage } from "@/lib/http-error";
import { chinaPcaCode } from "@/data/china-pca-code";
import { useAuthStore } from "@/stores/auth";
import PopupForm from "./popupForm.vue";
import SalesOrderPopupForm from "./salesOrderPopupForm.vue";
import EmptyTablePlaceholder from "@/components/custom/EmptyTablePlaceholder.vue";
import ConfirmDialog from "@/components/custom/ConfirmDialog.vue";
import SalesFollowUpDialog from "@/components/custom/SalesFollowUpDialog.vue";
import OperationFollowUpDialog from "@/components/custom/OperationFollowUpDialog.vue";
import type { Customer, CustomerFormPayload } from "@/types/customer";
import type { Contract, ContractFormPayload } from "@/types/contract";
import type { UserWithRole } from "@/types/user";

const authStore = useAuthStore();

const loading = ref(false);
const submitting = ref(false);
const convertingId = ref<number | null>(null);
const discardingId = ref<number | null>(null);
const batchDiscarding = ref(false);
const batchReassigning = ref(false);
const reassignResultOpen = ref(false);
const reassignResultItems = ref<BatchRankedReassignCustomersResponseItem[]>([]);
const error = ref("");
const customers = ref<Customer[]>([]);
const totalCount = ref(0);
const showSearch = ref(false);
const pageIndex = ref(0);
const pageSize = ref(10);
const selectedIds = ref<number[]>([]);
const confirmDialog = ref<InstanceType<typeof ConfirmDialog> | null>(null);
type OwnershipScope = "all" | "mine" | "sales" | "inside_sales" | "subordinates";
const ownershipScope = ref<OwnershipScope>("all");
const hasSubordinates = ref(false);
const ALL_OWNERSHIP_SCOPE_TABS: Array<{ value: OwnershipScope; label: string }> = [
  { value: "all", label: "全部客户" },
  { value: "mine", label: "我的" },
  { value: "sales", label: "销售部" },
  { value: "inside_sales", label: "电销部" },
  { value: "subordinates", label: "下属" },
];
const ownershipScopeTabs = computed(() => {
  if (!hasSubordinates.value) {
    return ALL_OWNERSHIP_SCOPE_TABS.filter((tab) => tab.value !== "subordinates");
  }
  return ALL_OWNERSHIP_SCOPE_TABS;
});
const showOwnershipTabs = computed(() => !isInsideSales.value);

const ownershipScopeDotClassMap: Record<OwnershipScope, string> = {
  all: "bg-[#ff4d6d]",
  mine: "bg-[#ef4444]",
  sales: "bg-[#f97316]",
  inside_sales: "bg-[#0ea5a4]",
  subordinates: "bg-[#f59e0b]",
};

const getOwnershipScopeDotClass = (scope: OwnershipScope) =>
  ownershipScopeDotClassMap[scope];

// 如果当前选中的 tab 因无下属而被移除，自动重置为"全部"
watch(ownershipScopeTabs, (tabs) => {
  const available = tabs.map((t) => t.value);
  if (!available.includes(ownershipScope.value)) {
    ownershipScope.value = "all";
  }
});

interface SearchForm {
  name: string;
  contactName: string;
  phone: string;
  weixin: string;
  ownerUserId: string;
  ownerUserName: string;
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
    ownerUserName: "",
    province: "",
    city: "",
    area: "",
  };
};

const searchForm = ref<SearchForm>(createEmptySearchForm());
const activeSearchForm = ref<SearchForm>(createEmptySearchForm());
const allUserOptions = ref<UserWithRole[]>([]);

const getUserDisplayName = (user: Pick<UserWithRole, "id" | "nickname" | "username">) =>
  (user.nickname || user.username || "").trim() || `用户 #${user.id}`;

const ownerFilterOptions = computed(() =>
  [...allUserOptions.value].sort((left, right) =>
    getUserDisplayName(left).localeCompare(getUserDisplayName(right), "zh-CN"),
  ),
);

const dialogOpen = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const editingCustomer = ref<Customer | null>(null);

const salesOrderDialogOpen = ref(false);
const salesOrderDialogMode = ref<"create" | "edit">("create");
const salesOrderSubmitting = ref(false);
const salesOrderReadonly = ref(false);
const salesOrderLoadingCustomerId = ref<number | null>(null);
const salesOrderCustomerId = ref<number | null>(null);
const editingSalesOrderContract = ref<Contract | null>(null);

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
const isAdmin = computed(() => isAdminUser(authStore.user));
const isInsideSales = computed(() => isInsideSalesUser(authStore.user));
const canBatchReassign = computed(() => isAdmin.value);
const currentUserId = computed(() => Number(authStore.user?.id || 0));
const followUpDropDays = ref(30);
const dealDropDays = ref(90);
const salesAssignDealDropDays = ref(30);
const countdownNowMs = ref(Date.now());
let countdownTimer: number | null = null;
const isPoolCustomer = (customer: Customer) => {
  if (customer.isInPool === true) return true;
  if (customer.ownerUserId === null || customer.ownerUserId === undefined)
    return true;
  return customer.status === "pool" || customer.status === "公海";
};
const canDiscardCustomer = (customer: Customer) =>
  !isPoolCustomer(customer) &&
  currentUserId.value > 0 &&
  Number(customer.ownerUserId || 0) === currentUserId.value;
const selectableCustomerIds = computed(() => customers.value.map((customer) => customer.id));
const allPageSelected = computed(
  () =>
    selectableCustomerIds.value.length > 0 &&
    selectableCustomerIds.value.every((id) => selectedIds.value.includes(id)),
);
const somePageSelected = computed(
  () =>
    selectableCustomerIds.value.some((id) => selectedIds.value.includes(id)) &&
    !allPageSelected.value,
);
const selectedCustomers = computed(() =>
  customers.value.filter(
    (customer) => selectedIds.value.includes(customer.id),
  ),
);
const selectedDiscardableCustomers = computed(() =>
  customers.value.filter(
    (customer) => selectedIds.value.includes(customer.id) && canDiscardCustomer(customer),
  ),
);

const toggleAllPage = (val: boolean | "indeterminate") => {
  const checked = val === true;
  if (checked) {
    selectedIds.value = [...selectableCustomerIds.value];
    return;
  }
  selectedIds.value = [];
};

const toggleRow = (id: number, val: boolean | "indeterminate") => {
  const checked = val === true;
  if (checked) {
    if (!selectedIds.value.includes(id)) {
      selectedIds.value = [...selectedIds.value, id];
    }
    return;
  }
  selectedIds.value = selectedIds.value.filter((item) => item !== id);
};

const renderOwner = (customer: Customer) => {
  if (isPoolCustomer(customer)) return "公海";
  return (
    customer.ownerUserName ||
    (customer.ownerUserId ? `用户 #${customer.ownerUserId}` : "未分配")
  );
};

const renderAssignmentLabel = (customer: Customer) =>
  customer.assignmentLabel || "-";

const isPendingConvertCustomer = (customer: Customer) =>
  (() => {
    const insideSalesUserId = Number(customer.insideSalesUserId || 0);
    if (insideSalesUserId <= 0 || customer.convertedAt) return false;
    if (isAdmin.value) {
      if (ownershipScope.value !== "inside_sales") return false;
      return (
        isPoolCustomer(customer) ||
        Number(customer.ownerUserId || 0) === insideSalesUserId
      );
    }
    if (
      !isInsideSales.value ||
      currentUserId.value <= 0 ||
      insideSalesUserId !== currentUserId.value
    ) {
      return false;
    }
    return (
      isPoolCustomer(customer) ||
      Number(customer.ownerUserId || 0) === insideSalesUserId
    );
  })();

const getPrimaryPhone = (customer: Customer) => {
  if (!customer.phones?.length) return "-";
  const primary = customer.phones.find((phone) => phone.isPrimary);
  return primary?.phone || customer.phones[0].phone;
};

const formatDateTime = (value?: string | null) => {
  if (!value) return "-";
  try {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "-";
    }
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const hours = String(date.getHours()).padStart(2, "0");
    const minutes = String(date.getMinutes()).padStart(2, "0");
    const seconds = String(date.getSeconds()).padStart(2, "0");
    return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
  } catch {
    return "-";
  }
};

const getFollowUpCountdownDisplay = (customer: Customer) =>
  getFollowUpDropCountdown(
    customer,
    followUpDropDays.value,
    countdownNowMs.value,
  );

const shouldUseSalesAssignDealCountdown = (customer: Customer) => {
  if (isInsideSales.value) return false;

  const insideSalesUserId = Number(customer.insideSalesUserId || 0);
  const ownerUserId = Number(customer.ownerUserId || 0);
  if (insideSalesUserId <= 0 || ownerUserId <= 0) return false;

  return insideSalesUserId !== ownerUserId;
};

const getDealCountdownDisplay = (customer: Customer) =>
  getDealDropCountdown(
    customer,
    dealDropDays.value,
    salesAssignDealDropDays.value,
    countdownNowMs.value,
    shouldUseSalesAssignDealCountdown(customer),
  );

const getDealCountdownBaseTime = (customer: Customer) =>
  shouldUseSalesAssignDealCountdown(customer)
    ? (customer.assignTime || customer.collectTime)
    : customer.collectTime;

const getCountdownCellClass = (isWarning: boolean) =>
  isWarning ? "font-medium text-destructive" : "text-muted-foreground";

const loadDropSettings = async () => {
  try {
    const settings = await getSystemSettings();
    followUpDropDays.value = Number(settings.followUpDropDays) > 0
      ? Number(settings.followUpDropDays)
      : 30;
    dealDropDays.value = Number(settings.dealDropDays) > 0
      ? Number(settings.dealDropDays)
      : 90;
    salesAssignDealDropDays.value = Number(settings.salesAssignDealDropDays) > 0
      ? Number(settings.salesAssignDealDropDays)
      : 30;
  } catch {
    followUpDropDays.value = 30;
    dealDropDays.value = 90;
    salesAssignDealDropDays.value = 30;
  }
};

const startCountdownTimer = () => {
  countdownNowMs.value = Date.now();
  if (countdownTimer !== null) return;
  countdownTimer = window.setInterval(() => {
    countdownNowMs.value = Date.now();
  }, 60 * 1000);
};

const stopCountdownTimer = () => {
  if (countdownTimer === null) return;
  window.clearInterval(countdownTimer);
  countdownTimer = null;
};

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
};

const normalizeRegionCode = (value: string): string | undefined => {
  if (!value || value === "all") return undefined;
  return value;
};

const normalizeOwnerUserId = (value: string): number | undefined => {
  if (!value || value === "all") return undefined;
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed <= 0) return undefined;
  return Math.floor(parsed);
};

const buildListParams = () => {
  return {
    page: pageIndex.value + 1,
    pageSize: pageSize.value,
    ownershipScope: ownershipScope.value,
    name: activeSearchForm.value.name || undefined,
    contactName: activeSearchForm.value.contactName || undefined,
    phone: activeSearchForm.value.phone || undefined,
    weixin: activeSearchForm.value.weixin || undefined,
    ownerUserId: normalizeOwnerUserId(activeSearchForm.value.ownerUserId),
    ownerUserName: activeSearchForm.value.ownerUserName || undefined,
    province: normalizeRegionCode(activeSearchForm.value.province),
    city: normalizeRegionCode(activeSearchForm.value.city),
    area: normalizeRegionCode(activeSearchForm.value.area),
  };
};

const fetchCustomers = async () => {
  loading.value = true;
  error.value = "";
  try {
    const result = await listMyCustomers(buildListParams());
    customers.value = result.items;
    totalCount.value = result.total;
    const validIds = new Set(result.items.map((item) => item.id));
    selectedIds.value = selectedIds.value.filter((id) => validIds.has(id));
  } catch (err) {
    customers.value = [];
    totalCount.value = 0;
    selectedIds.value = [];
    error.value = getRequestErrorMessage(err, "加载客户失败");
  } finally {
    loading.value = false;
  }
};

const openCreate = () => {
  dialogMode.value = "create";
  editingCustomer.value = null;
  dialogOpen.value = true;
};

const openEdit = (customer: Customer) => {
  dialogMode.value = "edit";
  editingCustomer.value = customer;
  dialogOpen.value = true;
};

const openSalesOrder = async (customer: Customer) => {
  salesOrderLoadingCustomerId.value = customer.id;
  salesOrderCustomerId.value = customer.id;
  salesOrderReadonly.value = false;
  editingSalesOrderContract.value = null;

  try {
    const result = await listContracts({
      customerId: customer.id,
      page: 1,
      pageSize: 1,
    });
    const existing = result.items[0] ?? null;
    if (existing && Number(existing.id) > 0) {
      salesOrderDialogMode.value = "edit";
      editingSalesOrderContract.value = existing;
      salesOrderReadonly.value = existing.auditStatus !== "pending";
    } else {
      salesOrderDialogMode.value = "create";
      editingSalesOrderContract.value = null;
      salesOrderReadonly.value = false;
    }
    salesOrderDialogOpen.value = true;
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "加载提单信息失败"));
  } finally {
    salesOrderLoadingCustomerId.value = null;
  }
};

const handleSalesOrderSubmit = async (payload: ContractFormPayload) => {
  salesOrderSubmitting.value = true;
  try {
    if (salesOrderDialogMode.value === "create") {
      await createContract(payload);
      toast.success("销售提单创建成功");
    } else if (editingSalesOrderContract.value) {
      const contractId = Number(editingSalesOrderContract.value.id);
      if (!Number.isFinite(contractId) || contractId <= 0) {
        toast.error("提单ID无效，请刷新后重试");
        return;
      }
      await updateContract(contractId, payload);
      toast.success("销售提单更新成功");
    }
    salesOrderDialogOpen.value = false;
    await fetchCustomers();
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "销售提单保存失败"));
  } finally {
    salesOrderSubmitting.value = false;
  }
};

const openFollowUp = (customer: Customer) => {
  followUpCustomerId.value = customer.id;
  followUpDialogOpen.value = true;
};

const handleFollowUpSubmit = () => {
  fetchCustomers();
};

const openOperationFollowUp = (customer: Customer) => {
  operationFollowUpCustomerId.value = customer.id;
  operationFollowUpDialogOpen.value = true;
};

const handleOperationFollowUpSubmit = () => {
  fetchCustomers();
};

const handleDiscard = async (customer: Customer) => {
  if (isPoolCustomer(customer)) return;
  const confirmed = await confirmDialog.value?.open({
    title: "丢弃客户到公海",
    description: `确定要将客户「${customer.name}」丢弃到公海吗？丢弃后该客户将不再归属于你。`,
    confirmText: "确认丢弃",
    variant: "warning",
  });
  if (!confirmed) return;

  discardingId.value = customer.id;
  try {
    await releaseCustomer(customer.id);
    selectedIds.value = selectedIds.value.filter((id) => id !== customer.id);
    toast.success("丢弃成功，客户已回到公海");
    await fetchCustomers();
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "丢弃失败"));
  } finally {
    discardingId.value = null;
  }
};

const handleBatchDiscard = async () => {
  if (selectedDiscardableCustomers.value.length === 0) return;

  const total = selectedDiscardableCustomers.value.length;
  const previewNames = selectedDiscardableCustomers.value
    .slice(0, 3)
    .map((item) => `「${item.name}」`)
    .join("、");
  const description =
    total > 3
      ? `确定要批量丢弃${previewNames}等 ${total} 个客户吗？`
      : `确定要批量丢弃${previewNames || `${total} 个客户`}吗？`;

  const confirmed = await confirmDialog.value?.open({
    title: "批量丢弃客户到公海",
    description,
    confirmText: "确认批量丢弃",
    variant: "warning",
  });
  if (!confirmed) return;

  batchDiscarding.value = true;
  try {
    const tasks = selectedDiscardableCustomers.value.map((customer) =>
      releaseCustomer(customer.id),
    );
    const results = await Promise.allSettled(tasks);
    const failed = results.filter((item) => item.status === "rejected");
    const successCount = results.length - failed.length;

    if (successCount > 0) {
      toast.success(`已成功丢弃 ${successCount} 个客户`);
    }
    if (failed.length > 0) {
      const firstFailed = failed[0];
      const reason =
        firstFailed.status === "rejected"
          ? getRequestErrorMessage(firstFailed.reason, "批量丢弃失败")
          : "批量丢弃失败";
      toast.error(`有 ${failed.length} 个客户丢弃失败：${reason}`);
    }
    selectedIds.value = [];
    await fetchCustomers();
  } finally {
    batchDiscarding.value = false;
  }
};

const handleBatchReassign = async () => {
  if (selectedCustomers.value.length === 0) return;

  const total = selectedCustomers.value.length;
  const previewNames = selectedCustomers.value
    .slice(0, 3)
    .map((item) => `「${item.name}」`)
    .join("、");
  const description =
    total > 3
      ? `确定要按昨日排名规则，重新分配${previewNames}等 ${total} 个客户吗？系统会按各自所属部门分别计算。`
      : `确定要按昨日排名规则，重新分配${previewNames || `${total} 个客户`}吗？系统会按各自所属部门分别计算。`;

  const confirmed = await confirmDialog.value?.open({
    title: "批量重新分配客户",
    description,
    confirmText: "确认重新分配",
  });
  if (!confirmed) return;

  batchReassigning.value = true;
  try {
    const result = await batchReassignCustomersByRanking(
      selectedCustomers.value.map((customer) => customer.id),
    );
    reassignResultItems.value = result.items;
    if (result.successCount > 0) {
      toast.success(`已完成 ${result.successCount} 个客户的重新分配`);
    }
    reassignResultOpen.value = true;
    if (result.failedCount > 0) {
      toast.error(`有 ${result.failedCount} 个客户重新分配失败，请查看明细`);
    }
    selectedIds.value = [];
    await fetchCustomers();
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "批量重新分配失败"));
  } finally {
    batchReassigning.value = false;
  }
};

const refreshList = () => {
  fetchCustomers();
};

const handleConvert = async (customer: Customer) => {
  convertingId.value = customer.id;
  try {
    await convertCustomer(customer.id);
    toast.success("转化成功，客户已按原分配规则分配");
    await fetchCustomers();
  } catch (err) {
    toast.error(getRequestErrorMessage(err, "转化失败"));
  } finally {
    convertingId.value = null;
  }
};

const handleOwnershipScopeChange = (scope: OwnershipScope) => {
  if (scope === ownershipScope.value) return;
  ownershipScope.value = scope;
  pageIndex.value = 0;
  selectedIds.value = [];
  fetchCustomers();
};

const handleSearchClick = () => {
  showSearch.value = !showSearch.value;
};

const handleSearch = () => {
  activeSearchForm.value = { ...searchForm.value };
  pageIndex.value = 0;
  fetchCustomers();
};

const clearSearch = () => {
  searchForm.value = createEmptySearchForm();
  activeSearchForm.value = createEmptySearchForm();
  pageIndex.value = 0;
  fetchCustomers();
};

const handlePageChange = (nextPage: number) => {
  if (nextPage === pageIndex.value) return;
  pageIndex.value = nextPage;
  fetchCustomers();
};

const handlePageSizeChange = (nextPageSize: number) => {
  const changed = nextPageSize !== pageSize.value;
  pageSize.value = nextPageSize;
  pageIndex.value = 0;
  if (changed) {
    fetchCustomers();
  }
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
};

const loadUserOptions = async () => {
  try {
    const users = await listUsers();
    allUserOptions.value = Array.isArray(users) ? users : [];
    const currentUserId = Number(authStore.user?.id || 0);
    hasSubordinates.value =
      currentUserId > 0 &&
      allUserOptions.value.some((user) => Number(user.parentId || 0) === currentUserId);
  } catch {
    allUserOptions.value = [];
    hasSubordinates.value = false;
  }
};

onMounted(async () => {
  startCountdownTimer();
  await loadDropSettings();
  await loadUserOptions();
  fetchCustomers();
});
onActivated(async () => {
  startCountdownTimer();
  await loadDropSettings();
  await loadUserOptions();
  fetchCustomers();
});
onDeactivated(() => {
  stopCountdownTimer();
});
onUnmounted(() => {
  stopCountdownTimer();
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
              <SelectTrigger class="h-9 w-44">
                <SelectValue placeholder="选择负责人" />
              </SelectTrigger>
              <SelectContent class="max-h-72">
                <SelectGroup>
                  <SelectItem value="all">全部负责人</SelectItem>
                  <SelectItem
                    v-for="user in ownerFilterOptions"
                    :key="user.id"
                    :value="String(user.id)"
                  >
                    {{ getUserDisplayName(user) }}
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
        <div v-if="showOwnershipTabs" class="mb-4 w-full overflow-x-auto pb-1">
          <div
            class="inline-flex w-max min-w-full items-center gap-1.5  p-1.5"
          >
            <button
              v-for="tab in ownershipScopeTabs"
              :key="tab.value"
              type="button"
              :aria-pressed="ownershipScope === tab.value"
              class="group inline-flex min-h-[26px] min-w-[84px] flex-none items-center justify-start gap-2.5 whitespace-nowrap rounded-[12px] border px-3.5 py-2.5 text-left ring-offset-background transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              :class="
                ownershipScope === tab.value
                  ? 'border-slate-900 bg-slate-900 text-white shadow-[0_12px_20px_-18px_rgba(15,23,42,0.9)]'
                  : 'border-slate-200 bg-white text-slate-600 shadow-[0_6px_14px_-14px_rgba(15,23,42,0.45)] hover:-translate-y-0.5 hover:border-slate-300 hover:text-slate-900 hover:shadow-[0_12px_20px_-18px_rgba(15,23,42,0.4)]'
              "
              @click="handleOwnershipScopeChange(tab.value)"
            >
              <span
                class="h-2.5 w-2.5 shrink-0 rounded-full transition-all duration-200"
                :class="
                  ownershipScope === tab.value
                    ? `${getOwnershipScopeDotClass(tab.value)} shadow-[0_0_0_4px_rgba(255,255,255,0.08)]`
                    : `${getOwnershipScopeDotClass(tab.value)} shadow-[0_0_0_3px_rgba(248,250,252,1)]`
                "
              />
              <span
                class="text-[14px] font-semibold leading-none tracking-[-0.02em]"
                :class="
                  ownershipScope === tab.value
                    ? 'text-white'
                    : 'text-slate-600 group-hover:text-slate-900'
                "
              >
                {{ tab.label }}
              </span>
            </button>
          </div>
        </div>

        <div class="mb-4 flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <Button size="sm" @click="refreshList">
              <RefreshCw class="h-4 w-4" />
            </Button>
            <Button size="sm" @click="openCreate">
              <Plus class="h-4 w-4" />
              <span>添加</span>
            </Button>
            <Button
              v-if="canBatchReassign"
              size="sm"
              variant="outline"
              :disabled="
                loading ||
                batchReassigning ||
                batchDiscarding ||
                selectedCustomers.length === 0 ||
                discardingId !== null
              "
              @click="handleBatchReassign"
            >
              <Loader2 v-if="batchReassigning" class="h-4 w-4 animate-spin" />
              <RefreshCw v-else class="h-4 w-4" />
              <span>{{
                batchReassigning
                  ? "重新分配中"
                  : `重新分配${selectedIds.length ? `(${selectedIds.length})` : ""}`
              }}</span>
            </Button>
            <Button
              size="sm"
              variant="outline"
              class="border-destructive/40 text-destructive hover:bg-destructive/10 hover:text-destructive"
              :disabled="
                loading ||
                batchReassigning ||
                batchDiscarding ||
                selectedDiscardableCustomers.length === 0 ||
                discardingId !== null
              "
              @click="handleBatchDiscard"
            >
              <Loader2 v-if="batchDiscarding" class="h-4 w-4 animate-spin" />
              <Trash2 v-else class="h-4 w-4" />
              <span>{{
                batchDiscarding
                  ? "批量丢弃中"
                  : `批量丢弃${selectedDiscardableCustomers.length ? `(${selectedDiscardableCustomers.length})` : ""}`
              }}</span>
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
                <TableHead class="w-12">
                  <div class="flex items-center justify-center">
                    <Checkbox
                      :checked="
                        allPageSelected || (somePageSelected && 'indeterminate')
                      "
                      class="border-black/70 data-[state=checked]:border-black data-[state=checked]:bg-black data-[state=checked]:text-white data-[state=indeterminate]:border-black data-[state=indeterminate]:bg-black data-[state=indeterminate]:text-white focus-visible:ring-black/30"
                      :disabled="
                        batchDiscarding ||
                        batchReassigning ||
                        loading
                      "
                      aria-label="全选客户"
                      @update:checked="toggleAllPage"
                    />
                  </div>
                </TableHead>
                <TableHead class="w-16 whitespace-nowrap">编号</TableHead>
                <TableHead>跟进倒计时</TableHead>
                <TableHead>签单倒计时</TableHead>
                <TableHead>客户名称</TableHead>
                <TableHead>法人</TableHead>
                <TableHead>联系人</TableHead>
                <TableHead>联系电话</TableHead>
                <TableHead>微信</TableHead>
                <TableHead>邮箱</TableHead>
                <TableHead>客户级别</TableHead>
                <TableHead>客户来源</TableHead>
                <TableHead>负责人</TableHead>
                <TableHead>分配方式</TableHead>
                <TableHead>省份</TableHead>
                <TableHead>城市</TableHead>
                <TableHead>区县</TableHead>
                <TableHead>备注</TableHead>
                <TableHead
                  class="sticky right-0 z-30 w-[180px] min-w-[180px] bg-muted/95 text-center border-l border-border before:absolute before:left-0 before:top-0 before:h-full before:w-px before:bg-border"
                >
                  <div class="inline-flex w-full items-center justify-center gap-1">
                    <SquarePen class="h-3.5 w-3.5 text-muted-foreground" />
                    <span>操作</span>
                  </div>
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="error">
                <TableCell
                  :colspan="19"
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
                  :data-state="
                    selectedIds.includes(customer.id) ? 'selected' : undefined
                  "
                >
                
                  <TableCell>
                    <div class="flex items-center justify-center">
                      <Checkbox
                        :checked="selectedIds.includes(customer.id)"
                        class="border-black/70 data-[state=checked]:border-black data-[state=checked]:bg-black data-[state=checked]:text-white data-[state=indeterminate]:border-black data-[state=indeterminate]:bg-black data-[state=indeterminate]:text-white focus-visible:ring-black/30"
                        :disabled="
                          batchDiscarding ||
                          batchReassigning ||
                          convertingId === customer.id ||
                          submitting ||
                          salesOrderSubmitting
                        "
                        aria-label="选择客户"
                        @update:checked="
                          (val: boolean | 'indeterminate') =>
                            toggleRow(customer.id, val)
                        "
                      />
                    </div>
                  </TableCell>
                  <TableCell class="text-muted-foreground">{{
                    customer.id
                  }}</TableCell>
                   <TableCell class="text-xs">
                    <span class="block whitespace-nowrap text-muted-foreground mb-2">
                      {{ formatDateTime(customer.nextTime) }}
                    </span>
                    <span
                      class="mt-1 block whitespace-nowrap"
                      :class="
                        getCountdownCellClass(
                          getFollowUpCountdownDisplay(customer).isWarning,
                        )
                      "
                    >
                      {{ getFollowUpCountdownDisplay(customer).text }}
                    </span>
                  </TableCell>
                  <TableCell class="text-xs">
                    <span class="block whitespace-nowrap text-muted-foreground mb-2">
                      {{ formatDateTime(getDealCountdownBaseTime(customer)) }}
                    </span>
                    <span
                      class="mt-1 block whitespace-nowrap"
                      :class="
                        getCountdownCellClass(
                          getDealCountdownDisplay(customer).isWarning,
                        )
                      "
                    >
                      {{ getDealCountdownDisplay(customer).text }}
                    </span>
                  </TableCell>
                  <TableCell class="font-medium">
                    <span class="block mb-2">{{ customer.name }}</span>
                    <Badge
                      v-if="isPendingConvertCustomer(customer)"
                      variant="secondary"
                      class="mt-1 w-fit whitespace-nowrap bg-amber-100 text-amber-700 hover:bg-amber-100"
                    >
                      待转化
                    </Badge>
                    <Badge
                      v-else
                      class="mt-1 w-fit whitespace-nowrap"
                      :variant="
                        customer.dealStatus === 'done' ? 'default' : 'secondary'
                      "
                      :class="
                        customer.dealStatus === 'done'
                          ? 'bg-green-100 text-green-700 hover:bg-green-200'
                          : ''
                      "
                    >
                      {{ customer.dealStatus === "done" ? "已成交" : "未成交" }}
                    </Badge>
                  </TableCell>
                 
                  <TableCell>{{ customer.legalName || "-" }}</TableCell>
                  <TableCell>{{ customer.contactName || "-" }}</TableCell>
                  <TableCell>{{ getPrimaryPhone(customer) }}</TableCell>
                  <TableCell>{{ customer.weixin || "-" }}</TableCell>
                  <TableCell>{{ customer.email || "-" }}</TableCell>
                  <TableCell>{{ customer.customerLevelName || "-" }}</TableCell>
                  <TableCell>{{
                    customer.customerSourceName || "-"
                  }}</TableCell>
                  <TableCell>
                    <Badge
                      variant="outline"
                      class="bg-background whitespace-nowrap text-muted-foreground"
                    >
                      {{ renderOwner(customer) }}
                    </Badge>
                  </TableCell>
                  <TableCell class="text-xs">
                    <Badge
                      variant="outline"
                      class="bg-background whitespace-nowrap text-muted-foreground"
                    >
                      {{ renderAssignmentLabel(customer) }}
                    </Badge>
                    <p
                      v-if="
                        customer.assignmentType === 'auto_assign' &&
                        customer.assignmentOperatorUserName
                      "
                      class="mt-1 whitespace-nowrap text-muted-foreground"
                    >
                      电销：{{ customer.assignmentOperatorUserName }}
                    </p>
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
                  <TableCell class="text-xs text-muted-foreground">
                    <p
                      class="max-w-[220px] truncate"
                      :title="
                        typeof customer.remark === 'string' &&
                        customer.remark.trim()
                          ? customer.remark
                          : undefined
                      "
                    >
                      {{
                        typeof customer.remark === "string" &&
                        customer.remark.trim()
                          ? customer.remark
                          : "-"
                      }}
                    </p>
                  </TableCell>
                  <TableCell
                    class="sticky right-0 z-10 w-[180px] min-w-[180px] border-l border-border bg-background text-center before:absolute before:left-0 before:top-0 before:h-full before:w-px before:bg-border"
                  >
                    <div
                      v-if="isPendingConvertCustomer(customer)"
                      class="ml-auto grid w-fit grid-cols-2 gap-1.5"
                    >
                      <Button
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2"
                        :disabled="
                          batchDiscarding ||
                          discardingId === customer.id ||
                          convertingId === customer.id ||
                          submitting
                        "
                        @click="handleConvert(customer)"
                      >
                        <Loader2
                          v-if="convertingId === customer.id"
                          class="h-4 w-4 flex-shrink-0 animate-spin"
                        />
                        <span>{{
                          convertingId === customer.id ? "转化中" : "转化"
                        }}</span>
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2"
                        :disabled="
                          batchDiscarding ||
                          discardingId === customer.id ||
                          convertingId === customer.id ||
                          submitting
                        "
                        @click="openEdit(customer)"
                      >
                        <SquarePen class="h-4 w-4 flex-shrink-0" />
                        <span>编辑</span>
                      </Button>
                    </div>
                    <div
                      v-else
                      class="ml-auto grid w-[168px] grid-cols-2 gap-1.5"
                    >
                      <Button
                        v-if="!isInsideSales"
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2"
                        :disabled="
                          batchDiscarding ||
                          discardingId === customer.id ||
                          convertingId === customer.id ||
                          submitting ||
                          salesOrderLoadingCustomerId === customer.id ||
                          salesOrderSubmitting
                        "
                        @click="openSalesOrder(customer)"
                      >
                        <Loader2
                          v-if="salesOrderLoadingCustomerId === customer.id"
                          class="h-4 w-4 flex-shrink-0 animate-spin"
                        />
                        <span>{{
                          salesOrderLoadingCustomerId === customer.id
                            ? "加载提单中"
                            : "销售提单"
                        }}</span>
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2"
                        :disabled="
                          batchDiscarding ||
                          discardingId === customer.id ||
                          convertingId === customer.id ||
                          submitting
                        "
                        @click="openFollowUp(customer)"
                      >
                        <span>销售跟进</span>
                      </Button>
                      <Button
                        v-if="isAdmin"
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2"
                        :disabled="
                          batchDiscarding ||
                          discardingId === customer.id ||
                          convertingId === customer.id ||
                          submitting
                        "
                        @click="openOperationFollowUp(customer)"
                      >
                        <span>运营跟进</span>
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2"
                        :disabled="
                          batchDiscarding ||
                          discardingId === customer.id ||
                          convertingId === customer.id ||
                          submitting
                        "
                        @click="openEdit(customer)"
                      >
                        <SquarePen class="h-4 w-4 flex-shrink-0" />
                        <span>编辑</span>
                      </Button>
                      <Button
                        v-if="canDiscardCustomer(customer)"
                        variant="ghost"
                        size="sm"
                        class="w-full justify-start gap-2 text-destructive hover:text-destructive"
                        :disabled="
                          batchDiscarding ||
                          discardingId === customer.id ||
                          convertingId === customer.id ||
                          submitting
                        "
                        @click="handleDiscard(customer)"
                      >
                        <Loader2
                          v-if="discardingId === customer.id"
                          class="h-4 w-4 flex-shrink-0 animate-spin"
                        />
                        <Trash2 v-else class="h-4 w-4 flex-shrink-0" />
                        <span>{{
                          discardingId === customer.id ? "丢弃中" : "丢弃"
                        }}</span>
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>

                <EmptyTablePlaceholder
                  v-if="customers.length === 0"
                  :colspan="19"
                />
              </template>
            </TableBody>
          </Table>
        </div>

        <div class="mt-4">
          <Pagination
            :current-page="pageIndex"
            :total-pages="totalPages"
            :page-size="pageSize"
            :selected-count="selectedIds.length"
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
    <ConfirmDialog ref="confirmDialog" />
    <Dialog :open="reassignResultOpen" @update:open="reassignResultOpen = $event">
      <DialogContent class="sm:max-w-[900px]">
        <DialogHeader>
          <DialogTitle>重新分配结果明细</DialogTitle>
        </DialogHeader>
        <div class="max-h-[60vh] overflow-y-auto rounded-md border border-border/60">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="w-20">状态</TableHead>
                <TableHead class="w-24">客户ID</TableHead>
                <TableHead class="w-48">客户名称</TableHead>
                <TableHead class="w-28">原负责人</TableHead>
                <TableHead class="w-28">新负责人</TableHead>
                <TableHead>结果说明</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <template v-if="reassignResultItems.length > 0">
                <TableRow v-for="item in reassignResultItems" :key="item.customerId">
                  <TableCell>
                    <Badge :variant="item.success ? 'default' : 'destructive'">
                      {{ item.success ? "成功" : "失败" }}
                    </Badge>
                  </TableCell>
                  <TableCell>{{ item.customerId }}</TableCell>
                  <TableCell>{{ item.customerName || "-" }}</TableCell>
                  <TableCell>{{ item.fromOwnerUserId ?? "-" }}</TableCell>
                  <TableCell>{{ item.toOwnerUserId ?? "-" }}</TableCell>
                  <TableCell :class="item.success ? 'text-muted-foreground' : 'text-destructive'">
                    {{ item.message || (item.success ? "重新分配成功" : "重新分配失败") }}
                  </TableCell>
                </TableRow>
              </template>
              <EmptyTablePlaceholder
                v-else
                :colspan="6"
                text="暂无重新分配结果"
              />
            </TableBody>
          </Table>
        </div>
      </DialogContent>
    </Dialog>
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
    <SalesOrderPopupForm
      v-model:open="salesOrderDialogOpen"
      :mode="salesOrderDialogMode"
      :contract="editingSalesOrderContract"
      :customer-id="salesOrderCustomerId"
      :readonly="salesOrderReadonly"
      :submitting="salesOrderSubmitting"
      @submit="handleSalesOrderSubmit"
    />
  </div>
</template>
