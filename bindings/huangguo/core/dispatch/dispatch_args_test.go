package dispatch

import (
	"context"
	"strings"
	"testing"

	"github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/sdk"
)

func testClient(t *testing.T) *sdk.Client {
	t.Helper()
	c, err := sdk.NewClient(sdk.Options{
		Transport: &argStubTransport{},
		Store:     sdk.NewMemStore(),
		Base:      "https://huangguoai.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

type argStubTransport struct{}

func (argStubTransport) Do(ctx context.Context, r *sdk.Request) (*sdk.Response, error) {
	return &sdk.Response{StatusCode: 200, Body: []byte(`{"status":1,"data":null}`)}, nil
}

// 缺省参数不得让 category/search/comments/postComments 切片越界崩溃；
// 应返回参数错误（配齐所需参数的方法才继续执行）。
func TestMissingArgsNoPanic(t *testing.T) {
	cases := []string{"category", "search", "comments", "postComments"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s panicked with missing args: %v", name, r)
				}
			}()
			_, err := Call(context.Background(), testClient(t), name, `[]`)
			if err == nil {
				t.Fatalf("%s: expected argument error for empty args", name)
			}
		})
	}
}

// 参数不足（只有第一个）也不能越界。
func TestPartialArgsNoPanic(t *testing.T) {
	cases := []struct {
		name string
		args string
	}{
		{"category", `["drama"]`},
		{"search", `["q"]`},
		{"comments", `[1,2]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s panicked: %v", tc.name, r)
				}
			}()
			_, err := Call(context.Background(), testClient(t), tc.name, tc.args)
			if err == nil || !strings.Contains(err.Error(), "argument") {
				t.Fatalf("%s: expected argument error, got %v", tc.name, err)
			}
		})
	}
}

// 完全缺失的必填参数必须报错（不能静默取零值导致误操作，如 collect([42]) 变成 on=false 发 DELETE）。
func TestMissingRequiredArgErrors(t *testing.T) {
	cases := []struct {
		name string
		args string
	}{
		{"collect", `[42]`},            // 缺 on
		{"like", `[42,2]`},             // 缺 on
		{"postLike", `[42]`},           // 缺 on
		{"play", `[42]`},               // 缺 ep
		{"commentSend", `[42,2,"hi"]`}, // replyID 是可选，不应在此报错
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Call(context.Background(), testClient(t), tc.name, tc.args)
			if tc.name == "commentSend" {
				// replyID 缺省应走零值，不报错。
				if err != nil {
					t.Fatalf("optional replyID should be allowed missing: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "argument") {
				t.Fatalf("%s: expected argument error, got %v", tc.name, err)
			}
		})
	}
}

// 显式 null 保留零值：兼容 TS 侧 `replyId ?? null` 的缺省约定。
func TestExplicitNullKeepsZeroValue(t *testing.T) {
	// commentSend 的 replyID 显式 null 应可用（可选参数，零值语义）。
	_, err := Call(context.Background(), testClient(t), "commentSend", `[42,2,"hi",null]`)
	if err != nil {
		t.Fatalf("explicit null optional arg should work: %v", err)
	}
}
