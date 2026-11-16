package sdk

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func req(t *testing.T, raw string) *http.Request {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Request{URL: u}
}

// 每跳重新做目标守卫：重定向到内网/回环必须被拒。
func TestMediaRedirectTargetGuard(t *testing.T) {
	via := []*http.Request{req(t, "https://cdn.example.com/a.ts")}
	to := req(t, "http://127.0.0.1/secret")
	err := mediaCheckRedirect(to, via)
	if err == nil || !strings.Contains(err.Error(), "private") {
		t.Fatalf("redirect to loopback must be rejected, got %v", err)
	}
	// 合法公网跳转放行。
	if err := mediaCheckRedirect(req(t, "https://cdn2.example.com/b.ts"), via); err != nil {
		t.Fatalf("public redirect should pass: %v", err)
	}
}

// 禁止 HTTPS → HTTP 降级。
func TestMediaRedirectNoDowngrade(t *testing.T) {
	via := []*http.Request{req(t, "https://cdn.example.com/a.ts")}
	err := mediaCheckRedirect(req(t, "http://cdn.example.com/a.ts"), via)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "downgrade") {
		t.Fatalf("https->http downgrade must be rejected, got %v", err)
	}
	// HTTP → HTTPS 允许。
	via2 := []*http.Request{req(t, "http://cdn.example.com/a.ts")}
	if err := mediaCheckRedirect(req(t, "https://cdn.example.com/b.ts"), via2); err != nil {
		t.Fatalf("http->https upgrade should pass: %v", err)
	}
}

// 最多跟随 5 跳：via 长度 5（即将跟第 5 跳）应放行，长度 6（第 6 跳）才拒绝。
func TestMediaRedirectMaxHops(t *testing.T) {
	via5 := make([]*http.Request, 5)
	for i := range via5 {
		via5[i] = req(t, "https://cdn.example.com/x")
	}
	if err := mediaCheckRedirect(req(t, "https://cdn.example.com/y"), via5); err != nil {
		t.Fatalf("5th hop should be allowed, got %v", err)
	}
	via6 := make([]*http.Request, 6)
	for i := range via6 {
		via6[i] = req(t, "https://cdn.example.com/x")
	}
	err := mediaCheckRedirect(req(t, "https://cdn.example.com/y"), via6)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "redirect") {
		t.Fatalf("must reject 6th hop, got %v", err)
	}
}

// edge 请求仍禁自动重定向（NoRedirect 保持 ErrUseLastResponse），媒体请求开启逐跳守卫。
func TestFetchSegmentEnablesGuardedRedirect(t *testing.T) {
	tr := &captureRequestTransport{}
	c, _ := NewClient(Options{Transport: tr, Store: NewMemStore(), Base: "https://huangguoai.com"})
	_, _ = c.FetchSegment(context.Background(), "https://cdn.example.com/seg.ts", nil)
	if tr.got == nil || !tr.got.FollowMediaRedirects {
		t.Fatal("FetchSegment must enable guarded media redirects")
	}
}

type captureRequestTransport struct{ got *Request }

func (c *captureRequestTransport) Do(ctx context.Context, r *Request) (*Response, error) {
	c.got = r
	return &Response{StatusCode: 200, Body: []byte("x")}, nil
}
