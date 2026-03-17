package handler

import (
	"backend/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	service service.UploadService
}

func NewUploadHandler(service service.UploadService) *UploadHandler {
	return &UploadHandler{service: service}
}

// UploadAvatar godoc
// @Summary     上传用户头像
// @Description 上传用户头像到阿里云OSS
// @Tags        users
// @Accept      multipart/form-data
// @Produce     json
// @Security    BearerAuth
// @Param       file formData file true "头像文件（支持 JPG、PNG、WEBP，最大 20MB）"
// @Success     200 {object} APIResponse{data=map[string]string}
// @Failure     400 {object} APIResponse "请求参数错误"
// @Failure     401 {object} APIResponse "未登录或登录已失效"
// @Failure     500 {object} APIResponse "服务器内部错误"
// @Router      /api/v1/users/avatar/upload [post]
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	if h.service == nil {
		Error(c, http.StatusInternalServerError, 40020, "上传服务未配置")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 40021, "请上传头像文件")
		return
	}

	url, err := h.service.UploadAvatar(c.Request.Context(), file)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidImageType):
			Error(c, http.StatusBadRequest, 40022, "仅支持 JPG、PNG、WEBP 图片")
		case errors.Is(err, service.ErrImageTooLarge):
			Error(c, http.StatusBadRequest, 40023, "图片大小不能超过20MB")
		case errors.Is(err, service.ErrUploadServiceNotConfigured):
			Error(c, http.StatusInternalServerError, 40024, "OSS上传服务未配置")
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 40025, "头像上传失败", err)
		}
		return
	}

	Success(c, gin.H{"url": url})
}
