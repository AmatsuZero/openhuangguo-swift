// Package model 定义 HGhub 业务响应的强类型（对照 src/types.ts）。
// 泛型 Paged<T> 实例化为具名类型，匿名嵌套结构提为具名 struct，供 schema AST 解析与生成。
package model

type Author struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar,omitempty"`
	Followers int    `json:"followers,omitempty"`
}

// DetailAuthor 是详情里的作者（TS 的 Author & {is_following?...}），单独具名。
type DetailAuthor struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar,omitempty"`
	IsFollowing bool   `json:"is_following,omitempty"`
	Following   int    `json:"following,omitempty"`
	LikeCount   int    `json:"like_count,omitempty"`
}

type VideoCard struct {
	ID            int64    `json:"id"`
	Title         string   `json:"title"`
	Cover         string   `json:"cover"`
	Description   string   `json:"description,omitempty"`
	Tags          []string `json:"tags"`
	Score         float64  `json:"score"`
	EpisodeCount  int      `json:"episode_count"`
	TotalEpisodes int      `json:"total_episodes"`
	Duration      string   `json:"duration,omitempty"`
	Hot           int64    `json:"hot"`
	LikeCount     int64    `json:"like_count"`
	CollectCount  int64    `json:"collect_count"`
	CommentCount  int64    `json:"comment_count"`
	IsFinished    bool     `json:"is_finished"`
	IsOriginal    int      `json:"is_original"`
	CreatedAt     string   `json:"created_at,omitempty"`
	LastEpAt      *int64   `json:"last_ep_at,omitempty"`
	NewAt         *int64   `json:"new_at,omitempty"`
	Author        *Author  `json:"author,omitempty"`
}

type Pagination struct {
	Page  int `json:"page"`
	Pages int `json:"pages"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// 分页具名实例（避免泛型，供 schema AST 解析）。
type VideoCardPage struct {
	Items      []VideoCard `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

type PostPage struct {
	Items      []Post     `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type CommentPage struct {
	EP         int        `json:"ep"`
	Revision   string     `json:"revision,omitempty"`
	Reset      bool       `json:"reset,omitempty"`
	Items      []Comment  `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type SearchResult struct {
	Items      []VideoCard `json:"items"`
	Pagination Pagination  `json:"pagination"`
	Keyword    string      `json:"keyword"`
}

type Episode struct {
	EpNum    int    `json:"ep_num"`
	Title    string `json:"title"`
	Duration int    `json:"duration"`
	Status   int    `json:"status"`
}

type TagLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type DetailCounts struct {
	Collect int64 `json:"collect"`
	Comment int64 `json:"comment"`
	Hot     int64 `json:"hot"`
	Like    int64 `json:"like"`
}

type Interaction struct {
	Collected bool `json:"collected"`
	Liked     bool `json:"liked"`
	LikedEP   int  `json:"liked_ep,omitempty"`
}

type Detail struct {
	ID            int64        `json:"id"`
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	Cover         string       `json:"cover"`
	Tags          []string     `json:"tags"`
	TagLinks      []TagLink    `json:"tag_links,omitempty"`
	Score         float64      `json:"score"`
	EpisodeCount  int          `json:"episode_count"`
	TotalEpisodes int          `json:"total_episodes"`
	Episodes      []Episode    `json:"episodes"`
	Counts        DetailCounts `json:"counts"`
	Author        DetailAuthor `json:"author"`
	Interaction   *Interaction `json:"interaction,omitempty"`
	VideoURL      string       `json:"video_url,omitempty"`
	VideoURLHevc  string       `json:"video_url_hevc,omitempty"`
	IsFinished    bool         `json:"is_finished"`
	IsOriginal    int          `json:"is_original"`
	VideoType     int          `json:"video_type,omitempty"`
	Duration      int          `json:"duration,omitempty"`
	CreatedAt     string       `json:"created_at,omitempty"`
}

type Streams struct {
	Main    string `json:"main,omitempty"`
	Hevc    string `json:"hevc,omitempty"`
	Preview string `json:"preview,omitempty"`
}

type PlayInfo struct {
	ID           int64    `json:"id"`
	EpNum        int      `json:"ep_num"`
	Duration     int      `json:"duration"`
	Title        string   `json:"title"`
	VideoURL     string   `json:"video_url"`
	VideoURLHevc string   `json:"video_url_hevc,omitempty"`
	Streams      *Streams `json:"streams,omitempty"`
}

type RankMetrics struct {
	Collect int64 `json:"collect"`
	Comment int64 `json:"comment"`
	Like    int64 `json:"like"`
	Play    int64 `json:"play"`
}

type RankItem struct {
	Rank         int64       `json:"rank"`
	VideoID      int64       `json:"video_id"`
	Title        string      `json:"title"`
	Cover        string      `json:"cover"`
	Tags         []string    `json:"tags"`
	Score        float64     `json:"score"`
	CateName     string      `json:"cate_name"`
	EpisodeCount int         `json:"episode_count"`
	IsFinished   bool        `json:"is_finished"`
	MetricLabel  string      `json:"metric_label"`
	MetricValue  int64       `json:"metric_value"`
	Metrics      RankMetrics `json:"metrics"`
}

type Comment struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Nickname    string    `json:"nickname"`
	Avatar      string    `json:"avatar"`
	AvatarColor string    `json:"avatar_color"`
	Text        string    `json:"text"`
	Time        string    `json:"time"`
	CreatedAt   string    `json:"created_at,omitempty"`
	LikeCount   int64     `json:"like_count"`
	Liked       bool      `json:"liked"`
	ReplyCount  int       `json:"reply_count"`
	ReplyID     int64     `json:"reply_id"`
	ReplyList   []Comment `json:"reply_list,omitempty"`
	IsPinned    bool      `json:"is_pinned,omitempty"`
	IPRegion    string    `json:"ip_region,omitempty"`
}

type PostAuthor struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar,omitempty"`
	IsFollowing bool   `json:"is_following,omitempty"`
}

