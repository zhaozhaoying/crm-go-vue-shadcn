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
// @Failure     500 {object} APIResponse "服务器内部错误"
// @Router      /api/v1/tasks/customer-drop/run [post]
func (h *CrontabHandler) RunAutoDropTask(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	if !canRunCustomerAutoDropTask(currentUserRole(c)) {
		Error(c, http.StatusForbidden, 10203, "仅管理员或财务经理可以手动执行自动掉库任务")
		return
	}

	if h.autoDropService == nil {
		Error(c, http.StatusInternalServerError, 10201, "自动掉库服务未配置")
		return
	}

	result, err := h.autoDropService.Run(c.Request.Context())
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10202, "执行自动掉库任务失败", err)
		return
	}
	Success(c, result)
}

func canRunCustomerAutoDropTask(role string) bool {
	switch role {
	case "admin", "管理员", "finance", "finance_manager", "财务", "财务经理":
		return true
	default:
		return false
	}
}
