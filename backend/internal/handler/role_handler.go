package handler

import (
	"backend/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	service service.RoleService
}

type CreateRoleRequest struct {
	Name  string `json:"name" binding:"required" example:"sales_staff"`
	Label string `json:"label" binding:"required" example:"销售员工"`
	Sort  int    `json:"sort" example:"10"`
}

type UpdateRoleRequest struct {
	Name  string `json:"name" binding:"required"`
	Label string `json:"label" binding:"required"`
	Sort  int    `json:"sort"`
}

func NewRoleHandler(service service.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

// List godoc
// @Summary     获取角色列表
// @Tags        roles
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} APIResponse{data=[]model.Role}
// @Router      /api/v1/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.service.List(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "获取角色列表失败")
		return
	}
	Success(c, roles)
}

// Create godoc
// @Summary     创建角色
// @Tags        roles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body CreateRoleRequest true "角色信息"
// @Success     200 {object} APIResponse{data=model.Role}
// @Router      /api/v1/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 50002, "参数错误: "+err.Error())
		return
	}
	role, err := h.service.Create(c.Request.Context(), req.Name, req.Label, req.Sort)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50003, "创建角色失败")
		return
	}
	Success(c, role)
}

// Update godoc
// @Summary     更新角色
// @Tags        roles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path int              true "角色ID"
// @Param       body body UpdateRoleRequest true "角色信息"
// @Success     200 {object} APIResponse{data=model.Role}
// @Router      /api/v1/roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 50004, "无效的角色ID")
		return
	}
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 50002, "参数错误: "+err.Error())
		return
	}
	role, err := h.service.Update(c.Request.Context(), id, req.Name, req.Label, req.Sort)
	if err != nil {
		if errors.Is(err, service.ErrRoleNotFound) {
			Error(c, http.StatusNotFound, 50005, "角色不存在")
			return
		}
		Error(c, http.StatusInternalServerError, 50006, "更新角色失败")
		return
	}
	Success(c, role)
}

// Delete godoc
// @Summary     删除角色
// @Tags        roles
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "角色ID"
// @Success     200 {object} APIResponse
// @Router      /api/v1/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 50004, "无效的角色ID")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		Error(c, http.StatusInternalServerError, 50007, "删除角色失败")
		return
	}
	Success(c, nil)
}
