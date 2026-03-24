<script setup lang="ts">
import { ref, onMounted } from "vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Plus, Trash2, Loader2, Save } from "lucide-vue-next";
import { toast } from "vue-sonner";
import {
  createCustomerLevel,
  createCustomerSource,
  createFollowMethod,
  deleteCustomerLevel,
  deleteCustomerSource,
  deleteFollowMethod,
  getFollowMethods,
  getSystemSettings,
  updateSystemSettings,
  type CustomerLevel,
  type CustomerSource,
  type FollowMethod,
  type SystemSettings,
} from "@/api/modules/systemSettings";
import { getVisitPurposeOptions, normalizeVisitPurposeOptions } from "@/constants/customerVisit";
import { getRequestErrorMessage } from "@/lib/http-error";

const loading = ref(false);
const savingRules = ref(false);
const savingVisitPurposes = ref(false);
const savingLevels = ref(false);
const savingSources = ref(false);
const savingMethods = ref(false);
const removingLevel = ref(false);
const removingSource = ref(false);
const removingMethod = ref(false);

const settings = ref<SystemSettings>({
  customerAutoDropEnabled: true,
  followUpDropDays: 30,
  dealDropDays: 90,
  claimFreezeDays: 7,
  holidayModeEnabled: false,
  customerLimit: 100,
  showFullContact: true,
  contractNumberPrefix: "zzy_",
  visitPurposes: getVisitPurposeOptions(),
  customerLevels: [],
  customerSources: [],
});

const followMethods = ref<FollowMethod[]>([]);

type EditableDictionaryItem = {
  id?: number;
  name: string;
  sort: number;
};

const sanitizeSort = (value: unknown): number => {
  const num = Number(value);
  return Number.isFinite(num) ? num : 0;
};

const dedupeDictionaryItems = <T extends EditableDictionaryItem>(items: T[]) => {
  const seen = new Set<string>();
  const unique: T[] = [];
  let removed = 0;

  for (const item of items) {
    const name = (item.name || "").trim();
    if (!name) continue;
    const key = name.toLowerCase();
    if (seen.has(key)) {
      removed += 1;
      continue;
    }
    seen.add(key);
    unique.push({
      ...item,
      name,
      sort: sanitizeSort(item.sort),
    });
  }

  return { unique, removed };
};

const ensureEditableRows = () => {
  if (settings.value.visitPurposes.length === 0) {
    settings.value.visitPurposes = [""];
  }
  if (settings.value.customerLevels.length === 0) {
    settings.value.customerLevels = [{ name: "", sort: 0 }];
  }
  if (settings.value.customerSources.length === 0) {
    settings.value.customerSources = [{ name: "", sort: 0 }];
  }
  if (followMethods.value.length === 0) {
    followMethods.value = [{ name: "", sort: 0 }];
  }
}

const loadSettings = async () => {
  loading.value = true;
  try {
    const data = await getSystemSettings();
    const levelResult = dedupeDictionaryItems(data.customerLevels || []);
    const sourceResult = dedupeDictionaryItems(data.customerSources || []);

    settings.value = {
      ...settings.value,
      ...data,
      visitPurposes: getVisitPurposeOptions(data.visitPurposes),
      customerLevels: levelResult.unique,
      customerSources: sourceResult.unique,
    };
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "加载系统设置失败"));
  }

  try {
    const methods = await getFollowMethods();
    followMethods.value = dedupeDictionaryItems(methods || []).unique;
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "加载跟进方式失败"));
  } finally {
    ensureEditableRows();
    loading.value = false;
  }
}

