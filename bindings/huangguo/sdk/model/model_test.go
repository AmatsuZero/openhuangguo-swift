package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVideoCardRoundTrip(t *testing.T) {
	raw := `{"id":1,"title":"t","cover":"c","description":"d","tags":["a"],"score":9,
      "episode_count":2,"total_episodes":3,"hot":100,"like_count":3,"collect_count":4,
      "comment_count":5,"is_finished":true,"is_original":1,"author":{"id":7,"name":"n","followers":8}}`
	var v VideoCard
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatal(err)
	}
	if v.ID != 1 || v.Author == nil || v.Author.ID != 7 || v.Author.Followers != 8 {
		t.Fatalf("bad: %+v", v)
	}
	out, _ := json.Marshal(v)
	for _, key := range []string{`"episode_count"`, `"total_episodes"`, `"is_finished"`, `"like_count"`} {
		if !strings.Contains(string(out), key) {
			t.Fatalf("missing key %s in %s", key, out)
		}
	}
}

func TestDetailAndPlayInfoNested(t *testing.T) {
	raw := `{"id":1,"title":"t","description":"d","cover":"c","tags":[],
      "score":9,"episode_count":1,"total_episodes":1,"episodes":[],
      "counts":{"collect":1,"comment":2,"hot":3,"like":4},
      "author":{"id":3,"name":"n","is_following":true,"following":5,"like_count":6},
      "is_finished":true,"is_original":0,"video_url":""}`
	var d Detail
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatal(err)
	}
	if d.Counts.Like != 4 || d.Author.ID != 3 || !d.Author.IsFollowing {
		t.Fatalf("bad: %+v", d)
	}
	var p PlayInfo
	if err := json.Unmarshal([]byte(`{"id":1,"ep_num":2,"video_url":"http://m","streams":{"main":"http://m"}}`), &p); err != nil {
		t.Fatal(err)
	}
	if p.EpNum != 2 || p.Streams == nil || p.Streams.Main != "http://m" {
		t.Fatalf("bad: %+v", p)
	}
}

func TestPagedAndRankAndComment(t *testing.T) {
	var page VideoCardPage
	if err := json.Unmarshal([]byte(`{"items":[{"id":1,"title":"t","cover":"c","tags":[],"score":1,"episode_count":1,"total_episodes":1,"hot":1,"like_count":1,"collect_count":1,"comment_count":1,"is_finished":true,"is_original":0}],"pagination":{"page":1,"pages":1,"size":20,"total":1}}`), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Pagination.Total != 1 {
		t.Fatalf("bad page: %+v", page)
	}
	var r RankItem
	if err := json.Unmarshal([]byte(`{"rank":1,"video_id":2,"title":"t","cover":"c","tags":[],"score":9,"cate_name":"cate","episode_count":1,"is_finished":true,"metric_label":"播放","metric_value":10,"metrics":{"collect":1,"comment":2,"like":3,"play":10}}`), &r); err != nil {
		t.Fatal(err)
	}
	if r.Metrics.Play != 10 || r.VideoID != 2 {
		t.Fatalf("bad rank: %+v", r)
	}
	var cp CommentPage
	if err := json.Unmarshal([]byte(`{"ep":1,"revision":"r","reset":true,"items":[{"id":1,"user_id":2,"nickname":"n","avatar":"a","avatar_color":"c","text":"t","time":"now","like_count":1,"liked":false,"reply_count":0,"reply_id":0}],"pagination":{"page":1,"pages":1,"size":20,"total":1}}`), &cp); err != nil {
		t.Fatal(err)
	}
	if cp.EP != 1 || !cp.Reset || len(cp.Items) != 1 || cp.Items[0].Nickname != "n" {
		t.Fatalf("bad comment page: %+v", cp)
	}
}

func TestPostAndUser(t *testing.T) {
	var p Post
	if err := json.Unmarshal([]byte(`{"id":1,"title":"t","cover":"c","tags":[],"hot":1,"author":{"id":2,"name":"n","is_following":true},"counts":{"like":1,"comment":2,"collect":3,"share":4}}`), &p); err != nil {
		t.Fatal(err)
	}
	if p.Author == nil || !p.Author.IsFollowing || p.Counts == nil || p.Counts.Share != 4 {
		t.Fatalf("bad post: %+v", p)
	}
	var u User
	if err := json.Unmarshal([]byte(`{"id":1,"username":"u","nickname":"n","avatar":"a","avatar_color":"c","counts":{"followers":5},"interactions":{"like_video_ids":[1,2],"favorite_video_ids":[],"follow_user_ids":[],"like_video_episodes":[{"video_id":1,"ep":2}]}}`), &u); err != nil {
		t.Fatal(err)
	}
	if u.Counts["followers"] != 5 || len(u.Interactions.LikeVideoIDs) != 2 || u.Interactions.LikeVideoEpisodes[0].EP != 2 {
		t.Fatalf("bad user: %+v", u)
	}
}
