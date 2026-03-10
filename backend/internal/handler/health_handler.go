package handler

import (
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	serviceName string
}

type HealthPayload struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
}

func NewHealthHandler(serviceName string) *HealthHandler {
	return &HealthHandler{
		serviceName: serviceName,
	}
}

// GetHealth godoc
// @Summary     健康检查
// @Description 检查服务是否正常运行
// @Tags        health
// @Produce     json
// @Success     200 {object} APIResponse{data=HealthPayload}
// @Router      /api/health [get]
func (h *HealthHandler) GetHealth(c *gin.Context) {
	Success(c, HealthPayload{
		Status:    "ok",
		Service:   h.serviceName,
		Timestamp: time.Now().UTC(),
	})
}
