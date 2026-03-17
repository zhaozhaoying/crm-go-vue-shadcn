package service

import (
	"backend/internal/config"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	ErrUploadServiceNotConfigured = errors.New("upload service not configured")
	ErrInvalidImageType           = errors.New("invalid image type")
	ErrImageTooLarge              = errors.New("image too large")
	ErrImageUploadFailed          = errors.New("image upload failed")
)

const maxAvatarSize int64 = 20 * 1024 * 1024

type UploadService interface {
	UploadAvatar(ctx context.Context, file *multipart.FileHeader) (string, error)
}

type uploadService struct {
	bucket   *oss.Bucket
	endpoint string
	bucketID string
	basePath string
}

func NewUploadService(cfg config.Config) (UploadService, error) {
	if cfg.OSSEndpoint == "" || cfg.OSSAccessKeyID == "" || cfg.OSSAccessKeySecret == "" || cfg.OSSBucketName == "" {
		return nil, ErrUploadServiceNotConfigured
	}

	client, err := oss.New(cfg.OSSEndpoint, cfg.OSSAccessKeyID, cfg.OSSAccessKeySecret)
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket(cfg.OSSBucketName)
	if err != nil {
		return nil, err
	}

	basePath := strings.TrimSpace(cfg.OSSBasePath)
	basePath = strings.Trim(basePath, "/")
	if basePath == "" {
		basePath = "avatars"
	}

	return &uploadService{
		bucket:   bucket,
		endpoint: cfg.OSSEndpoint,
		bucketID: cfg.OSSBucketName,
		basePath: basePath,
	}, nil
}

func (s *uploadService) UploadAvatar(_ context.Context, file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", ErrInvalidImageType
	}
	if file.Size <= 0 || file.Size > maxAvatarSize {
		return "", ErrImageTooLarge
	}

	contentType := strings.ToLower(strings.TrimSpace(file.Header.Get("Content-Type")))
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return "", ErrInvalidImageType
	}

	src, err := file.Open()
	if err != nil {
		return "", ErrImageUploadFailed
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
	}

	objectKey := path.Join(
		s.basePath,
		time.Now().Format("2006/01/02"),
		fmt.Sprintf("%d%s", time.Now().UnixNano(), ext),
	)

	if err := s.bucket.PutObject(objectKey, src, oss.ContentType(contentType)); err != nil {
		return "", ErrImageUploadFailed
	}

	url := fmt.Sprintf("https://%s.%s/%s", s.bucketID, s.endpoint, objectKey)
	return url, nil
}
