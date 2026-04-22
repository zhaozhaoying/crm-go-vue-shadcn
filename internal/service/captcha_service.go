package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

var (
	ErrCaptchaInvalid         = errors.New("captcha invalid")
	ErrCaptchaExpired         = errors.New("captcha expired")
	ErrCaptchaTooManyAttempts = errors.New("captcha too many attempts")
)

const (
	defaultCaptchaTTL         = 2 * time.Minute
	defaultCaptchaMaxAttempts = 5
	defaultCaptchaLength      = 4
	defaultCleanupInterval    = 30 * time.Second
	captchaCharset            = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

type CaptchaService interface {
	Generate(ctx context.Context, fingerprint string) (*CaptchaChallenge, error)
	VerifyAndConsume(ctx context.Context, captchaID, captchaCode, fingerprint string) error
}

type CaptchaChallenge struct {
	CaptchaID    string `json:"captchaId"`
	CaptchaImage string `json:"captchaImage"`
	ExpiresAt    int64  `json:"expiresAt"`
}

type captchaRecord struct {
	answerHash      [32]byte
	fingerprintHash [32]byte
	expiresAt       time.Time
	failedAttempts  int
}

type inMemoryCaptchaService struct {
	mu          sync.Mutex
	records     map[string]captchaRecord
	ttl         time.Duration
	maxAttempts int
}

func NewCaptchaService(ttl time.Duration, maxAttempts int) CaptchaService {
	if ttl <= 0 {
		ttl = defaultCaptchaTTL
	}
	if maxAttempts <= 0 {
		maxAttempts = defaultCaptchaMaxAttempts
	}

	svc := &inMemoryCaptchaService{
		records:     make(map[string]captchaRecord),
		ttl:         ttl,
		maxAttempts: maxAttempts,
	}
	go svc.cleanupLoop()
	return svc
}

func (s *inMemoryCaptchaService) Generate(_ context.Context, fingerprint string) (*CaptchaChallenge, error) {
	captchaID, err := generateTokenID()
	if err != nil {
		return nil, err
	}

	answer, err := randomString(defaultCaptchaLength, captchaCharset)
	if err != nil {
		return nil, err
	}

	imageData, err := buildCaptchaImageData(answer)
	if err != nil {
		return nil, err
	}

	record := captchaRecord{
		answerHash:      hashNormalized(answer),
		fingerprintHash: hashNormalized(fingerprint),
		expiresAt:       time.Now().Add(s.ttl),
	}

	s.mu.Lock()
	s.records[captchaID] = record
	s.cleanExpiredLocked(time.Now())
	s.mu.Unlock()

	return &CaptchaChallenge{
		CaptchaID:    captchaID,
		CaptchaImage: imageData,
		ExpiresAt:    record.expiresAt.Unix(),
	}, nil
}

func (s *inMemoryCaptchaService) VerifyAndConsume(_ context.Context, captchaID, captchaCode, fingerprint string) error {
	now := time.Now()
	id := strings.TrimSpace(captchaID)
	if id == "" {
		return ErrCaptchaInvalid
	}

	codeHash := hashNormalized(captchaCode)
	fingerprintHash := hashNormalized(fingerprint)

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[id]
	if !ok {
		return ErrCaptchaExpired
	}
	if now.After(record.expiresAt) {
		delete(s.records, id)
		return ErrCaptchaExpired
	}
	if subtle.ConstantTimeCompare(record.fingerprintHash[:], fingerprintHash[:]) != 1 {
		delete(s.records, id)
		return ErrCaptchaInvalid
	}
	if subtle.ConstantTimeCompare(record.answerHash[:], codeHash[:]) != 1 {
		record.failedAttempts++
		if record.failedAttempts >= s.maxAttempts {
			delete(s.records, id)
			return ErrCaptchaTooManyAttempts
		}
		s.records[id] = record
		return ErrCaptchaInvalid
	}

	delete(s.records, id)
	return nil
}

func (s *inMemoryCaptchaService) cleanupLoop() {
	ticker := time.NewTicker(defaultCleanupInterval)
	defer ticker.Stop()

	for now := range ticker.C {
		s.mu.Lock()
		s.cleanExpiredLocked(now)
		s.mu.Unlock()
	}
}

func (s *inMemoryCaptchaService) cleanExpiredLocked(now time.Time) {
	for id, record := range s.records {
		if now.After(record.expiresAt) {
			delete(s.records, id)
		}
	}
}

func hashNormalized(value string) [32]byte {
	return sha256.Sum256([]byte(strings.ToUpper(strings.TrimSpace(value))))
}

func randomString(length int, charset string) (string, error) {
	if length <= 0 {
		return "", nil
	}
	var builder strings.Builder
	builder.Grow(length)
	maxIdx := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, maxIdx)
		if err != nil {
			return "", err
		}
		builder.WriteByte(charset[n.Int64()])
	}
	return builder.String(), nil
}

