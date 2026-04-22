package handler

import (
	"backend/internal/authctx"
	"backend/internal/service"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service        service.AuthService
	provider       authctx.Provider
	captchaService service.CaptchaService
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username" binding:"required" example:"admin"`
	Password    string `json:"password" binding:"required" example:"admin123"`
	CaptchaID   string `json:"captchaId" binding:"required" example:"a1b2c3d4"`
	CaptchaCode string `json:"captchaCode" binding:"required" example:"A9K3"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token            string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken     string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIs..."`
	ExpiresInSeconds int64  `json:"expiresInSeconds" example:"86400"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type VerifyResetIdentityRequest struct {
	Username string `json:"username" binding:"required"`
	Contact  string `json:"contact" binding:"required"`
}

type VerifyResetIdentityResponse struct {
	ResetToken string `json:"resetToken"`
}

type ResetPasswordRequest struct {
	ResetToken  string `json:"resetToken" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

type ResetPasswordDirectRequest struct {
	Username    string `json:"username" binding:"required"`
	Contact     string `json:"contact" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type CaptchaResponse struct {
	CaptchaID    string `json:"captchaId" example:"a1b2c3d4"`
	CaptchaImage string `json:"captchaImage" example:"data:image/svg+xml;base64,PHN2ZyB4bWxucz0iLi4u\""`
	ExpiresAt    int64  `json:"expiresAt" example:"1710000000"`
}

func NewAuthHandler(service service.AuthService, provider authctx.Provider, captchaService service.CaptchaService) *AuthHandler {
	return &AuthHandler{
		service:        service,
		provider:       provider,
		captchaService: captchaService,
	}
}

// Captcha godoc
// @Summary     获取登录验证码
// @Description 获取一次性验证码挑战，用于登录前校验
// @Tags        auth
// @Produce     json
// @Success     200 {object} APIResponse{data=CaptchaResponse}
// @Failure     500 {object} APIResponse "服务器内部错误"
// @Router      /api/v1/auth/captcha [get]
func (h *AuthHandler) Captcha(c *gin.Context) {
	challenge, err := h.captchaService.Generate(c.Request.Context(), buildCaptchaFingerprint(c))
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 20016, "验证码生成失败", err)
		return
	}

	Success(c, CaptchaResponse{
		CaptchaID:    challenge.CaptchaID,
		CaptchaImage: challenge.CaptchaImage,
		ExpiresAt:    challenge.ExpiresAt,
	})
}

// Login godoc
// @Summary     用户登录
// @Description 使用用户名和密码登录，返回访问令牌和刷新令牌
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     LoginRequest true "登录信息"
// @Success     200  {object} APIResponse{data=LoginResponse}
// @Failure     400  {object} APIResponse "请求参数错误"
// @Failure     401  {object} APIResponse "未登录或登录已失效"
// @Router      /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 20001, "登录参数错误", err)
		return
	}

	if err := h.captchaService.VerifyAndConsume(c.Request.Context(), req.CaptchaID, req.CaptchaCode, buildCaptchaFingerprint(c)); err != nil {
		if errors.Is(err, service.ErrCaptchaExpired) {
			Error(c, http.StatusBadRequest, 20014, "验证码已过期，请刷新后重试")
			return
		}
		if errors.Is(err, service.ErrCaptchaTooManyAttempts) {
			Error(c, http.StatusTooManyRequests, 20015, "验证码错误次数过多，请刷新后重试")
			return
		}
		if errors.Is(err, service.ErrCaptchaInvalid) {
			Error(c, http.StatusBadRequest, 20013, "验证码错误")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 20017, "验证码校验失败", err)
		return
	}

	tokens, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredential) {
			Error(c, http.StatusUnauthorized, 20002, "用户名或密码错误")
			return
		}
		if errors.Is(err, service.ErrUserDisabled) {
			Error(c, http.StatusForbidden, 20003, "账号已被禁用")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 20004, "登录失败", err)
		return
	}

	Success(c, LoginResponse{
		Token:            tokens.Token,
		RefreshToken:     tokens.RefreshToken,
		ExpiresInSeconds: tokens.ExpiresInSeconds,
	})
}

// Refresh godoc
// @Summary     刷新访问令牌
// @Description 使用刷新令牌换取新的访问令牌和刷新令牌
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     RefreshTokenRequest true "刷新令牌"
// @Success     200  {object} APIResponse{data=LoginResponse}
// @Failure     400  {object} APIResponse "请求参数错误"
// @Failure     401  {object} APIResponse "未登录或登录已失效"
// @Router      /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 20005, "刷新令牌参数错误", err)
		return
	}

	tokens, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			Error(c, http.StatusUnauthorized, 20006, "刷新令牌无效")
			return
		}
		if errors.Is(err, service.ErrTokenRevoked) {
			Error(c, http.StatusUnauthorized, 20007, "刷新令牌已失效")
			return
		}
		if errors.Is(err, service.ErrUserDisabled) {
			Error(c, http.StatusForbidden, 20003, "账号已被禁用")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 20008, "刷新令牌失败", err)
		return
	}

	Success(c, LoginResponse{
		Token:            tokens.Token,
		RefreshToken:     tokens.RefreshToken,
		ExpiresInSeconds: tokens.ExpiresInSeconds,
	})
}

// Logout godoc
// @Summary     退出登录
// @Description 将当前访问令牌拉黑，并可选作废刷新令牌
// @Tags        auth
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body     LogoutRequest false "可选刷新令牌"
// @Success     200  {object} APIResponse
// @Failure     401  {object} APIResponse "未登录或登录已失效"
// @Router      /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	_ = c.ShouldBindJSON(&req)

	accessJTI, _ := c.Get("tokenJTI")
	accessExp, _ := c.Get("tokenExp")

	jti, _ := accessJTI.(string)
	exp := int64(0)
	switch v := accessExp.(type) {
	case int64:
		exp = v
	case int:
		exp = int64(v)
	case float64:
		exp = int64(v)
	}

	if err := h.service.Logout(c.Request.Context(), jti, exp, req.RefreshToken); err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 20009, "退出登录失败", err)
		return
	}

	Success(c, gin.H{"success": true})
}

// Me godoc
// @Summary     获取当前登录用户信息
// @Description 根据当前访问令牌返回完整用户资料
// @Tags        auth
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} APIResponse{data=authctx.CurrentUser}
// @Failure     401 {object} APIResponse "未登录或登录已失效"
// @Router      /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	currentUser, err := h.provider.GetCurrentUser(c.Request.Context(), c)
	if err != nil {
		if errors.Is(err, authctx.ErrUnauthorized) {
			Error(c, http.StatusUnauthorized, 20010, "未登录或登录已失效")
			return
		}
		if errors.Is(err, authctx.ErrUserNotFound) {
			Error(c, http.StatusUnauthorized, 20011, "用户不存在或已被删除")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 20012, "获取当前用户失败", err)
		return
	}

	Success(c, currentUser)
}

// VerifyResetIdentity godoc
// @Summary     验证重置密码身份
// @Description 验证用户名与邮箱/手机号，成功后返回重置令牌（15分钟有效）
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     VerifyResetIdentityRequest true "验证信息"
// @Success     200  {object} APIResponse{data=VerifyResetIdentityResponse}
// @Failure     400  {object} APIResponse
// @Router      /api/v1/auth/reset-password/verify [post]
func (h *AuthHandler) VerifyResetIdentity(c *gin.Context) {
	var req VerifyResetIdentityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 20018, "参数错误", err)
		return
	}
	token, err := h.service.VerifyResetIdentity(c.Request.Context(), req.Username, req.Contact)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredential) {
			Error(c, http.StatusBadRequest, 20019, "账号或联系方式不正确")
			return
		}
		if errors.Is(err, service.ErrContactMismatch) {
			Error(c, http.StatusBadRequest, 20019, "账号或联系方式不正确")
			return
		}
		if errors.Is(err, service.ErrUserNoContact) {
			Error(c, http.StatusBadRequest, 20020, "该账号未绑定邮箱或手机号，无法重置密码")
			return
		}
		if errors.Is(err, service.ErrUserDisabled) {
			Error(c, http.StatusForbidden, 20003, "账号已被禁用")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 20021, "验证身份失败", err)
		return
	}
	Success(c, VerifyResetIdentityResponse{ResetToken: token})
}

// ResetPassword godoc
// @Summary     重置密码
// @Description 使用重置令牌设置新密码
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     ResetPasswordRequest true "重置信息"
// @Success     200  {object} APIResponse
// @Failure     400  {object} APIResponse
// @Router      /api/v1/auth/reset-password/confirm [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 20021, "参数错误", err)
		return
	}
	if err := h.service.ResetPassword(c.Request.Context(), req.ResetToken, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidResetToken) {
			Error(c, http.StatusBadRequest, 20022, "重置令牌无效或已过期，请重新验证")
			return
		}
		if errors.Is(err, service.ErrInvalidPassword) {
			Error(c, http.StatusBadRequest, 40011, "密码至少6位")
			return
		}
		if errors.Is(err, service.ErrUserDisabled) {
			Error(c, http.StatusForbidden, 20003, "账号已被禁用")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 20023, "重置密码失败", err)
		return
	}
	Success(c, gin.H{"success": true})
}

// ResetPasswordDirect godoc
// @Summary     一步重置密码
// @Description 验证账号与邮箱/手机号，通过后直接设置新密码
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     ResetPasswordDirectRequest true "重置信息"
// @Success     200  {object} APIResponse
// @Failure     400  {object} APIResponse
// @Router      /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPasswordDirect(c *gin.Context) {
	var req ResetPasswordDirectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 20024, "参数错误", err)
		return
	}
	if err := h.service.ResetPasswordDirect(c.Request.Context(), req.Username, req.Contact, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidCredential) || errors.Is(err, service.ErrContactMismatch) {
			Error(c, http.StatusBadRequest, 20019, "账号或联系方式不正确")
			return
		}
		if errors.Is(err, service.ErrUserNoContact) {
			Error(c, http.StatusBadRequest, 20020, "该账号未绑定邮箱或手机号，无法重置密码")
			return
		}
		if errors.Is(err, service.ErrUserDisabled) {
			Error(c, http.StatusForbidden, 20003, "账号已被禁用")
			return
		}
		if errors.Is(err, service.ErrWeakPassword) {
			Error(c, http.StatusBadRequest, 20025, "密码须为字母+数字组合，至少6位")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 20026, "重置密码失败", err)
		return
	}
	Success(c, gin.H{"success": true})
}

func buildCaptchaFingerprint(c *gin.Context) string {
	ip := strings.TrimSpace(c.ClientIP())
	userAgent := strings.TrimSpace(c.GetHeader("User-Agent"))
	return ip + "|" + userAgent
}
