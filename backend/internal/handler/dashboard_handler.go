package handler

import (
	"backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

// GetOverview godoc
// @Summary     获取仪表盘概览
// @Tags        dashboard
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} APIResponse{data=model.DashboardOverview}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/dashboard/overview [get]
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	overview, err := h.service.GetOverview(c.Request.Context(), userID, currentUserRole(c))
	if err != nil {
		Error(c, http.StatusInternalServerError, 70001, "failed to load dashboard overview")
		return
	}
	Success(c, overview)
}
