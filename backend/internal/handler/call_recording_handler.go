package handler

import (
	"backend/internal/authctx"
	"backend/internal/model"
	"backend/internal/service"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CallRecordingHandler struct {
	service      *service.CallRecordingService
	syncService  *service.CallRecordingSyncService
	authProvider authctx.Provider
	audioClient  *http.Client
}

func NewCallRecordingHandler(
	service *service.CallRecordingService,
	syncService *service.CallRecordingSyncService,
	authProvider authctx.Provider,
) *CallRecordingHandler {
	return &CallRecordingHandler{
		service:      service,
		syncService:  syncService,
		authProvider: authProvider,
		audioClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

type ImportCallRecordingsRequest struct {
	Items []ImportCallRecordingItem `json:"items"`
}

type SyncCallRecordingsRequest struct {
	StartTimeBegin  string `json:"startTimeBegin"`
	StartTimeFinish string `json:"startTimeFinish"`
	MinTime         string `json:"minTime"`
	Limit           int    `json:"limit"`
}

type ImportCallRecordingItem struct {
	AgentCode        int64   `json:"agentCode"`
	CallStatus       int     `json:"callStatus"`
	CallStatusName   string  `json:"callStatusName"`
	CallType         int     `json:"callType"`
	CalleeAttr       string  `json:"calleeAttr"`
	CallerAttr       string  `json:"callerAttr"`
	CreateTime       int64   `json:"createTime"`
	DeptName         string  `json:"deptName"`
	Duration         int     `json:"duration"`
	EndTime          int64   `json:"endTime"`
	EnterpriseName   string  `json:"enterpriseName"`
	FinishStatus     int     `json:"finishStatus"`
	FinishStatusName string  `json:"finishStatusName"`
	Handle           int     `json:"handle"`
	ID               string  `json:"id"`
	InterfaceID      string  `json:"interfaceId"`
	InterfaceName    string  `json:"interfaceName"`
	LineName         string  `json:"lineName"`
	Mobile           string  `json:"mobile"`
	Mode             int     `json:"mode"`
	MoveBatchCode    *string `json:"moveBatchCode"`
	OctCustomerID    *string `json:"octCustomerId"`
	Phone            string  `json:"phone"`
	Postage          float64 `json:"postage"`
	PreRecordURL     string  `json:"preRecordUrl"`
	RealName         string  `json:"realName"`
	StartTime        int64   `json:"startTime"`
	Status           int     `json:"status"`
	TelA             string  `json:"telA"`
	TelB             string  `json:"telB"`
	TelX             string  `json:"telX"`
	TenantCode       string  `json:"tenantCode"`
	UpdateTime       int64   `json:"updateTime"`
	UserID           string  `json:"userId"`
	WorkNum          *string `json:"workNum"`
}

func (h *CallRecordingHandler) List(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	minDuration, _ := strconv.Atoi(strings.TrimSpace(c.Query("minDuration")))
	maxDuration, _ := strconv.Atoi(strings.TrimSpace(c.Query("maxDuration")))
	result, err := h.service.List(c.Request.Context(), model.CallRecordingListFilter{
		ShowAll:                 true,
		ViewerHanghangCRMMobile: "",
		Keyword:                 strings.TrimSpace(c.Query("keyword")),
		MinDuration:             minDuration,
		MaxDuration:             maxDuration,
		Page:                    page,
		PageSize:                pageSize,
	})
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 91203, "加载通话录音失败", err)
		return
	}
	Success(c, result)
}

func (h *CallRecordingHandler) StreamAudio(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	recording, err := h.service.GetByID(c.Request.Context(), c.Param("id"), true, "")
	if err != nil {
		if errors.Is(err, service.ErrCallRecordingNotFound) {
			Error(c, http.StatusNotFound, 91206, "未找到通话录音")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 91207, "加载通话录音失败", err)
		return
	}

	audioURL := strings.TrimSpace(recording.PreRecordURL)
	if audioURL == "" {
		Error(c, http.StatusNotFound, 91208, "当前记录没有录音地址")
		return
	}

	parsedURL, err := url.Parse(audioURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		Error(c, http.StatusBadRequest, 91209, "录音地址无效")
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 91210, "创建录音请求失败", err)
		return
	}
	req.Header.Set("User-Agent", "crm-go-vue-shadcn/audio-proxy")

	resp, err := h.audioClient.Do(req)
	if err != nil {
		ErrorWithDetail(c, http.StatusBadGateway, 91211, "加载录音文件失败", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		Error(c, http.StatusBadGateway, 91212, "远程录音文件不可用")
		return
	}

	copyHeaderIfPresent(c.Writer.Header(), resp.Header, "Content-Type")
	copyHeaderIfPresent(c.Writer.Header(), resp.Header, "Content-Length")
	copyHeaderIfPresent(c.Writer.Header(), resp.Header, "Accept-Ranges")
	copyHeaderIfPresent(c.Writer.Header(), resp.Header, "Cache-Control")
	if c.Writer.Header().Get("Content-Type") == "" {
		c.Writer.Header().Set("Content-Type", "audio/wav")
	}
	c.Status(resp.StatusCode)
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		_ = c.Error(err)
	}
}

