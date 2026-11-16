package sdk

import (
	"bytes"
	"context"
	"crypto/aes"
	"strings"
	"testing"
)

// encryptCover 用与 DecryptCover 相同的 key/IV 生成 AES-CBC 密文（PKCS7 padding），仅供测试。
func encryptCover(t *testing.T, plaintext []byte) []byte {
	t.Helper()
	pad := 16 - len(plaintext)%16
	padded := append(append([]byte{}, plaintext...), bytes.Repeat([]byte{byte(pad)}, pad)...)
	block, err := aes.NewCipher([]byte(imgKey))
	if err != nil {
		t.Fatal(err)
	}
	out := make([]byte, len(padded))
	prev := []byte(imgIV)
	for i := 0; i < len(padded); i += 16 {
		for j := 0; j < 16; j++ {
			padded[i+j] ^= prev[j]
		}
		block.Encrypt(out[i:i+16], padded[i:i+16])
		prev = out[i : i+16]
	}
	return out
}

func TestDecryptCover(t *testing.T) {
	// 明文 JPEG 透传。
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0x10, 'J', 'F', 'I', 'F', 0, 1}
	got, err := DecryptCover(jpeg)
	if err != nil || !bytes.Equal(got, jpeg) {
		t.Fatalf("plaintext passthrough: %v, %v", got, err)
	}

	// 密文解密 + 去 padding。
	plain := []byte("hello cover")
	got, err = DecryptCover(encryptCover(t, plain))
	if err != nil || !bytes.Equal(got, plain) {
		t.Fatalf("decrypt: %q, %v", got, err)
	}

	// 无效 padding 原样保留。
	badPad := []byte("0123456789abcde\x05") // 16 字节，末字节 5 但前 4 字节非 5
	got, err = DecryptCover(encryptCover(t, badPad))
	if err != nil || !bytes.Equal(got, badPad) {
		t.Fatalf("invalid padding should be preserved: %q, %v", got, err)
	}

	// 非法长度（非 16 倍数且非明文图片）报错。
	if _, err := DecryptCover([]byte("short")); err == nil {
		t.Fatal("expected error for invalid ciphertext length")
	}
}

func TestRewritePlaylist(t *testing.T) {
	in := "\ufeff#EXTM3U\n" +
		"#EXT-X-CONTENT-STEERING:URI=\"x\"\n" +
		"#EXT-X-DATERANGE:ID=\"a\",X-ASSET-LIST=\"b\",DURATION=1\n" +
		"#EXT-X-KEY:METHOD=AES-128,URI=\"key.bin\"\n" +
		"seg-1.ts\n" +
		"sub/list.m3u8\n"
	got := RewritePlaylist("https://cdn.example/master.m3u8", "hg", "localhost", in)

	if !strings.HasPrefix(got, "#EXTM3U") {
		t.Fatalf("BOM not stripped: %q", got)
	}
	if strings.Contains(got, "CONTENT-STEERING") {
		t.Fatalf("CONTENT-STEERING not removed: %q", got)
	}
	if !strings.Contains(got, "#EXT-X-DATERANGE:ID=\"a\",DURATION=1") {
		t.Fatalf("DATERANGE not filtered: %q", got)
	}
	if strings.Contains(got, "X-ASSET-LIST") {
		t.Fatalf("X-ASSET-LIST not removed: %q", got)
	}
	// KEY URI / 分片 / 子清单都应改写为 hg://localhost/seg?u=<pct>，且 URI 已绝对化。
	// NON_ALPHANUMERIC 会转义 '.' 和 '-'（与 Rust percent_encoding 一致）。
	for _, want := range []string{
		"hg://localhost/seg?u=https%3A%2F%2Fcdn%2Eexample%2Fkey%2Ebin",
		"hg://localhost/seg?u=https%3A%2F%2Fcdn%2Eexample%2Fseg%2D1%2Ets",
		"hg://localhost/seg?u=https%3A%2F%2Fcdn%2Eexample%2Fsub%2Flist%2Em3u8",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

// capturingTransport 记录请求头并返回预置响应（不依赖回环 server）。
type capturingTransport struct {
	gotHeaders map[string]string
	resp       *Response
}

func (c *capturingTransport) Do(ctx context.Context, r *Request) (*Response, error) {
	c.gotHeaders = r.Headers
	return c.resp, nil
}

func TestFetchSegmentRange(t *testing.T) {
	body := []byte{0x00, 0x01, 0xff, 0xfe, 0x80}
	ct := &capturingTransport{resp: &Response{
		StatusCode: 206,
		Headers: map[string]string{
			"Content-Range":  "bytes 0-4/100",
			"Accept-Ranges":  "bytes",
			"Content-Length": "5",
		},
		Body: body,
	}}
	c, _ := NewClient(Options{Transport: ct, Store: NewMemStore(), Base: "https://huangguoai.com"})
	res, err := c.FetchSegment(context.Background(), "https://cdn.example/seg.ts", map[string]string{"Range": "bytes=0-4"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != 206 || res.Headers["Content-Range"] != "bytes 0-4/100" {
		t.Fatalf("bad segment result: %+v", res)
	}
	if !bytes.Equal(res.Body, body) {
		t.Fatalf("body roundtrip failed: %v", res.Body)
	}
	if ct.gotHeaders["Range"] != "bytes=0-4" {
		t.Fatalf("Range not forwarded: %+v", ct.gotHeaders)
	}
	if _, ok := ct.gotHeaders["Authorization"]; ok {
		t.Fatalf("media request must not auto-carry Authorization: %+v", ct.gotHeaders)
	}
}
