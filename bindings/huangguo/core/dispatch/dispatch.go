// Package dispatch 把 SDK 方法按名字和 JSON 参数调用，供 gomobile 绑定共用。
package dispatch

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/sdk"
)

type method func(ctx context.Context, c *sdk.Client, args []json.RawMessage) (any, error)

var methods = map[string]method{
	"login": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var username, password string
		if err := args2(a, &username, &password); err != nil {
			return nil, err
		}
		return c.Login(ctx, username, password)
	},
	"register": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var username, password, email string
		if err := arg(a, 2, &email); err != nil {
			return nil, err
		}
		if err := args2(a, &username, &password); err != nil {
			return nil, err
		}
		return c.Register(ctx, username, password, email)
	},
	"logout": func(ctx context.Context, c *sdk.Client, _ []json.RawMessage) (any, error) {
		return nil, c.Logout(ctx)
	},
	"me": func(ctx context.Context, c *sdk.Client, _ []json.RawMessage) (any, error) {
		return c.Me(ctx)
	},
	"feed": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var page, size int
		if err := args2(a, &page, &size); err != nil {
			return nil, err
		}
		return c.Feed(ctx, page, size)
	},
	"category": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var cate string
		var page, size int
		if err := arg(a, 0, &cate); err != nil {
			return nil, err
		}
		if err := args2(a[1:], &page, &size); err != nil {
			return nil, err
		}
		return c.Category(ctx, cate, page, size)
	},
	"rank": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var kind string
		if err := arg(a, 0, &kind); err != nil {
			return nil, err
		}
		return c.Rank(ctx, kind)
	},
	"search": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var q string
		var page, size int
		if err := arg(a, 0, &q); err != nil {
			return nil, err
		}
		if err := args2(a[1:], &page, &size); err != nil {
			return nil, err
		}
		return c.Search(ctx, q, page, size)
	},
	"posts": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var page, size int
		if err := args2(a, &page, &size); err != nil {
			return nil, err
		}
		return c.Posts(ctx, page, size)
	},
	"post": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var id int64
		if err := arg(a, 0, &id); err != nil {
			return nil, err
		}
		return c.Post(ctx, id)
	},
	"video": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var id int64
		if err := arg(a, 0, &id); err != nil {
			return nil, err
		}
		return c.Video(ctx, id)
	},
	"play": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var id int64
		var ep int
		if err := arg(a, 0, &id); err != nil {
			return nil, err
		}
		if err := arg(a, 1, &ep); err != nil {
			return nil, err
		}
		return c.Play(ctx, id, ep)
	},
	"comments": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var id int64
		var ep, page, size int
		if err := arg(a, 0, &id); err != nil {
			return nil, err
		}
		if err := arg(a, 1, &ep); err != nil {
			return nil, err
		}
		if err := args2(a[2:], &page, &size); err != nil {
			return nil, err
		}
		return c.Comments(ctx, id, ep, page, size)
	},
	"commentSend": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var id int64
		var ep int
		var text string
		var replyID int64
		if err := arg(a, 0, &id); err != nil {
			return nil, err
		}
		if err := arg(a, 1, &ep); err != nil {
			return nil, err
		}
		if err := arg(a, 2, &text); err != nil {
			return nil, err
		}
		if err := arg(a, 3, &replyID); err != nil {
			return nil, err
		}
		return c.CommentSend(ctx, id, ep, text, replyID)
	},
	"like": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var id int64
		var ep int
		var on bool
		if err := arg(a, 0, &id); err != nil {
			return nil, err
		}
		if err := arg(a, 1, &ep); err != nil {
			return nil, err
		}
		if err := arg(a, 2, &on); err != nil {
			return nil, err
		}
		return c.Like(ctx, id, ep, on)
	},
	"collect": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var id int64
		var on bool
		if err := arg(a, 0, &id); err != nil {
			return nil, err
		}
		if err := arg(a, 1, &on); err != nil {
			return nil, err
		}
		return c.Collect(ctx, id, on)
	},
	"follow": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var uid int64
		var on bool
		if err := arg(a, 0, &uid); err != nil {
			return nil, err
		}
		if err := arg(a, 1, &on); err != nil {
			return nil, err
		}
		return c.Follow(ctx, uid, on)
	},
	"postLike": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var id int64
		var on bool
		if err := arg(a, 0, &id); err != nil {
			return nil, err
		}
		if err := arg(a, 1, &on); err != nil {
			return nil, err
		}
		return c.PostLike(ctx, id, on)
	},
	"postCollect": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var id int64
		var on bool
		if err := arg(a, 0, &id); err != nil {
			return nil, err
		}
		if err := arg(a, 1, &on); err != nil {
			return nil, err
		}
		return c.PostCollect(ctx, id, on)
	},
	"postComments": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var id int64
		var page, size int
		if err := arg(a, 0, &id); err != nil {
			return nil, err
		}
		if err := args2(a[1:], &page, &size); err != nil {
			return nil, err
		}
		return c.PostComments(ctx, id, page, size)
	},
	"setBase": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var base string
		if err := arg(a, 0, &base); err != nil {
			return nil, err
		}
		return nil, c.SetBase(base)
	},
	"decryptCover": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var enc string
		if err := arg(a, 0, &enc); err != nil {
			return nil, err
		}
		data, err := base64.StdEncoding.DecodeString(enc)
		if err != nil {
			return nil, err
		}
		out, err := sdk.DecryptCover(data)
		return out, err // []byte 经 envelope 的 JSON 自动 base64
	},
	"rewritePlaylist": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var upstream, scheme, host, text string
		if err := arg(a, 0, &upstream); err != nil {
			return nil, err
		}
		if err := arg(a, 1, &scheme); err != nil {
			return nil, err
		}
		if err := arg(a, 2, &host); err != nil {
			return nil, err
		}
		if err := arg(a, 3, &text); err != nil {
			return nil, err
		}
		return sdk.RewritePlaylist(upstream, scheme, host, text), nil
	},
	"fetchSegment": func(ctx context.Context, c *sdk.Client, a []json.RawMessage) (any, error) {
		var url string
		var headers map[string]string
		if err := arg(a, 0, &url); err != nil {
			return nil, err
		}
		if err := arg(a, 1, &headers); err != nil {
			return nil, err
		}
		return c.FetchSegment(ctx, url, headers)
	},
}

