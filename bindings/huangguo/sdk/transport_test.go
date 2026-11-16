package sdk

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// fakeTransport 是可注入的 Transport，供后续 Task 的 edge / api 测试复用。
// 它只记录 Authorization 头并返回预置响应，不走网络。
type fakeTransport struct {
	gotAuth   string
	gotURL    string
	gotMethod string
	resp      *Response
	err       error
}

func (f *fakeTransport) Do(ctx context.Context, r *Request) (*Response, error) {
	f.gotAuth = r.Headers["Authorization"]
	f.gotURL = r.URL
	f.gotMethod = r.Method
	if f.err != nil {
		return nil, f.err
	}
	if f.resp != nil {
		return f.resp, nil
	}
	return &Response{StatusCode: 200, Body: []byte("null")}, nil
}

// recordingRoundTripper 捕获真实 HTTPTransport 发出的 *http.Request 头，
// 用于证明 HTTPTransport 不像 hongguo 那样拒绝 Authorization: Bearer。
type recordingRoundTripper struct {
	gotAuth string
	gotCT   string
}

func (rt *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.gotAuth = req.Header.Get("Authorization")
	rt.gotCT = req.Header.Get("Content-Type")
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func TestTransportAllowsBearer(t *testing.T) {
	rt := &recordingRoundTripper{}
	tr := NewHTTPTransport(TransportConfig{})
	tr.client.Transport = rt

	_, err := tr.Do(context.Background(), &Request{
		Method: "POST",
		URL:    "https://huangguoai.com/api/edge/request",
		Headers: map[string]string{
			"Authorization": "Bearer tok",
			"Content-Type":  "application/json",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rt.gotAuth != "Bearer tok" {
		t.Fatalf("Authorization not forwarded: %q", rt.gotAuth)
	}
	if rt.gotCT != "application/json" {
		t.Fatalf("Content-Type not forwarded: %q", rt.gotCT)
	}
}

func TestTargetGuards(t *testing.T) {
	tr := NewHTTPTransport(TransportConfig{})
	cases := []struct {
		url  string
		want string
	}{
		{"https://u:p@example.com/", "userinfo"},
		{"http://127.0.0.1:1/", "private"},
		{"http://192.168.1.1/", "private"},
		{"http://10.0.0.1/", "private"},
		{"http://localhost/", "private"},
	}
	for _, tc := range cases {
		_, err := tr.Do(context.Background(), &Request{Method: "GET", URL: tc.url})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: expected error containing %q, got %v", tc.url, tc.want, err)
		}
	}
}
