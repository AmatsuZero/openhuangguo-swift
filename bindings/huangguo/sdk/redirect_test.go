package sdk

import (
	"context"
	"errors"
	"io"
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

func TestSignedMediaHeadersAreNotForwardedToAnotherOrigin(t *testing.T) {
	from := req(t, "https://cdn.example.com/main.m3u8")
	to := req(t, "https://other.example.com/part.ts")
	to.Header = http.Header{"X-Cdn-Auth": []string{"private"}, "X-Cdn-App": []string{"player"}, "Range": []string{"bytes=0-9"}}
	if err := mediaCheckRedirect(to, []*http.Request{from}); err != nil {
		t.Fatal(err)
	}
	if to.Header.Get("X-Cdn-Auth") != "" || to.Header.Get("X-Cdn-App") != "" || to.Header.Get("Range") != "bytes=0-9" {
		t.Fatalf("unsafe redirect headers: %v", to.Header)
	}
}

// Exercise net/http's redirect chain: it copies the initial request headers on
// every hop, so checking only the immediately preceding origin leaks them again.
func TestMediaRedirectDoesNotRestoreCustomCredentials(t *testing.T) {
	tr := NewHTTPTransport(TransportConfig{})
	var seen []http.Header
	tr.client.Transport = redirectRoundTripper(func(r *http.Request) (*http.Response, error) {
		seen = append(seen, r.Header.Clone())
		location := ""
		switch len(seen) {
		case 1:
			location = "https://other.example/child.m3u8"
		case 2:
			location = "https://other.example/final.m3u8"
		}
		status := 200
		headers := make(http.Header)
		if location != "" {
			status = 302
			headers.Set("Location", location)
		}
		return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader("#EXTM3U\n")), Request: r}, nil
	})
	c, _ := NewClient(Options{Transport: tr, Store: NewMemStore()})
	result, err := c.FetchSegment(context.Background(), "https://cdn.example/master.m3u8", map[string]string{
		"X-Playback-Ticket": "secret", "X-Cdn-Auth": "signed", "Authorization": "Bearer private", "Cookie": "sid=private", "Range": "bytes=0-99",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.URL != "https://other.example/final.m3u8" {
		t.Fatalf("lost final manifest URL: %s", result.URL)
	}
	if len(seen) != 3 || seen[0].Get("X-Playback-Ticket") != "secret" {
		t.Fatalf("unexpected initial request or redirect count: %v", seen)
	}
	for hop, headers := range seen[1:] {
		for _, name := range []string{"X-Playback-Ticket", "X-Cdn-Auth", "Authorization", "Cookie"} {
			if headers.Get(name) != "" {
				t.Errorf("hop %d forwarded %s", hop+1, name)
			}
		}
		if headers.Get("Range") != "bytes=0-99" {
			t.Errorf("hop %d lost Range", hop+1)
		}
	}
}

type redirectRoundTripper func(*http.Request) (*http.Response, error)

func (f redirectRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMediaRedirectPreservesHeadersForEquivalentOrigins(t *testing.T) {
	from := req(t, "https://CDN.example:443/master.m3u8")
	to := req(t, "https://cdn.example/child.m3u8")
	to.Header = http.Header{"X-Playback-Ticket": []string{"secret"}, "X-Cdn-Auth": []string{"signed"}}
	if err := mediaCheckRedirect(to, []*http.Request{from}); err != nil {
		t.Fatal(err)
	}
	if to.Header.Get("X-Cdn-Auth") != "signed" || to.Header.Get("X-Playback-Ticket") != "secret" {
		t.Fatal("same-origin redirect lost playback headers")
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

// FetchSegment 保留底层错误链：transport 返回取消错误时 errors.Is 必须识别。
func TestFetchSegmentErrorChain(t *testing.T) {
	cErr := &cancelTransport{}
	c, _ := NewClient(Options{Transport: cErr, Store: NewMemStore(), Base: "https://huangguoai.com"})
	_, err := c.FetchSegment(context.Background(), "https://cdn.example.com/seg.ts", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("media cancellation cause lost: %v", err)
	}
}

type cancelTransport struct{}

func (cancelTransport) Do(ctx context.Context, r *Request) (*Response, error) {
	return nil, context.Canceled
}
