package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

type FollowRecordHandler struct {
	service *service.FollowRecordService
}

func NewFollowRecordHandler(service *service.FollowRecordService) *FollowRecordHandler {
	return &FollowRecordHandler{service: service}
}

type CreateFollowRecordRequest struct {
	CustomerID       int64  `json:"customerId" binding:"required"`
	Content          string `json:"content" binding:"required"`
	NextFollowTime   string `json:"nextFollowTime"`
	AppointmentTime  string `json:"appointmentTime"`
	ShootingTime     string `json:"shootingTime"`
	CustomerLevelID  int    `json:"customerLevelId"`
	CustomerSourceID int    `json:"customerSourceId"`
	FollowMethodID   int    `json:"followMethodId"`
}

// CreateOperationFollowRecord 创建运营跟进记录
// @Summary 创建运营跟进记录
// @Tags 跟进记录
// @Accept json
// @Produce json
// @Param request body CreateFollowRecordRequest true "创建运营跟进记录"
// @Success 200 {object} APIResponse
// @Router /api/v1/operation-follow-records [post]
func (h *FollowRecordHandler) CreateOperationFollowRecord(c *gin.Context) {
	var req CreateFollowRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 11001, "请求参数错误", err)
		return
	}

	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	// 校验跟进内容
	content := strings.TrimSpace(req.Content)
	if len(content) < 10 {
		Error(c, http.StatusBadRequest, 11002, "跟进内容必须至少10个字")
		return
	}
	if strings.Contains(content, "跟进") {
		Error(c, http.StatusBadRequest, 11003, `跟进内容不能包含"跟进"两个字`)
		return
	}
	// 检查是否包含英文字母
	for _, r := range content {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			Error(c, http.StatusBadRequest, 11004, "跟进内容不能包含英文单词")
			return
		}
	}

	if req.FollowMethodID <= 0 {
		Error(c, http.StatusBadRequest, 11005, "请选择跟进方式")
		return
	}

	nextFollowTime, err := parseFollowTime(req.NextFollowTime)
	if err != nil {
		Error(c, http.StatusBadRequest, 11006, "无效的下次跟进时间格式")
		return
	}

	appointmentTime, err := parseFollowTime(req.AppointmentTime)
	if err != nil {
		Error(c, http.StatusBadRequest, 11007, "无效的约见时间格式")
		return
	}

	shootingTime, err := parseFollowTime(req.ShootingTime)
	if err != nil {
		Error(c, http.StatusBadRequest, 11008, "无效的拍摄时间格式")
		return
	}

	input := model.FollowRecordCreateInput{
		CustomerID:       req.CustomerID,
		Content:          req.Content,
		NextFollowTime:   nextFollowTime,
		AppointmentTime:  appointmentTime,
		ShootingTime:     shootingTime,
		CustomerLevelID:  req.CustomerLevelID,
		CustomerSourceID: req.CustomerSourceID,
		FollowMethodID:   req.FollowMethodID,
		OperatorUserID:   userID,
	}

	id, err := h.service.CreateOperationFollowRecord(input)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 11009, "创建运营跟进记录失败", err)
		return
	}

	Success(c, gin.H{"id": id})
}

// CreateSalesFollowRecord 创建销售跟进记录
// @Summary 创建销售跟进记录
// @Tags 跟进记录
// @Accept json
// @Produce json
// @Param request body CreateFollowRecordRequest true "创建销售跟进记录"
// @Success 200 {object} APIResponse
// @Router /api/v1/sales-follow-records [post]
func (h *FollowRecordHandler) CreateSalesFollowRecord(c *gin.Context) {
	var req CreateFollowRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 11001, "请求参数错误", err)
		return
	}

	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	// 校验跟进内容
	content := strings.TrimSpace(req.Content)
	if len(content) < 10 {
		Error(c, http.StatusBadRequest, 11002, "跟进内容必须至少10个字")
		return
	}
	if strings.Contains(content, "跟进") {
		Error(c, http.StatusBadRequest, 11003, `跟进内容不能包含"跟进"两个字`)
		return
	}
	// 检查是否包含英文字母
	for _, r := range content {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			Error(c, http.StatusBadRequest, 11004, "跟进内容不能包含英文单词")
			return
		}
	}

	if req.FollowMethodID <= 0 {
		Error(c, http.StatusBadRequest, 11005, "请选择跟进方式")
		return
	}

	nextFollowTime, err := parseFollowTime(req.NextFollowTime)
	if err != nil {
		Error(c, http.StatusBadRequest, 11006, "无效的下次跟进时间格式")
		return
	}

	input := model.FollowRecordCreateInput{
		CustomerID:       req.CustomerID,
		Content:          req.Content,
		NextFollowTime:   nextFollowTime,
		CustomerLevelID:  req.CustomerLevelID,
		CustomerSourceID: req.CustomerSourceID,
		FollowMethodID:   req.FollowMethodID,
		OperatorUserID:   userID,
	}

	id, err := h.service.CreateSalesFollowRecord(input)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 11010, "创建销售跟进记录失败", err)
		return
	}

	Success(c, gin.H{"id": id})
}

