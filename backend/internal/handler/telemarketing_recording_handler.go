package handler

import (
	"backend/internal/authctx"
	"backend/internal/model"
	"backend/internal/service"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type TelemarketingRecordingHandler struct {
	service      *service.TelemarketingRecordingService
	authProvider authctx.Provider
}

func NewTelemarketingRecordingHandler(
	service *service.TelemarketingRecordingService,
	authProvider authctx.Provider,
) *TelemarketingRecordingHandler {
	return &TelemarketingRecordingHandler{
		service:      service,
		authProvider: authProvider,
	}
}

type SyncTelemarketingRecordingsRequest struct {
	PageSize   int    `json:"pageSize"`
	TimePeriod string `json:"timePeriod"`
}

// List godoc
// @Summary 获取电销录音库列表
// @Tags 电销录音库
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "关键词：工号/姓名/号码"
// @Param startDate query string false "开始日期，格式：YYYY-MM-DD"
// @Param endDate query string false "结束日期，格式：YYYY-MM-DD"
// @Param minDuration query int false "最小时长（秒）"
// @Param maxDuration query int false "最大时长（秒）"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} APIResponse
// @Router /api/v1/telemarketing-recordings [get]
func (h *TelemarketingRecordingHandler) List(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	role := currentUserRole(c)
	if !canViewCallRecordings(role) {
		Error(c, http.StatusForbidden, 91230, "当前角色无权查看电销录音库")
		return
	}

	showAll := canViewAllCallRecordings(role)
	viewerMihuaWorkNumber, err := h.resolveViewerMihuaWorkNumber(c)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 91231, "加载当前用户信息失败", err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	minDuration, _ := strconv.Atoi(strings.TrimSpace(c.Query("minDuration")))
	maxDuration, _ := strconv.Atoi(strings.TrimSpace(c.Query("maxDuration")))

	result, err := h.service.List(c.Request.Context(), model.TelemarketingRecordingListFilter{
		ShowAll:           showAll,
		ViewerMihuaWorkNo: viewerMihuaWorkNumber,
		Keyword:           strings.TrimSpace(c.Query("keyword")),
		StartDate:         strings.TrimSpace(c.Query("startDate")),
		EndDate:           strings.TrimSpace(c.Query("endDate")),
		MinDuration:       minDuration,
		MaxDuration:       maxDuration,
		Page:              page,
		PageSize:          pageSize,
	})
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 91232, "加载电销录音库失败", err)
		return
	}
	Success(c, result)
}

// GetDetail godoc
// @Summary 获取电销录音详情
// @Tags 电销录音库
// @Produce json
// @Security BearerAuth
// @Param id path string true "录音ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/telemarketing-recordings/{id} [get]
func (h *TelemarketingRecordingHandler) GetDetail(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	role := currentUserRole(c)
	if !canViewCallRecordings(role) {
		Error(c, http.StatusForbidden, 91233, "当前角色无权查看电销录音详情")
		return
	}

	showAll := canViewAllCallRecordings(role)
	viewerMihuaWorkNumber, err := h.resolveViewerMihuaWorkNumber(c)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 91234, "加载当前用户信息失败", err)
		return
	}

	result, err := h.service.GetDetail(
		c.Request.Context(),
		c.Param("id"),
		showAll,
		viewerMihuaWorkNumber,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTelemarketingRecordingNotFound):
			Error(c, http.StatusNotFound, 91235, "未找到对应的电销录音")
		case errors.Is(err, service.ErrMiHuaTelemarketingRecordingConfigNeeded):
			Error(c, http.StatusBadRequest, 91236, "缺少米话电销录音配置，请先在系统设置中配置 MIHUA_CALL_RECORD_TOKEN，并确认 backend/.env 中已配置 MIHUA_TELEMARKETING_RECORDING_LIST_URL 和 MIHUA_CALL_RECORD_SOURCE_ORIGIN")
		case errors.Is(err, service.ErrMiHuaTelemarketingRecordingRequestFail):
			ErrorWithDetail(c, http.StatusBadGateway, 91237, "请求米话录音播放地址失败", err)
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 91238, "加载电销录音详情失败", err)
		}
		return
	}

	Success(c, result)
}

// Sync godoc
// @Summary 同步电销录音库
// @Tags 电销录音库
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} APIResponse
// @Router /api/v1/telemarketing-recordings/sync [post]
func (h *TelemarketingRecordingHandler) Sync(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	if !canImportCallRecordings(currentUserRole(c)) {
		Error(c, http.StatusForbidden, 91239, "当前角色无权同步电销录音库")
		return
	}

	var req SyncTelemarketingRecordingsRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorWithDetail(c, http.StatusBadRequest, 91240, "请求参数错误", err)
			return
		}
	}

	result, err := h.service.Sync(c.Request.Context(), service.SyncTelemarketingRecordingsInput{
		PageSize:   req.PageSize,
		TimePeriod: strings.TrimSpace(req.TimePeriod),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMiHuaTelemarketingRecordingConfigNeeded):
			Error(c, http.StatusBadRequest, 91241, "缺少米话电销录音配置，请先在系统设置中配置 MIHUA_CALL_RECORD_TOKEN，并确认 backend/.env 中已配置 MIHUA_TELEMARKETING_RECORDING_LIST_URL 和 MIHUA_CALL_RECORD_SOURCE_ORIGIN")
		case errors.Is(err, service.ErrMiHuaTelemarketingRecordingRequestFail):
			ErrorWithDetail(c, http.StatusBadGateway, 91242, "请求米话电销录音列表失败", err)
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 91243, "同步电销录音库失败", err)
		}
		return
	}

	Success(c, result)
}

func (h *TelemarketingRecordingHandler) resolveViewerMihuaWorkNumber(c *gin.Context) (string, error) {
	if h == nil || h.authProvider == nil {
		return "", nil
	}
	user, err := h.authProvider.GetCurrentUser(c.Request.Context(), c)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", nil
	}
	return strings.TrimSpace(user.MihuaWorkNumber), nil
}
