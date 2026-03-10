import { computed, onBeforeUnmount, ref } from "vue"

import {
  cancelExternalCompanySearchTask,
  connectExternalCompanySearchStream,
  createExternalCompanySearchTasks,
  getExternalCompanySearchTask,
  listExternalCompanySearchEvents
} from "@/api/modules/externalCompanySearch"
import { getRequestErrorMessage } from "@/lib/http-error"
import type {
  CreateExternalCompanySearchTasksRequest,
  ExternalCompanySearchEvent,
  ExternalCompanySearchFailedPayload,
  ExternalCompanySearchResultSavedPayload,
  ExternalCompanySearchTask,
  ExternalCompanySearchTaskProgressPayload
} from "@/types/externalCompanySearch"
import {
  EXTERNAL_COMPANY_SEARCH_EVENT_TYPE,
  EXTERNAL_COMPANY_SEARCH_TASK_STATUS
} from "@/types/externalCompanySearch"

export interface UseExternalCompanySearchTaskOptions {
  autoReconnect?: boolean
  reconnectBaseDelayMs?: number
  reconnectMaxDelayMs?: number
  eventBufferSize?: number
}

const DEFAULT_RECONNECT_BASE_DELAY_MS = 1000
const DEFAULT_RECONNECT_MAX_DELAY_MS = 10000
const DEFAULT_EVENT_BUFFER_SIZE = 200

