package handler

import (
	"backend/internal/model"
	"backend/internal/repository"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	activityLogRepo  *repository.ActivityLogRepository
	notificationRepo repository.NotificationRepository
}

type NotificationMarkReadRequest struct {
	Keys []string `json:"keys"`
}

func NewNotificationHandler(activityLogRepo *repository.ActivityLogRepository, notificationRepo repository.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{
		activityLogRepo:  activityLogRepo,
		notificationRepo: notificationRepo,
	}
}

// ListActivityLogs godoc
// @Summary     获取通知中心活动日志
// @Tags        notifications
// @Produce     json
// @Security    BearerAuth
// @Param       limit query int false "返回条数，默认50，最大200"
// @Success     200 {object} APIResponse{data=[]model.ActivityLog}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/notifications/activity-logs [get]
func (h *NotificationHandler) ListActivityLogs(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	role := currentUserRole(c)
	showAll := isGlobalRole(role)

	logs, err := h.activityLogRepo.ListRecent(c.Request.Context(), limit, userID, showAll)
	if err != nil {
		Error(c, http.StatusInternalServerError, 80001, "failed to list activity logs")
		return
	}

	Success(c, logs)
}

// ListReadKeys godoc
// @Summary     获取已读通知键
// @Tags        notifications
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Router      /api/v1/notifications/read-keys [get]
func (h *NotificationHandler) ListReadKeys(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}

	keys, err := h.notificationRepo.ListReadKeys(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 80002, "failed to list read keys")
		return
	}
	Success(c, gin.H{"keys": keys})
}

// MarkAsRead godoc
// @Summary     批量标记通知已读
// @Tags        notifications
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body NotificationMarkReadRequest true "通知键列表"
// @Success     200 {object} APIResponse
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Router      /api/v1/notifications/mark-read [post]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}

	var req NotificationMarkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 80003, "invalid request body")
		return
	}
	if len(req.Keys) == 0 {
		Success(c, nil)
		return
	}

	if err := h.notificationRepo.MarkAsRead(c.Request.Context(), userID, req.Keys); err != nil {
		Error(c, http.StatusInternalServerError, 80004, "failed to mark as read")
		return
	}
	Success(c, nil)
}

func isGlobalRole(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin", "管理员", "finance", "finance_manager", "财务", "财务经理":
		return true
	default:
		return false
	}
}

func (h *NotificationHandler) CreateNotificationRead(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}

	var req struct {
		Key string `json:"key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Key == "" {
		Error(c, http.StatusBadRequest, 80003, "invalid request body")
		return
	}

	if err := h.notificationRepo.MarkAsRead(c.Request.Context(), userID, []string{req.Key}); err != nil {
		Error(c, http.StatusInternalServerError, 80004, "failed to mark as read")
		return
	}
	Success(c, nil)
}

// UnreadCount godoc
// @Summary     获取未读通知数量
// @Tags        notifications
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Router      /api/v1/notifications/unread-count [get]
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}

	role := currentUserRole(c)
	showAll := isGlobalRole(role)

	logs, err := h.activityLogRepo.ListRecent(c.Request.Context(), 200, userID, showAll)
	if err != nil {
		Error(c, http.StatusInternalServerError, 80001, "failed to list activity logs")
		return
	}

	readKeys, err := h.notificationRepo.ListReadKeys(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 80002, "failed to list read keys")
		return
	}

	readSet := make(map[string]struct{}, len(readKeys))
	for _, key := range readKeys {
		readSet[key] = struct{}{}
	}

	unread := 0
	for _, log := range logs {
		key := model.ActivityLogNotificationKey(log)
		if _, ok := readSet[key]; !ok {
			unread++
		}
	}

	Success(c, gin.H{"count": unread})
}
