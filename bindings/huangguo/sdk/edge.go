package sdk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

// edgeEnvelope 是 HGhub 所有业务接口收敛到的请求信封。
type edgeEnvelope struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Body        string `json:"body"`
	ContentType string `json:"content_type"`
}

// edge 是唯一线上出口。全程使用同一份 (token, base) 快照。
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

	enc, ctype := "", ""
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		enc = base64.StdEncoding.EncodeToString(b)
		ctype = "application/json"
	}
	payload, err := json.Marshal(edgeEnvelope{Method: method, Path: path, Body: enc, ContentType: ctype})
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json", // 外层信封的媒体类型
		"Referer":      base + "/",
	}
	if s.Token != "" {
		headers["Authorization"] = "Bearer " + s.Token
	}

	resp, err := c.transport.Do(ctx, &Request{
		Method:     "POST",
		URL:        base + "/api/edge/request",
		Headers:    headers,
		Body:       payload,
		NoRedirect: true, // edge 首版禁自动重定向
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransport, err)
	}

	// 先判 HTTP 状态，再解析信封。
	if err := classifyHTTPStatus(resp.StatusCode); err != nil {
		return nil, err
	}
	data, err := parseEdgeResponse(resp.Body)
	if err != nil {
		return nil, err
	}

	// 提交结果前校验 generation 未变：请求期间若切域名/登出，丢弃本次响应。
	cur, _ := c.store.Load()
	if cur.Generation != s.Generation {
		return nil, errors.New("session generation changed during request")
	}
	return data, nil
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