func (h *CallRecordingHandler) Import(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	if !canImportCallRecordings(currentUserRole(c)) {
		Error(c, http.StatusForbidden, 91213, "当前角色无权导入通话录音")
		return
	}

	var req ImportCallRecordingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 91214, "请求参数错误", err)
		return
	}
	if len(req.Items) == 0 {
		Error(c, http.StatusBadRequest, 91215, "请至少提供一条通话录音数据")
		return
	}

	items := make([]model.CallRecordingUpsertInput, 0, len(req.Items))
	for _, item := range req.Items {
		if strings.TrimSpace(item.ID) == "" {
			Error(c, http.StatusBadRequest, 91216, "通话录音ID不能为空")
			return
		}
		items = append(items, model.CallRecordingUpsertInput{
			ID:               item.ID,
			AgentCode:        item.AgentCode,
			CallStatus:       item.CallStatus,
			CallStatusName:   item.CallStatusName,
			CallType:         item.CallType,
			CalleeAttr:       item.CalleeAttr,
			CallerAttr:       item.CallerAttr,
			CreateTime:       item.CreateTime,
			DeptName:         item.DeptName,
			Duration:         item.Duration,
			EndTime:          item.EndTime,
			EnterpriseName:   item.EnterpriseName,
			FinishStatus:     item.FinishStatus,
			FinishStatusName: item.FinishStatusName,
			Handle:           item.Handle,
			InterfaceID:      item.InterfaceID,
			InterfaceName:    item.InterfaceName,
			LineName:         item.LineName,
			Mobile:           item.Mobile,
			Mode:             item.Mode,
			MoveBatchCode:    item.MoveBatchCode,
			OctCustomerID:    item.OctCustomerID,
			Phone:            item.Phone,
			Postage:          item.Postage,
			PreRecordURL:     item.PreRecordURL,
			RealName:         item.RealName,
			StartTime:        item.StartTime,
			Status:           item.Status,
			TelA:             item.TelA,
			TelB:             item.TelB,
			TelX:             item.TelX,
			TenantCode:       item.TenantCode,
			UpdateTime:       item.UpdateTime,
			UserID:           item.UserID,
			WorkNum:          item.WorkNum,
		})
	}

	saved, err := h.service.UpsertBatch(c.Request.Context(), items)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 91217, "导入通话录音失败", err)
		return
	}
	Success(c, gin.H{"saved": len(saved)})
}

func (h *CallRecordingHandler) Sync(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	if h.syncService == nil {
		Error(c, http.StatusInternalServerError, 91219, "通话录音同步服务未配置")
		return
	}

	var req SyncCallRecordingsRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorWithDetail(c, http.StatusBadRequest, 91220, "请求参数错误", err)
			return
		}
	}

	result, err := h.syncService.Sync(c.Request.Context(), service.SyncCallRecordingsInput{
		StartTimeBegin:  req.StartTimeBegin,
		StartTimeFinish: req.StartTimeFinish,
		MinTime:         req.MinTime,
		Limit:           req.Limit,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFeigeCallRecordingCookieRequired):
			Error(c, http.StatusBadRequest, 91221, "缺少飞鸽通话录音 cookie，请先在 backend/.env 中配置 FEIGE_CALL_RECORDING_COOKIE")
		case errors.Is(err, service.ErrFeigeCallRecordingInvalidDate):
			Error(c, http.StatusBadRequest, 91222, "通话录音同步日期格式错误，应为 YYYY-MM-DD")
		case errors.Is(err, service.ErrFeigeCallRecordingRequestFailed):
			ErrorWithDetail(c, http.StatusBadGateway, 91223, "请求飞鸽通话录音接口失败", err)
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 91224, "同步通话录音失败", err)
		}
		return
	}

	Success(c, result)
}

func (h *CallRecordingHandler) resolveViewerHanghangCRMMobile(c *gin.Context) (string, error) {
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
	return strings.TrimSpace(user.HanghangCRMMobile), nil
}

func canViewCallRecordings(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin", "管理员",
		"finance", "finance_manager", "财务", "财务经理",
		"sales_director", "销售总监",
		"sales_manager", "销售经理",
		"sales_staff", "销售员工",
		"sales_inside", "sale_inside", "inside销售", "电销员工",
		"sales_outside", "sale_outside", "outside销售":
		return true
	default:
		return false
	}
}

func canViewAllCallRecordings(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin", "管理员", "finance", "finance_manager", "财务", "财务经理":
		return true
	default:
		return false
	}
}

func canImportCallRecordings(role string) bool {
	return canViewAllCallRecordings(role)
}

func copyHeaderIfPresent(dst http.Header, src http.Header, key string) {
	value := strings.TrimSpace(src.Get(key))
	if value == "" {
		return
	}
	dst.Set(key, value)
}
