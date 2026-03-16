package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type TokenBlacklistChecker interface {
	IsAccessTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}

func JWTAuth(secret string, checker TokenBlacklistChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, errCode, errMessage := resolveBearerToken(c)
		if tokenString == "" {
			handler.Error(c, http.StatusUnauthorized, errCode, errMessage)
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			handler.Error(c, http.StatusUnauthorized, 30003, "登录状态已失效或令牌无效")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			handler.Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
			c.Abort()
			return
		}

		if typ, _ := claims["typ"].(string); typ != "access" {
			handler.Error(c, http.StatusUnauthorized, 30003, "登录状态已失效或令牌无效")
			c.Abort()
			return
		}

		tokenJTI, _ := claims["jti"].(string)
		if tokenJTI == "" {
			handler.Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
			c.Abort()
			return
		}

		if checker != nil {
			blacklisted, err := checker.IsAccessTokenBlacklisted(c.Request.Context(), tokenJTI)
			if err != nil {
				handler.ErrorWithDetail(c, http.StatusInternalServerError, 30005, "登录状态校验失败", err)
				c.Abort()
				return
			}
			if blacklisted {
				handler.Error(c, http.StatusUnauthorized, 30003, "登录状态已失效或令牌无效")
				c.Abort()
				return
			}
		}

		var userID int64
		switch sub := claims["sub"].(type) {
		case float64:
			userID = int64(sub)
		case int64:
			userID = sub
		case int:
			userID = int64(sub)
		case string:
			parsed, err := strconv.ParseInt(sub, 10, 64)
			if err != nil {
				handler.Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
				c.Abort()
				return
			}
			userID = parsed
		default:
			handler.Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
			c.Abort()
			return
		}

		var tokenExp int64
		switch exp := claims["exp"].(type) {
		case float64:
			tokenExp = int64(exp)
		case int64:
			tokenExp = exp
		case int:
			tokenExp = int64(exp)
		case string:
			parsed, err := strconv.ParseInt(exp, 10, 64)
			if err == nil {
				tokenExp = parsed
			}
		}
		if tokenExp == 0 {
			tokenExp = time.Now().Unix()
		}

		c.Set("userID", userID)
		c.Set("username", claims["username"])
		c.Set("tokenJTI", tokenJTI)
		c.Set("tokenExp", tokenExp)
		if role, ok := claims["role"].(string); ok {
			c.Set("role", role)
		}
		c.Next()
	}
}

func resolveBearerToken(c *gin.Context) (string, int, string) {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return "", 30002, "认证头格式错误"
		}
		return strings.TrimSpace(parts[1]), 0, ""
	}

	upgrade := strings.ToLower(strings.TrimSpace(c.GetHeader("Upgrade")))
	if upgrade == "websocket" {
		if token := strings.TrimSpace(c.Query("token")); token != "" {
			return token, 0, ""
		}
	}

	return "", 30001, "缺少认证信息"
}
