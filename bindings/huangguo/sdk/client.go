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

// Client 在一个互斥锁下持有权威的 {base, session} 内存状态；TokenStore 是持久化后端，
// 所有读写都在同一临界区内写穿（write-through），避免"新 base + 旧 token"或
// 旧登录覆盖登出的竞态（见 commitLogin / SetBase / Logout）。
type Client struct {
	mu            sync.Mutex
	transport     Transport
	store         TokenStore
	base          string
	session       Session // 进程内权威会话，镜像 store 但读取无需 I/O
	credTrust     map[string]bool
	logger        Logger
	retryAttempts int
}

// NewClient 校验依赖并定默认值，不发请求。
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
	sess, _ := o.Store.Load()
	return &Client{
		transport:     o.Transport,
		store:         o.Store,
		base:          base,
		session:       sess,
		credTrust:     trust,
		logger:        o.Logger,
		retryAttempts: o.RetryAttempts,
	}, nil
}

// snapshot 在一个临界区内原子返回 (session, base)，供 edge 全程使用同一份。
func (c *Client) snapshot() (Session, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.session, c.base
}

// ExportSession 供 gomobile token 桥接读取当前会话快照。
func (c *Client) ExportSession() Session {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.session
}

// commitLogin 在 generation 与 base 都未变时才提交登录/注册结果。
// 比较与写入在同一临界区，杜绝"检查通过后被并发 Logout/SetBase 抢先、旧 token 又被写回"。
// 持久化失败时内存状态已推进，按设计上报错误、不回滚。
func (c *Client) commitLogin(gen int64, base string, s Session) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session.Generation != gen || c.base != base {
		return errors.New("会话已变化，请重新登录")
	}
	c.session = s
	return c.store.Save(s)
}

// commitSession 在 generation 未变时对当前会话做局部更新（如 Me 回写 user）。
func (c *Client) commitSession(gen int64, fn func(*Session)) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session.Generation != gen {
		return errors.New("会话已变化")
	}
	fn(&c.session)
	return c.store.Save(c.session)
}

// clearSession 原子地递增 generation 并清空会话（登出），与 SetBase 共用同一把锁。
func (c *Client) clearSession() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session.Generation++
	c.session.Origin = ""
	c.session.Token = ""
	c.session.User = nil
	return c.store.Save(c.session)
}
