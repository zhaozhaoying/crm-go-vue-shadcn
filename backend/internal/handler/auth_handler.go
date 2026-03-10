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
// @Description 获取一次性验证码 challenge，用于登录前校验
// @Tags        auth
// @Produce     json
// @Success     200 {object} APIResponse{data=CaptchaResponse}
// @Failure     500 {object} APIResponse
// @Router      /api/v1/auth/captcha [get]
func (h *AuthHandler) Captcha(c *gin.Context) {
	challenge, err := h.captchaService.Generate(c.Request.Context(), buildCaptchaFingerprint(c))
	if err != nil {
		Error(c, http.StatusInternalServerError, 20016, "验证码生成失败")
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
// @Description 使用用户名和密码登录，返回JWT token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     LoginRequest true "登录信息"
// @Success     200  {object} APIResponse{data=LoginResponse}
// @Failure     400  {object} APIResponse
// @Failure     401  {object} APIResponse
// @Router      /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 20001, "invalid request: "+err.Error())
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
		Error(c, http.StatusInternalServerError, 20016, "验证码校验失败")
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
		Error(c, http.StatusInternalServerError, 20004, "登录失败")
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
// @Description 使用 refresh token 换取新的 access token 和 refresh token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     RefreshTokenRequest true "刷新令牌"
// @Success     200  {object} APIResponse{data=LoginResponse}
// @Failure     400  {object} APIResponse
// @Failure     401  {object} APIResponse
// @Router      /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 20005, "invalid request: "+err.Error())
		return
	}

	tokens, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			Error(c, http.StatusUnauthorized, 20006, "refresh token 无效")
			return
		}
		if errors.Is(err, service.ErrTokenRevoked) {
			Error(c, http.StatusUnauthorized, 20007, "refresh token 已失效")
			return
		}
		if errors.Is(err, service.ErrUserDisabled) {
			Error(c, http.StatusForbidden, 20003, "账号已被禁用")
			return
		}
		Error(c, http.StatusInternalServerError, 20008, "刷新 token 失败")
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
// @Description 将当前 access token 拉黑，可选作废 refresh token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body     LogoutRequest false "可选 refresh token"
// @Success     200  {object} APIResponse
// @Failure     401  {object} APIResponse
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
		Error(c, http.StatusInternalServerError, 20009, "退出登录失败")
		return
	}

	Success(c, gin.H{"success": true})
}

// Me godoc
// @Summary     获取当前登录用户信息
// @Description 根据当前 access token 返回完整用户资料
// @Tags        auth
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} APIResponse{data=authctx.CurrentUser}
// @Failure     401 {object} APIResponse
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
		Error(c, http.StatusInternalServerError, 20012, "获取当前用户失败")
		return
	}

	Success(c, currentUser)
}

func buildCaptchaFingerprint(c *gin.Context) string {
	ip := strings.TrimSpace(c.ClientIP())
	userAgent := strings.TrimSpace(c.GetHeader("User-Agent"))
	return ip + "|" + userAgent
}
