package handler

import (
	"backend/internal/service"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserService
}

var userMobileRegex = regexp.MustCompile(`^1\d{10}$`)

type CreateUserRequest struct {
	Username          string `json:"username" binding:"required,min=3,max=32" example:"zhangsan"`
	Password          string `json:"password" binding:"required,min=6,max=64" example:"123456"`
	Nickname          string `json:"nickname" example:"张三"`
	Email             string `json:"email" example:"zhangsan@example.com"`
	Mobile            string `json:"mobile" binding:"required" example:"13800138000"`
	HanghangCRMMobile string `json:"hanghangCrmMobile" example:"13800138001"`
	MihuaWorkNumber   string `json:"mihuaWorkNumber" example:"1001"`
	Avatar            string `json:"avatar"`
	RoleID            int64  `json:"roleId" binding:"required" example:"1"`
	ParentID          *int64 `json:"parentId"`
}

type UpdateUserRequest struct {
	Username          string `json:"username" binding:"required,min=3,max=32"`
	Password          string `json:"password"`
	Nickname          string `json:"nickname"`
	Email             string `json:"email"`
	Mobile            string `json:"mobile" binding:"required"`
	HanghangCRMMobile string `json:"hanghangCrmMobile"`
	MihuaWorkNumber   string `json:"mihuaWorkNumber"`
	Avatar            string `json:"avatar"`
	RoleID            int64  `json:"roleId" binding:"required"`
	ParentID          *int64 `json:"parentId"`
	Status            string `json:"status" example:"enabled"`
}

func normalizeRequestHanghangCRMMobile(value string, required bool) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		if required {
			return "", errors.New("请填写航航CRM手机号")
		}
		return "", nil
	}
	if !userMobileRegex.MatchString(trimmed) {
		return "", errors.New("航航CRM手机号必须为11位数字")
	}
	return trimmed, nil
}

type BatchDisableUsersRequest struct {
	UserIDs []int64 `json:"userIds" binding:"required,min=1"`
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// List godoc
// @Summary     获取用户列表
// @Tags        users
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} APIResponse{data=[]model.UserWithRole}
// @Router      /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.service.List(c.Request.Context())
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 40001, "获取用户列表失败", err)
		return
	}
	Success(c, users)
}

// Search godoc
// @Summary     搜索用户
// @Tags        users
// @Produce     json
// @Security    BearerAuth
// @Param       keyword query string false "搜索关键词"
// @Success     200 {object} APIResponse{data=[]model.UserWithRole}
// @Router      /api/v1/users/search [get]
func (h *UserHandler) Search(c *gin.Context) {
	keyword := c.Query("keyword")
	users, err := h.service.Search(c.Request.Context(), keyword)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 40014, "搜索用户失败", err)
		return
	}
	Success(c, users)
}

// ListTelemarketingUsers godoc
// @Summary     获取电销用户列表
// @Tags        users
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} APIResponse{data=[]model.UserWithRole}
// @Router      /api/v1/users/telemarketing [get]
func (h *UserHandler) ListTelemarketingUsers(c *gin.Context) {
	users, err := h.service.ListTelemarketingUsers(c.Request.Context())
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 40015, "获取电销用户列表失败", err)
		return
	}
	Success(c, users)
}

// GetByID godoc
// @Summary     获取用户详情
// @Tags        users
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "用户ID"
// @Success     200 {object} APIResponse{data=model.User}
// @Router      /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 40002, "无效的用户ID")
		return
	}
	user, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, http.StatusNotFound, 40003, "用户不存在")
		return
	}
	Success(c, user)
}

// Create godoc
// @Summary     创建用户
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body CreateUserRequest true "用户信息"
// @Success     200 {object} APIResponse{data=model.User}
// @Router      /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 40004, "参数错误", err)
		return
	}
	if !userMobileRegex.MatchString(req.Mobile) {
		Error(c, http.StatusBadRequest, 40004, "参数错误: 手机号必须为11位数字")
		return
	}
	crmMobile, err := normalizeRequestHanghangCRMMobile(req.HanghangCRMMobile, false)
	if err != nil {
		Error(c, http.StatusBadRequest, 40004, "参数错误: "+err.Error())
		return
	}
	user, err := h.service.Create(c.Request.Context(), service.CreateUserInput{
		Username:          req.Username,
		Password:          req.Password,
		Nickname:          req.Nickname,
		Email:             req.Email,
		Mobile:            req.Mobile,
		HanghangCRMMobile: crmMobile,
		MihuaWorkNumber:   strings.TrimSpace(req.MihuaWorkNumber),
		Avatar:            req.Avatar,
		RoleID:            req.RoleID,
		ParentID:          req.ParentID,
	})
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			Error(c, http.StatusConflict, 40005, "用户名已存在")
			return
		}
		if errors.Is(err, service.ErrInvalidRole) {
			Error(c, http.StatusBadRequest, 40006, "无效的角色")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 40007, "创建用户失败", err)
		return
	}
	Success(c, user)
}

// Update godoc
// @Summary     更新用户
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path int              true "用户ID"
// @Param       body body UpdateUserRequest true "更新信息"
// @Success     200 {object} APIResponse{data=model.User}
// @Router      /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 40002, "无效的用户ID")
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 40004, "参数错误", err)
		return
	}
	if !userMobileRegex.MatchString(req.Mobile) {
		Error(c, http.StatusBadRequest, 40004, "参数错误: 手机号必须为11位数字")
		return
	}
	crmMobile, err := normalizeRequestHanghangCRMMobile(req.HanghangCRMMobile, false)
	if err != nil {
		Error(c, http.StatusBadRequest, 40004, "参数错误: "+err.Error())
		return
	}
	user, err := h.service.Update(c.Request.Context(), id, service.UpdateUserInput{
		Username:          req.Username,
		Password:          req.Password,
		Nickname:          req.Nickname,
		Email:             req.Email,
		Mobile:            req.Mobile,
		HanghangCRMMobile: crmMobile,
		MihuaWorkNumber:   strings.TrimSpace(req.MihuaWorkNumber),
		Avatar:            req.Avatar,
		RoleID:            req.RoleID,
		ParentID:          req.ParentID,
		Status:            req.Status,
	})
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			Error(c, http.StatusNotFound, 40003, "用户不存在")
			return
		}
		if errors.Is(err, service.ErrInvalidRole) {
			Error(c, http.StatusBadRequest, 40006, "无效的角色")
			return
		}
		if errors.Is(err, service.ErrInvalidPassword) {
			Error(c, http.StatusBadRequest, 40011, "密码至少6位")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 40008, "更新用户失败", err)
		return
	}
	Success(c, user)
}

// Delete godoc
// @Summary     删除用户
// @Tags        users
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "用户ID"
// @Success     200 {object} APIResponse
// @Router      /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 40002, "无效的用户ID")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 40010, "删除用户失败", err)
		return
	}
	Success(c, nil)
}

// BatchDisable godoc
// @Summary     批量禁用用户
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body BatchDisableUsersRequest true "批量禁用信息"
// @Success     200 {object} APIResponse{data=map[string]int64}
// @Router      /api/v1/users/batch/disable [put]
func (h *UserHandler) BatchDisable(c *gin.Context) {
	var req BatchDisableUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 40004, "参数错误", err)
		return
	}

	affected, err := h.service.BatchDisable(c.Request.Context(), req.UserIDs)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserIDs) {
			Error(c, http.StatusBadRequest, 40012, "用户ID列表不能为空且必须为正整数")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 40013, "批量禁用用户失败", err)
		return
	}
	Success(c, gin.H{"affected": affected})
}