type PostCounts struct {
	Like    int64 `json:"like"`
	Comment int64 `json:"comment"`
	Collect int64 `json:"collect"`
	Share   int64 `json:"share,omitempty"`
	Hot     int64 `json:"hot,omitempty"`
}

type Post struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Desc        string      `json:"desc,omitempty"`
	Content     string      `json:"content,omitempty"`
	Cover       string      `json:"cover"`
	Tags        []string    `json:"tags"`
	Hot         int64       `json:"hot"`
	VideoURL    string      `json:"video_url,omitempty"`
	VideoID     int64       `json:"video_id,omitempty"`
	Author      *PostAuthor `json:"author,omitempty"`
	Counts      *PostCounts `json:"counts,omitempty"`
	CreatedAt   string      `json:"created_at,omitempty"`
	PublishTime string      `json:"publish_time,omitempty"`
}

type LikeVideoEpisode struct {
	VideoID int64 `json:"video_id"`
	EP      int   `json:"ep"`
}

type UserInteractions struct {
	LikeVideoIDs      []int64            `json:"like_video_ids"`
	FavoriteVideoIDs  []int64            `json:"favorite_video_ids"`
	FollowUserIDs     []int64            `json:"follow_user_ids"`
	LikeVideoEpisodes []LikeVideoEpisode `json:"like_video_episodes"`
}

type User struct {
	ID           int64             `json:"id"`
	Username     string            `json:"username"`
	Nickname     string            `json:"nickname"`
	Avatar       string            `json:"avatar"`
	AvatarColor  string            `json:"avatar_color"`
	Sign         string            `json:"sign,omitempty"`
	Desc         string            `json:"desc,omitempty"`
	UserType     string            `json:"user_type,omitempty"`
	Counts       map[string]int64  `json:"counts,omitempty"`
	Interactions *UserInteractions `json:"interactions,omitempty"`
}