const saveRules = async () => {
  const prefix = settings.value.contractNumberPrefix.trim();
  if (!prefix) {
    toast.error("合同编号前缀必填");
    return;
  }
  const claimFreezeDays = Math.max(0, Number(settings.value.claimFreezeDays) || 0);

  savingRules.value = true;
  try {
    await updateSystemSettings({
      customerAutoDropEnabled: settings.value.customerAutoDropEnabled,
      followUpDropDays: settings.value.followUpDropDays,
      dealDropDays: settings.value.dealDropDays,
      claimFreezeDays,
      holidayModeEnabled: settings.value.holidayModeEnabled,
      customerLimit: settings.value.customerLimit,
      showFullContact: settings.value.showFullContact,
      contractNumberPrefix: prefix,
    });
    settings.value.contractNumberPrefix = prefix;
    settings.value.claimFreezeDays = claimFreezeDays;
    toast.success("规则设置已保存");
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "保存失败"));
  } finally {
    savingRules.value = false;
  }
}

const addVisitPurpose = () => {
  settings.value.visitPurposes.push("");
}

const removeVisitPurpose = (index: number) => {
  if (settings.value.visitPurposes.length <= 1) return;
  settings.value.visitPurposes.splice(index, 1);
}

const saveVisitPurposes = async () => {
  const visitPurposes = normalizeVisitPurposeOptions(settings.value.visitPurposes);
  if (visitPurposes.length === 0) {
    toast.error("至少保留一个拜访目的");
    return;
  }

  savingVisitPurposes.value = true;
  try {
    await updateSystemSettings({
      visitPurposes,
    });
    settings.value.visitPurposes = visitPurposes;
    toast.success("拜访目的已保存");
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "保存失败"));
  } finally {
    savingVisitPurposes.value = false;
  }
}

const addCustomerLevel = () => {
  settings.value.customerLevels.push({ name: "", sort: 0 });
}

const removeCustomerLevel = async (index: number) => {
  const level = settings.value.customerLevels[index];
  if (!level) return;

  if (typeof level.id !== "number") {
    if (settings.value.customerLevels.length <= 1) return;
    settings.value.customerLevels.splice(index, 1);
    return;
  }

  removingLevel.value = true;
  try {
    const latest = await getSystemSettings();
    const normalized = (level.name || "").trim().toLowerCase();
    const matchedIds = (latest.customerLevels || [])
      .filter((item) => {
        if (typeof item.id !== "number") return false;
        if (item.id === level.id) return true;
        if (!normalized) return false;
        return (item.name || "").trim().toLowerCase() === normalized;
      })
      .map((item) => item.id as number);

    for (const id of new Set(matchedIds)) {
      await deleteCustomerLevel(id);
    }
    await loadSettings();
    toast.success("客户级别已删除");
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "删除失败"));
  } finally {
    removingLevel.value = false;
  }
}

const saveCustomerLevels = async () => {
  savingLevels.value = true;
  try {
    const latest = await getSystemSettings();
    const existingLevelIds = (latest.customerLevels || [])
      .map((level) => level.id)
      .filter((id): id is number => typeof id === "number");
    for (const id of existingLevelIds) {
      await deleteCustomerLevel(id);
    }

    const { unique: newLevels } = dedupeDictionaryItems(settings.value.customerLevels);
    for (const level of newLevels) {
      await createCustomerLevel({
        name: level.name,
        sort: sanitizeSort(level.sort),
      });
    }

    await loadSettings();
    toast.success("客户级别已保存");
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "保存失败"));
  } finally {
    savingLevels.value = false;
  }
}

const addCustomerSource = () => {
  settings.value.customerSources.push({ name: "", sort: 0 });
}

const removeCustomerSource = async (index: number) => {
  const source = settings.value.customerSources[index];
  if (!source) return;

  if (typeof source.id !== "number") {
    if (settings.value.customerSources.length <= 1) return;
    settings.value.customerSources.splice(index, 1);
    return;
  }

  removingSource.value = true;
  try {
    const latest = await getSystemSettings();
    const normalized = (source.name || "").trim().toLowerCase();
    const matchedIds = (latest.customerSources || [])
      .filter((item) => {
        if (typeof item.id !== "number") return false;
        if (item.id === source.id) return true;
        if (!normalized) return false;
        return (item.name || "").trim().toLowerCase() === normalized;
      })
      .map((item) => item.id as number);

    for (const id of new Set(matchedIds)) {
      await deleteCustomerSource(id);
    }
    await loadSettings();
    toast.success("客户来源已删除");
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "删除失败"));
  } finally {
    removingSource.value = false;
  }
}

