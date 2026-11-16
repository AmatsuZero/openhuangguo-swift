package sdk

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// parseProxyURL 解析并校验代理 URL：只支持 HTTP/HTTPS，排除 SOCKS 等，
// 凭据只用于代理认证；错误里的凭据必须脱敏。
func parseProxyURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("unsupported proxy scheme %q for %s", u.Scheme, redactProxyURL(u))
	}
	if u.Host == "" {
		return nil, fmt.Errorf("proxy URL missing host: %s", redactProxyURL(u))
	}
	return u, nil
}

// redactProxyURL 把代理 URL 里的 userinfo 换成 *****，避免日志/错误泄漏凭据。
func redactProxyURL(u *url.URL) string {
	if u == nil {
		return "<nil>"
	}
	scheme := u.Scheme
	if scheme != "" {
		scheme += "://"
	}
	host := u.Host
	if u.User != nil {
		host = "***:***@" + host
	}
	return scheme + host + u.Path
}

// noProxyBypass 判断 host 是否命中 NO_PROXY（命中则直连）。
// 遵循常见语义： "*" 全匹配；前导 "." 匹配子域；否则精确匹配 host[:port]。
func noProxyBypass(host, noProxy string) bool {
	if noProxy == "" {
		return false
	}
	for _, entry := range strings.Split(noProxy, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if entry == "*" {
			return true
		}
		rule := strings.ToLower(entry)
		// 去掉端口，host 与规则都按主机名比较。
		if h, _, err := netSplitHostPort(rule); err == nil {
			rule = h
		}
		h := strings.ToLower(host)
		if strings.HasPrefix(rule, ".") {
			if h == strings.TrimPrefix(rule, ".") || strings.HasSuffix(h, rule) {
				return true
			}
		} else if h == rule {
			return true
		}
	}
	return false
}

// netSplitHostPort 只拆 host:port， IPv6 用最后一个冒号；无端口返回原串。
func netSplitHostPort(s string) (string, string, error) {
	// 简单场景：不含 ':' 直接返回；含 ':' 用 strings.LastIndex 处理 host:port。
	if !strings.Contains(s, ":") {
		return s, "", nil
	}
	i := strings.LastIndex(s, ":")
	return s[:i], s[i+1:], nil
}

// envProxyURL 按优先级读取环境变量：HGHUB_PROXY > HTTPS_PROXY/HTTP_PROXY > ALL_PROXY。
// 环境变量回退遵循 NO_PROXY；显式代理不走这里（见 makeProxyFunc）。
func envProxyURL(req *http.Request) (*url.URL, error) {
	host := req.URL.Hostname()
	if noProxyBypass(host, os.Getenv("NO_PROXY")) || noProxyBypass(host, os.Getenv("no_proxy")) {
		return nil, nil
	}
	raw := os.Getenv("HGHUB_PROXY")
	if raw == "" {
		if strings.EqualFold(req.URL.Scheme, "https") {
			raw = os.Getenv("HTTPS_PROXY")
		} else {
			raw = os.Getenv("HTTP_PROXY")
		}
	}
	if raw == "" {
		raw = os.Getenv("ALL_PROXY")
	}
	if raw == "" {
		return nil, nil
	}
	return parseProxyURL(raw)
}

// makeProxyFunc 返回给 http.Transport.Proxy 的函数。
// cfg.Proxy: nil=环境回退；指向 ""=强制直连；指向 URL=显式代理（不受 NO_PROXY 影响）。
func makeProxyFunc(cfg *string) func(*http.Request) (*url.URL, error) {
	if cfg != nil {
		if *cfg == "" {
			return func(*http.Request) (*url.URL, error) { return nil, nil }
		}
		u, err := parseProxyURL(*cfg)
		if err != nil {
			// 解析失败时返回的函数把错误交给 net/http，避免构造期 panic。
			return func(*http.Request) (*url.URL, error) { return nil, err }
		}
		// 显式代理绕过 NO_PROXY。
		return func(*http.Request) (*url.URL, error) { return u, nil }
	}
	return envProxyURL
}
