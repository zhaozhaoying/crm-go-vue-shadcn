package handler

import (
	"backend/internal/service"
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
// @Param period query string false "周期类型: all=总排名, month=月排名, week=周排名"
// @Success 200 {object} APIResponse
// @Router /api/v1/ranking-leaderboard [get]
func (h *RankingLeaderboardHandler) ListRankingLeaderboard(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	period := strings.TrimSpace(c.Query("period"))
	if period == "" {
		period = "all"
	}

	result, err := h.service.ListRankingLeaderboard(c.Request.Context(), period)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 91116, "加载排名榜单失败", err)
		return
	}
	Success(c, result)
}
