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

type ResourcePoolHandler struct {
	service service.ResourcePoolService
}

type ResourcePoolSearchRequest struct {
	Region          string   `json:"region"`
	Address         string   `json:"address"`
	Radius          int      `json:"radius"`
	Keyword         string   `json:"keyword"`
	CenterLatitude  *float64 `json:"centerLatitude"`
	CenterLongitude *float64 `json:"centerLongitude"`
}

type ResourcePoolBatchConvertRequest struct {
	ResourceIDs []int64 `json:"resourceIds"`
}

func NewResourcePoolHandler(service service.ResourcePoolService) *ResourcePoolHandler {
	return &ResourcePoolHandler{service: service}
}

// List godoc
// @Summary     获取资源池列表
// @Description 分页获取资源池信息
// @Tags        resource-pool
// @Produce     json
// @Security    BearerAuth
// @Param       keyword query string false "关键词(名称/电话/地址)"
// @Param       hasPhone query bool false "是否有电话"
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.ResourcePoolListResult}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/resource-pool [get]
func (h *ResourcePoolHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	var hasPhone *bool
	rawHasPhone := strings.TrimSpace(strings.ToLower(c.Query("hasPhone")))
	if rawHasPhone != "" {
		switch rawHasPhone {
		case "1", "true", "yes", "on":
			v := true
			hasPhone = &v
		case "0", "false", "no", "off":
			v := false
			hasPhone = &v
		}
	}

	result, err := h.service.List(c.Request.Context(), model.ResourcePoolListFilter{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		HasPhone: hasPhone,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		Error(c, http.StatusInternalServerError, 70001, "failed to list resource pool")
		return
	}
	Success(c, result)
}

// SearchAndStore godoc
// @Summary     百度地图检索并入资源池
// @Description 根据区域/地址检索周边企业并保存到资源池
// @Tags        resource-pool
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body ResourcePoolSearchRequest true "检索参数"
// @Success     200 {object} APIResponse{data=model.ResourcePoolSearchResult}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Router      /api/v1/resource-pool/search [post]
func (h *ResourcePoolHandler) SearchAndStore(c *gin.Context) {
	userID, err := authctx.GetUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, 70002, "invalid token claims")
		return
	}

	var req ResourcePoolSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 70003, "invalid request body")
		return
	}

	result, svcErr := h.service.SearchAndStore(c.Request.Context(), userID, model.ResourcePoolSearchInput{
		Region:          req.Region,
		Address:         req.Address,
		Radius:          req.Radius,
		Keyword:         req.Keyword,
		CenterLatitude:  req.CenterLatitude,
		CenterLongitude: req.CenterLongitude,
	})
	if svcErr != nil {
		switch {
		case errors.Is(svcErr, service.ErrResourcePoolInvalidInput):
			Error(c, http.StatusBadRequest, 70004, "请至少提供区域或地址")
		case errors.Is(svcErr, service.ErrResourcePoolProviderNotConfigured):
			Error(c, http.StatusServiceUnavailable, 70005, "百度地图服务未配置")
		case errors.Is(svcErr, service.ErrResourcePoolLocationNotFound):
			Error(c, http.StatusBadRequest, 70006, "未找到查询位置")
		case errors.Is(svcErr, service.ErrResourcePoolSearchFailed):
			Error(c, http.StatusBadGateway, 70007, "百度地图检索失败")
		default:
			Error(c, http.StatusInternalServerError, 70008, "地图资源检索失败")
		}
		return
	}

	Success(c, result)
}

// ConvertToCustomer godoc
// @Summary     资源池线索一键转客户
// @Description 将资源池中的单条线索转换为客户并绑定给当前操作人
// @Tags        resource-pool
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "资源池ID"
// @Success     200 {object} APIResponse{data=model.ResourcePoolConvertResult}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/resource-pool/{id}/convert [post]
func (h *ResourcePoolHandler) ConvertToCustomer(c *gin.Context) {
	userID, err := authctx.GetUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, 70002, "invalid token claims")
		return
	}

	resourceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || resourceID <= 0 {
		Error(c, http.StatusBadRequest, 70009, "invalid resource id")
		return
	}

	result, svcErr := h.service.ConvertToCustomer(c.Request.Context(), userID, resourceID)
	if svcErr != nil {
		switch {
		case errors.Is(svcErr, service.ErrResourcePoolItemNotFound):
			Error(c, http.StatusNotFound, 70010, "resource pool item not found")
		case errors.Is(svcErr, service.ErrResourcePoolNoConvertiblePhone):
			Error(c, http.StatusBadRequest, 70011, "地图资源电话不可用于创建客户，请补充手机号后重试")
		case errors.Is(svcErr, service.ErrResourcePoolConvertFailed):
			Error(c, http.StatusConflict, 70012, "地图资源转客户失败：客户信息冲突")
		default:
			Error(c, http.StatusInternalServerError, 70013, "地图资源转客户失败")
		}
		return
	}

	Success(c, result)
}

// ConvertBatchToCustomer godoc
// @Summary     地图资源批量转客户
// @Description 将地图资源中的多条线索批量转换为客户并绑定给当前操作人
// @Tags        resource-pool
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body ResourcePoolBatchConvertRequest true "批量转客户参数"
// @Success     200 {object} APIResponse{data=model.ResourcePoolBatchConvertResult}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Router      /api/v1/resource-pool/convert/batch [post]
func (h *ResourcePoolHandler) ConvertBatchToCustomer(c *gin.Context) {
	userID, err := authctx.GetUserIDFromContext(c)
	if err != nil {
		Error(c, http.StatusUnauthorized, 70002, "invalid token claims")
		return
	}

	var req ResourcePoolBatchConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 70014, "invalid request body")
		return
	}

	result, svcErr := h.service.ConvertBatchToCustomer(c.Request.Context(), userID, req.ResourceIDs)
	if svcErr != nil {
		switch {
		case errors.Is(svcErr, service.ErrResourcePoolInvalidInput):
			Error(c, http.StatusBadRequest, 70015, "请提供有效的地图资源ID列表")
		default:
			Error(c, http.StatusInternalServerError, 70016, "地图资源批量转客户失败")
		}
		return
	}

	Success(c, result)
}
