package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"backend/internal/config"
)

const defaultPerPage = 100

type seatStatisticsResponse struct {
	Code int    `json:"code"`
	Info string `json:"info"`
	Data struct {
		List            []json.RawMessage `json:"list"`
		GroupTotalCount int               `json:"groupTotalCount"`
	} `json:"data"`
}

func main() {
	page := flag.Int("page", 1, "米话 page 参数")
	perPage := flag.Int("per-page", defaultPerPage, "米话 per_page 参数")
	timeout := flag.Duration("timeout", 15*time.Second, "单次请求超时时间")
	flag.Parse()

	cfg := config.Load()
	if strings.TrimSpace(cfg.MiHuaCallRecordListURL) == "" ||
		strings.TrimSpace(cfg.MiHuaCallRecordToken) == "" ||
		strings.TrimSpace(cfg.MiHuaCallRecordOrigin) == "" {
		fmt.Fprintln(os.Stderr, "缺少米话配置：MIHUA_CALL_RECORD_LIST_URL / MIHUA_CALL_RECORD_TOKEN / MIHUA_CALL_RECORD_SOURCE_ORIGIN")
		os.Exit(1)
	}

	baseURL, err := url.Parse(cfg.MiHuaCallRecordListURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析 MIHUA_CALL_RECORD_LIST_URL 失败: %v\n", err)
		os.Exit(1)
	}
	originURL, err := url.Parse(cfg.MiHuaCallRecordOrigin)
	if err != nil || strings.TrimSpace(originURL.Scheme) == "" || strings.TrimSpace(originURL.Host) == "" {
		fmt.Fprintf(os.Stderr, "解析 MIHUA_CALL_RECORD_SOURCE_ORIGIN 失败: %v\n", err)
		os.Exit(1)
	}

	requestURL := *baseURL
	query := requestURL.Query()
	query.Set("page", strconv.Itoa(*page))
	query.Set("per_page", strconv.Itoa(*perPage))
	requestURL.RawQuery = query.Encode()

	origin := originURL.Scheme + "://" + originURL.Host
	referer := origin + "/"
	client := &http.Client{Timeout: *timeout}

	body, statusCode, err := doRequest(
		context.Background(),
		client,
		&requestURL,
		cfg.MiHuaCallRecordToken,
		origin,
		referer,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "请求米话失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("URL: %s\n", sanitizeURL(&requestURL))
	fmt.Printf("Status: %d\n", statusCode)
	fmt.Printf("Body: %s\n", truncateBody(string(body), 1600))

	var parsed seatStatisticsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "解析响应失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("解析结果: code=%d info=%s listCount=%d groupTotalCount=%d\n",
		parsed.Code,
		strings.TrimSpace(parsed.Info),
		len(parsed.Data.List),
		parsed.Data.GroupTotalCount,
	)
}

func doRequest(
	ctx context.Context,
	client *http.Client,
	requestURL *url.URL,
	token, origin, referer string,
) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("token", token)
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("origin", origin)
	req.Header.Set("referer", referer)
	req.Header.Set("source", "client.web")
	req.Header.Set("nonce", generateNonce())
	req.Header.Set("user-agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return body, resp.StatusCode, nil
}

func sanitizeURL(rawURL *url.URL) string {
	if rawURL == nil {
		return ""
	}
	cloned := *rawURL
	query := cloned.Query()
	if query.Has("token") {
		query.Set("token", "[REDACTED]")
	}
	cloned.RawQuery = query.Encode()
	return cloned.String()
}

func truncateBody(body string, limit int) string {
	trimmed := strings.TrimSpace(body)
	if len(trimmed) <= limit {
		return trimmed
	}
	return trimmed[:limit] + "...(truncated)"
}

func generateNonce() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buffer)
}
