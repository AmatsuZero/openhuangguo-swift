package sdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestEdgeHTTPStatusBeforeEnvelope(t *testing.T) {
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
	auth    string
	ct      string
	referer string
	accept  string
	body    []byte
}

func (rt *captureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.auth = req.Header.Get("Authorization")
	rt.ct = req.Header.Get("Content-Type")
	rt.referer = req.Header.Get("Referer")
	rt.accept = req.Header.Get("Accept")
	if req.Body != nil {
		rt.body, _ = io.ReadAll(req.Body)
	}
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"status":1,"data":{}}`)),
	}, nil
}

func TestEdgeEnvelopeFields(t *testing.T) {
	rt := &captureRoundTripper{}
	tr := NewHTTPTransport(TransportConfig{})
	tr.client.Transport = rt
	store := NewMemStore()
	_ = store.Save(Session{Origin: "https://huangguoai.com", Token: "tok", Generation: 1})
	c, _ := NewClient(Options{Transport: tr, Store: store, Base: "https://huangguoai.com"})

	// 无 body：两空串。
	if _, err := c.edge(context.Background(), "GET", "/api/videos", nil, false); err != nil {
		t.Fatal(err)
	}
	var env edgeEnvelope
	if err := json.Unmarshal(rt.body, &env); err != nil {
		t.Fatal(err)
	}
	if env.Method != "GET" || env.Path != "/api/videos" || env.Body != "" || env.ContentType != "" {
		t.Fatalf("bad no-body envelope: %+v", env)
	}
	if rt.auth != "Bearer tok" {
		t.Fatalf("Authorization = %q", rt.auth)
	}
	if rt.ct != "application/json" {
		t.Fatalf("outer Content-Type = %q", rt.ct)
	}
	if rt.referer != "https://huangguoai.com/" {
		t.Fatalf("Referer = %q", rt.referer)
	}

	// 有 body：base64 + application/json。
	payload := map[string]any{"username": "u", "password": "p"}
	if _, err := c.edge(context.Background(), "POST", "/api/account/loginByPassword", payload, true); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(rt.body, &env); err != nil {
		t.Fatal(err)
	}
	if env.Method != "POST" || env.Path != "/api/account/loginByPassword" || env.ContentType != "application/json" {
		t.Fatalf("bad body envelope: %+v", env)
	}
	decoded, err := base64.StdEncoding.DecodeString(env.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(decoded), `"password":"p"`) {
		t.Fatalf("decoded body = %s", decoded)
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
