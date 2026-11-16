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

// SetBase 切换线路：规范化后更新 base、generation++、清空 token。
// 返回 store.Save 结果——持久化失败必须上报（不谎称成功），但 base 与 generation
// 在内存已推进，不回滚；宿主须把失败视为硬错误重试/上报（跨重启退出不保证生效）。
func (c *Client) SetBase(raw string) error {
	b, err := normalizeBase(raw)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.base = b
	c.mu.Unlock()
	s, _ := c.store.Load()
	s.Generation++
	s.Origin = ""
	s.Token = ""
	s.User = nil
	return c.store.Save(s)
}
