package handler

import (
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	serviceName string
	version     string
	gitCommit   string
	buildTime   string
	startedAt   time.Time
}

type HealthPayload struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	GitCommit string    `json:"gitCommit"`
	BuildTime string    `json:"buildTime"`
	StartedAt time.Time `json:"startedAt"`
	Timestamp time.Time `json:"timestamp"`
}

func NewHealthHandler(serviceName, version, gitCommit, buildTime string) *HealthHandler {
	return &HealthHandler{
		serviceName: serviceName,
		version:     version,
		gitCommit:   gitCommit,
		buildTime:   buildTime,
		startedAt:   time.Now().UTC(),
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
		Version:   h.version,
		GitCommit: h.gitCommit,
		BuildTime: h.buildTime,
		StartedAt: h.startedAt,
		Timestamp: time.Now().UTC(),
	})
}