// Call 调用名为 name 的方法；raw 是参数数组的 JSON（可为空）。
func Call(ctx context.Context, c *sdk.Client, name, raw string) (any, error) {
	m, ok := methods[name]
	if !ok {
		return nil, fmt.Errorf("unknown method %q", name)
	}
	var a []json.RawMessage
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &a); err != nil {
			return nil, fmt.Errorf("arguments: %v", err)
		}
	}
	v, err := m(ctx, c, a)
	if err != nil && ctx.Err() != nil && !errors.Is(err, ctx.Err()) {
		err = fmt.Errorf("%w: %v", ctx.Err(), err)
	}
	return v, err
}

// arg 把第 i 个参数解码到 v；缺省或 null 时保留 v 的零值。
func arg(a []json.RawMessage, i int, v any) error {
	if i >= len(a) || string(a[i]) == "null" {
		return nil
	}
	if err := json.Unmarshal(a[i], v); err != nil {
		return fmt.Errorf("argument %d: %v", i+1, err)
	}
	return nil
}

func args2(a []json.RawMessage, v0, v1 any) error {
	if err := arg(a, 0, v0); err != nil {
		return err
	}
	return arg(a, 1, v1)
}

// ErrorCode 是绑定使用的错误码及其对应错误。
type ErrorCode struct {
	Err  error
	Code string
}

// ErrorCodes 把 SDK 哨兵错误映射成绑定错误码（取消排最前）。
var ErrorCodes = []ErrorCode{
	{context.Canceled, "CANCELLED"},
	{sdk.ErrAuth, "AUTHENTICATION_REQUIRED"},
	{sdk.ErrCode, "API_ERROR"},
	{sdk.ErrInvalidBase, "INVALID_BASE"},
	{sdk.ErrRetryable, "RETRYABLE"},
	{sdk.ErrTransport, "TRANSPORT"},
}

// Error 把 err 转成 {code, message, retryable, retryAfterMs}。
func Error(err error) map[string]any {
	code := "UNKNOWN"
	for _, e := range ErrorCodes {
		if errors.Is(err, e.Err) {
			code = e.Code
			break
		}
	}
	return map[string]any{
		"code":         code,
		"message":      err.Error(),
		"retryable":    errors.Is(err, sdk.ErrRetryable),
		"retryAfterMs": 0,
	}
}
