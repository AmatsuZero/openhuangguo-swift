package schema

import (
	"reflect"
	"testing"
)

func TestWords(t *testing.T) {
	for in, want := range map[string][]string{
		"CoverURL":      {"Cover", "URL"},
		"PPTRecommend":  {"PPT", "Recommend"},
		"PUGC":          {"PUGC"},
		"UGCVideo":      {"UGC", "Video"},
		"SeriesUpdated": {"Series", "Updated"},
	} {
		if got := Words(in); !reflect.DeepEqual(got, want) {
			t.Errorf("Words(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEnumPrefix(t *testing.T) {
	consts := func(names ...string) []Const {
		var out []Const
		for _, n := range names {
			out = append(out, Const{GoName: n})
		}
		return out
	}
	tests := []struct {
		names []string
		want  string
	}{
		{[]string{"ContentUnknown", "ContentShortSeries", "ContentPUGC"}, "Content"},
		{[]string{"SeriesUpdating", "SeriesUpdatedToday"}, "Series"},
		{[]string{"RankNew", "Rank"}, ""},
		{[]string{"Only"}, ""},
	}
	for _, tt := range tests {
		if got := EnumPrefix(consts(tt.names...)); got != tt.want {
			t.Errorf("EnumPrefix(%v) = %q, want %q", tt.names, got, tt.want)
		}
	}
}

// TestLoadModelOnly 解析 bindings/huangguo/sdk/model，Queries 必须为空。
func TestLoadModelOnly(t *testing.T) {
	s, err := LoadModelOnly("../..") // gen-huangguo/../.. = bindings/huangguo
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Queries) != 0 {
		t.Fatalf("huangguo should have no queries, got %+v", s.Queries)
	}
	v, ok := s.Lookup("VideoCard")
	if !ok || !v.Struct {
		t.Fatalf("VideoCard = %+v", v)
	}
	found := false
	for _, f := range s.AllFields(v) {
		if f.JSONName == "episode_count" {
			found = true
		}
	}
	if !found {
		t.Fatal("VideoCard missing episode_count")
	}
	if _, ok := s.Lookup("PlayInfo"); !ok {
		t.Fatal("PlayInfo not found")
	}
}