const saveCustomerSources = async () => {
  savingSources.value = true;
  try {
    const latest = await getSystemSettings();
    const existingSourceIds = (latest.customerSources || [])
      .map((source) => source.id)
      .filter((id): id is number => typeof id === "number");
    for (const id of existingSourceIds) {
      await deleteCustomerSource(id);
    }

    const { unique: newSources } = dedupeDictionaryItems(settings.value.customerSources);
    for (const source of newSources) {
      await createCustomerSource({
        name: source.name,
        sort: sanitizeSort(source.sort),
      });
    }

    await loadSettings();
    toast.success("客户来源已保存");
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "保存失败"));
  } finally {
    savingSources.value = false;
  }
}

const addFollowMethod = () => {
  followMethods.value.push({ name: "", sort: 0 });
}

const removeFollowMethod = async (index: number) => {
  const method = followMethods.value[index];
  if (!method) return;

  if (typeof method.id !== "number") {
    if (followMethods.value.length <= 1) return;
    followMethods.value.splice(index, 1);
    return;
  }

  removingMethod.value = true;
  try {
    const latestMethods = await getFollowMethods();
    const normalized = (method.name || "").trim().toLowerCase();
    const matchedIds = (latestMethods || [])
      .filter((item) => {
        if (typeof item.id !== "number") return false;
        if (item.id === method.id) return true;
        if (!normalized) return false;
        return (item.name || "").trim().toLowerCase() === normalized;
      })
      .map((item) => item.id as number);

    for (const id of new Set(matchedIds)) {
      await deleteFollowMethod(id);
    }
    await loadSettings();
    toast.success("跟进方式已删除");
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "删除失败"));
  } finally {
    removingMethod.value = false;
  }
}

const saveFollowMethods = async () => {
  savingMethods.value = true;
  try {
    const latestMethods = await getFollowMethods();
    const existingMethodIds = (latestMethods || [])
      .map((method) => method.id)
      .filter((id): id is number => typeof id === "number");
    for (const id of existingMethodIds) {
      await deleteFollowMethod(id);
    }

    const { unique: newMethods } = dedupeDictionaryItems(followMethods.value);
    for (const method of newMethods) {
      await createFollowMethod({
        name: method.name,
        sort: sanitizeSort(method.sort),
      });
    }

    await loadSettings();
    toast.success("跟进方式已保存");
  } catch (error) {
    toast.error(getRequestErrorMessage(error, "保存失败"));
  } finally {
    savingMethods.value = false;
  }
}

onMounted(() => {
  loadSettings();
});
</script>

