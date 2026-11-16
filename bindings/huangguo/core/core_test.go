package core

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/sdk"
)

type stubTransport struct{}

func (stubTransport) Do(ctx context.Context, r *sdk.Request) (*sdk.Response, error) {
	return &sdk.Response{StatusCode: 200, Body: []byte(`{"status":1,"data":{"items":[],"pagination":{"page":1,"pages":0,"size":20,"total":0}}}`)}, nil
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	store := sdk.NewMemStore()
	sc, err := sdk.NewClient(sdk.Options{Transport: stubTransport{}, Store: store, Base: "https://huangguoai.com"})
	if err != nil {
		t.Fatal(err)
	}
	return &Client{c: sc, store: store}
}

type testCallback struct{ ch chan string }

func (cb *testCallback) Done(s string) { cb.ch <- s }

func TestStartEnvelope(t *testing.T) {
	cl := newTestClient(t)
	cb := &testCallback{ch: make(chan string, 1)}
	call := cl.Start("feed", "[1,20]", cb)
	defer call.Cancel()

	select {
	case got := <-cb.ch:
		if !strings.Contains(got, `"result"`) {
			t.Fatalf("expected result envelope, got %s", got)
		}
		var env map[string]any
		if err := json.Unmarshal([]byte(got), &env); err != nil {
			t.Fatal(err)
		}
		result, ok := env["result"].(map[string]any)
		if !ok || result["pagination"] == nil {
			t.Fatalf("bad result: %s", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("callback timeout")
	}
}

func TestStartErrorEnvelope(t *testing.T) {
	cl := newTestClient(t)
	cb := &testCallback{ch: make(chan string, 1)}
	call := cl.Start("nonexistent", "[]", cb)
	defer call.Cancel()
	select {
	case got := <-cb.ch:
		if !strings.Contains(got, `"error"`) || !strings.Contains(got, `"code"`) {
			t.Fatalf("expected error envelope, got %s", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("callback timeout")
	}
}

func TestTokenBridgeRoundTrip(t *testing.T) {
	cl := newTestClient(t)
	in := `{"origin":"https://huangguoai.com","token":"t","generation":1}`
	if err := cl.SetTokenJSON(in); err != nil {
		t.Fatal(err)
	}
	got, err := cl.GetTokenJSON()
	if err != nil {
		t.Fatal(err)
	}
	var sess struct {
		Origin     string `json:"origin"`
		Token      string `json:"token"`
		Generation int64  `json:"generation"`
	}
	if err := json.Unmarshal([]byte(got), &sess); err != nil {
		t.Fatal(err)
	}
	if sess.Origin != "https://huangguoai.com" || sess.Token != "t" || sess.Generation != 1 {
		t.Fatalf("bad token roundtrip: %s", got)
	}
}

func TestNewClientConfigJSON(t *testing.T) {
	cl, err := NewClient(`{"base":"https://huangguoai.ai","credTrustOrigins":{"https://huangguoai.ai":true}}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cl == nil || cl.c == nil || cl.store == nil {
		t.Fatal("client not initialized")
	}
	if _, err := NewClient(`not json`, nil); err == nil {
		t.Fatal("expected error for invalid configJSON")
	}
}
