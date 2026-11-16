package sdk

import (
	"context"
	"testing"
	"time"
)

// 登录响应拿到后、提交前若 generation 已变（并发 Logout/SetBase），
// 必须拒绝提交，不能把已登出的 token 写回去。
func TestCommitLoginRejectsStaleGeneration(t *testing.T) {
	store := NewMemStore()
	c, _ := NewClient(Options{Transport: &fakeTransport{}, Store: store, Base: "https://huangguoai.com"})
	before, beforeBase := c.snapshot()

	// 模拟并发切域名：generation++ 且清空 token。
	if err := c.SetBase("https://huangguoai.ai"); err != nil {
		t.Fatal(err)
	}

	stale := Session{Origin: beforeBase, Token: "old-token", Generation: before.Generation}
	if err := c.commitLogin(before.Generation, beforeBase, stale); err == nil {
		t.Fatal("stale login must be rejected after generation changed")
	}
	s, base := c.snapshot()
	if base != "https://huangguoai.ai" || s.Token != "" {
		t.Fatalf("stale login resurrected state: base=%s sess=%+v", base, s)
	}
}

// commitLogin 在 generation/base 都没变时应正常提交。
func TestCommitLoginSucceedsWhenStable(t *testing.T) {
	c, _ := NewClient(Options{Transport: &fakeTransport{}, Store: NewMemStore(), Base: "https://huangguoai.com"})
	before, beforeBase := c.snapshot()
	next := Session{Origin: beforeBase, Token: "tok", Generation: before.Generation}
	if err := c.commitLogin(before.Generation, beforeBase, next); err != nil {
		t.Fatal(err)
	}
	s, _ := c.snapshot()
	if s.Token != "tok" {
		t.Fatalf("commit failed: %+v", s)
	}
}

// blockingStore 在 Save 里暂停，用来验证 SetBase 持有锁期间
// snapshot() 不会观察到"新 base + 旧 token"的中间态。
type blockingStore struct {
	*MemStore
	entered chan struct{}
	release chan struct{}
}

func (b *blockingStore) Save(s Session) error {
	select {
	case <-b.entered:
	default:
		close(b.entered)
	}
	<-b.release
	return b.MemStore.Save(s)
}

func TestSetBaseAtomicSnapshot(t *testing.T) {
	bs := &blockingStore{MemStore: NewMemStore(), entered: make(chan struct{}), release: make(chan struct{})}
	_ = bs.MemStore.Save(Session{Origin: "https://huangguoai.com", Token: "old", Generation: 1})
	c, _ := NewClient(Options{Transport: &fakeTransport{}, Store: bs, Base: "https://huangguoai.com"})

	setDone := make(chan error, 1)
	go func() { setDone <- c.SetBase("https://huangguoai.ai") }()
	<-bs.entered // SetBase 正在 Save，应持有会话锁

	snapDone := make(chan struct{})
	go func() {
		c.snapshot() // 必须阻塞到 SetBase 完成，不能看到中间态
		close(snapDone)
	}()
	select {
	case <-snapDone:
		t.Fatal("snapshot observed state mid-SetBase (race window)")
	case <-time.After(60 * time.Millisecond):
		// 预期：被会话锁挡住。
	}

	close(bs.release)
	if err := <-setDone; err != nil {
		t.Fatal(err)
	}
	<-snapDone
	s, base := c.snapshot()
	if base != "https://huangguoai.ai" || s.Token != "" || s.Generation != 2 {
		t.Fatalf("bad final state: base=%s sess=%+v", base, s)
	}
}

// 并发 Logout 后旧 token 不得残留。
func TestLogoutClearsTokenAtomically(t *testing.T) {
	c, _ := NewClient(Options{Transport: &fakeTransport{}, Store: NewMemStore(), Base: "https://huangguoai.com"})
	_ = c.commitLogin(0, "https://huangguoai.com", Session{Origin: "https://huangguoai.com", Token: "tok", Generation: 0})
	if err := c.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
	s, _ := c.snapshot()
	if s.Token != "" || s.Generation != 1 {
		t.Fatalf("logout left stale token: %+v", s)
	}
}
