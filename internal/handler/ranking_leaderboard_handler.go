package handler

import (
	"backend/internal/service"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type RankingLeaderboardHandler struct {
	service service.SalesDailyScoreService
}

func NewRankingLeaderboardHandler(service service.SalesDailyScoreService) *RankingLeaderboardHandler {
	return &RankingLeaderboardHandler{service: service}
}

// ListRankingLeaderboard godoc
// @Summary 获取排名榜单
// @Tags 排名榜单
// @Produce json
// @Security BearerAuth
// @Param period query string false "周期类型: month=月排名, week=周排名, day=日排名"
// @Param startDate query string false "开始日期，格式：YYYY-MM-DD"
// @Param endDate query string false "结束日期，格式：YYYY-MM-DD"
// @Param sync query boolean false "是否同步今日榜单数据，默认 true"
// @Success 200 {object} APIResponse
// @Router /api/v1/ranking-leaderboard [get]
func (h *RankingLeaderboardHandler) ListRankingLeaderboard(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	period := strings.TrimSpace(c.Query("period"))
	if period == "" {
		period = "day"
	}

	syncOnToday := true
	if rawSync := strings.TrimSpace(c.Query("sync")); rawSync != "" {
		syncOnToday = parseBoolQuery(rawSync)
	}

	result, err := h.service.ListRankingLeaderboard(
		c.Request.Context(),
		period,
		strings.TrimSpace(c.Query("startDate")),
		strings.TrimSpace(c.Query("endDate")),
		syncOnToday,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRankingLeaderboardInvalidPeriod), errors.Is(err, service.ErrRankingLeaderboardInvalidRange):
			ErrorWithDetail(c, http.StatusBadRequest, 91116, "排名榜单查询参数无效", err)
		case errors.Is(err, service.ErrMiHuaTelemarketingConfigRequired):
			Error(c, http.StatusBadRequest, 91117, "缺少米话电销排名配置，请先在系统设置中配置 MIHUA_CALL_RECORD_TOKEN，并确认根目录 .env 中已配置 MIHUA_CALL_RECORD_LIST_URL 和 MIHUA_CALL_RECORD_SOURCE_ORIGIN")
		case errors.Is(err, service.ErrMiHuaTelemarketingRequestFailed):
			ErrorWithDetail(c, http.StatusBadGateway, 91118, "请求米话电销坐席统计失败", err)
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 91119, "加载排名榜单失败", err)
		}
		return
	}
	Success(c, result)
}

// GetRankingLeaderboardDetail godoc
// @Summary 获取排名榜单详情
// @Tags 排名榜单
// @Produce json
// @Security BearerAuth
// @Param identityKey path string true "榜单身份标识"
// @Param period query string false "周期类型: month=月排名, week=周排名, day=日排名"
// @Param startDate query string false "开始日期，格式：YYYY-MM-DD"
// @Param endDate query string false "结束日期，格式：YYYY-MM-DD"
// @Param sync query boolean false "是否同步今日榜单数据，默认 true"
// @Success 200 {object} APIResponse
// @Router /api/v1/ranking-leaderboard/{identityKey} [get]
func (h *RankingLeaderboardHandler) GetRankingLeaderboardDetail(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	identityKey := strings.TrimSpace(c.Param("identityKey"))
	if identityKey == "" {
		Error(c, http.StatusBadRequest, 91120, "无效的榜单身份标识")
		return
	}

	period := strings.TrimSpace(c.Query("period"))
	if period == "" {
		period = "day"
	}

	syncOnToday := true
	if rawSync := strings.TrimSpace(c.Query("sync")); rawSync != "" {
		syncOnToday = parseBoolQuery(rawSync)
	}

	result, err := h.service.GetRankingLeaderboardDetail(
		c.Request.Context(),
		period,
		strings.TrimSpace(c.Query("startDate")),
		strings.TrimSpace(c.Query("endDate")),
		identityKey,
		syncOnToday,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRankingLeaderboardNotFound):
			Error(c, http.StatusNotFound, 91121, "未找到对应榜单详情")
		case errors.Is(err, service.ErrRankingLeaderboardInvalidPeriod), errors.Is(err, service.ErrRankingLeaderboardInvalidRange):
			ErrorWithDetail(c, http.StatusBadRequest, 91122, "榜单详情查询参数无效", err)
		case errors.Is(err, service.ErrMiHuaTelemarketingConfigRequired):
			Error(c, http.StatusBadRequest, 91123, "缺少米话电销排名配置，请先在系统设置中配置 MIHUA_CALL_RECORD_TOKEN，并确认根目录 .env 中已配置 MIHUA_CALL_RECORD_LIST_URL 和 MIHUA_CALL_RECORD_SOURCE_ORIGIN")
		case errors.Is(err, service.ErrMiHuaTelemarketingRequestFailed):
			ErrorWithDetail(c, http.StatusBadGateway, 91124, "请求米话电销坐席统计失败", err)
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 91125, "加载榜单详情失败", err)
		}
		return
	}
	Success(c, result)
}
