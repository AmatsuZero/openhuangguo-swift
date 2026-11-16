package sdk

import (
	"fmt"
	"net"
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
		// 不能 %w 原始解析错误：net/url 把整条原始 URL（含 user:pass）放进错误消息。
		// 用脱敏后的 userinfo 拼一条通用错误。
		redacted := redactRawProxyURL(raw)
		return nil, fmt.Errorf("invalid proxy URL %s", redacted)
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

// redactRawProxyURL 在 URL 解析失败时，按字符串规则抹掉 scheme:// 之后、
// 最后一个 @ 之前的 userinfo（该段可能畸形、无法用 url.Parse 解析）。
func redactRawProxyURL(raw string) string {
	s := raw
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if at := strings.LastIndex(s, "@"); at >= 0 {
		s = "***:***@" + s[at+1:]
	}
	return s
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

// noProxyBypass 判断 host[:port] 是否命中 NO_PROXY（命中则直连）。
// 支持：* 全匹配、CIDR（IP 或 IP/netmask）、前导点子域、精确 host、host:port。
func noProxyBypass(hostPort, noProxy string) bool {
	if noProxy == "" {
		return false
	}
	host, port := splitHostPort(hostPort)
	host = strings.ToLower(host)
	for _, entry := range strings.Split(noProxy, ",") {
		rule := strings.ToLower(strings.TrimSpace(entry))
		if rule == "" {
			continue
		}
		if rule == "*" {
			return true
		}
		// CIDR：IP 规则（含 /）按网段匹配。
		if strings.Contains(rule, "/") {
			if _, cidr, err := net.ParseCIDR(rule); err == nil {
				if ip := net.ParseIP(host); ip != nil && cidr.Contains(ip) {
					return true
				}
			}
			continue
		}
		// host:port 规则：端口必须一致，且 host 部分精确。
		if ruleHost, rulePort, ok := splitHostPortRule(rule); ok && rulePort != "" {
			if port == rulePort && host == ruleHost {
				return true
			}
			continue
		}
		if strings.HasPrefix(rule, ".") {
			base := strings.TrimPrefix(rule, ".")
			if host == base || strings.HasSuffix(host, rule) {
				return true
			}
		} else if host == rule {
			return true
		}
	}
	return false
}

// splitHostPort 拆 "host" 或 "host:port"（IPv6 用方括号）。
func splitHostPort(s string) (string, string) {
	if strings.HasPrefix(s, "[") {
		if i := strings.Index(s, "]"); i >= 0 {
			h := s[1:i]
			if i+1 < len(s) && s[i+1] == ':' {
				return h, s[i+2:]
			}
			return h, ""
		}
	}
	if i := strings.LastIndex(s, ":"); i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

// splitHostPortRule 判断规则是否为 host:port 形式（不含 CIDR /，不含前导点）。
func splitHostPortRule(rule string) (string, string, bool) {
	if strings.Contains(rule, "/") || strings.HasPrefix(rule, ".") {
		return "", "", false
	}
	h, p := splitHostPort(rule)
	return h, p, p != ""
}

// envProxyURL 按优先级读取环境变量：HGHUB_PROXY > HTTPS_PROXY/HTTP_PROXY > ALL_PROXY。
// 环境变量回退遵循 NO_PROXY；显式代理不走这里（见 makeProxyFunc）。
func envProxyURL(req *http.Request) (*url.URL, error) {
	hostPort := req.URL.Host // 含端口，供 NO_PROXY 的 host:port 规则匹配
	if noProxyBypass(hostPort, os.Getenv("NO_PROXY")) || noProxyBypass(hostPort, os.Getenv("no_proxy")) {
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