func randomInt(min, max int) (int, error) {
	if max <= min {
		return min, nil
	}
	width := max - min + 1
	n, err := rand.Int(rand.Reader, big.NewInt(int64(width)))
	if err != nil {
		return 0, err
	}
	return min + int(n.Int64()), nil
}

func randomColorHex(min, max int) (string, error) {
	r, err := randomInt(min, max)
	if err != nil {
		return "", err
	}
	g, err := randomInt(min, max)
	if err != nil {
		return "", err
	}
	b, err := randomInt(min, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b), nil
}

func randomColorRGBA(min, max int, alpha float64) (string, error) {
	r, err := randomInt(min, max)
	if err != nil {
		return "", err
	}
	g, err := randomInt(min, max)
	if err != nil {
		return "", err
	}
	b, err := randomInt(min, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("rgba(%d,%d,%d,%.2f)", r, g, b, alpha), nil
}

func buildCaptchaImageData(answer string) (string, error) {
	const (
		width  = 120
		height = 40
	)

	bgColor, err := randomColorHex(238, 252)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height))
	builder.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="%s"/>`, width, height, bgColor))

	for i := 0; i < 4; i++ {
		x1, err := randomInt(0, width)
		if err != nil {
			return "", err
		}
		y1, err := randomInt(0, height)
		if err != nil {
			return "", err
		}
		x2, err := randomInt(0, width)
		if err != nil {
			return "", err
		}
		y2, err := randomInt(0, height)
		if err != nil {
			return "", err
		}
		lineColor, err := randomColorRGBA(120, 210, 0.35)
		if err != nil {
			return "", err
		}
		builder.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1"/>`, x1, y1, x2, y2, lineColor))
	}

	runes := []rune(answer)
	for i, ch := range runes {
		xOffset, err := randomInt(-2, 2)
		if err != nil {
			return "", err
		}
		yOffset, err := randomInt(-3, 3)
		if err != nil {
			return "", err
		}
		rotateDeg, err := randomInt(-20, 20)
		if err != nil {
			return "", err
		}
		fontSize, err := randomInt(18, 24)
		if err != nil {
			return "", err
		}
		textColor, err := randomColorHex(60, 150)
		if err != nil {
			return "", err
		}

		x := 12 + i*26 + xOffset
		y := 27 + yOffset
		builder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" fill="%s" font-size="%d" font-family="Arial, sans-serif" font-weight="700" transform="rotate(%d %d %d)">%c</text>`, x, y, textColor, fontSize, rotateDeg, x, y, ch))
	}

	for i := 0; i < 24; i++ {
		cx, err := randomInt(0, width)
		if err != nil {
			return "", err
		}
		cy, err := randomInt(0, height)
		if err != nil {
			return "", err
		}
		dotColor, err := randomColorRGBA(120, 220, 0.45)
		if err != nil {
			return "", err
		}
		builder.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="1" fill="%s"/>`, cx, cy, dotColor))
	}

	builder.WriteString(`</svg>`)
	encoded := base64.StdEncoding.EncodeToString([]byte(builder.String()))
	return "data:image/svg+xml;base64," + encoded, nil
}
