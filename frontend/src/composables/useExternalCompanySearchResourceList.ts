import { computed, onBeforeUnmount, ref, watch } from "vue"

import { listAllExternalCompanySearchResults } from "@/api/modules/externalCompanySearch"
import { getRequestErrorMessage } from "@/lib/http-error"
import type { ExternalCompanySearchResultItem } from "@/types/externalCompanySearch"

const DEFAULT_PAGE_SIZE = 10
const DEFAULT_REFRESH_DELAY_MS = 400
const DEFAULT_SEARCH_DEBOUNCE_MS = 300

const normalizeKeyword = (value: string) => String(value || "").trim()

export function useExternalCompanySearchResourceList(initialPageSize = DEFAULT_PAGE_SIZE) {
  const resourceLoading = ref(false)
  const resourceListError = ref("")
  const resourceItems = ref<ExternalCompanySearchResultItem[]>([])
  const resourceTotal = ref(0)
  const resourcePage = ref(1)
  const resourcePageSize = ref(initialPageSize)
  const resourceSearch = ref("")
  const resourcePlatformFilter = ref("all")
  const resourceNewOnly = ref(false)

  let resourceListRefreshTimer: ReturnType<typeof setTimeout> | null = null
  let resourceSearchDebounceTimer: ReturnType<typeof setTimeout> | null = null
  let resourceListRequestToken = 0

  const resourceTotalPages = computed(() =>
    Math.max(1, Math.ceil(resourceTotal.value / resourcePageSize.value))
  )
  const selectedResourcePlatform = computed(() =>
    resourcePlatformFilter.value === "all"
      ? undefined
      : Number(resourcePlatformFilter.value)
  )

  const loadResourceList = async (
    page = resourcePage.value,
    pageSize = resourcePageSize.value
  ) => {
    const requestToken = ++resourceListRequestToken
    resourceLoading.value = true
    resourceListError.value = ""
    try {
      const response = await listAllExternalCompanySearchResults({
        search: normalizeKeyword(resourceSearch.value) || undefined,
        platform: selectedResourcePlatform.value,
        newOnly: resourceNewOnly.value || undefined,
        page,
        pageSize
      })
      if (requestToken !== resourceListRequestToken) return
      resourceItems.value = response.items
      resourceTotal.value = response.total
      resourcePage.value = response.page
      resourcePageSize.value = response.pageSize
    } catch (error) {
      if (requestToken !== resourceListRequestToken) return
      resourceListError.value = getRequestErrorMessage(error, "加载资源列表失败")
    } finally {
      if (requestToken !== resourceListRequestToken) return
      resourceLoading.value = false
    }
  }

  const scheduleResourceListRefresh = () => {
    if (resourceListRefreshTimer) return
    resourceListRefreshTimer = setTimeout(() => {
      resourceListRefreshTimer = null
      void loadResourceList()
    }, DEFAULT_REFRESH_DELAY_MS)
  }

  const handleResourceSearch = () => {
    if (resourceSearchDebounceTimer) {
      clearTimeout(resourceSearchDebounceTimer)
      resourceSearchDebounceTimer = null
    }
    void loadResourceList(1, resourcePageSize.value)
  }

  const handleResultPageChange = (nextPage: number) => {
    if (nextPage + 1 === resourcePage.value) return
    void loadResourceList(nextPage + 1, resourcePageSize.value)
  }

  const handleResultPageSizeChange = (nextPageSize: number) => {
    void loadResourceList(1, nextPageSize)
  }

  watch(
    () => normalizeKeyword(resourceSearch.value),
    (nextKeyword, previousKeyword) => {
      if (nextKeyword === previousKeyword) return
      if (resourceSearchDebounceTimer) {
        clearTimeout(resourceSearchDebounceTimer)
      }
      resourceSearchDebounceTimer = setTimeout(() => {
        resourceSearchDebounceTimer = null
        void loadResourceList(1, resourcePageSize.value)
      }, DEFAULT_SEARCH_DEBOUNCE_MS)
    }
  )

  watch([resourcePlatformFilter, resourceNewOnly], () => {
    handleResourceSearch()
  })

  onBeforeUnmount(() => {
    if (resourceListRefreshTimer) {
      clearTimeout(resourceListRefreshTimer)
      resourceListRefreshTimer = null
    }
    if (resourceSearchDebounceTimer) {
      clearTimeout(resourceSearchDebounceTimer)
      resourceSearchDebounceTimer = null
    }
  })

  return {
    resourceLoading,
    resourceListError,
    resourceItems,
    resourceTotal,
    resourcePage,
    resourcePageSize,
    resourceSearch,
    resourcePlatformFilter,
    resourceNewOnly,
    resourceTotalPages,
    loadResourceList,
    scheduleResourceListRefresh,
    handleResourceSearch,
    handleResultPageChange,
    handleResultPageSizeChange
  }
}
