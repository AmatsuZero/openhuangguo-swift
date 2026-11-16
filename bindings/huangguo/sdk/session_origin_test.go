package sdk

import (
	"context"
	"errors"
	"testing"
)

// token 属于 .com，但客户端 base 切到 .ai（两个域名都在信任名单里）。
// edge 必须拒绝把 .com 的 Bearer 发给 .ai，否则跨域泄漏凭据。
func TestEdgeRejectsMismatchedTokenOrigin(t *testing.T) {
	store := NewMemStore()
	_ = store.Save(Session{Origin: "https://huangguoai.com", Token: "tok-com", Generation: 7})
	fake := &fakeTransport{resp: &Response{StatusCode: 200, Body: []byte(`{"status":1,"data":{}}`)}}
	c, err := NewClient(Options{Transport: fake, Store: store, Base: "https://huangguoai.ai"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Me(context.Background())
	if err == nil || !errors.Is(err, ErrAuth) {
		t.Fatalf("expected auth error for cross-origin token, got %v", err)
	}
	if fake.gotAuth != "" {
		t.Fatalf("Bearer must not be sent cross-origin, got %q", fake.gotAuth)
	}
}

// 空 origin 但带 token 也无法验证绑定，必须拒绝。
func TestEdgeRejectsTokenWithEmptyOrigin(t *testing.T) {
	store := NewMemStore()
	_ = store.Save(Session{Token: "tok", Generation: 1})
	fake := &fakeTransport{resp: &Response{StatusCode: 200, Body: []byte(`{"status":1,"data":{}}`)}}
	c, _ := NewClient(Options{Transport: fake, Store: store, Base: "https://huangguoai.com"})
	if _, err := c.Me(context.Background()); err == nil {
		t.Fatal("expected rejection for token without origin binding")
	}
	if fake.gotAuth != "" {
		t.Fatalf("Bearer must not be sent, got %q", fake.gotAuth)
	}
}

// 恢复会话时即校验 origin：不匹配直接拒绝，不让宿主误以为恢复成功。
func TestRestoreSessionRejectsMismatch(t *testing.T) {
	c, _ := NewClient(Options{Transport: &fakeTransport{}, Store: NewMemStore(), Base: "https://huangguoai.ai"})
	if err := c.RestoreSession(Session{Origin: "https://huangguoai.com", Token: "t", Generation: 1}); err == nil {
		t.Fatal("expected restore rejection for cross-origin session")
	}
	// 匹配 origin 的会话应可恢复。
	if err := c.RestoreSession(Session{Origin: "https://huangguoai.ai", Token: "t", Generation: 1}); err != nil {
		t.Fatalf("matching origin restore failed: %v", err)
	}
}
