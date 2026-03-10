import { computed, onBeforeUnmount, ref } from "vue"

import { listExternalCompanySearchTasks } from "@/api/modules/externalCompanySearch"
import { getRequestErrorMessage } from "@/lib/http-error"
import type { ExternalCompanySearchTask } from "@/types/externalCompanySearch"

const DEFAULT_PAGE_SIZE = 10
const DEFAULT_REFRESH_DELAY_MS = 400

const normalizeKeyword = (value: string) => String(value || "").trim()

export function useExternalCompanySearchTaskList(initialPageSize = DEFAULT_PAGE_SIZE) {
  const taskLoading = ref(false)
  const taskListError = ref("")
  const taskItems = ref<ExternalCompanySearchTask[]>([])
  const taskTotal = ref(0)
  const taskPageIndex = ref(0)
  const taskPageSize = ref(initialPageSize)
  const taskKeyword = ref("")

  let taskListRefreshTimer: ReturnType<typeof setTimeout> | null = null

  const taskTotalPages = computed(() =>
    Math.max(1, Math.ceil(taskTotal.value / taskPageSize.value))
  )

  const buildTaskListParams = () => ({
    page: taskPageIndex.value + 1,
    pageSize: taskPageSize.value,
    keyword: normalizeKeyword(taskKeyword.value) || undefined
  })

  const loadTaskList = async () => {
    taskLoading.value = true
    taskListError.value = ""
    try {
      const response = await listExternalCompanySearchTasks(buildTaskListParams())
      taskItems.value = response.items
      taskTotal.value = response.total
      taskPageIndex.value = Math.max(0, response.page - 1)
      taskPageSize.value = response.pageSize
    } catch (error) {
      taskListError.value = getRequestErrorMessage(error, "加载任务列表失败")
    } finally {
      taskLoading.value = false
    }
  }

  const scheduleTaskListRefresh = () => {
    if (taskListRefreshTimer) return
    taskListRefreshTimer = setTimeout(() => {
      taskListRefreshTimer = null
      void loadTaskList()
    }, DEFAULT_REFRESH_DELAY_MS)
  }

  const handleTaskSearch = () => {
    taskPageIndex.value = 0
    void loadTaskList()
  }

  const handleTaskPageChange = (nextPage: number) => {
    if (nextPage === taskPageIndex.value) return
    taskPageIndex.value = nextPage
    void loadTaskList()
  }

  const handleTaskPageSizeChange = (nextPageSize: number) => {
    taskPageSize.value = nextPageSize
    taskPageIndex.value = 0
    void loadTaskList()
  }

  onBeforeUnmount(() => {
    if (!taskListRefreshTimer) return
    clearTimeout(taskListRefreshTimer)
    taskListRefreshTimer = null
  })

  return {
    taskLoading,
    taskListError,
    taskItems,
    taskTotal,
    taskPageIndex,
    taskPageSize,
    taskKeyword,
    taskTotalPages,
    loadTaskList,
    scheduleTaskListRefresh,
    handleTaskSearch,
    handleTaskPageChange,
    handleTaskPageSizeChange
  }
}
