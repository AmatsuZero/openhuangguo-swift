package sdk

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// normalizeBase 把 base 规范化为 HTTPS 纯 origin（scheme://host[:port]）：
// 无 userinfo / path / query / fragment，禁内网回环。返回可直接比较的完整 origin。
func normalizeBase(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid base: %w", err)
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return "", errors.New("base must use https")
	}
	if u.User != nil {
		return "", errors.New("base must not contain userinfo")
	}
	if u.Path != "" && u.Path != "/" {
		return "", errors.New("base must be a bare origin (no path)")
	}
	if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return "", errors.New("base must not contain query or fragment")
	}
	host := u.Hostname()
	if host == "" {
		return "", errors.New("base must have a host")
	}
	if isInternalHost(host) {
		return "", errors.New("base must not be a private/loopback host")
	}
	return "https://" + u.Host, nil
}

// allowedCredentialTarget 用规范化后的完整 origin（含端口）比较凭据信任名单。
// 显式接收 origin 参数，避免读取可变的 c.base；调用方传 edge 快照里的 base。
func (c *Client) allowedCredentialTarget(origin string) bool {
	return c.credTrust[origin]
}

// RestoreSession 供宿主把持久化快照灌回 store；恢复时校验 origin 绑定，
// 拒绝把另一 base 的 token 恢复到当前客户端（防御跨域误用旧 Bearer）。
func (c *Client) RestoreSession(s Session) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if s.Token != "" && s.Origin != c.base {
		return fmt.Errorf("%w: restored token origin %q does not match base %q", ErrAuth, s.Origin, c.base)
	}
	c.session = s
	return c.store.Save(s)
}

// SetBase 切换线路：规范化后在同一临界区内更新 base、generation++、清空 token。
// 内存先推进、再写穿 store；持久化失败上报（不谎称成功）、不回滚；宿主须视为硬错误。
// 同一把锁保证任何 snapshot 都看不到"新 base + 旧 token"的中间态。
func (c *Client) SetBase(raw string) error {
	b, err := normalizeBase(raw)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.base = b
	c.session.Generation++
	c.session.Origin = ""
	c.session.Token = ""
	c.session.User = nil
	return c.store.Save(c.session)
}
