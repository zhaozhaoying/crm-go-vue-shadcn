package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowOrigin string) gin.HandlerFunc {
	allowedOrigins := parseAllowedOrigins(allowOrigin)

	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		originAllowed := isOriginAllowed(origin, allowedOrigins)

		if originAllowed {
			if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
			c.Writer.Header().Set("Vary", "Origin")
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			if origin != "" && !originAllowed {
				c.AbortWithStatus(403)
				return
			}
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func parseAllowedOrigins(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}

	parts := strings.Split(trimmed, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := normalizeOrigin(part)
		if item == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}

func normalizeOrigin(origin string) string {
	normalized := strings.TrimSpace(origin)
	normalized = strings.TrimSuffix(normalized, "/")
	return strings.ToLower(normalized)
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	normalizedOrigin := normalizeOrigin(origin)
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if normalizedOrigin == allowed {
			return true
		}
	}
	return false
}
