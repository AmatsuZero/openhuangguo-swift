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

// 畸形代理 URL 的解析错误不得泄漏原始 URL 里的凭据。
func TestProxyParseErrorRedactsCredentials(t *testing.T) {
	_, err := parseProxyURL("http://user:supersecret%zz@proxy.example:8080")
	if err == nil {
		t.Fatal("expected parse error")
	}
	if strings.Contains(err.Error(), "supersecret") {
		t.Fatalf("proxy credential leaked in parse error: %s", err)
	}
}

// NO_PROXY 支持端口、CIDR、前导点子域、精确匹配。
func TestNoProxySemantics(t *testing.T) {
	cases := []struct {
		hostPort string
		rule     string
		want     bool
	}{
		{"localhost:8080", "localhost:8080", true},  // 端口精确
		{"localhost:9090", "localhost:8080", false}, // 端口不符不走 bypass
		{"api.example.com", ".example.com", true},   // 前导点子域
		{"example.com", ".example.com", true},       // 含裸域
		{"notexample.com", ".example.com", false},   // 不能后缀误匹配
		{"10.1.2.3", "10.0.0.0/8", true},            // CIDR
		{"11.1.2.3", "10.0.0.0/8", false},
		{"192.168.1.5", "192.168.0.0/16", true},
		{"example.com", "*", true},                          // 全匹配
		{"example.com", "foo, example.com", true},           // 规则列表（带空格）
		{"cdn.example.com:443", ".example.com:443", true},   // 前导点 + 端口
		{"api.example.com:8080", ".example.com:443", false}, // 端口不符
	}
	for _, tc := range cases {
		if got := noProxyBypass(tc.hostPort, tc.rule); got != tc.want {
			t.Errorf("noProxyBypass(%q, %q) = %v, want %v", tc.hostPort, tc.rule, got, tc.want)
		}
	}
}

// 环境回退时按 scheme 补默认端口：https://example.com 应命中 NO_PROXY=example.com:443。
func TestEnvProxyNoProxyDefaultPort(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://env-proxy.example:3128")
	t.Setenv("NO_PROXY", "example.com:443")
	t.Setenv("HGHUB_PROXY", "")
	t.Setenv("ALL_PROXY", "")
	req := &http.Request{URL: &url.URL{Scheme: "https", Host: "example.com"}} // 无显式端口
	if got, _ := makeProxyFunc(nil)(req); got != nil {
		t.Fatalf("https default port 443 should hit NO_PROXY, got %v", got)
	}
	// 不同端口仍走代理。
	req2 := &http.Request{URL: &url.URL{Scheme: "https", Host: "example.com:8443"}}
	if got, _ := makeProxyFunc(nil)(req2); got == nil {
		t.Fatal("non-443 port should use proxy")
	}
}
