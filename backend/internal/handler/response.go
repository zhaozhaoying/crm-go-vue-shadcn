package handler

import "github.com/gin-gonic/gin"

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
