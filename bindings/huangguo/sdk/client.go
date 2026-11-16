package sdk

import (
	"errors"
	"sync"
)

type Logger interface {
	Printf(format string, args ...any)
}

type Options struct {
	Transport     Transport
	Store         TokenStore
	Base          string
	Proxy         *string // nil=走环境；指向 ""=强制直连；指向 URL=显式代理
	CredTrust     map[string]bool
	Logger        Logger
	RetryAttempts int
}

type Client struct {
	mu            sync.Mutex // 保护 base / snapshot 读取
	transport     Transport
	store         TokenStore
	base          string
	credTrust     map[string]bool
	logger        Logger
	retryAttempts int
}

// NewClient 校验依赖并定默认值，不发请求。
// base 规范化（normalizeBase 在 Task 3 实现，本 Task 先内联最小校验）。
// 默认 credTrust = {https://huangguoai.com, https://huangguoai.ai}。
func NewClient(o Options) (*Client, error) {
	if o.Transport == nil {
		return nil, errors.New("Transport required")
	}
	if o.Store == nil {
		return nil, errors.New("Store required")
	}
	baseRaw := o.Base
	if baseRaw == "" {
		baseRaw = "https://huangguoai.com"
	}
	base, err := normalizeBase(baseRaw)
	if err != nil {
		return nil, err
	}
	trust := o.CredTrust
	if trust == nil {
		trust = map[string]bool{
			"https://huangguoai.com": true,
			"https://huangguoai.ai":  true,
		}
	}
	return &Client{
		transport:     o.Transport,
		store:         o.Store,
		base:          base,
		credTrust:     trust,
		logger:        o.Logger,
		retryAttempts: o.RetryAttempts,
	}, nil
}

// snapshot 原子读 {origin, token, generation} + 当前 base，供 edge 全程使用同一份。
func (c *Client) snapshot() (Session, string) {
	c.mu.Lock()
	base := c.base
	c.mu.Unlock()
	s, _ := c.store.Load()
	return s, base
}
