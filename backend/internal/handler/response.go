package handler

import (
	"backend/internal/errmsg"
	"strings"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, APIResponse{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, code int, message string) {
	c.JSON(statusCode, APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func ErrorWithDetail(c *gin.Context, statusCode int, code int, message string, err error) {
	if err == nil {
		Error(c, statusCode, code, message)
		return
	}

	detail := normalizeErrorDetail(err.Error())
	if detail == "" || strings.Contains(message, detail) {
		Error(c, statusCode, code, message)
		return
	}

	Error(c, statusCode, code, message+": "+detail)
}

func normalizeErrorDetail(detail string) string {
	detail = errmsg.Normalize(detail)
	if len(detail) > 300 {
		return detail[:300] + "..."
	}
	return detail
}
