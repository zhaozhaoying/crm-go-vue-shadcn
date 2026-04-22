package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SystemSettingHandler struct {
	service *service.SystemSettingService
}

func NewSystemSettingHandler(service *service.SystemSettingService) *SystemSettingHandler {
	return &SystemSettingHandler{service: service}
}

// GetSettings godoc
// @Summary 获取系统设置
// @Tags 系统设置
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.SystemSettingsResponse
// @Router /api/v1/settings [get]
func (h *SystemSettingHandler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetAllSettings()
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10100, "加载系统设置失败", err)
		return
	}
	if !canManageSystemSettings(currentUserRole(c)) {
		settings.MihuaCallRecordToken = ""
		settings.HanghangCrmCloudToken = ""
	}
	Success(c, settings)
}

// UpdateSettings godoc
// @Summary 更新系统设置
// @Tags 系统设置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param settings body model.UpdateSystemSettingsRequest true "系统设置"
// @Success 200 {object} map[string]string
// @Router /api/v1/settings [put]
func (h *SystemSettingHandler) UpdateSettings(c *gin.Context) {
	if !ensureManageSystemSettings(c) {
		return
	}

	var req model.UpdateSystemSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10101, "请求参数错误", err)
		return
	}

	if err := h.service.UpdateSettings(&req); err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10102, "更新系统设置失败", err)
		return
	}

	Success(c, gin.H{"message": "设置已更新"})
}

// GetCustomerLevels godoc
// @Summary 获取客户级别列表
// @Tags 系统设置
// @Security BearerAuth
// @Produce json
// @Success 200 {array} model.CustomerLevel
// @Router /api/v1/settings/customer-levels [get]
func (h *SystemSettingHandler) GetCustomerLevels(c *gin.Context) {
	levels, err := h.service.GetCustomerLevels()
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10103, "加载客户级别失败", err)
		return
	}
	Success(c, levels)
}

// CreateCustomerLevel godoc
// @Summary 创建客户级别
// @Tags 系统设置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param level body model.CustomerLevelRequest true "客户级别"
// @Success 201 {object} model.CustomerLevel
// @Router /api/v1/settings/customer-levels [post]
func (h *SystemSettingHandler) CreateCustomerLevel(c *gin.Context) {
	if !ensureManageSystemSettings(c) {
		return
	}

	var req model.CustomerLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10104, "请求参数错误", err)
		return
	}

	level, err := h.service.CreateCustomerLevel(&req)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10105, "创建客户级别失败", err)
		return
	}

	Success(c, level)
}

// UpdateCustomerLevel godoc
// @Summary 更新客户级别
// @Tags 系统设置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "级别ID"
// @Param level body model.CustomerLevelRequest true "客户级别"
// @Success 200 {object} map[string]string
// @Router /api/v1/settings/customer-levels/{id} [put]
func (h *SystemSettingHandler) UpdateCustomerLevel(c *gin.Context) {
	if !ensureManageSystemSettings(c) {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, 10106, "无效的客户级别ID")
		return
	}

	var req model.CustomerLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10107, "请求参数错误", err)
		return
	}

	if err := h.service.UpdateCustomerLevel(id, &req); err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10108, "更新客户级别失败", err)
		return
	}

	Success(c, gin.H{"message": "更新成功"})
}

// DeleteCustomerLevel godoc
// @Summary 删除客户级别
// @Tags 系统设置
// @Security BearerAuth
// @Param id path int true "级别ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/settings/customer-levels/{id} [delete]
func (h *SystemSettingHandler) DeleteCustomerLevel(c *gin.Context) {
	if !ensureManageSystemSettings(c) {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, 10109, "无效的客户级别ID")
		return
	}

	if err := h.service.DeleteCustomerLevel(id); err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10110, "删除客户级别失败", err)
		return
	}

	Success(c, gin.H{"message": "删除成功"})
}

// GetCustomerSources godoc
// @Summary 获取客户来源列表
// @Tags 系统设置
// @Security BearerAuth
// @Produce json
// @Success 200 {array} model.CustomerSource
// @Router /api/v1/settings/customer-sources [get]
func (h *SystemSettingHandler) GetCustomerSources(c *gin.Context) {
	sources, err := h.service.GetCustomerSources()
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10111, "加载客户来源失败", err)
		return
	}
	Success(c, sources)
}

// CreateCustomerSource godoc
// @Summary 创建客户来源
// @Tags 系统设置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param source body model.CustomerSourceRequest true "客户来源"
// @Success 201 {object} model.CustomerSource
// @Router /api/v1/settings/customer-sources [post]
func (h *SystemSettingHandler) CreateCustomerSource(c *gin.Context) {
	if !ensureManageSystemSettings(c) {
		return
	}

	var req model.CustomerSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10112, "请求参数错误", err)
		return
	}

	source, err := h.service.CreateCustomerSource(&req)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10113, "创建客户来源失败", err)
		return
	}

	Success(c, source)
}

// UpdateCustomerSource godoc
// @Summary 更新客户来源
// @Tags 系统设置
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "来源ID"
// @Param source body model.CustomerSourceRequest true "客户来源"
// @Success 200 {object} map[string]string
// @Router /api/v1/settings/customer-sources/{id} [put]
func (h *SystemSettingHandler) UpdateCustomerSource(c *gin.Context) {
	if !ensureManageSystemSettings(c) {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, 10114, "无效的客户来源ID")
		return
	}

	var req model.CustomerSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10115, "请求参数错误", err)
		return
	}

	if err := h.service.UpdateCustomerSource(id, &req); err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10116, "更新客户来源失败", err)
		return
	}

	Success(c, gin.H{"message": "更新成功"})
}

// DeleteCustomerSource godoc
// @Summary 删除客户来源
// @Tags 系统设置
// @Security BearerAuth
// @Param id path int true "来源ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/settings/customer-sources/{id} [delete]
func (h *SystemSettingHandler) DeleteCustomerSource(c *gin.Context) {
	if !ensureManageSystemSettings(c) {
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, 10117, "无效的客户来源ID")
		return
	}

	if err := h.service.DeleteCustomerSource(id); err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10118, "删除客户来源失败", err)
		return
	}

	Success(c, gin.H{"message": "删除成功"})
}

func ensureManageSystemSettings(c *gin.Context) bool {
	if canManageSystemSettings(currentUserRole(c)) {
		return true
	}
	Error(c, http.StatusForbidden, 10119, "仅管理员可以管理系统设置")
	return false
}

func canManageSystemSettings(role string) bool {
	switch role {
	case "admin", "管理员":
		return true
	default:
		return false
	}
}
