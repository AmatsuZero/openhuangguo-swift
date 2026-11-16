package sdk

import (
	"context"
	"crypto/aes"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// 封面 AES 常量（硬编码于 HGhub media.rs）。
const (
	imgKey = "f5d965df75336270"
	imgIV  = "97b60394abc2fbe1"
)

// imageKind 识别明文图片类型（照 media.rs:87-104）。
func imageKind(buf []byte) string {
	if len(buf) < 12 {
		return ""
	}
	if buf[0] == 0xFF && buf[1] == 0xD8 {
		return "image/jpeg"
	}
	if buf[0] == 0x89 && buf[1] == 0x50 && buf[2] == 0x4E && buf[3] == 0x47 {
		return "image/png"
	}
	if buf[0] == 'G' && buf[1] == 'I' && buf[2] == 'F' && buf[3] == '8' {
		return "image/gif"
	}
	if buf[0] == 'R' && buf[1] == 'I' && buf[2] == 'F' && buf[3] == 'F' &&
		buf[8] == 'W' && buf[9] == 'E' && buf[10] == 'B' && buf[11] == 'P' {
		return "image/webp"
	}
	return ""
}

// aesCBCDecrypt 手写 AES-128-CBC 块解密 + 与前一块 XOR（照 media.rs:56-75）。
func aesCBCDecrypt(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data)%16 != 0 {
		return nil, errors.New("封面密文长度非法")
	}
	block, err := aes.NewCipher([]byte(imgKey))
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	prev := []byte(imgIV)
	for i := 0; i < len(data); i += 16 {
		block.Decrypt(out[i:i+16], data[i:i+16])
		for j := 0; j < 16; j++ {
			out[i+j] ^= prev[j]
		}
		prev = data[i : i+16]
	}
	return out, nil
}

// stripPadding 条件去 PKCS7 padding；无效 padding 原样保留（照 media.rs:77-84）。
func stripPadding(buf []byte) []byte {
	if len(buf) == 0 {
		return buf
	}
	pad := int(buf[len(buf)-1])
	if pad > 0 && pad <= 16 && len(buf) >= pad {
		all := true
		for _, b := range buf[len(buf)-pad:] {
			if int(b) != pad {
				all = false
				break
			}
		}
		if all {
			return buf[:len(buf)-pad]
		}
	}
	return buf
}

// DecryptCover 明文图片直接透传；否则 AES-128-CBC 解密 + 条件去 padding。
func DecryptCover(data []byte) ([]byte, error) {
	if imageKind(data) != "" {
		return data, nil
	}
	dec, err := aesCBCDecrypt(data)
	if err != nil {
		return nil, err
	}
	return stripPadding(dec), nil
}

// pctEncode 对齐 Rust utf8_percent_encode(NON_ALPHANUMERIC)：仅字母数字不转义。
func pctEncode(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0F])
		}
	}
	return b.String()
}

// absolutize 解析相对 URI（照 media.rs:134-143）。
func absolutize(base, target string) string {
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	b, err := url.Parse(base)
	if err != nil {
		return target
	}
	r, err := url.Parse(target)
	if err != nil {
		return target
	}
	return b.ResolveReference(r).String()
}

// hlsAttributes 按逗号拆分属性，尊重引号内的逗号（照 media.rs:148-163）。
func hlsAttributes(line string) (string, []string, bool) {
	idx := strings.IndexByte(line, ':')
	if idx < 0 {
		return "", nil, false
	}
	tag, rest := line[:idx], line[idx+1:]
	attrs := make([]string, 0)
	quoted := false
	start := 0
	for i := 0; i < len(rest); i++ {
		ch := rest[i]
		if ch == '"' {
			quoted = !quoted
		}
		if ch == ',' && !quoted {
			attrs = append(attrs, rest[start:i])
			start = i + 1
		}
	}
	attrs = append(attrs, rest[start:])
	return tag, attrs, true
}

// mapURIAttributes 改写带引号的 URI="..." 属性（照 media.rs:165-185）。
func mapURIAttributes(line string, rewrite func(string) string) string {
	tag, attrs, ok := hlsAttributes(line)
	if !ok {
		return line
	}
	mapped := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		key, value, found := strings.Cut(attr, "=")
		if found && strings.TrimSpace(key) == "URI" {
			v := strings.TrimSpace(value)
			if len(v) >= 2 && strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`) {
				mapped = append(mapped, key+`="`+rewrite(v[1:len(v)-1])+`"`)
				continue
			}
		}
		mapped = append(mapped, attr)
	}
	return tag + ":" + strings.Join(mapped, ",")
}

// RewritePlaylist 把 HLS 清单的 KEY/MAP/分片/子清单 URI 改写为可逆的本地协议映射。
// 上游清单完整 URL 作为所有相对 URI 的解析基准；localScheme+host 拼本地前缀。
// 行为矩阵对齐 media.rs：BOM、CONTENT-STEERING 删除、DATERANGE 外部资源清理、URI 属性绝对化改写、分片/子清单改写。
func RewritePlaylist(upstreamURL, localScheme, host, text string) string {
	prefix := localScheme + "://" + host
	text = strings.TrimPrefix(text, "\ufeff")
	var out strings.Builder
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimRight(raw, " \t\r")
		if line == "" {
			out.WriteByte('\n')
			continue
		}
		if strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "#EXT-X-CONTENT-STEERING:") {
				continue
			}
			if strings.HasPrefix(line, "#EXT-X-DATERANGE:") {
				if tag, attrs, ok := hlsAttributes(line); ok {
					safe := make([]string, 0, len(attrs))
					for _, attr := range attrs {
						key := strings.TrimSpace(strings.SplitN(attr, "=", 2)[0])
						key = strings.ToUpper(key)
						if strings.Contains(key, "URI") || key == "X-ASSET-LIST" {
							continue
						}
						safe = append(safe, attr)
					}
					out.WriteString(tag + ":" + strings.Join(safe, ",") + "\n")
				}
				continue
			}
			if strings.Contains(line, `URI="`) {
				abs := mapURIAttributes(line, func(uri string) string { return absolutize(upstreamURL, uri) })
				out.WriteString(mapURIAttributes(abs, func(uri string) string {
					return prefix + "/seg?u=" + pctEncode(uri)
				}) + "\n")
			} else {
				out.WriteString(line + "\n")
			}
		} else {
			abs := absolutize(upstreamURL, line)
			out.WriteString(prefix + "/seg?u=" + pctEncode(abs) + "\n")
		}
	}
	return out.String()
}

// SegmentResult 经 core 的 Start/callback JSON 传出；Body 经 JSON 自动 base64。
type SegmentResult struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

// FetchSegment 用共享 transport 拉取分片/KEY，透传 Range 等请求头，返回 206 与响应头。
// 媒体请求不自动带鉴权头；独立大小上限（32MB）；支持 ctx 取消。
func (c *Client) FetchSegment(ctx context.Context, url string, reqHeader map[string]string) (SegmentResult, error) {
	headers := make(map[string]string, len(reqHeader))
	for k, v := range reqHeader {
		headers[k] = v
	}
	resp, err := c.transport.Do(ctx, &Request{
		Method:               "GET",
		URL:                  url,
		Headers:              headers,
		MaxResponseBytes:     32 << 20,
		FollowMediaRedirects: true, // 逐跳目标守卫 + 5 跳上限 + 禁降级
	})
	if err != nil {
		return SegmentResult{}, fmt.Errorf("%w: %v", ErrTransport, err)
	}
	return SegmentResult{Status: resp.StatusCode, Headers: resp.Headers, Body: resp.Body}, nil
}
