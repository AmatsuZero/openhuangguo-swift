package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/gen-huangguo/schema"
)

func TestGenerateHuangguoTypes(t *testing.T) {
	// 测试在 gen-huangguo 目录运行：huangguo 根 = ..（内部拼 <root>/sdk/model）。
	s, err := schema.LoadModelOnly("..")
	if err != nil {
		t.Fatal(err)
	}
	swift := swiftTypes(s)
	kotlin := kotlinTypes(s)
	for _, want := range []string{"struct VideoCard", "struct Detail", "struct PlayInfo"} {
		if !strings.Contains(swift, want) {
			t.Fatalf("missing %s in Swift output", want)
		}
	}
	for _, want := range []string{"data class VideoCard", "data class Detail", "data class PlayInfo"} {
		if !strings.Contains(kotlin, want) {
			t.Fatalf("missing %s in Kotlin output", want)
		}
	}
}

// TestGeneratedUpToDate 保证提交的 Types.swift / Types.kt 与 Go model 源码一致。
func TestGeneratedUpToDate(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	out, err := generate(root)
	if err != nil {
		t.Fatal(err)
	}
	for path, want := range out {
		got, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Errorf("%s is stale: run go run . --repo-root <repo>", path)
		}
	}
}

func TestCamel(t *testing.T) {
	for in, want := range map[string]string{
		"episode_count": "episodeCount", "video_url": "videoUrl", "pageSize": "pageSize", "key": "key",
	} {
		if got := camel(in); got != want {
			t.Errorf("camel(%q) = %q, want %q", in, got, want)
		}
	}
}
