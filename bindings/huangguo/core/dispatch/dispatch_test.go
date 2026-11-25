package dispatch

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/sdk"
)

type stubTransport struct{}

func (stubTransport) Do(ctx context.Context, r *sdk.Request) (*sdk.Response, error) {
	body := `{"status":1,"data":{"items":[],"pagination":{"page":1,"pages":0,"size":20,"total":0}}}`
	if strings.Contains(r.URL, "/playback") || strings.Contains(r.URL, "/play") {
		body = `{"status":1,"data":{"id":1,"ep_num":2,"duration":0,"title":"t","video_url":"https://cdn/v.mp4"}}`
	} else if strings.Contains(r.URL, "loginByPassword") || strings.Contains(r.URL, "registerByPassword") {
		body = `{"status":1,"data":{"token":"tok","user":{"id":1,"username":"u"}}}`
	}
	return &sdk.Response{StatusCode: 200, Body: []byte(body)}, nil
}

func newTestClient(t *testing.T) *sdk.Client {
	t.Helper()
	c, err := sdk.NewClient(sdk.Options{Transport: stubTransport{}, Store: sdk.NewMemStore(), Base: "https://huangguoai.com"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCallAllMethods(t *testing.T) {
	c := newTestClient(t)
	ctx := context.Background()
	cases := []struct {
		name string
		args string
	}{
		{"login", `["u","p"]`},
		{"register", `["u","p",""]`},
		{"logout", `[]`},
		{"me", `[]`},
		{"feed", `[1,20]`},
		{"category", `["drama",1,20]`},
		{"rank", `["hot"]`},
		{"search", `["q",1,20]`},
		{"posts", `[1,20]`},
		{"post", `[1]`},
		{"video", `[1]`},
		{"play", `[1,2]`},
		{"comments", `[1,2,1,20]`},
		{"commentSend", `[1,2,"hi",0]`},
		{"like", `[1,2,true]`},
		{"collect", `[1,true]`},
		{"follow", `[1,false]`},
		{"postLike", `[1,true]`},
		{"postCollect", `[1,false]`},
		{"postComments", `[1,1,20]`},
		{"setBase", `["https://huangguoai.ai"]`},
		{"decryptCover", `[ "` + base64.StdEncoding.EncodeToString([]byte{0xFF, 0xD8, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}) + `" ]`},
		{"rewritePlaylist", `["https://cdn/m.m3u8","hg","localhost","#EXTM3U\nseg.ts\n"]`},
		{"fetchSegment", `["https://cdn/seg.ts",{"Range":"bytes=0-4"}]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Call(ctx, c, tc.name, tc.args)
			if err != nil {
				t.Fatalf("Call(%s) error: %v", tc.name, err)
			}
		})
	}
}

func TestCallUnknownMethod(t *testing.T) {
	_, err := Call(context.Background(), newTestClient(t), "nope", "[]")
	if err == nil || !strings.Contains(err.Error(), "unknown method") {
		t.Fatalf("expected unknown method error, got %v", err)
	}
}

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{
		{context.Canceled, "CANCELLED"},
		{sdk.ErrAuth, "AUTHENTICATION_REQUIRED"},
		{sdk.ErrCode, "API_ERROR"},
		{sdk.ErrInvalidBase, "INVALID_BASE"},
		{sdk.ErrRetryable, "RETRYABLE"},
		{sdk.ErrTransport, "TRANSPORT"},
		{errors.New("random"), "UNKNOWN"},
	}
	for _, tc := range cases {
		e := Error(tc.err)
		if e["code"] != tc.code {
			t.Errorf("Error(%v).code = %v, want %s", tc.err, e["code"], tc.code)
		}
	}
}

// 多层包装后错误码仍按优先级正确映射：取消优先于传输/重试。
func TestErrorCodePriorityThroughWrapping(t *testing.T) {
	wrapped := fmt.Errorf("%w: %w", context.Canceled, errors.New("transport boom"))
	if got := Error(wrapped)["code"]; got != "CANCELLED" {
		t.Fatalf("cancel should win, got %v", got)
	}
	retryable := fmt.Errorf("%w: %w: %w", sdk.ErrRetryable, sdk.ErrTransport, errors.New("reset"))
	if got := Error(retryable)["code"]; got != "RETRYABLE" {
		t.Fatalf("retryable should map, got %v", got)
	}
	if Error(retryable)["retryable"] != true {
		t.Fatal("retryable flag should be true")
	}
}
