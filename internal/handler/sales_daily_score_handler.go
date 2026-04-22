package handler

import (
	"backend/internal/service"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SalesDailyScoreHandler struct {
	service service.SalesDailyScoreService
}

func NewSalesDailyScoreHandler(service service.SalesDailyScoreService) *SalesDailyScoreHandler {
	return &SalesDailyScoreHandler{service: service}
}

// ListDailyRankings godoc
// @Summary 获取每日排名列表
// @Tags 每日排名
// @Produce json
// @Security BearerAuth
// @Param scoreDate query string false "积分日期，格式：YYYY-MM-DD"
// @Success 200 {object} APIResponse
// @Router /api/v1/sales-daily-scores [get]
func (h *SalesDailyScoreHandler) ListDailyRankings(c *gin.Context) {
	actorUserID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	result, err := h.service.ListDailyRankings(
		c.Request.Context(),
		strings.TrimSpace(c.Query("scoreDate")),
		actorUserID,
		currentUserRole(c),
	)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 91103, "加载每日排名失败", err)
		return
	}
	Success(c, result)
}

// GetDailyScoreDetail godoc
// @Summary 获取每日积分详情
// @Tags 每日排名
// @Produce json
// @Security BearerAuth
// @Param userId path int true "用户ID"
// @Param scoreDate query string false "积分日期，格式：YYYY-MM-DD"
// @Success 200 {object} APIResponse
// @Router /api/v1/sales-daily-scores/{userId} [get]
func (h *SalesDailyScoreHandler) GetDailyScoreDetail(c *gin.Context) {
	actorUserID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil || userID <= 0 {
		Error(c, http.StatusBadRequest, 91105, "无效的用户ID")
		return
	}

	result, err := h.service.GetDailyScoreDetail(
		c.Request.Context(),
		strings.TrimSpace(c.Query("scoreDate")),
		userID,
		actorUserID,
		currentUserRole(c),
	)
	if err != nil {
		if errors.Is(err, service.ErrSalesDailyScoreNotFound) {
			Error(c, http.StatusNotFound, 91106, "未找到对应日期的积分详情")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 91107, "加载积分详情失败", err)
		return
	}
	Success(c, result)
}
