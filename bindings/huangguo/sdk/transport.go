package sdk

import (
	"bytes"
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

// Transport 是内部 HTTP 抽象（不 gomobile 绑定）。测试注入 fake 实现。
type Transport interface {
	Do(ctx context.Context, req *Request) (*Response, error)
}

type Request struct {
	Method           string
	URL              string
	Headers          map[string]string
	Body             []byte
	MaxResponseBytes int64 // 媒体设大值，API 用默认 8MB
}

type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// TransportConfig 由 configJSON 构造。Proxy: nil=走环境；指向 ""=强制直连；指向 URL=显式代理。
type TransportConfig struct {
	Proxy *string
}

// HTTPTransport 是 Transport 的真实实现，持有可注入的 *http.Client（测试塞 fake RoundTripper）。
type HTTPTransport struct {
	client *http.Client
	cfg    TransportConfig
}

func NewHTTPTransport(cfg TransportConfig) *HTTPTransport {
	return &HTTPTransport{client: &http.Client{Timeout: 30 * time.Second}, cfg: cfg}
}

// guardTarget 对目标 URL 做轻守卫：拒绝 userinfo、非 http(s)、内网/回环目标。
// 代理 URL 不走这里（分开校验，见 Task 4）。
func guardTarget(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid target URL: %w", err)
	}
	if u.User != nil {
		return errors.New("reject target with userinfo")
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	default:
		return fmt.Errorf("unsupported target scheme %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("reject target with empty host")
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return errors.New("reject private/loopback target")
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
			ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return errors.New("reject private/loopback target")
		}
	}
	return nil
}

// Do 发送请求。目标守卫在开头；Authorization（Bearer）原样放行，区别于 hongguo。
// 响应默认 8MB 上限；媒体可通过 req.MaxResponseBytes 提高。
// 代理解析与首版禁重定向策略在 Task 4 补齐。
func (t *HTTPTransport) Do(ctx context.Context, req *Request) (*Response, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	if err := guardTarget(req.URL); err != nil {
		return nil, err
	}
	method := req.Method
	if method == "" {
		method = http.MethodGet
	}
	hreq, err := http.NewRequestWithContext(ctx, method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return nil, err
	}
	for k, v := range req.Headers {
		hreq.Header.Set(k, v)
	}
	resp, err := t.client.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	limit := int64(8 << 20)
	if req.MaxResponseBytes > 0 {
		limit = req.MaxResponseBytes
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("response exceeds %d bytes", limit)
	}
	headers := make(map[string]string)
	for k, values := range resp.Header {
		if len(values) > 0 {
			headers[k] = values[0]
		}
	}
	return &Response{StatusCode: resp.StatusCode, Headers: headers, Body: body}, nil
}
