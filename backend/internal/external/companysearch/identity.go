package companysearch

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"
)

func NewTaskNo() string {
	return buildNumber("ecs")
}

func NewCompanyNo() string {
	return buildNumber("ec")
}

func NormalizeKeyword(raw string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(raw)), " "))
}

func BuildDedupeKey(platformCompanyID, companyURL, companyName, city string) string {
	if value := strings.TrimSpace(platformCompanyID); value != "" {
		return "id:" + value
	}
	if value := normalizeURLForDedupe(companyURL); value != "" {
		return "url:" + value
	}
	fallback := strings.Join([]string{
		NormalizeKeyword(companyName),
		NormalizeKeyword(city),
	}, "|")
	if fallback == "|" {
		return "hash:" + shortSHA1(time.Now().UTC().Format(time.RFC3339Nano))
	}
	return "hash:" + shortSHA1(fallback)
}

func NormalizeAbsoluteURL(baseURL, raw string) string {
	raw = ProcessURLProtocol(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err == nil && parsed.IsAbs() {
		return parsed.String()
	}
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return raw
	}
	ref, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return base.ResolveReference(ref).String()
}

func ExtractMadeInChinaCompanyID(rawURL string) string {
	normalized := normalizeURLForDedupe(rawURL)
	if normalized == "" {
		return ""
	}
	parsed, err := url.Parse(normalized)
	if err != nil {
		return shortSHA1(normalized)
	}
	segment := path.Base(strings.Trim(parsed.Path, "/"))
	segment = strings.TrimSpace(strings.TrimSuffix(segment, path.Ext(segment)))
	if segment == "" || segment == "." || segment == "/" {
		return shortSHA1(normalized)
	}
	if len(segment) > 100 {
		return shortSHA1(segment)
	}
	return segment
}

func buildNumber(prefix string) string {
	return fmt.Sprintf("%s_%s_%s", prefix, time.Now().UTC().Format("20060102150405"), randomHex(4))
}

func randomHex(byteCount int) string {
	if byteCount <= 0 {
		byteCount = 4
	}
	buf := make([]byte, byteCount)
	if _, err := rand.Read(buf); err != nil {
		return shortSHA1(fmt.Sprintf("fallback-%d", time.Now().UTC().UnixNano()))
	}
	return hex.EncodeToString(buf)
}

func shortSHA1(raw string) string {
	sum := sha1.Sum([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:8])
}

func normalizeURLForDedupe(raw string) string {
	raw = ProcessURLProtocol(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	parsed.Fragment = ""
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Path = strings.TrimRight(parsed.EscapedPath(), "/")
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String()
}
