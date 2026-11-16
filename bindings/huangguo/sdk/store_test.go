package sdk

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionOriginBindingAndGeneration(t *testing.T) {
	m := NewMemStore()
	want := Session{Origin: "https://huangguoai.com", Token: "tok", Generation: 1, User: json.RawMessage(`{}`)}
	if err := m.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := m.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Token != "tok" || got.Origin != want.Origin || got.Generation != 1 {
		t.Fatalf("bad roundtrip: %+v", got)
	}
	if err := m.Clear(); err != nil {
		t.Fatal(err)
	}
	got, _ = m.Load()
	if got.Token != "" || got.Generation != 0 {
		t.Fatalf("not cleared: %+v", got)
	}
}

func TestFileStoreAtomicAndPerms(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	fs := NewFileStore(path)
	if err := fs.Save(Session{Origin: "https://huangguoai.com", Token: "tok", Generation: 2}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("perm = %o, want 600", info.Mode().Perm())
	}
	got, err := fs.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Token != "tok" || got.Generation != 2 {
		t.Fatalf("bad: %+v", got)
	}
	if err := fs.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("file not removed: %v", err)
	}
}

type failingStore struct {
	saveErr  error
	clearErr error
}

func (f *failingStore) Load() (Session, error) { return Session{}, nil }
func (f *failingStore) Save(Session) error     { return f.saveErr }
func (f *failingStore) Clear() error           { return f.clearErr }

func TestNormalizeBaseAndCredentialTrust(t *testing.T) {
	for _, raw := range []string{
		"http://huangguoai.com",        // 非 HTTPS
		"https://u:p@huangguoai.com",   // userinfo
		"https://huangguoai.com/x",     // path
		"https://huangguoai.com?x=1",   // query
		"https://huangguoai.com/#frag", // fragment
		"https://127.0.0.1",            // 内网
	} {
		if _, err := normalizeBase(raw); err == nil {
			t.Fatalf("should reject %s", raw)
		}
	}
	for _, raw := range []string{"https://huangguoai.com", "https://huangguoai.ai", "https://mirror.example:8443"} {
		if _, err := normalizeBase(raw); err != nil {
			t.Fatalf("should accept %s: %v", raw, err)
		}
	}
	c, _ := NewClient(Options{Transport: &fakeTransport{}, Store: NewMemStore(), Base: "https://huangguoai.ai"})
	if !c.allowedCredentialTarget("https://huangguoai.ai") {
		t.Fatal("huangguoai.ai should be trusted")
	}
	if c.allowedCredentialTarget("https://mirror.example:8443") {
		t.Fatal("mirror should not be trusted by default")
	}
}

func TestSetBaseClearsTokenAndBumpsGeneration(t *testing.T) {
	m := NewMemStore()
	_ = m.Save(Session{Origin: "https://huangguoai.com", Token: "tok", Generation: 1})
	c, err := NewClient(Options{Transport: &fakeTransport{}, Store: m, Base: "https://huangguoai.com"})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetBase("https://huangguoai.ai"); err != nil {
		t.Fatal(err)
	}
	s, _ := m.Load()
	if s.Token != "" || s.Origin != "" || s.Generation != 2 {
		t.Fatalf("token not cleared / generation not bumped: %+v", s)
	}
	if _, base := c.snapshot(); base != "https://huangguoai.ai" {
		t.Fatalf("base = %q", base)
	}

	// 持久化失败必须上报，不谎称成功（Clear/Save 失败一律硬错误）。
	c2, _ := NewClient(Options{Transport: &fakeTransport{}, Store: &failingStore{saveErr: errors.New("disk full")}, Base: "https://huangguoai.com"})
	if err := c2.SetBase("https://huangguoai.ai"); err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("expected save error, got %v", err)
	}
	// base 与 generation 仍在内存推进，失败仅上报（不回滚）。
	if _, base := c2.snapshot(); base != "https://huangguoai.ai" {
		t.Fatalf("base should still advance on save failure, got %q", base)
	}
}
