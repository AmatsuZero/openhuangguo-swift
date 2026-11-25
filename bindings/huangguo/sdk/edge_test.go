package sdk

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestEdgeHTTPStatusBeforeResponseParse(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		body    string
		wantErr bool
		errIs   error
	}{
		{"success envelope", 200, `{"status":1,"data":{}}`, false, nil},
		{"empty body success", 200, ``, false, nil},
		{"null success", 200, `null`, false, nil},
		{"data null success", 200, `{"status":1,"data":null}`, false, nil},
		{"bare {} fails", 200, `{}`, true, ErrCode},
		{"401 fails", 401, ``, true, ErrAuth},
		{"403 fails", 403, ``, true, ErrAuth},
		{"500 fails", 500, ``, true, ErrRetryable},
		{"408 fails", 408, ``, true, ErrRetryable},
		{"429 fails", 429, ``, true, ErrRetryable},
		{"business status 0 fails", 200, `{"status":0,"msg":"bad"}`, true, ErrCode},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeTransport{resp: &Response{StatusCode: tc.status, Body: []byte(tc.body)}}
			c, _ := NewClient(Options{Transport: fake, Store: NewMemStore(), Base: "https://huangguoai.com"})
			got, err := c.edge(context.Background(), "GET", "/api/videos", nil, false)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got result %s", got)
				}
				if tc.errIs != nil && !errors.Is(err, tc.errIs) {
					t.Fatalf("expected errors.Is %v, got %v", tc.errIs, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

type captureRoundTripper struct {
	method  string
	url     string
	auth    string
	ct      string
	referer string
	accept  string
	body    []byte
}

func (rt *captureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.method = req.Method
	rt.url = req.URL.String()
	rt.auth = req.Header.Get("Authorization")
	rt.ct = req.Header.Get("Content-Type")
	rt.referer = req.Header.Get("Referer")
	rt.accept = req.Header.Get("Accept")
	rt.body = nil
	if req.Body != nil {
		rt.body, _ = io.ReadAll(req.Body)
	}
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"status":1,"data":{}}`)),
	}, nil
}

func TestEdgeDirectRequestFields(t *testing.T) {
	rt := &captureRoundTripper{}
	tr := NewHTTPTransport(TransportConfig{})
	tr.client.Transport = rt
	store := NewMemStore()
	_ = store.Save(Session{Origin: "https://huangguoai.com", Token: "tok", Generation: 1})
	c, _ := NewClient(Options{Transport: tr, Store: store, Base: "https://huangguoai.com"})

	// 无 body：明文 GET，直连 path，无 Content-Type。
	if _, err := c.edge(context.Background(), "GET", "/api/videos", nil, false); err != nil {
		t.Fatal(err)
	}
	if rt.method != "GET" || rt.url != "https://huangguoai.com/api/videos" {
		t.Fatalf("request = %s %s", rt.method, rt.url)
	}
	if len(rt.body) != 0 || rt.ct != "" {
		t.Fatalf("GET body = %q, Content-Type = %q", rt.body, rt.ct)
	}
	if rt.auth != "Bearer tok" {
		t.Fatalf("Authorization = %q", rt.auth)
	}
	if rt.accept != "application/json" {
		t.Fatalf("Accept = %q", rt.accept)
	}
	if rt.referer != "https://huangguoai.com/" {
		t.Fatalf("Referer = %q", rt.referer)
	}

	// 有 body：原始 JSON，非 base64。
	payload := map[string]any{"username": "u", "password": "p"}
	if _, err := c.edge(context.Background(), "POST", "/api/account/loginByPassword", payload, true); err != nil {
		t.Fatal(err)
	}
	if rt.method != "POST" || rt.url != "https://huangguoai.com/api/account/loginByPassword" {
		t.Fatalf("request = %s %s", rt.method, rt.url)
	}
	if rt.ct != "application/json" {
		t.Fatalf("Content-Type = %q", rt.ct)
	}
	if !strings.Contains(string(rt.body), `"password":"p"`) {
		t.Fatalf("body = %s", rt.body)
	}
}

func TestEdgeRejectsCredentialOnUntrustedBase(t *testing.T) {
	fake := &fakeTransport{resp: &Response{StatusCode: 200, Body: []byte(`{"status":1,"data":{}}`)}}
	c, _ := NewClient(Options{
		Transport: fake,
		Store:     NewMemStore(),
		Base:      "https://mirror.example",
		CredTrust: map[string]bool{}, // 镜像不收密码/Bearer
	})
	if _, err := c.edge(context.Background(), "POST", "/api/account/loginByPassword", map[string]string{}, true); err == nil {
		t.Fatal("expected credential rejection on untrusted base")
	}
	// 带 token 的 GET 同样被镜像拒绝。
	store := NewMemStore()
	_ = store.Save(Session{Origin: "https://mirror.example", Token: "tok", Generation: 1})
	c2, _ := NewClient(Options{Transport: fake, Store: store, Base: "https://mirror.example", CredTrust: map[string]bool{}})
	if _, err := c2.edge(context.Background(), "GET", "/api/account/me", nil, false); err == nil {
		t.Fatal("expected rejection for Bearer on untrusted base")
	}
}
