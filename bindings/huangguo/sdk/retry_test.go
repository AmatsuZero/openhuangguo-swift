package sdk

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"
)

// retryTransport 前 failN 次返回指定错误，之后成功。
type retryTransport struct {
	failN int
	fails int
	err   error
	calls int
}

func (r *retryTransport) Do(ctx context.Context, req *Request) (*Response, error) {
	r.calls++
	if r.fails < r.failN {
		r.fails++
		return nil, r.err
	}
	return &Response{StatusCode: 200, Body: []byte(`{"status":1,"data":{}}`)}, nil
}

func transientErr() error {
	return &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNRESET}
}

// 默认重试次数 >1：临时网络错误应被重试并最终成功。
func TestDefaultRetryAttempts(t *testing.T) {
	tr := &retryTransport{failN: 2, err: transientErr()}
	c, err := NewClient(Options{Transport: tr, Store: NewMemStore(), Base: "https://huangguoai.com"})
	if err != nil {
		t.Fatal(err)
	}
	if c.retryAttempts <= 1 {
		t.Fatalf("default retry attempts must be >1, got %d", c.retryAttempts)
	}
	_, err = c.call(context.Background(), "GET", "/api/videos", nil, false, true)
	if err != nil {
		t.Fatalf("expected success after retries: %v", err)
	}
	if tr.calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", tr.calls)
	}
}

// 写操作（retryable=false）即使临时错误也不得重试。
func TestWriteNotRetried(t *testing.T) {
	tr := &retryTransport{failN: 5, err: transientErr()}
	c, _ := NewClient(Options{Transport: tr, Store: NewMemStore(), Base: "https://huangguoai.com"})
	_, err := c.call(context.Background(), "POST", "/api/account/loginByPassword", nil, true, false)
	if err == nil {
		t.Fatal("expected error")
	}
	if tr.calls != 1 {
		t.Fatalf("write must not be retried, got %d calls", tr.calls)
	}
}

// 错误链保留：transport 返回 context.Canceled，errors.Is 必须能识别。
func TestTransportErrorChainPreserved(t *testing.T) {
	tr := &retryTransport{failN: 1, err: context.Canceled}
	c, _ := NewClient(Options{Transport: tr, Store: NewMemStore(), Base: "https://huangguoai.com"})
	_, err := c.edge(context.Background(), "GET", "/api/videos", nil, false)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation cause lost: %v", err)
	}
	if !errors.Is(err, ErrTransport) {
		t.Fatalf("transport sentinel lost: %v", err)
	}
}

// 退避带抖动：不同 attempt 的延迟不一致且随时间增长。
func TestRetryBackoffHasJitter(t *testing.T) {
	// retryDelay 带 ±50% 抖动：attempt1 base=100ms → [50ms,150ms]；attempt2 base=200ms → [100ms,300ms]。
	for i := 0; i < 20; i++ {
		d1 := retryDelay(1)
		if d1 < 50*time.Millisecond || d1 > 150*time.Millisecond {
			t.Fatalf("attempt 1 delay out of range: %v", d1)
		}
		d2 := retryDelay(2)
		if d2 < 100*time.Millisecond || d2 > 300*time.Millisecond {
			t.Fatalf("attempt 2 delay out of range: %v", d2)
		}
	}
}

// tempNetError：非超时但临时的 net.Error。
type tempNetError struct{}

func (tempNetError) Error() string   { return "temporary network failure" }
func (tempNetError) Timeout() bool   { return false }
func (tempNetError) Temporary() bool { return true }

func TestTransientNonTimeoutNetError(t *testing.T) {
	if !isTransientTransportError(tempNetError{}) {
		t.Fatal("non-timeout temporary net.Error should be retryable")
	}
	// 取消绝不重试。
	if isTransientTransportError(context.Canceled) {
		t.Fatal("context.Canceled must not be transient")
	}
}
