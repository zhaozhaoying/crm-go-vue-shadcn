package handler

import (
	"backend/internal/service"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type TelemarketingDailyScoreHandler struct {
	service service.SalesDailyScoreService
}

func NewTelemarketingDailyScoreHandler(service service.SalesDailyScoreService) *TelemarketingDailyScoreHandler {
	return &TelemarketingDailyScoreHandler{service: service}
}

// ListTelemarketingDailyRankings godoc
// @Summary 获取电销每日排名列表
// @Tags 每日排名
// @Produce json
// @Security BearerAuth
// @Param scoreDate query string false "积分日期，格式：YYYY-MM-DD"
// @Param sync query boolean false "是否同步今日榜单数据，默认 true"
// @Success 200 {object} APIResponse
// @Router /api/v1/telemarketing-rankings [get]
func (h *TelemarketingDailyScoreHandler) ListTelemarketingDailyRankings(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	syncOnToday := true
	if rawSync := strings.TrimSpace(c.Query("sync")); rawSync != "" {
		syncOnToday = parseBoolQuery(rawSync)
	}

	result, err := h.service.ListTelemarketingDailyRankings(
		c.Request.Context(),
		strings.TrimSpace(c.Query("scoreDate")),
		syncOnToday,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMiHuaTelemarketingConfigRequired):
			Error(c, http.StatusBadRequest, 91108, "缺少米话电销排名配置，请先在系统设置中配置 MIHUA_CALL_RECORD_TOKEN，并确认根目录 .env 中已配置 MIHUA_CALL_RECORD_LIST_URL 和 MIHUA_CALL_RECORD_SOURCE_ORIGIN")
		case errors.Is(err, service.ErrMiHuaTelemarketingRequestFailed):
			ErrorWithDetail(c, http.StatusBadGateway, 91109, "请求米话电销坐席统计失败", err)
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 91110, "加载电销每日排名失败", err)
		}
		return
	}
	Success(c, result)
}

// GetTelemarketingDailyScoreDetail godoc
// @Summary 获取电销每日积分详情
// @Tags 每日排名
// @Produce json
// @Security BearerAuth
// @Param seatWorkNumber path string true "坐席工号"
// @Param scoreDate query string false "积分日期，格式：YYYY-MM-DD"
// @Param sync query boolean false "是否同步今日榜单数据，默认 true"
// @Success 200 {object} APIResponse
// @Router /api/v1/telemarketing-rankings/{seatWorkNumber} [get]
func (h *TelemarketingDailyScoreHandler) GetTelemarketingDailyScoreDetail(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	seatWorkNumber := strings.TrimSpace(c.Param("seatWorkNumber"))
	if seatWorkNumber == "" {
		Error(c, http.StatusBadRequest, 91111, "无效的坐席工号")
		return
	}

	syncOnToday := true
	if rawSync := strings.TrimSpace(c.Query("sync")); rawSync != "" {
		syncOnToday = parseBoolQuery(rawSync)
	}

	result, err := h.service.GetTelemarketingDailyScoreDetail(
		c.Request.Context(),
		strings.TrimSpace(c.Query("scoreDate")),
		seatWorkNumber,
		syncOnToday,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTelemarketingDailyScoreNotFound):
			Error(c, http.StatusNotFound, 91112, "未找到对应工号的电销积分详情")
		case errors.Is(err, service.ErrMiHuaTelemarketingConfigRequired):
			Error(c, http.StatusBadRequest, 91113, "缺少米话电销排名配置，请先在系统设置中配置 MIHUA_CALL_RECORD_TOKEN，并确认根目录 .env 中已配置 MIHUA_CALL_RECORD_LIST_URL 和 MIHUA_CALL_RECORD_SOURCE_ORIGIN")
		case errors.Is(err, service.ErrMiHuaTelemarketingRequestFailed):
			ErrorWithDetail(c, http.StatusBadGateway, 91114, "请求米话电销坐席统计失败", err)
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 91115, "加载电销积分详情失败", err)
		}
		return
	}
	Success(c, result)
}
