package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/sdk/model"
)

// AuthResult 是登录/注册结果，token 已写入 store，仅回传 user。
type AuthResult struct {
	User json.RawMessage `json:"user"`
}

// call 是 edge 的重试封装：仅无 body 的 GET 与取播放地址的 POST 可重试；
// 可重试错误（408/425/429/5xx）按退避重试，其他错误立即返回。
func (c *Client) call(ctx context.Context, method, path string, body any, needCred, retryable bool) ([]byte, error) {
	attempts := c.retryAttempts
	if attempts <= 0 {
		attempts = 1
	}
	if !retryable {
		attempts = 1
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		data, err := c.edge(ctx, method, path, body, needCred)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if !errors.Is(err, ErrRetryable) || i+1 >= attempts {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(i+1) * 100 * time.Millisecond):
		}
	}
	return nil, lastErr
}

func decodeJSON[T any](data []byte) (*T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// ---- 账号 ----

func (c *Client) Login(ctx context.Context, username, password string) (*AuthResult, error) {
	before, base := c.snapshot()
	data, err := c.call(ctx, "POST", "/api/account/loginByPassword",
		map[string]string{"username": username, "password": password}, true, false)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Token string          `json:"token"`
		User  json.RawMessage `json:"user"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.Token == "" {
		return nil, errors.New("登录返回缺少 token")
	}
	if err := c.commitLogin(before.Generation, base,
		Session{Origin: base, Token: resp.Token, Generation: before.Generation, User: resp.User}); err != nil {
		return nil, err
	}
	return &AuthResult{User: resp.User}, nil
}

func (c *Client) Register(ctx context.Context, username, password, email string) (*AuthResult, error) {
	before, base := c.snapshot()
	body := map[string]string{
		"username":         username,
		"password":         password,
		"confirm_password": password,
	}
	if email != "" {
		body["email"] = email
	}
	data, err := c.call(ctx, "POST", "/api/account/registerByPassword", body, true, false)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Token string          `json:"token"`
		User  json.RawMessage `json:"user"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.Token == "" {
		return nil, errors.New("注册返回缺少 token")
	}
	if err := c.commitLogin(before.Generation, base,
		Session{Origin: base, Token: resp.Token, Generation: before.Generation, User: resp.User}); err != nil {
		return nil, err
	}
	return &AuthResult{User: resp.User}, nil
}

// Logout 忽略服务端错误（对齐 HGhub），但本地清理失败必须上报。
// 在同一临界区内 generation++ 并清空会话，杜绝并发请求读到旧 token。
func (c *Client) Logout(ctx context.Context) error {
	_, _ = c.edge(ctx, "POST", "/api/account/logout", nil, false)
	return c.clearSession()
}

func (c *Client) Me(ctx context.Context) (*model.User, error) {
	before, _ := c.snapshot()
	data, err := c.call(ctx, "GET", "/api/account/me", nil, false, true)
	if err != nil {
		return nil, err
	}
	u, err := decodeJSON[model.User](data)
	if err != nil {
		return nil, err
	}
	// 在 generation 未变时才回写 user；比较与写在 commitSession 的同一临界区内。
	if err := c.commitSession(before.Generation, func(s *Session) {
		if s.Token == "" {
			return
		}
		s.User, _ = json.Marshal(u)
	}); err != nil {
		return nil, errors.New("读取账号期间线路已变化")
	}
	return u, nil
}

// ---- 列表 / 详情 ----

func (c *Client) Feed(ctx context.Context, page, size int) (*model.VideoCardPage, error) {
	data, err := c.call(ctx, "GET", fmt.Sprintf("/api/videos?page=%d&size=%d", page, size), nil, false, true)
	if err != nil {
		return nil, err
	}
	return decodeJSON[model.VideoCardPage](data)
}

func (c *Client) Category(ctx context.Context, cate string, page, size int) (*model.VideoCardPage, error) {
	path := fmt.Sprintf("/api/videos/category/%s?page=%d&size=%d", url.PathEscape(cate), page, size)
	data, err := c.call(ctx, "GET", path, nil, false, true)
	if err != nil {
		return nil, err
	}
	return decodeJSON[model.VideoCardPage](data)
}

func (c *Client) Rank(ctx context.Context, kind string) (*model.RankPage, error) {
	if kind != "recommend" && kind != "potential" {
		kind = "hot"
	}
	data, err := c.call(ctx, "GET", "/api/ranks/"+kind, nil, false, true)
	if err != nil {
		return nil, err
	}
	return decodeJSON[model.RankPage](data)
}

func (c *Client) Search(ctx context.Context, q string, page, size int) (*model.SearchResult, error) {
	path := fmt.Sprintf("/api/search?q=%s&page=%d&size=%d", url.QueryEscape(q), page, size)
	data, err := c.call(ctx, "GET", path, nil, false, true)
	if err != nil {
		return nil, err
	}
	return decodeJSON[model.SearchResult](data)
}

func (c *Client) Posts(ctx context.Context, page, size int) (*model.PostPage, error) {
	data, err := c.call(ctx, "GET", fmt.Sprintf("/api/posts?page=%d&size=%d", page, size), nil, false, true)
	if err != nil {
		return nil, err
	}
	return decodeJSON[model.PostPage](data)
}

func (c *Client) Post(ctx context.Context, id int64) (*model.Post, error) {
	data, err := c.call(ctx, "GET", fmt.Sprintf("/api/posts/%d", id), nil, false, true)
	if err != nil {
		return nil, err
	}
	return decodeJSON[model.Post](data)
}

func (c *Client) Video(ctx context.Context, id int64) (*model.Detail, error) {
	data, err := c.call(ctx, "GET", fmt.Sprintf("/api/videos/%d", id), nil, false, true)
	if err != nil {
		return nil, err
	}
	return decodeJSON[model.Detail](data)
}

// Play 优先新接口 /episodes/{ep}/playback，失败或无地址回退旧 GET /play。
func (c *Client) Play(ctx context.Context, id int64, ep int) (*model.PlayInfo, error) {
	primaryData, primaryErr := c.call(ctx, "POST",
		fmt.Sprintf("/api/videos/%d/episodes/%d/playback", id, ep), struct{}{}, false, true)
	var primaryInfo *model.PlayInfo
	if primaryErr == nil {
		if info, err := decodeJSON[model.PlayInfo](primaryData); err == nil {
			normalizeStreams(info)
			primaryInfo = info
			if info.VideoURL != "" {
				return info, nil
			}
			// 无可用地址，继续回退旧接口。
		} else {
			primaryErr = err
		}
	}

	legacyData, legacyErr := c.call(ctx, "GET",
		fmt.Sprintf("/api/videos/%d/play?ep=%d&page_texts=0", id, ep), nil, false, true)
	if legacyErr == nil {
		info, err := decodeJSON[model.PlayInfo](legacyData)
		if err == nil {
			normalizeStreams(info)
			if info.VideoURL != "" || primaryErr != nil {
				return info, nil
			}
		}
	}
	if primaryErr != nil {
		return nil, primaryErr
	}
	if primaryInfo != nil {
		return primaryInfo, nil
	}
	return nil, errors.New("play: no usable result")
}

// normalizeStreams 把 streams.main/hevc 回填到 video_url / video_url_hevc（原字段空时）。
func normalizeStreams(p *model.PlayInfo) {
	if p.Streams == nil {
		return
	}
	if p.VideoURL == "" && p.Streams.Main != "" {
		p.VideoURL = p.Streams.Main
	}
	if p.VideoURLHevc == "" && p.Streams.Hevc != "" {
		p.VideoURLHevc = p.Streams.Hevc
	}
}

// ---- 评论 / 互动 ----

func (c *Client) Comments(ctx context.Context, id int64, ep, page, size int) (*model.CommentPage, error) {
	path := fmt.Sprintf("/api/videos/%d/comments?ep=%d&page=%d&size=%d", id, ep, page, size)
	data, err := c.call(ctx, "GET", path, nil, false, true)
	if err != nil {
		return nil, err
	}
	return decodeJSON[model.CommentPage](data)
}

func (c *Client) CommentSend(ctx context.Context, id int64, ep int, text string, replyID int64) (json.RawMessage, error) {
	body := map[string]any{"text": text}
	path := fmt.Sprintf("/api/videos/%d/comments?ep=%d", id, ep)
	if replyID > 0 {
		body["reply_id"] = replyID
		path = fmt.Sprintf("/api/videos/%d/comments", id)
	}
	data, err := c.call(ctx, "POST", path, body, true, false)
	return json.RawMessage(data), err
}

func (c *Client) Like(ctx context.Context, id int64, ep int, on bool) (json.RawMessage, error) {
	return c.interaction(ctx, on, fmt.Sprintf("/api/videos/%d/like?ep=%d", id, ep), nil)
}

func (c *Client) Collect(ctx context.Context, id int64, on bool) (json.RawMessage, error) {
	return c.interaction(ctx, on, fmt.Sprintf("/api/videos/%d/collect", id), nil)
}

func (c *Client) Follow(ctx context.Context, uid int64, on bool) (json.RawMessage, error) {
	return c.interaction(ctx, on, fmt.Sprintf("/api/users/%d/follow", uid), nil)
}

func (c *Client) PostLike(ctx context.Context, id int64, on bool) (json.RawMessage, error) {
	return c.interaction(ctx, on, fmt.Sprintf("/api/posts/%d/like", id), nil)
}

func (c *Client) PostCollect(ctx context.Context, id int64, on bool) (json.RawMessage, error) {
	return c.interaction(ctx, on, fmt.Sprintf("/api/posts/%d/collect", id), nil)
}

func (c *Client) interaction(ctx context.Context, on bool, path string, body any) (json.RawMessage, error) {
	method := "DELETE"
	if on {
		method = "POST"
	}
	data, err := c.call(ctx, method, path, body, true, false)
	return json.RawMessage(data), err
}

func (c *Client) PostComments(ctx context.Context, id int64, page, size int) (*model.CommentPage, error) {
	path := fmt.Sprintf("/api/posts/%d/comments?page=%d&size=%d", id, page, size)
	data, err := c.call(ctx, "GET", path, nil, false, true)
	if err != nil {
		return nil, err
	}
	return decodeJSON[model.CommentPage](data)
}
