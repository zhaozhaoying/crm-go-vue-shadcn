package handler

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type ExternalCompanySearchHandler struct {
	service        service.ExternalCompanySearchService
	hub            *service.ExternalCompanySearchHub
	frontendOrigin string
}

type CreateExternalCompanySearchTasksRequest struct {
	Platforms     []int  `json:"platforms" binding:"required"`
	Keyword       string `json:"keyword" binding:"required"`
	RegionKeyword string `json:"regionKeyword"`
	PageLimit     int    `json:"pageLimit"`
	TargetCount   int    `json:"targetCount"`
	Priority      int    `json:"priority"`
	SearchOptions any    `json:"searchOptions"`
}

type ExternalCompanySearchTasksResponse struct {
	Items []model.ExternalCompanySearchTask `json:"items"`
}

type ExternalCompanySearchCancelResponse struct {
	TaskID int64 `json:"taskId"`
}

func NewExternalCompanySearchHandler(
	service service.ExternalCompanySearchService,
	hub *service.ExternalCompanySearchHub,
	frontendOrigin string,
) *ExternalCompanySearchHandler {
	return &ExternalCompanySearchHandler{service: service, hub: hub, frontendOrigin: strings.TrimSpace(frontendOrigin)}
}

// CreateTasks godoc
// @Summary     创建外部企业抓取任务
// @Tags        external-company-search
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body CreateExternalCompanySearchTasksRequest true "抓取任务"
// @Success     200 {object} APIResponse{data=handler.ExternalCompanySearchTasksResponse}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Router      /api/v1/external-company-search/tasks [post]
func (h *ExternalCompanySearchHandler) CreateTasks(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 80001, "invalid token claims")
		return
	}
	var req CreateExternalCompanySearchTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 80002, "invalid request body")
		return
	}
	searchOptionsJSON := ""
	if req.SearchOptions != nil {
		data, err := json.Marshal(req.SearchOptions)
		if err != nil {
			Error(c, http.StatusBadRequest, 80003, "invalid search options")
			return
		}
		searchOptionsJSON = string(data)
	}
	items, err := h.service.CreateTasks(c.Request.Context(), model.ExternalCompanySearchTaskCreateInput{
		Platforms:     req.Platforms,
		Keyword:       req.Keyword,
		RegionKeyword: req.RegionKeyword,
		PageLimit:     req.PageLimit,
		TargetCount:   req.TargetCount,
		Priority:      req.Priority,
		SearchOptions: searchOptionsJSON,
		CreatedBy:     userID,
	}, currentUserRole(c))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, gin.H{"items": items})
}

// ListTasks godoc
// @Summary     获取外部企业抓取任务列表
// @Tags        external-company-search
// @Produce     json
// @Security    BearerAuth
// @Param       platform query int false "平台"
// @Param       status query int false "状态"
// @Param       keyword query string false "关键词"
// @Param       page query int false "页码"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.ExternalCompanySearchTaskListResult}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/external-company-search/tasks [get]
func (h *ExternalCompanySearchHandler) ListTasks(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 80001, "invalid token claims")
		return
	}
	filter := model.ExternalCompanySearchTaskListFilter{
		Platform:  parsePositiveInt(c.Query("platform")),
		Status:    parsePositiveInt(c.Query("status")),
		Keyword:   strings.TrimSpace(c.Query("keyword")),
		CreatedBy: userID,
		Page:      parsePositiveIntWithDefault(c.Query("page"), 1),
		PageSize:  parsePositiveIntWithDefault(c.Query("pageSize"), 20),
	}
	result, err := h.service.ListTasks(c.Request.Context(), filter, currentUserRole(c))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, result)
}

// GetTask godoc
// @Summary     获取外部企业抓取任务详情
// @Tags        external-company-search
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "任务ID"
// @Success     200 {object} APIResponse{data=model.ExternalCompanySearchTask}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/external-company-search/tasks/{id} [get]
func (h *ExternalCompanySearchHandler) GetTask(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 80001, "invalid token claims")
		return
	}
	taskID, ok := parseIDParam(c, "id")
	if !ok {
		Error(c, http.StatusBadRequest, 80004, "invalid task id")
		return
	}
	task, err := h.service.GetTask(c.Request.Context(), taskID, userID, currentUserRole(c))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, task)
}

