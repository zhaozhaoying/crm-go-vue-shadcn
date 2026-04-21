package handler

import (
	"backend/internal/service"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CrontabHandler struct {
	autoDropService                     service.CustomerAutoDropService
	hanghangCRMDailyUserCallStatService service.HanghangCRMDailyUserCallStatService
	salesDailyScoreService              service.SalesDailyScoreService
	location                            *time.Location
}

func NewCrontabHandler(
	autoDropService service.CustomerAutoDropService,
	hanghangCRMDailyUserCallStatService service.HanghangCRMDailyUserCallStatService,
	salesDailyScoreService service.SalesDailyScoreService,
	location *time.Location,
) *CrontabHandler {
	if location == nil {
		location = time.Local
	}
	return &CrontabHandler{
		autoDropService:                     autoDropService,
		hanghangCRMDailyUserCallStatService: hanghangCRMDailyUserCallStatService,
		salesDailyScoreService:              salesDailyScoreService,
		location:                            location,
	}
}

type RunHanghangCRMDailyUserCallStatSyncTaskResponse struct {
	CallStatSync   service.SyncHanghangCRMDailyUserCallStatResult `json:"callStatSync"`
	SalesScoreSync service.SyncSalesDailyScoreResult              `json:"salesScoreSync"`
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
	if !canRunCrontabTask(currentUserRole(c)) {
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

// RunHanghangCRMDailyUserCallStatSyncTask godoc
// @Summary     执行航航CRM每日用户通话统计同步任务
// @Tags        crontab
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} APIResponse{data=handler.RunHanghangCRMDailyUserCallStatSyncTaskResponse}
// @Failure     400 {object} APIResponse "航航CRM配置缺失"
// @Failure     401 {object} APIResponse "未登录或登录已失效"
// @Failure     500 {object} APIResponse "服务器内部错误"
// @Failure     502 {object} APIResponse "请求航航CRM失败"
// @Router      /api/v1/tasks/hanghang-crm-daily-user-call-stats/run [post]
func (h *CrontabHandler) RunHanghangCRMDailyUserCallStatSyncTask(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	if h.hanghangCRMDailyUserCallStatService == nil {
		Error(c, http.StatusInternalServerError, 91098, "航航CRM通话统计服务未配置")
		return
	}
	if h.salesDailyScoreService == nil {
		Error(c, http.StatusInternalServerError, 91100, "销售每日考核服务未配置")
		return
	}

	today := time.Now().In(h.location).Format("2006-01-02")
	result, err := h.hanghangCRMDailyUserCallStatService.SyncDailyUserCallStats(
		c.Request.Context(),
		service.SyncHanghangCRMDailyUserCallStatInput{
			SortBy:     []string{"bindNum"},
			SortDesc:   []bool{true},
			CensusType: 0,
			Limit:      10,
			StartTime:  today,
			EndTime:    today,
			UserIDs:    []int64{},
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrHanghangCRMCloudTokenRequired):
			Error(c, http.StatusBadRequest, 91002, "缺少航航CRM cloud-token，请先在系统设置中配置 HANGHANG_CRM_CLOUD_TOKEN")
		case errors.Is(err, service.ErrHanghangCRMRequestFailed):
			ErrorWithDetail(c, http.StatusBadGateway, 91005, "请求航航CRM失败", err)
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 91099, "同步航航CRM通话统计失败", err)
		}
		return
	}

	scoreResult, err := h.salesDailyScoreService.SyncDailyScores(c.Request.Context(), result.StatDate)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 91101, "同步销售每日考核得分失败", err)
		return
	}

	Success(c, RunHanghangCRMDailyUserCallStatSyncTaskResponse{
		CallStatSync:   result,
		SalesScoreSync: scoreResult,
	})
}

func canRunCrontabTask(role string) bool {
	switch role {
	case "admin", "管理员", "finance", "finance_manager", "财务", "财务经理":
		return true
	default:
		return false
	}
}
