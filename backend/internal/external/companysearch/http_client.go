package companysearch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultUserAgent = "Mozilla/5.0"

type HTTPClientConfig struct {
	Timeout               time.Duration
	ConnectTimeout        time.Duration
	ResponseHeaderTimeout time.Duration
	RetryCount            int
	RetryWait             time.Duration
	UserAgent             string
	DisableCompression    bool
	ProxyURL              string
}

type RequestOptions struct {
	Query   map[string]string
	Headers map[string]string
}

type HTTPStatusError struct {
	URL        string
	StatusCode int
	Body       string
}

func (e *HTTPStatusError) Error() string {
	if e == nil {
		return ""
	}
	if e.Body == "" {
		return fmt.Sprintf("GET %s returned status %d", e.URL, e.StatusCode)
	}
	return fmt.Sprintf("GET %s returned status %d: %s", e.URL, e.StatusCode, e.Body)
}

type HTTPClient struct {
	client     *http.Client
	retryCount int
	retryWait  time.Duration
	userAgent  string
}

func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:               15 * time.Second,
		ConnectTimeout:        15 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		RetryCount:            2,
		RetryWait:             300 * time.Millisecond,
		UserAgent:             defaultUserAgent,
	}
}

func NewHTTPClient(cfg HTTPClientConfig) *HTTPClient {
	defaults := DefaultHTTPClientConfig()
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaults.Timeout
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = defaults.ConnectTimeout
	}
	if cfg.ResponseHeaderTimeout <= 0 {
		cfg.ResponseHeaderTimeout = defaults.ResponseHeaderTimeout
	}
	if cfg.RetryWait <= 0 {
		cfg.RetryWait = defaults.RetryWait
	}
	if strings.TrimSpace(cfg.UserAgent) == "" {
		cfg.UserAgent = defaults.UserAgent
	}
	if cfg.RetryCount < 0 {
		cfg.RetryCount = 0
	}
	proxyFunc := http.ProxyFromEnvironment
	if strings.TrimSpace(cfg.ProxyURL) != "" {
		if proxyURL, err := url.Parse(strings.TrimSpace(cfg.ProxyURL)); err == nil && proxyURL.Scheme != "" && proxyURL.Host != "" {
			proxyFunc = http.ProxyURL(proxyURL)
		}
	}

	transport := &http.Transport{
		Proxy: proxyFunc,
		DialContext: (&net.Dialer{
			Timeout: cfg.ConnectTimeout,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   cfg.ConnectTimeout,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		DisableCompression:    cfg.DisableCompression,
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		retryCount: cfg.RetryCount,
		retryWait:  cfg.RetryWait,
		userAgent:  cfg.UserAgent,
	}
}

func NewDefaultHTTPClient() *HTTPClient {
	return NewHTTPClient(DefaultHTTPClientConfig())
}

func (c *HTTPClient) Get(ctx context.Context, rawURL string, opts RequestOptions) (string, error) {
	if c == nil {
		return "", errors.New("http client is nil")
	}

	requestURL, err := buildRequestURL(rawURL, opts.Query)
	if err != nil {
		return "", err
	}

	var lastErr error
	attempts := c.retryCount + 1
	for attempt := 1; attempt <= attempts; attempt++ {
		body, err := c.doGet(ctx, requestURL, opts.Headers)
		if err == nil {
			return body, nil
		}

		lastErr = err
		if !shouldRetry(err) || attempt == attempts {
			break
		}
		if waitErr := waitWithContext(ctx, time.Duration(attempt)*c.retryWait); waitErr != nil {
			return "", waitErr
		}
	}

	return "", lastErr
}

func (c *HTTPClient) doGet(ctx context.Context, requestURL string, headers map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	for key, value := range headers {
		if strings.TrimSpace(key) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	if c.userAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", &HTTPStatusError{
			URL:        requestURL,
			StatusCode: resp.StatusCode,
			Body:       truncateForError(string(data), 256),
		}
	}

	return string(data), nil
}

func buildRequestURL(rawURL string, query map[string]string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", fmt.Errorf("parse url %q: %w", rawURL, err)
	}

	if len(query) == 0 {
		return parsed.String(), nil
	}

	values := parsed.Query()
	for key, value := range query {
		if strings.TrimSpace(key) == "" {
			continue
		}
		values.Set(key, value)
	}
	parsed.RawQuery = values.Encode()
	return parsed.String(), nil
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	var statusErr *HTTPStatusError
	if errors.As(err, &statusErr) {
		return statusErr.StatusCode == http.StatusTooManyRequests ||
			statusErr.StatusCode == http.StatusBadGateway ||
			statusErr.StatusCode == http.StatusServiceUnavailable ||
			statusErr.StatusCode == http.StatusGatewayTimeout ||
			statusErr.StatusCode >= http.StatusInternalServerError
	}

	var netErr net.Error
	return errors.As(err, &netErr)
}

func waitWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func truncateForError(raw string, limit int) string {
	if limit <= 0 || len(raw) <= limit {
		return raw
	}
	return raw[:limit] + "..."
}