// ListResults godoc
// @Summary     获取任务结果列表
// @Tags        external-company-search
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "任务ID"
// @Param       search query string false "搜索关键词"
// @Param       platform query int false "平台"
// @Param       newOnly query bool false "是否只看新发掘"
// @Param       page query int false "页码"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.ExternalCompanySearchResultListResult}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/external-company-search/tasks/{id}/results [get]
func (h *ExternalCompanySearchHandler) ListResults(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 80001, "invalid token claims")
		return
	}
	taskID, ok := parseIDParam(c, "id")
	if !ok {
		Error(c, http.StatusBadRequest, 80004, "invalid task id")
		return
	}
	result, err := h.service.ListTaskResults(c.Request.Context(), taskID, userID, currentUserRole(c), model.ExternalCompanySearchResultListFilter{
		Search:   strings.TrimSpace(c.Query("search")),
		Platform: parsePositiveInt(c.Query("platform")),
		NewOnly:  parseBoolQuery(c.Query("newOnly")),
		Page:     parsePositiveIntWithDefault(c.Query("page"), 1),
		PageSize: parsePositiveIntWithDefault(c.Query("pageSize"), 20),
	})
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, result)
}

// ListAllResults godoc
// @Summary     获取全部任务结果列表
// @Tags        external-company-search
// @Produce     json
// @Security    BearerAuth
// @Param       search query string false "搜索关键词"
// @Param       platform query int false "平台"
// @Param       newOnly query bool false "是否只看新发掘"
// @Param       page query int false "页码"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.ExternalCompanySearchResultListResult}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/external-company-search/results [get]
func (h *ExternalCompanySearchHandler) ListAllResults(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 80001, "invalid token claims")
		return
	}
	result, err := h.service.ListResults(c.Request.Context(), userID, currentUserRole(c), model.ExternalCompanySearchResultListFilter{
		Search:   strings.TrimSpace(c.Query("search")),
		Platform: parsePositiveInt(c.Query("platform")),
		NewOnly:  parseBoolQuery(c.Query("newOnly")),
		Page:     parsePositiveIntWithDefault(c.Query("page"), 1),
		PageSize: parsePositiveIntWithDefault(c.Query("pageSize"), 20),
	})
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, result)
}

// ListEvents godoc
// @Summary     获取任务事件列表
// @Tags        external-company-search
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "任务ID"
// @Param       afterSeq query int false "事件序号起点"
// @Param       limit query int false "返回条数"
// @Success     200 {object} APIResponse{data=model.ExternalCompanySearchEventListResult}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/external-company-search/tasks/{id}/events [get]
func (h *ExternalCompanySearchHandler) ListEvents(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 80001, "invalid token claims")
		return
	}
	taskID, ok := parseIDParam(c, "id")
	if !ok {
		Error(c, http.StatusBadRequest, 80004, "invalid task id")
		return
	}
	result, err := h.service.ListTaskEvents(c.Request.Context(), taskID, userID, currentUserRole(c), parsePositiveInt64(c.Query("afterSeq")), parsePositiveIntWithDefault(c.Query("limit"), 100))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, result)
}

// CancelTask godoc
// @Summary     取消抓取任务
// @Tags        external-company-search
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "任务ID"
// @Success     200 {object} APIResponse{data=handler.ExternalCompanySearchCancelResponse}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/external-company-search/tasks/{id}/cancel [post]
func (h *ExternalCompanySearchHandler) CancelTask(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 80001, "invalid token claims")
		return
	}
	taskID, ok := parseIDParam(c, "id")
	if !ok {
		Error(c, http.StatusBadRequest, 80004, "invalid task id")
		return
	}
	if err := h.service.CancelTask(c.Request.Context(), taskID, userID, currentUserRole(c)); err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, gin.H{"taskId": taskID})
}

