package sdk

import "encoding/json"

// Session 是 token 快照。Origin 绑定 + Generation 防止跨 base 误用旧 token。
type Session struct {
	Origin     string          `json:"origin"`
	Token      string          `json:"token"`
	Generation int64           `json:"generation"`
	User       json.RawMessage `json:"user"`
}

// TokenStore 是 Go SDK 内部 token 存储接口（多返回值，不受 gomobile 限制）。
// MemStore / FileStore 实现在 Task 3。
type TokenStore interface {
	Load() (Session, error)
	Save(Session) error
	Clear() error
}
