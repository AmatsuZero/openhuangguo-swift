package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
)

// 哨兵错误供 dispatch 映射错误码、供重试分类。
var (
	ErrAuth        = errors.New("authentication required")
	ErrCode        = errors.New("api error")
	ErrTransport   = errors.New("transport error")
	ErrInvalidBase = errors.New("invalid base")
	ErrRetryable   = errors.New("retryable error")
)

// CodeError 是业务 status != 1 的错误。
type CodeError struct {
	Code    int
	Message string
	Name    string
	Kind    string
}

func (e *CodeError) Error() string { return e.Message }

// Is 让 errors.Is(err, ErrCode) 成立。
func (e *CodeError) Is(target error) bool { return target == ErrCode }

// edge 是唯一线上出口，按 api.go 给出的明文路径直连。全程使用同一份 (token, base) 快照。
func (c *Client) edge(ctx context.Context, method, path string, body any, needCred bool) ([]byte, error) {
	s, base := c.snapshot()

	// 凭据检查用快照 base：needCred（发密码）或带 Bearer 时，目标必须在信任名单内；
	// 镜像不收密码/Bearer。base 全程用快照值，不再读 c.base。
	if (needCred || s.Token != "") && !c.allowedCredentialTarget(base) {
		return nil, fmt.Errorf("%w: credentials not allowed for %s", ErrInvalidBase, base)
	}

	// Origin 绑定：带 token 时必须与其签发的 origin 一致，否则拒绝发送 Bearer。
	// 防止从 Keychain 恢复的 .com token 被发给同为受信任的 .ai base。
	if s.Token != "" && s.Origin != base {
		return nil, fmt.Errorf("%w: token origin %q does not match base %q", ErrAuth, s.Origin, base)
	}

	var raw []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		raw = b
	}

	headers := map[string]string{
		"Accept":  "application/json",
		"Referer": base + "/",
	}
	if raw != nil {
		headers["Content-Type"] = "application/json"
	}
	if s.Token != "" {
		headers["Authorization"] = "Bearer " + s.Token
	}

	resp, err := c.transport.Do(ctx, &Request{
		Method:  method,
		URL:     base + path,
		Headers: headers,
		Body:    raw,
	})
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			// 取消优先：保留原始取消原因，便于 errors.Is(err, context.Canceled)。
			return nil, fmt.Errorf("%w: %w", ctxErr, err)
		}
		// 临时网络故障按可重试分类，否则作为永久传输错误。
		if isTransientTransportError(err) {
			return nil, fmt.Errorf("%w: %w: %w", ErrRetryable, ErrTransport, err)
		}
		return nil, fmt.Errorf("%w: %w", ErrTransport, err)
	}

	// 先判 HTTP 状态，再解析响应。
	if err := classifyHTTPStatus(resp.StatusCode); err != nil {
		return nil, err
	}
	data, err := parseEdgeResponse(resp.Body)
	if err != nil {
		return nil, err
	}

	// 提交结果前校验 generation 未变：请求期间若切域名/登出，丢弃本次响应。
	// 从内存快照读（受同一把锁保护），用当前 base 再核对一次 origin 绑定。
	cur, curBase := c.snapshot()
	if cur.Generation != s.Generation || curBase != base {
		return nil, errors.New("session generation changed during request")
	}
	if s.Token != "" && s.Origin != curBase {
		return nil, fmt.Errorf("%w: token origin %q no longer matches base %q", ErrAuth, s.Origin, curBase)
	}
	return data, nil
}

// isTransientTransportError 判断网络层错误是否值得重试。
// 不包含 context.Canceled（用户取消，绝不重试）；包含超时、连接重置/拒绝、DNS 失败。
func isTransientTransportError(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() {
			return true
		}
		// 非超时的临时错误（如某些代理/连接层的 Temporary()=true）也应重试。
		// 用接口断言避免直接调用已废弃的 net.Error.Temporary()。
		if te, ok := ne.(interface{ Temporary() bool }); ok && te.Temporary() {
			return true
		}
	}
	// syscall 连接类临时错误。
	for _, code := range []syscall.Errno{
		syscall.ECONNRESET, syscall.ECONNREFUSED, syscall.ETIMEDOUT,
		syscall.EHOSTUNREACH, syscall.ENETUNREACH, syscall.EPIPE,
	} {
		if errors.Is(err, code) {
			return true
		}
	}
	// DNS 查询失败。
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr)
}

// classifyHTTPStatus：401/403→ErrAuth；408/425/429/5xx→可重试；其余非 2xx→错误。
func classifyHTTPStatus(status int) error {
	switch {
	case status >= 200 && status < 300:
		return nil
	case status == 401 || status == 403:
		return fmt.Errorf("%w: HTTP %d", ErrAuth, status)
	case status == 408 || status == 425 || status == 429 || status >= 500:
		return fmt.Errorf("%w: HTTP %d", ErrRetryable, status)
	default:
		return fmt.Errorf("HTTP %d", status)
	}
}

// parseEdgeResponse 精确对齐 HGhub net.rs:578-613：
// 空串 / null → 成功返回 null；status==1 → 取 data（可为 {} 或 null）；
// 裸 {}（无 status）→ 失败；status!=1 → 取非空 msg，否则"接口返回失败"。
func parseEdgeResponse(raw []byte) ([]byte, error) {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" {
		return []byte("null"), nil
	}
	var v struct {
		Status *int            `json:"status"`
		Data   json.RawMessage `json:"data"`
		Msg    string          `json:"msg"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	if v.Status != nil && *v.Status == 1 {
		if len(v.Data) == 0 {
			return []byte("null"), nil
		}
		return v.Data, nil
	}
	if v.Msg != "" {
		return nil, &CodeError{Message: v.Msg}
	}
	return nil, &CodeError{Message: "接口返回失败"}
}
