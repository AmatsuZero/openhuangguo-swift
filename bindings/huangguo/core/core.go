// Package core 是 Android / iOS 绑定的 Go 入口，由 gomobile bind 编译成
// OpenHuangguoCore.xcframework / AAR。
//
// gomobile 只导出字符串、数字和接口等简单类型，所以这里只有一个按名字调用的
// 入口：Client.Start(method, argsJSON, callback) 在 goroutine 里调用 SDK，
// 结束时把 {"result": …} 或 {"error": {code, message, retryable, retryAfterMs}}
// 交给 callback.Done。复杂数据（含 SegmentResult）都走这个 JSON 通道。
package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/core/dispatch"
	"github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/sdk"
)

// Logger 接收诊断日志。
type Logger interface {
	Log(line string)
}

// Callback 接收一次调用的结果，在 Go 的线程里调用，每次恰好一次。
type Callback interface {
	Done(resultJSON string)
}

// Client 是一个 SDK 客户端，可多线程使用。
type Client struct {
	c     *sdk.Client
	store sdk.TokenStore // 持有 store 引用，供 token 桥接访问
}

// NewClient 只收 configJSON + logger（gomobile 可绑定）；内部建 SDK Client。
// configJSON: {base, proxy, credTrustOrigins}；logger 可为 nil。
func NewClient(configJSON string, logger Logger) (*Client, error) {
	var cfg struct {
		Base      string          `json:"base"`
		Proxy     *string         `json:"proxy"`
		CredTrust map[string]bool `json:"credTrustOrigins"`
	}
	if configJSON != "" {
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, jsonError(err)
		}
	}
	store := sdk.NewMemStore() // 默认内存；原生层用 SetTokenJSON 回写/恢复
	var slog sdk.Logger
	if logger != nil {
		slog = logf{logger}
	}
	c, err := sdk.NewClient(sdk.Options{
		Transport: sdk.NewHTTPTransport(sdk.TransportConfig{Proxy: cfg.Proxy}),
		Store:     store,
		Base:      cfg.Base,
		CredTrust: cfg.CredTrust,
		Logger:    slog,
	})
	if err != nil {
		return nil, jsonError(err)
	}
	return &Client{c: c, store: store}, nil
}

// SetTokenJSON 仅用于启动恢复：把原生持久化快照灌入内存 store。
// 不得在会话中调用覆盖 SDK 正在管理的 generation。
func (cl *Client) SetTokenJSON(s string) error {
	var sess sdk.Session
	if err := json.Unmarshal([]byte(s), &sess); err != nil {
		return err
	}
	return cl.c.RestoreSession(sess)
}

// GetTokenJSON 输出 {origin,token,generation,user}，供原生层写 Keychain/Keystore。
func (cl *Client) GetTokenJSON() (string, error) {
	sess := cl.c.ExportSession()
	b, err := json.Marshal(sess)
	return string(b), err
}

// Start 在后台开始调用 method；argsJSON 是参数数组的 JSON。
func (cl *Client) Start(method, argsJSON string, callback Callback) *Call {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		v, err := dispatch.Call(ctx, cl.c, method, argsJSON)
		callback.Done(envelope(v, err))
	}()
	return &Call{cancel: cancel}
}

// Call 是一次进行中的调用。
type Call struct {
	cancel context.CancelFunc
}

// Cancel 取消调用；已结束的调用不受影响。
func (c *Call) Cancel() { c.cancel() }

func envelope(v any, err error) string {
	var out any
	if err != nil {
		out = map[string]any{"error": dispatch.Error(err)}
	} else {
		out = map[string]any{"result": v}
	}
	b, err := json.Marshal(out)
	if err != nil {
		b, _ = json.Marshal(map[string]any{"error": dispatch.Error(err)})
	}
	return string(b)
}

// jsonError 返回信息是错误对象 JSON 的错误，供另一端解析出错误码。
func jsonError(err error) error {
	b, _ := json.Marshal(dispatch.Error(err))
	return errors.New(string(b))
}

type logf struct{ l Logger }

func (l logf) Printf(format string, args ...any) { l.l.Log(fmt.Sprintf(format, args...)) }