// ListOperationFollowRecords 获取运营跟进记录列表（按客户ID）
// @Summary 获取运营跟进记录列表
// @Tags 跟进记录
// @Accept json
// @Produce json
// @Param customerId query int64 true "客户ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} APIResponse{data=model.OperationFollowRecordListResult}
// @Router /api/v1/operation-follow-records [get]
func (h *FollowRecordHandler) ListOperationFollowRecords(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Query("customerId"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 11011, "无效的客户ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	filter := model.FollowRecordListFilter{
		CustomerID: customerID,
		Page:       page,
		PageSize:   pageSize,
	}

	result, err := h.service.ListOperationFollowRecords(filter)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 11012, "加载运营跟进记录失败", err)
		return
	}

	Success(c, result)
}

// ListAllOperationFollowRecords 获取所有运营跟进记录列表
// @Summary 获取所有运营跟进记录列表
// @Tags 跟进记录
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} APIResponse{data=model.OperationFollowRecordListResult}
// @Router /api/v1/operation-follow-records/all [get]
func (h *FollowRecordHandler) ListAllOperationFollowRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := h.service.ListAllOperationFollowRecords(page, pageSize)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 11013, "加载全部运营跟进记录失败", err)
		return
	}

	Success(c, result)
}

// ListSalesFollowRecords 获取销售跟进记录列表（按客户ID）
// @Summary 获取销售跟进记录列表
// @Tags 跟进记录
// @Accept json
// @Produce json
// @Param customerId query int64 true "客户ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} APIResponse{data=model.SalesFollowRecordListResult}
// @Router /api/v1/sales-follow-records [get]
func (h *FollowRecordHandler) ListSalesFollowRecords(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Query("customerId"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 11011, "无效的客户ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	filter := model.FollowRecordListFilter{
		CustomerID: customerID,
		Page:       page,
		PageSize:   pageSize,
	}

	result, err := h.service.ListSalesFollowRecords(filter)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 11014, "加载销售跟进记录失败", err)
		return
	}

	Success(c, result)
}

// ListAllSalesFollowRecords 获取所有销售跟进记录列表
// @Summary 获取所有销售跟进记录列表
// @Tags 跟进记录
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} APIResponse{data=model.SalesFollowRecordListResult}
// @Router /api/v1/sales-follow-records/all [get]
func (h *FollowRecordHandler) ListAllSalesFollowRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := h.service.ListAllSalesFollowRecords(page, pageSize)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 11015, "加载全部销售跟进记录失败", err)
		return
	}

	Success(c, result)
}

// ListFollowMethods 获取所有跟进方式
// @Summary 获取所有跟进方式
// @Tags 跟进记录
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=[]model.FollowMethod}
// @Router /api/v1/follow-methods [get]
func (h *FollowRecordHandler) ListFollowMethods(c *gin.Context) {
	methods, err := h.service.ListFollowMethods()
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 11016, "加载跟进方式失败", err)
		return
	}

	Success(c, methods)
}

// CreateFollowMethod 创建跟进方式
// @Summary 创建跟进方式
// @Tags 跟进记录
// @Accept json
// @Produce json
// @Param request body model.FollowMethodRequest true "创建跟进方式"
// @Success 200 {object} APIResponse
// @Router /api/v1/follow-methods [post]
func (h *FollowRecordHandler) CreateFollowMethod(c *gin.Context) {
	var req model.FollowMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 11017, "请求参数错误", err)
		return
	}

	id, err := h.service.CreateFollowMethod(req)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 11018, "创建跟进方式失败", err)
		return
	}

	Success(c, gin.H{"id": id})
}

// UpdateFollowMethod 更新跟进方式
// @Summary 更新跟进方式
// @Tags 跟进记录
// @Accept json
// @Produce json
// @Param id path int true "跟进方式ID"
// @Param request body model.FollowMethodRequest true "更新跟进方式"
// @Success 200 {object} APIResponse
// @Router /api/v1/follow-methods/{id} [put]
func (h *FollowRecordHandler) UpdateFollowMethod(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, 11019, "无效的跟进方式ID")
		return
	}

	var req model.FollowMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 11017, "请求参数错误", err)
		return
	}

	if err := h.service.UpdateFollowMethod(id, req); err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 11020, "更新跟进方式失败", err)
		return
	}

	Success(c, nil)
}

// DeleteFollowMethod 删除跟进方式
// @Summary 删除跟进方式
// @Tags 跟进记录
// @Accept json
// @Produce json
// @Param id path int true "跟进方式ID"
// @Success 200 {object} APIResponse
// @Router /api/v1/follow-methods/{id} [delete]
func (h *FollowRecordHandler) DeleteFollowMethod(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, 11019, "无效的跟进方式ID")
		return
	}

	if err := h.service.DeleteFollowMethod(id); err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 11021, "删除跟进方式失败", err)
		return
	}

	Success(c, nil)
}

func parseFollowTime(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return &parsed, nil
		}
	}

	return nil, errors.New("invalid follow time format")
}
