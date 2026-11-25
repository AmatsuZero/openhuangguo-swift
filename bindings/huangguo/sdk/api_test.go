package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// scriptedTransport 按顺序返回预置响应，并记录每次请求，用于 Play 回退测试。
type scriptedTransport struct {
	resps    []*Response
	errs     []error
	requests []*Request
}

func (s *scriptedTransport) Do(ctx context.Context, r *Request) (*Response, error) {
	s.requests = append(s.requests, r)
	i := len(s.requests) - 1
	if i < len(s.errs) && s.errs[i] != nil {
		return nil, s.errs[i]
	}
	if i < len(s.resps) && s.resps[i] != nil {
		return s.resps[i], nil
	}
	return &Response{StatusCode: 200, Body: []byte(`{"status":1,"data":null}`)}, nil
}

func edgeOK(data string) *Response {
	return &Response{StatusCode: 200, Body: []byte(fmt.Sprintf(`{"status":1,"data":%s}`, data))}
}

func TestFeedParsesFixture(t *testing.T) {
	fake := &fakeTransport{resp: &Response{StatusCode: 200, Body: mustRead(t, "testdata/feed.json")}}
	c, _ := NewClient(Options{Transport: fake, Store: NewMemStore(), Base: "https://huangguoai.com"})
	page, err := c.Feed(context.Background(), 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) == 0 || page.Pagination.Total != 100 {
		t.Fatalf("fields lost: %+v", page)
	}
	if page.Items[0].Author == nil || page.Items[0].Author.ID != 7 {
		t.Fatalf("nested author lost: %+v", page.Items[0])
	}
}

