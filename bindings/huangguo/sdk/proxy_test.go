package sdk

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// 解析显式代理：scheme 必须是 http/https，拒绝 SOCKS，凭据脱敏。
func TestParseProxyURL(t *testing.T) {
	u, err := parseProxyURL("http://user:pass@proxy.example:8080")
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "proxy.example:8080" || u.User.Username() != "user" {
		t.Fatalf("bad proxy: %v", u)
	}
	if got := redactProxyURL(u); strings.Contains(got, "pass") || !strings.Contains(got, "***") {
		t.Fatalf("credentials not redacted: %s", got)
	}
	for _, bad := range []string{"socks5://127.0.0.1:1080", "ftp://proxy.example", "://bad"} {
		if _, err := parseProxyURL(bad); err == nil {
			t.Fatalf("should reject proxy %s", bad)
		}
	}
}

// 三态：显式 URL 强代理；显式空串强制直连；nil 走环境回退。
func TestProxyResolution(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://env-proxy.example:3128")
	t.Setenv("HTTPS_PROXY", "http://env-https.example:3129")
	t.Setenv("NO_PROXY", "huangguoai.ai")
	req := &http.Request{URL: &url.URL{Scheme: "https", Host: "huangguoai.com"}}
	noProxyReq := &http.Request{URL: &url.URL{Scheme: "https", Host: "huangguoai.ai"}}

	// nil → 环境回退，遵循 NO_PROXY。
	fn := makeProxyFunc(nil)
	if got, _ := fn(req); got == nil || got.Host != "env-https.example:3129" {
		t.Fatalf("env proxy: %v", got)
	}
	if got, _ := fn(noProxyReq); got != nil {
		t.Fatalf("NO_PROXY should bypass env proxy: %v", got)
	}

	// 显式代理：不受 NO_PROXY 影响。
	explicit := "http://explicit.example:8080"
	fn = makeProxyFunc(&explicit)
	if got, _ := fn(noProxyReq); got == nil || got.Host != "explicit.example:8080" {
		t.Fatalf("explicit proxy should ignore NO_PROXY: %v", got)
	}

	// 显式空串 → 强制直连，忽略环境。
	direct := ""
	fn = makeProxyFunc(&direct)
	if got, _ := fn(req); got != nil {
		t.Fatalf("explicit direct should bypass env: %v", got)
	}
}

// HGHUB_PROXY 优先级最高，ALL_PROXY 兜底。
func TestProxyEnvPriority(t *testing.T) {
	t.Setenv("HGHUB_PROXY", "http://hghub.example:9000")
	t.Setenv("HTTPS_PROXY", "http://env-https.example:3129")
	t.Setenv("ALL_PROXY", "http://all.example:8081")
	t.Setenv("NO_PROXY", "")
	req := &http.Request{URL: &url.URL{Scheme: "https", Host: "huangguoai.com"}}
	if got, _ := makeProxyFunc(nil)(req); got == nil || got.Host != "hghub.example:9000" {
		t.Fatalf("HGHUB_PROXY priority: %v", got)
	}
	t.Setenv("HGHUB_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	if got, _ := makeProxyFunc(nil)(req); got == nil || got.Host != "all.example:8081" {
		t.Fatalf("ALL_PROXY fallback: %v", got)
	}
}

// 代理凭据出现在错误里必须脱敏。
func TestProxyErrorRedacted(t *testing.T) {
	_, err := parseProxyURL("socks5://user:supersecret@proxy.example:1080")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "supersecret") {
		t.Fatalf("proxy credential leaked in error: %s", err)
	}
}

// 构造出的 HTTPTransport 必须把代理函数真正装到 http.Client 上。
func TestHTTPTransportInstallsProxy(t *testing.T) {
	explicit := "http://explicit.example:8080"
	tr := NewHTTPTransport(TransportConfig{Proxy: &explicit})
	ht, ok := tr.client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", tr.client.Transport)
	}
	got, err := ht.Proxy(&http.Request{URL: &url.URL{Scheme: "https", Host: "huangguoai.ai"}})
	if err != nil || got == nil || got.Host != "explicit.example:8080" {
		t.Fatalf("proxy not installed: %v %v", got, err)
	}
}