export function useExternalCompanySearchTask(options: UseExternalCompanySearchTaskOptions = {}) {
  const createdTasks = ref<ExternalCompanySearchTask[]>([])
  const task = ref<ExternalCompanySearchTask | null>(null)
  const events = ref<ExternalCompanySearchEvent[]>([])
  const lastSeq = ref(0)

  const creating = ref(false)
  const loadingTask = ref(false)
  const canceling = ref(false)
  const streaming = ref(false)

  const actionError = ref("")
  const streamError = ref("")

  const autoReconnect = options.autoReconnect !== false
  const reconnectBaseDelayMs = options.reconnectBaseDelayMs ?? DEFAULT_RECONNECT_BASE_DELAY_MS
  const reconnectMaxDelayMs = options.reconnectMaxDelayMs ?? DEFAULT_RECONNECT_MAX_DELAY_MS
  const eventBufferSize = Math.max(20, options.eventBufferSize ?? DEFAULT_EVENT_BUFFER_SIZE)

  let socket: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let manualDisconnect = false
  let reconnectAttempts = 0

  const currentTaskId = computed(() => task.value?.id ?? null)
  const isTerminalTask = computed(() => {
    const status = task.value?.status
    return status === EXTERNAL_COMPANY_SEARCH_TASK_STATUS.COMPLETED ||
      status === EXTERNAL_COMPANY_SEARCH_TASK_STATUS.FAILED ||
      status === EXTERNAL_COMPANY_SEARCH_TASK_STATUS.CANCELED
  })

  const safeParsePayload = <T>(event: ExternalCompanySearchEvent): T | null => {
    if (!event.payload) return null
    try {
      return JSON.parse(event.payload) as T
    } catch {
      return null
    }
  }

  const mergeTaskProgress = (payload: Partial<ExternalCompanySearchTaskProgressPayload>) => {
    if (!task.value) return
    task.value = {
      ...task.value,
      status: payload.status ?? task.value.status,
      pageNo: payload.pageNo ?? task.value.pageNo,
      progressPercent: payload.progressPercent ?? task.value.progressPercent,
      fetchedCount: payload.fetchedCount ?? task.value.fetchedCount,
      savedCount: payload.savedCount ?? task.value.savedCount,
      duplicateCount: payload.duplicateCount ?? task.value.duplicateCount,
      failedCount: payload.failedCount ?? task.value.failedCount
    }
  }

  const clearReconnectTimer = () => {
    if (!reconnectTimer) return
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }

  const disconnectStream = (manual = true) => {
    manualDisconnect = manual
    clearReconnectTimer()
    streaming.value = false
    if (!socket) return
    const currentSocket = socket
    socket = null
    currentSocket.onclose = null
    currentSocket.onerror = null
    currentSocket.onmessage = null
    currentSocket.close()
  }

  const appendEvent = (event: ExternalCompanySearchEvent) => {
    if (event.taskId !== currentTaskId.value) return
    if (event.seqNo <= lastSeq.value) return

    lastSeq.value = event.seqNo
    events.value.push(event)
    if (events.value.length > eventBufferSize) {
      events.value.splice(0, events.value.length - eventBufferSize)
    }
  }

  const refreshTask = async (taskId = currentTaskId.value) => {
    if (!taskId) return null
    loadingTask.value = true
    try {
      const nextTask = await getExternalCompanySearchTask(taskId)
      task.value = nextTask
      return nextTask
    } catch (error) {
      actionError.value = getRequestErrorMessage(error, "加载任务失败")
      throw error
    } finally {
      loadingTask.value = false
    }
  }

  const catchUpEvents = async (taskId = currentTaskId.value) => {
    if (!taskId) return

    let afterSeq = lastSeq.value
    let shouldRefreshTask = false

    for (;;) {
      const response = await listExternalCompanySearchEvents(taskId, afterSeq, 100)
      if (!response.items.length) {
        lastSeq.value = Math.max(lastSeq.value, response.nextSeq)
        break
      }

      for (const event of response.items) {
        appendEvent(event)
        shouldRefreshTask = true
      }

      afterSeq = response.nextSeq
      if (response.items.length < 100) break
    }

    if (shouldRefreshTask) {
      await refreshTask(taskId)
    }
  }

  const handleStreamEvent = (event: ExternalCompanySearchEvent) => {
    appendEvent(event)

    switch (event.eventType) {
      case EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.TASK_STARTED: {
        mergeTaskProgress({ status: EXTERNAL_COMPANY_SEARCH_TASK_STATUS.RUNNING })
        break
      }
      case EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.TASK_PROGRESS:
      case EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.TASK_COMPLETED: {
        const payload = safeParsePayload<ExternalCompanySearchTaskProgressPayload>(event)
        if (payload) {
          mergeTaskProgress(payload)
        } else if (event.eventType === EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.TASK_COMPLETED) {
          mergeTaskProgress({
            status: EXTERNAL_COMPANY_SEARCH_TASK_STATUS.COMPLETED,
            progressPercent: 100
          })
        }
        if (event.eventType === EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.TASK_COMPLETED) {
          void refreshTask()
        }
        break
      }
      case EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.TASK_FAILED: {
        const payload = safeParsePayload<ExternalCompanySearchFailedPayload>(event)
        mergeTaskProgress({
          status: EXTERNAL_COMPANY_SEARCH_TASK_STATUS.FAILED,
          pageNo: payload?.pageNo,
          fetchedCount: payload?.fetchedCount,
          savedCount: payload?.savedCount,
          duplicateCount: payload?.duplicateCount,
          failedCount: payload?.failedCount
        })
        if (task.value) {
          task.value = {
            ...task.value,
            errorMessage: payload?.errorMessage || event.message || task.value.errorMessage
          }
        }
        break
      }
      case EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.TASK_CANCELED: {
        mergeTaskProgress({ status: EXTERNAL_COMPANY_SEARCH_TASK_STATUS.CANCELED })
        break
      }
      case EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.RESULT_SAVED: {
        const payload = safeParsePayload<ExternalCompanySearchResultSavedPayload>(event)
        if (payload && task.value) {
          task.value = {
            ...task.value,
            savedCount: task.value.savedCount + 1,
            duplicateCount: Math.max(task.value.duplicateCount, payload.duplicateCount)
          }
        }
        break
      }
      default:
        break
    }

    if (
      event.eventType === EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.TASK_COMPLETED ||
      event.eventType === EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.TASK_FAILED ||
      event.eventType === EXTERNAL_COMPANY_SEARCH_EVENT_TYPE.TASK_CANCELED
    ) {
      disconnectStream(true)
    }
  }

  const reconnectStream = async () => {
    clearReconnectTimer()
    if (!autoReconnect || manualDisconnect || !currentTaskId.value || isTerminalTask.value) {
      return
    }

    reconnectAttempts += 1
    const delay = Math.min(
      reconnectBaseDelayMs * Math.max(1, 2 ** (reconnectAttempts - 1)),
      reconnectMaxDelayMs
    )

    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      void connectStream()
    }, delay)
  }

  const connectStream = async () => {
    const taskId = currentTaskId.value
    if (!taskId || isTerminalTask.value) return

    disconnectStream(false)
    manualDisconnect = false
    streamError.value = ""

    try {
      await catchUpEvents(taskId)
    } catch (error) {
      streamError.value = getRequestErrorMessage(error, "补拉任务事件失败")
      void reconnectStream()
      return
    }

    const nextSocket = connectExternalCompanySearchStream({
      taskId,
      afterSeq: lastSeq.value,
      onEvent: handleStreamEvent,
      onError: () => {
        streamError.value = "实时连接异常，准备重连"
      },
      onClose: () => {
        streaming.value = false
        if (!manualDisconnect) {
          void reconnectStream()
        }
      }
    })

    nextSocket.onopen = () => {
      socket = nextSocket
      streaming.value = true
      streamError.value = ""
      reconnectAttempts = 0
    }
  }

  const openTask = async (taskId: number) => {
    disconnectStream(true)
    actionError.value = ""
    streamError.value = ""
    task.value = null
    events.value = []
    lastSeq.value = 0

    await refreshTask(taskId)
    await catchUpEvents(taskId)

    if (!isTerminalTask.value) {
      await connectStream()
    }
  }

  const syncTaskState = async (taskId = currentTaskId.value) => {
    if (!taskId) return

    await refreshTask(taskId)
    await catchUpEvents(taskId)

    if (!streaming.value && !isTerminalTask.value) {
      await connectStream()
    }
  }

  const createAndWatch = async (
    input: CreateExternalCompanySearchTasksRequest,
    preferredTaskId?: number
  ) => {
    creating.value = true
    actionError.value = ""
    try {
      const response = await createExternalCompanySearchTasks(input)
      createdTasks.value = response.items
      const nextTask = preferredTaskId
        ? response.items.find((item) => item.id === preferredTaskId) ?? response.items[0]
        : response.items[0]
      if (nextTask) {
        await openTask(nextTask.id)
      }
      return response.items
    } catch (error) {
      actionError.value = getRequestErrorMessage(error, "创建抓取任务失败")
      throw error
    } finally {
      creating.value = false
    }
  }

  const cancelTask = async (taskId = currentTaskId.value) => {
    if (!taskId) return
    canceling.value = true
    actionError.value = ""
    try {
      await cancelExternalCompanySearchTask(taskId)
      await refreshTask(taskId)
    } catch (error) {
      actionError.value = getRequestErrorMessage(error, "取消任务失败")
      throw error
    } finally {
      canceling.value = false
    }
  }

  onBeforeUnmount(() => {
    disconnectStream(true)
  })

  return {
    createdTasks,
    task,
    events,
    lastSeq,
    creating,
    loadingTask,
    canceling,
    streaming,
    isTerminalTask,
    actionError,
    streamError,
    createAndWatch,
    openTask,
    refreshTask,
    syncTaskState,
    catchUpEvents,
    connectStream,
    disconnectStream,
    cancelTask
  }
}