// StreamTask godoc
// @Summary     订阅任务实时事件流
// @Tags        external-company-search
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "任务ID"
// @Param       afterSeq query int false "事件序号起点"
// @Success     200 {object} APIResponse
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/external-company-search/tasks/{id}/stream [get]
func (h *ExternalCompanySearchHandler) StreamTask(c *gin.Context) {
	if !h.allowWebSocketOrigin(c.Request.Header.Get("Origin")) {
		Error(c, http.StatusForbidden, 80005, "websocket origin not allowed")
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 80001, "invalid token claims")
		return
	}
	taskID, ok := parseIDParam(c, "id")
	if !ok {
		Error(c, http.StatusBadRequest, 80004, "invalid task id")
		return
	}
	if _, err := h.service.GetTask(c.Request.Context(), taskID, userID, currentUserRole(c)); err != nil {
		h.handleServiceError(c, err)
		return
	}
	afterSeq := parsePositiveInt64(c.Query("afterSeq"))
	server := websocket.Server{Handler: websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()
		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		history, err := h.service.ListTaskEvents(ctx, taskID, userID, currentUserRole(c), afterSeq, 500)
		if err != nil {
			_ = websocket.JSON.Send(conn, gin.H{"error": err.Error()})
			return
		}
		for _, event := range history.Items {
			if err := websocket.JSON.Send(conn, event); err != nil {
				return
			}
			afterSeq = event.SeqNo
		}

		eventCh, unsubscribe := h.hub.Subscribe(taskID)
		defer unsubscribe()

		readDone := make(chan struct{})
		go func() {
			defer close(readDone)
			for {
				var message string
				if err := websocket.Message.Receive(conn, &message); err != nil {
					return
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-readDone:
				return
			case event, ok := <-eventCh:
				if !ok {
					return
				}
				if event.SeqNo <= afterSeq {
					continue
				}
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := websocket.JSON.Send(conn, event); err != nil {
					return
				}
				afterSeq = event.SeqNo
			}
		}
	})}
	server.ServeHTTP(c.Writer, c.Request)
}

func (h *ExternalCompanySearchHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrExternalCompanySearchKeywordRequired):
		Error(c, http.StatusBadRequest, 80006, "keyword is required")
	case errors.Is(err, service.ErrExternalCompanySearchPlatformRequired):
		Error(c, http.StatusBadRequest, 80007, "platform is required")
	case errors.Is(err, service.ErrExternalCompanySearchPlatformUnsupported):
		Error(c, http.StatusBadRequest, 80008, err.Error())
	case errors.Is(err, repository.ErrExternalCompanySearchTaskNotFound):
		Error(c, http.StatusNotFound, 80009, "task not found")
	case errors.Is(err, service.ErrExternalCompanySearchTaskForbidden):
		Error(c, http.StatusForbidden, 80010, "task access forbidden")
	default:
		Error(c, http.StatusInternalServerError, 80099, "external company search request failed")
	}
}

func (h *ExternalCompanySearchHandler) allowWebSocketOrigin(origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" || strings.TrimSpace(h.frontendOrigin) == "" {
		return true
	}
	current, err := url.Parse(origin)
	if err != nil {
		return false
	}
	for _, rawAllowed := range strings.Split(h.frontendOrigin, ",") {
		allowedOrigin := strings.TrimSpace(rawAllowed)
		if allowedOrigin == "" {
			continue
		}
		allowed, parseErr := url.Parse(allowedOrigin)
		if parseErr != nil {
			continue
		}
		if strings.EqualFold(allowed.Scheme, current.Scheme) && strings.EqualFold(allowed.Host, current.Host) {
			return true
		}
	}
	return false
}

func parseIDParam(c *gin.Context, key string) (int64, bool) {
	value, err := strconv.ParseInt(strings.TrimSpace(c.Param(key)), 10, 64)
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

func parsePositiveInt(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return 0
	}
	return value
}

func parsePositiveIntWithDefault(raw string, defaultValue int) int {
	value := parsePositiveInt(raw)
	if value <= 0 {
		return defaultValue
	}
	return value
}

func parsePositiveInt64(raw string) int64 {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return 0
	}
	return value
}