<template>
  <div class="container mx-auto p-6 max-w-7xl">
    <div v-if="loading" class="flex items-center justify-center py-20">
      <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
    </div>

    <div v-else class="grid gap-6 lg:grid-cols-2">
      <!-- 左侧：系统规则 -->
      <div class="space-y-6">
        <!-- 客户掉库规则 -->
        <Card>
          <CardHeader class="pb-4">
            <div class="flex items-center justify-between">
              <div>
                <CardTitle class="text-base">客户管理规则</CardTitle>
              </div>
              <Button @click="saveRules" :disabled="savingRules" size="sm" class="gap-1.5">
                <Loader2 v-if="savingRules" class="h-3.5 w-3.5 animate-spin" />
                <Save v-else class="h-3.5 w-3.5" />
                保存
              </Button>
            </div>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="flex items-center justify-between">
              <div class="space-y-1">
                <Label for="customerAutoDropEnabled" class="text-sm">自动掉库开关</Label>
              </div>
              <Switch
                id="customerAutoDropEnabled"
                v-model:checked="settings.customerAutoDropEnabled"
              />
            </div>

            <div class="flex items-center justify-between gap-4">
              <Label for="followUpDropDays" class="text-sm whitespace-nowrap">未跟进自动掉库</Label>
              <div class="relative w-32">
                <Input
                  id="followUpDropDays"
                  v-model.number="settings.followUpDropDays"
                  type="number"
                  min="1"
                  :disabled="!settings.customerAutoDropEnabled"
                  class="h-9 pr-8"
                />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">天</span>
              </div>
            </div>

            <div class="flex items-center justify-between gap-4">
              <Label for="dealDropDays" class="text-sm whitespace-nowrap">未签单自动掉库</Label>
              <div class="relative w-32">
                <Input
                  id="dealDropDays"
                  v-model.number="settings.dealDropDays"
                  type="number"
                  min="1"
                  :disabled="!settings.customerAutoDropEnabled"
                  class="h-9 pr-8"
                />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">天</span>
              </div>
            </div>

            <div class="flex items-center justify-between gap-4">
              <div class="space-y-1">
                <Label for="claimFreezeDays" class="text-sm whitespace-nowrap">回捡冷冻期</Label>
              </div>
              <div class="relative w-32">
                <Input
                  id="claimFreezeDays"
                  v-model.number="settings.claimFreezeDays"
                  type="number"
                  min="0"
                  class="h-9 pr-8"
                />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">天</span>
              </div>
            </div>

            <div class="flex items-center justify-between">
              <Label for="holidayMode" class="text-sm">节假日不掉库</Label>
              <Switch
                id="holidayMode"
                v-model:checked="settings.holidayModeEnabled"
                :disabled="!settings.customerAutoDropEnabled"
              />
            </div>

            <div class="flex items-center justify-between gap-4">
              <Label for="customerLimit" class="text-sm whitespace-nowrap">个人客户池上限</Label>
              <div class="relative w-32">
                <Input id="customerLimit" v-model.number="settings.customerLimit" type="number" min="1" class="h-9 pr-8" />
                <span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground pointer-events-none">个</span>
              </div>
            </div>

            <div class="flex items-center justify-between gap-4">
              <Label for="contractNumberPrefix" class="text-sm whitespace-nowrap">合同编号前缀</Label>
              <div class="w-32">
                <Input id="contractNumberPrefix" v-model="settings.contractNumberPrefix" class="h-9" placeholder="例如 zzy_" />
              </div>
            </div>
          </CardContent>
        </Card>
        <!-- 客户级别 -->
        <Card>
          <CardHeader class="pb-4">
            <div class="flex items-center justify-between">
              <div>
                <CardTitle class="text-base">客户级别</CardTitle>
                <CardDescription class="text-xs mt-1">配置客户意向级别分类</CardDescription>
              </div>
              <Button @click="saveCustomerLevels" :disabled="savingLevels" size="sm" class="gap-1.5">
                <Loader2 v-if="savingLevels" class="h-3.5 w-3.5 animate-spin" />
                <Save v-else class="h-3.5 w-3.5" />
                保存
              </Button>
            </div>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center justify-between">
              <Label class="text-sm">级别列表</Label>
              <Button type="button" variant="outline" size="sm" @click="addCustomerLevel">
                <Plus class="h-3.5 w-3.5 mr-1" />
                新增级别
              </Button>
            </div>

            <div class="space-y-2 rounded-md border p-3">
              <div v-for="(level, idx) in settings.customerLevels" :key="idx"
                class="grid gap-2 grid-cols-[1fr_100px_auto]">
                <Input v-model="level.name" placeholder="级别名称" class="h-9" />
                <Input v-model.number="level.sort" type="number" placeholder="排序" class="h-9" />
                <Button type="button" variant="outline" size="icon" class="h-9 w-9"
                  :disabled="settings.customerLevels.length <= 1 || removingLevel || savingLevels" @click="removeCustomerLevel(idx)">
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- 右侧：字典配置 -->
      <div class="space-y-6">
        <!-- 拜访目的 -->
        <Card>
          <CardHeader class="pb-4">
            <div class="flex items-center justify-between">
              <div>
                <CardTitle class="text-base">拜访目的</CardTitle>
                <CardDescription class="text-xs mt-1">配置上门拜访签到的目的选项</CardDescription>
              </div>
              <Button @click="saveVisitPurposes" :disabled="savingVisitPurposes" size="sm" class="gap-1.5">
                <Loader2 v-if="savingVisitPurposes" class="h-3.5 w-3.5 animate-spin" />
                <Save v-else class="h-3.5 w-3.5" />
                保存
              </Button>
            </div>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center justify-between">
              <Label class="text-sm">目的列表</Label>
              <Button type="button" variant="outline" size="sm" @click="addVisitPurpose">
                <Plus class="h-3.5 w-3.5 mr-1" />
                新增目的
              </Button>
            </div>

            <div class="space-y-2 rounded-md border p-3">
              <div v-for="(_, idx) in settings.visitPurposes" :key="idx" class="grid gap-2 grid-cols-[1fr_auto]">
                <Input v-model="settings.visitPurposes[idx]" placeholder="例如：初次拜访" class="h-9" />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  class="h-9 w-9"
                  :disabled="settings.visitPurposes.length <= 1 || savingVisitPurposes"
                  @click="removeVisitPurpose(idx)"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- 客户来源 -->
        <Card>
          <CardHeader class="pb-4">
            <div class="flex items-center justify-between">
              <div>
                <CardTitle class="text-base">客户来源</CardTitle>
                <CardDescription class="text-xs mt-1">配置客户获取渠道标识</CardDescription>
              </div>
              <Button @click="saveCustomerSources" :disabled="savingSources" size="sm" class="gap-1.5">
                <Loader2 v-if="savingSources" class="h-3.5 w-3.5 animate-spin" />
                <Save v-else class="h-3.5 w-3.5" />
                保存
              </Button>
            </div>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center justify-between">
              <Label class="text-sm">来源列表</Label>
              <Button type="button" variant="outline" size="sm" @click="addCustomerSource">
                <Plus class="h-3.5 w-3.5 mr-1" />
                新增来源
              </Button>
            </div>

            <div class="space-y-2 rounded-md border p-3">
              <div v-for="(source, idx) in settings.customerSources" :key="idx"
                class="grid gap-2 grid-cols-[1fr_100px_auto]">
                <Input v-model="source.name" placeholder="来源名称" class="h-9" />
                <Input v-model.number="source.sort" type="number" placeholder="排序" class="h-9" />
                <Button type="button" variant="outline" size="icon" class="h-9 w-9"
                  :disabled="settings.customerSources.length <= 1 || removingSource || savingSources" @click="removeCustomerSource(idx)">
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- 跟进方式 -->
        <Card>
          <CardHeader class="pb-4">
            <div class="flex items-center justify-between">
              <div>
                <CardTitle class="text-base">跟进方式</CardTitle>
                <CardDescription class="text-xs mt-1">配置客户跟进方式选项</CardDescription>
              </div>
              <Button @click="saveFollowMethods" :disabled="savingMethods" size="sm" class="gap-1.5">
                <Loader2 v-if="savingMethods" class="h-3.5 w-3.5 animate-spin" />
                <Save v-else class="h-3.5 w-3.5" />
                保存
              </Button>
            </div>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="flex items-center justify-between">
              <Label class="text-sm">方式列表</Label>
              <Button type="button" variant="outline" size="sm" @click="addFollowMethod">
                <Plus class="h-3.5 w-3.5 mr-1" />
                新增方式
              </Button>
            </div>

            <div class="space-y-2 rounded-md border p-3">
              <div v-for="(method, idx) in followMethods" :key="idx"
                class="grid gap-2 grid-cols-[1fr_100px_auto]">
                <Input v-model="method.name" placeholder="方式名称" class="h-9" />
                <Input v-model.number="method.sort" type="number" placeholder="排序" class="h-9" />
                <Button type="button" variant="outline" size="icon" class="h-9 w-9"
                  :disabled="followMethods.length <= 1 || removingMethod || savingMethods" @click="removeFollowMethod(idx)">
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
