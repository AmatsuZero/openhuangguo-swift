// Command gen-huangguo 根据 bindings/huangguo/sdk/model 的源码生成
// iOS 的 Types.swift 与 Android 的 Types.kt（黄果无请求 DTO，只生成响应模型）。
//
// 用法：go run . --repo-root <仓库根>
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/AmatsuZero/openhuangguo-swift/bindings/huangguo/gen-huangguo/schema"
)

const (
	kotlinFile = "bindings/huangguo/android/src/main/kotlin/dev/openhuangguo/Types.kt"
	swiftFile  = "bindings/huangguo/ios/Sources/OpenHuangguo/Types.swift"
)

const header = "由 go run ./bindings/huangguo/gen-huangguo 生成，请勿手工修改。"

func main() {
	repoRoot := flag.String("repo-root", ".", "仓库根目录")
	flag.Parse()
	out, err := generate(*repoRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gen-huangguo:", err)
		os.Exit(1)
	}
	for path, content := range out {
		full := filepath.Join(*repoRoot, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "gen-huangguo:", err)
			os.Exit(1)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "gen-huangguo:", err)
			os.Exit(1)
		}
	}
}

// generate 返回仓库根 repoRoot 对应的输出，键是相对 repoRoot 的路径。
func generate(repoRoot string) (map[string]string, error) {
	s, err := schema.LoadModelOnly(filepath.Join(repoRoot, "bindings", "huangguo"))
	if err != nil {
		return nil, err
	}
	return map[string]string{
		filepath.FromSlash(kotlinFile): kotlinTypes(s),
		filepath.FromSlash(swiftFile):  swiftTypes(s),
	}, nil
}

// camel 把 JSON 名转成驼峰：episode_count → episodeCount。
func camel(json string) string {
	parts := strings.Split(json, "_")
	var b strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		if i == 0 {
			r[0] = unicode.ToLower(r[0])
		} else {
			r[0] = unicode.ToUpper(r[0])
		}
		b.WriteString(string(r))
	}
	return b.String()
}

func constWords(t schema.Type, c schema.Const) []string {
	return schema.Words(strings.TrimPrefix(c.GoName, schema.EnumPrefix(t.Consts)))
}

func docLines(doc string) []string {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return nil
	}
	return strings.Split(doc, "\n")
}

func indentLines(lines []string, indent, prefix string) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(strings.TrimRight(indent+prefix+l, " ") + "\n")
	}
	return b.String()
}

func isStruct(s *schema.Schema, r schema.Ref) bool {
	if r.Kind != schema.Named {
		return false
	}
	t, ok := s.Lookup(r.Name)
	return ok && t.Struct
}

func enumOf(s *schema.Schema, r schema.Ref) (schema.Ref, bool) {
	if r.Kind != schema.Named {
		return schema.Ref{}, false
	}
	t, ok := s.Lookup(r.Name)
	if !ok || t.Struct {
		return schema.Ref{}, false
	}
	return t.Underlying, true
}