func TestAllEndpointsPaths(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name   string
		invoke func(c *Client) error
		want   string // "METHOD path"
	}{
		{"Login", func(c *Client) error { _, e := c.Login(ctx, "u", "p"); return e }, "POST /api/account/loginByPassword"},
		{"Register", func(c *Client) error { _, e := c.Register(ctx, "u", "p", ""); return e }, "POST /api/account/registerByPassword"},
		{"Logout", func(c *Client) error { return c.Logout(ctx) }, "POST /api/account/logout"},
		{"Me", func(c *Client) error { _, e := c.Me(ctx); return e }, "GET /api/account/me"},
		{"Feed", func(c *Client) error { _, e := c.Feed(ctx, 1, 20); return e }, "GET /api/videos?page=1&size=20"},
		{"Category", func(c *Client) error { _, e := c.Category(ctx, "drama", 2, 10); return e }, "GET /api/videos/category/drama?page=2&size=10"},
		{"Rank", func(c *Client) error { _, e := c.Rank(ctx, "bad"); return e }, "GET /api/ranks/hot"},
		{"Search", func(c *Client) error { _, e := c.Search(ctx, "a b&c", 1, 20); return e }, "GET /api/search?q=a+b%26c&page=1&size=20"},
		{"Posts", func(c *Client) error { _, e := c.Posts(ctx, 1, 20); return e }, "GET /api/posts?page=1&size=20"},
		{"Post", func(c *Client) error { _, e := c.Post(ctx, 5); return e }, "GET /api/posts/5"},
		{"Video", func(c *Client) error { _, e := c.Video(ctx, 5); return e }, "GET /api/videos/5"},
		{"Comments", func(c *Client) error { _, e := c.Comments(ctx, 5, 2, 1, 20); return e }, "GET /api/videos/5/comments?ep=2&page=1&size=20"},
		{"CommentSend", func(c *Client) error { _, e := c.CommentSend(ctx, 5, 2, "hi", 0); return e }, "POST /api/videos/5/comments?ep=2"},
		{"CommentSendReply", func(c *Client) error { _, e := c.CommentSend(ctx, 5, 2, "hi", 9); return e }, "POST /api/videos/5/comments"},
		{"LikeOn", func(c *Client) error { _, e := c.Like(ctx, 5, 2, true); return e }, "POST /api/videos/5/like?ep=2"},
		{"LikeOff", func(c *Client) error { _, e := c.Like(ctx, 5, 2, false); return e }, "DELETE /api/videos/5/like?ep=2"},
		{"CollectOn", func(c *Client) error { _, e := c.Collect(ctx, 5, true); return e }, "POST /api/videos/5/collect"},
		{"FollowOff", func(c *Client) error { _, e := c.Follow(ctx, 7, false); return e }, "DELETE /api/users/7/follow"},
		{"PostLike", func(c *Client) error { _, e := c.PostLike(ctx, 3, true); return e }, "POST /api/posts/3/like"},
		{"PostCollect", func(c *Client) error { _, e := c.PostCollect(ctx, 3, false); return e }, "DELETE /api/posts/3/collect"},
		{"PostComments", func(c *Client) error { _, e := c.PostComments(ctx, 3, 1, 20); return e }, "GET /api/posts/3/comments?page=1&size=20"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := &scriptedTransport{}
			store := NewMemStore()
			_ = store.Save(Session{Origin: "https://huangguoai.com", Token: "tok", Generation: 1})
			c, _ := NewClient(Options{Transport: tr, Store: store, Base: "https://huangguoai.com"})
			_ = tc.invoke(c)
			if len(tr.requests) == 0 {
				t.Fatal("no request recorded")
			}
			// 明文直连：method 与 path 直接落在请求上。
			got := tr.requests[0].Method + " " + strings.TrimPrefix(tr.requests[0].URL, "https://huangguoai.com")
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPlayFallbackAndStreams(t *testing.T) {
	ctx := context.Background()

	// 1. 新接口 streams.main 回填 VideoURL，不发旧接口。
	tr1 := &scriptedTransport{resps: []*Response{edgeOK(`{"id":1,"ep_num":2,"duration":0,"title":"t","video_url":"","streams":{"main":"https://cdn/main.m3u8"}}`)}}
	c1, _ := NewClient(Options{Transport: tr1, Store: NewMemStore(), Base: "https://huangguoai.com"})
	p1, err := c1.Play(ctx, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if p1.VideoURL != "https://cdn/main.m3u8" || len(tr1.requests) != 1 {
		t.Fatalf("streams backfill failed: %+v, requests=%d", p1, len(tr1.requests))
	}

	// 2. 新接口无地址 → 回退旧 GET。
	tr2 := &scriptedTransport{resps: []*Response{
		edgeOK(`{"id":1,"ep_num":2,"duration":0,"title":"t","video_url":""}`),
		edgeOK(`{"id":1,"ep_num":2,"duration":0,"title":"t","video_url":"https://cdn/legacy.mp4"}`),
	}}
	c2, _ := NewClient(Options{Transport: tr2, Store: NewMemStore(), Base: "https://huangguoai.com"})
	p2, err := c2.Play(ctx, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if p2.VideoURL != "https://cdn/legacy.mp4" || len(tr2.requests) != 2 {
		t.Fatalf("legacy fallback failed: %+v, requests=%d", p2, len(tr2.requests))
	}
	if !strings.Contains(tr2.requests[1].URL, "/api/videos/1/play?ep=2&page_texts=0") {
		t.Fatalf("legacy URL = %s", tr2.requests[1].URL)
	}

	// 3. 新接口失败（不可重试）→ 用旧接口结果。
	tr3 := &scriptedTransport{
		resps: []*Response{nil, edgeOK(`{"id":1,"ep_num":2,"duration":0,"title":"t","video_url":"https://cdn/legacy2.mp4"}`)},
		errs:  []error{errors.New("boom"), nil},
	}
	c3, _ := NewClient(Options{Transport: tr3, Store: NewMemStore(), Base: "https://huangguoai.com"})
	p3, err := c3.Play(ctx, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if p3.VideoURL != "https://cdn/legacy2.mp4" || len(tr3.requests) != 2 {
		t.Fatalf("primary failure fallback: %+v, requests=%d", p3, len(tr3.requests))
	}
}

func TestPlayAcceptsCurrentStringIdentityAndRetainsCDNHeaders(t *testing.T) {
	tr := &scriptedTransport{resps: []*Response{edgeOK(`{"id":"9007199254740993","ep":2,"streams":{"main":"https://cdn.example/master.m3u8"},"manifest_method":"GET","cdn":{"m3u8Enabled":true,"authKey":"x-cdn-auth","authValue":"signed","app":"player"}}`)}}
	c, _ := NewClient(Options{Transport: tr, Store: NewMemStore()})
	p, err := c.Play(context.Background(), 9007199254740993, 2)
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(p)
	var data map[string]any
	json.Unmarshal(encoded, &data)
	headers, _ := data["media_headers"].(map[string]any)
	if p.ID != 9007199254740993 || p.EpNum != 2 || p.VideoURL != "https://cdn.example/master.m3u8" || len(tr.requests) != 1 || headers["x-cdn-auth"] != "signed" || headers["x-cdn-app"] != "player" || data["manifest_method"] != "GET" {
		t.Fatalf("lost current playback contract: %s requests=%d", encoded, len(tr.requests))
	}
}

func TestLoginStoresSession(t *testing.T) {
	tr := &scriptedTransport{resps: []*Response{edgeOK(`{"token":"newtok","user":{"id":1,"username":"u"}}`)}}
	store := NewMemStore()
	c, _ := NewClient(Options{Transport: tr, Store: store, Base: "https://huangguoai.com"})
	res, err := c.Login(context.Background(), "u", "p")
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(res.User) {
		t.Fatalf("bad user json: %s", res.User)
	}
	s, _ := store.Load()
	if s.Token != "newtok" || s.Origin != "https://huangguoai.com" || s.Generation != 0 {
		t.Fatalf("session not stored: %+v", s)
	}
}
