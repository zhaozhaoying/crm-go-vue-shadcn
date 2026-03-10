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
		Error(c, http.StatusInternalServerError, 10100, "failed to load settings")
		return
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
	var req model.UpdateSystemSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10101, "invalid request body")
		return
	}

	if err := h.service.UpdateSettings(&req); err != nil {
		Error(c, http.StatusInternalServerError, 10102, "failed to update settings")
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
		Error(c, http.StatusInternalServerError, 10103, "failed to load customer levels")
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
	var req model.CustomerLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10104, "invalid request body")
		return
	}

	level, err := h.service.CreateCustomerLevel(&req)
	if err != nil {
		Error(c, http.StatusInternalServerError, 10105, "failed to create customer level")
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, 10106, "invalid customer level id")
		return
	}

	var req model.CustomerLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10107, "invalid request body")
		return
	}

	if err := h.service.UpdateCustomerLevel(id, &req); err != nil {
		Error(c, http.StatusInternalServerError, 10108, "failed to update customer level")
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, 10109, "invalid customer level id")
		return
	}

	if err := h.service.DeleteCustomerLevel(id); err != nil {
		Error(c, http.StatusInternalServerError, 10110, "failed to delete customer level")
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
		Error(c, http.StatusInternalServerError, 10111, "failed to load customer sources")
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
	var req model.CustomerSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10112, "invalid request body")
		return
	}

	source, err := h.service.CreateCustomerSource(&req)
	if err != nil {
		Error(c, http.StatusInternalServerError, 10113, "failed to create customer source")
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, 10114, "invalid customer source id")
		return
	}

	var req model.CustomerSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10115, "invalid request body")
		return
	}

	if err := h.service.UpdateCustomerSource(id, &req); err != nil {
		Error(c, http.StatusInternalServerError, 10116, "failed to update customer source")
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, 10117, "invalid customer source id")
		return
	}

	if err := h.service.DeleteCustomerSource(id); err != nil {
		Error(c, http.StatusInternalServerError, 10118, "failed to delete customer source")
		return
	}

	Success(c, gin.H{"message": "删除成功"})
}
