package handler

import (
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CrontabHandler struct {
	autoDropService service.CustomerAutoDropService
}

func NewCrontabHandler(autoDropService service.CustomerAutoDropService) *CrontabHandler {
	return &CrontabHandler{
		autoDropService: autoDropService,
	}
}

// RunAutoDropTask godoc
// @Summary     执行客户自动掉库任务
// @Tags        crontab
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} APIResponse{data=service.CustomerAutoDropTaskResult}
// @Failure     500 {object} APIResponse
// @Router      /api/v1/tasks/customer-drop/run [post]
func (h *CrontabHandler) RunAutoDropTask(c *gin.Context) {
	if h.autoDropService == nil {
		Error(c, http.StatusInternalServerError, 10031, "auto drop service unavailable")
		return
	}

	result, err := h.autoDropService.Run(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, 10033, "failed to execute auto drop task")
		return
	}
	Success(c, result)
}
