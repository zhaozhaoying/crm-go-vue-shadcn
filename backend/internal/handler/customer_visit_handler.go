package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CustomerVisitHandler struct {
	service *service.CustomerVisitService
}

func NewCustomerVisitHandler(service *service.CustomerVisitService) *CustomerVisitHandler {
	return &CustomerVisitHandler{service: service}
}

type CreateCustomerVisitRequest struct {
	CustomerName  string  `json:"customerName"`
	Inviter       string  `json:"inviter"`
	CheckInLat    float64 `json:"checkInLat"`
	CheckInLng    float64 `json:"checkInLng"`
	Province      any     `json:"province"`
	City          any     `json:"city"`
	Area          any     `json:"area"`
	DetailAddress string  `json:"detailAddress"`
	Images        string  `json:"images"`
	VisitPurpose  string  `json:"visitPurpose"`
	Remark        string  `json:"remark"`
}

// Create 创建上门拜访记录
// @Summary 创建上门拜访签到
// @Tags 上门拜访
// @Accept json
// @Produce json
// @Param request body CreateCustomerVisitRequest true "创建上门拜访"
// @Success 200 {object} APIResponse
// @Router /api/v1/customer-visits [post]
func (h *CustomerVisitHandler) Create(c *gin.Context) {
	var req CreateCustomerVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 12001, "请求参数错误", err)
		return
	}

	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	images := strings.TrimSpace(req.Images)
	if images == "" {
		images = "[]"
	}

	// 签到日期默认当天
	visitDate := time.Now().Format("2006-01-02")

	input := model.CustomerVisitCreateInput{
		OperatorUserID: userID,
		CustomerName:   strings.TrimSpace(req.CustomerName),
		Inviter:        strings.TrimSpace(req.Inviter),
		CheckInIP:      strings.TrimSpace(c.ClientIP()),
		CheckInLat:     req.CheckInLat,
		CheckInLng:     req.CheckInLng,
		Province:       normalizeVisitRegionField(req.Province),
		City:           normalizeVisitRegionField(req.City),
		Area:           normalizeVisitRegionField(req.Area),
		DetailAddress:  strings.TrimSpace(req.DetailAddress),
		Images:         images,
		VisitPurpose:   strings.TrimSpace(req.VisitPurpose),
		Remark:         strings.TrimSpace(req.Remark),
		VisitDate:      visitDate,
	}

	id, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrCustomerVisitAlreadyCheckedInToday) {
			Error(c, http.StatusConflict, 12002, "同一公司同一IP当天已打卡，不能重复提交")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 12004, "创建上门拜访记录失败", err)
		return
	}

	Success(c, gin.H{"id": id})
}

// List 获取上门拜访列表
// @Summary 获取上门拜访列表
// @Tags 上门拜访
// @Accept json
// @Produce json
// @Param keyword query string false "关键词"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} APIResponse{data=model.CustomerVisitListResult}
// @Router /api/v1/customer-visits [get]
func (h *CustomerVisitHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	keyword := strings.TrimSpace(c.Query("keyword"))
	startTime, endTime, err := parseCustomerVisitListTimeRange(
		c.Query("startTime"),
		c.Query("endTime"),
	)
	if err != nil {
		Error(c, http.StatusBadRequest, 12003, err.Error())
		return
	}

	filter := model.CustomerVisitListFilter{
		OperatorUserID: userID,
		CanViewAll:     canViewAllCustomerVisits(currentUserRole(c)),
		Keyword:        keyword,
		StartTime:      startTime,
		EndTime:        endTime,
		Page:           page,
		PageSize:       pageSize,
	}

	result, err := h.service.List(filter)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 12005, "加载上门拜访记录失败", err)
		return
	}

	Success(c, result)
}

func normalizeVisitRegionField(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" || trimmed == "0" || trimmed == "0.0" {
			return ""
		}
		return trimmed
	case float64:
		if v == 0 {
			return ""
		}
		return strconv.FormatInt(int64(v), 10)
	case float32:
		if v == 0 {
			return ""
		}
		return strconv.FormatInt(int64(v), 10)
	case int:
		if v == 0 {
			return ""
		}
		return strconv.Itoa(v)
	case int64:
		if v == 0 {
			return ""
		}
		return strconv.FormatInt(v, 10)
	case int32:
		if v == 0 {
			return ""
		}
		return strconv.FormatInt(int64(v), 10)
	default:
		raw := strings.TrimSpace(fmt.Sprint(v))
		if raw == "" || raw == "0" || raw == "<nil>" {
			return ""
		}
		return raw
	}
}

func canViewAllCustomerVisits(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin", "管理员", "finance", "finance_manager", "财务", "财务经理":
		return true
	default:
		return false
	}
}

func parseCustomerVisitListTimeRange(
	startRaw, endRaw string,
) (*time.Time, *time.Time, error) {
	startTime, err := parseCustomerVisitFilterTime(startRaw, false)
	if err != nil {
		return nil, nil, fmt.Errorf("开始时间格式错误")
	}
	endTime, err := parseCustomerVisitFilterTime(endRaw, true)
	if err != nil {
		return nil, nil, fmt.Errorf("结束时间格式错误")
	}
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		return nil, nil, fmt.Errorf("开始时间不能晚于结束时间")
	}
	return startTime, endTime, nil
}

func parseCustomerVisitFilterTime(raw string, isEnd bool) (*time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	layouts := []struct {
		layout   string
		dateOnly bool
	}{
		{layout: time.RFC3339, dateOnly: false},
		{layout: "2006-01-02T15:04:05", dateOnly: false},
		{layout: "2006-01-02 15:04:05", dateOnly: false},
		{layout: "2006-01-02", dateOnly: true},
	}

	for _, item := range layouts {
		parsed, err := time.ParseInLocation(item.layout, trimmed, time.Local)
		if err != nil {
			continue
		}
		if item.dateOnly && isEnd {
			parsed = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		return &parsed, nil
	}

	return nil, fmt.Errorf("invalid time")
}
