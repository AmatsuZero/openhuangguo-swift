// 由 go run ./bindings/huangguo/gen-huangguo 生成，请勿手工修改。

import Foundation

public struct Author: Codable, Hashable, Sendable {
    public var id: Int
    public var name: String
    public var avatar: String
    public var followers: Int

    public init(id: Int = 0, name: String = "", avatar: String = "", followers: Int = 0) {
        self.id = id
        self.name = name
        self.avatar = avatar
        self.followers = followers
    }

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case avatar
        case followers
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decodeIfPresent(Int.self, forKey: .id) ?? 0
        self.name = try c.decodeIfPresent(String.self, forKey: .name) ?? ""
        self.avatar = try c.decodeIfPresent(String.self, forKey: .avatar) ?? ""
        self.followers = try c.decodeIfPresent(Int.self, forKey: .followers) ?? 0
    }
}

public struct Comment: Codable, Hashable, Sendable {
    public var id: Int
    public var userId: Int
    public var nickname: String
    public var avatar: String
    public var avatarColor: String
    public var text: String
    public var time: String
    public var createdAt: String
    public var likeCount: Int
    public var liked: Bool
    public var replyCount: Int
    public var replyId: Int
    public var replyList: [Comment]
    public var isPinned: Bool
    public var ipRegion: String

    public init(id: Int = 0, userId: Int = 0, nickname: String = "", avatar: String = "", avatarColor: String = "", text: String = "", time: String = "", createdAt: String = "", likeCount: Int = 0, liked: Bool = false, replyCount: Int = 0, replyId: Int = 0, replyList: [Comment] = [], isPinned: Bool = false, ipRegion: String = "") {
        self.id = id
        self.userId = userId
        self.nickname = nickname
        self.avatar = avatar
        self.avatarColor = avatarColor
        self.text = text
        self.time = time
        self.createdAt = createdAt
        self.likeCount = likeCount
        self.liked = liked
        self.replyCount = replyCount
        self.replyId = replyId
        self.replyList = replyList
        self.isPinned = isPinned
        self.ipRegion = ipRegion
    }

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case nickname
        case avatar
        case avatarColor = "avatar_color"
        case text
        case time
        case createdAt = "created_at"
        case likeCount = "like_count"
        case liked
        case replyCount = "reply_count"
        case replyId = "reply_id"
        case replyList = "reply_list"
        case isPinned = "is_pinned"
        case ipRegion = "ip_region"
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decodeIfPresent(Int.self, forKey: .id) ?? 0
        self.userId = try c.decodeIfPresent(Int.self, forKey: .userId) ?? 0
        self.nickname = try c.decodeIfPresent(String.self, forKey: .nickname) ?? ""
        self.avatar = try c.decodeIfPresent(String.self, forKey: .avatar) ?? ""
        self.avatarColor = try c.decodeIfPresent(String.self, forKey: .avatarColor) ?? ""
        self.text = try c.decodeIfPresent(String.self, forKey: .text) ?? ""
        self.time = try c.decodeIfPresent(String.self, forKey: .time) ?? ""
        self.createdAt = try c.decodeIfPresent(String.self, forKey: .createdAt) ?? ""
        self.likeCount = try c.decodeIfPresent(Int.self, forKey: .likeCount) ?? 0
        self.liked = try c.decodeIfPresent(Bool.self, forKey: .liked) ?? false
        self.replyCount = try c.decodeIfPresent(Int.self, forKey: .replyCount) ?? 0
        self.replyId = try c.decodeIfPresent(Int.self, forKey: .replyId) ?? 0
        self.replyList = try c.decodeIfPresent([Comment].self, forKey: .replyList) ?? []
        self.isPinned = try c.decodeIfPresent(Bool.self, forKey: .isPinned) ?? false
        self.ipRegion = try c.decodeIfPresent(String.self, forKey: .ipRegion) ?? ""
    }
}

public struct CommentPage: Codable, Hashable, Sendable {
    public var ep: Int
    public var revision: String
    public var reset: Bool
    public var items: [Comment]
    public var pagination: Pagination

    public init(ep: Int = 0, revision: String = "", reset: Bool = false, items: [Comment] = [], pagination: Pagination = Pagination()) {
        self.ep = ep
        self.revision = revision
        self.reset = reset
        self.items = items
        self.pagination = pagination
    }

    enum CodingKeys: String, CodingKey {
        case ep
        case revision
        case reset
        case items
        case pagination
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.ep = try c.decodeIfPresent(Int.self, forKey: .ep) ?? 0
        self.revision = try c.decodeIfPresent(String.self, forKey: .revision) ?? ""
        self.reset = try c.decodeIfPresent(Bool.self, forKey: .reset) ?? false
        self.items = try c.decodeIfPresent([Comment].self, forKey: .items) ?? []
        self.pagination = try c.decodeIfPresent(Pagination.self, forKey: .pagination) ?? Pagination()
    }
}

public struct Detail: Codable, Hashable, Sendable {
    public var id: Int
    public var title: String
    public var description: String
    public var cover: String
    public var tags: [String]
    public var tagLinks: [TagLink]
    public var score: Double
    public var episodeCount: Int
    public var totalEpisodes: Int
    public var episodes: [Episode]
    public var counts: DetailCounts
    public var author: DetailAuthor
    public var interaction: Interaction?
    public var videoUrl: String
    public var videoUrlHevc: String
    public var isFinished: Bool
    public var isOriginal: Int
    public var videoType: Int
    public var duration: Int
    public var createdAt: String

    public init(id: Int = 0, title: String = "", description: String = "", cover: String = "", tags: [String] = [], tagLinks: [TagLink] = [], score: Double = 0, episodeCount: Int = 0, totalEpisodes: Int = 0, episodes: [Episode] = [], counts: DetailCounts = DetailCounts(), author: DetailAuthor = DetailAuthor(), interaction: Interaction? = nil, videoUrl: String = "", videoUrlHevc: String = "", isFinished: Bool = false, isOriginal: Int = 0, videoType: Int = 0, duration: Int = 0, createdAt: String = "") {
        self.id = id
        self.title = title
        self.description = description
        self.cover = cover
        self.tags = tags
        self.tagLinks = tagLinks
        self.score = score
        self.episodeCount = episodeCount
        self.totalEpisodes = totalEpisodes
        self.episodes = episodes
        self.counts = counts
        self.author = author
        self.interaction = interaction
        self.videoUrl = videoUrl
        self.videoUrlHevc = videoUrlHevc
        self.isFinished = isFinished
        self.isOriginal = isOriginal
        self.videoType = videoType
        self.duration = duration
        self.createdAt = createdAt
    }

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case description
        case cover
        case tags
        case tagLinks = "tag_links"
        case score
        case episodeCount = "episode_count"
        case totalEpisodes = "total_episodes"
        case episodes
        case counts
        case author
        case interaction
        case videoUrl = "video_url"
        case videoUrlHevc = "video_url_hevc"
        case isFinished = "is_finished"
        case isOriginal = "is_original"
        case videoType = "video_type"
        case duration
        case createdAt = "created_at"
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decodeIfPresent(Int.self, forKey: .id) ?? 0
        self.title = try c.decodeIfPresent(String.self, forKey: .title) ?? ""
        self.description = try c.decodeIfPresent(String.self, forKey: .description) ?? ""
        self.cover = try c.decodeIfPresent(String.self, forKey: .cover) ?? ""
        self.tags = try c.decodeIfPresent([String].self, forKey: .tags) ?? []
        self.tagLinks = try c.decodeIfPresent([TagLink].self, forKey: .tagLinks) ?? []
        self.score = try c.decodeIfPresent(Double.self, forKey: .score) ?? 0
        self.episodeCount = try c.decodeIfPresent(Int.self, forKey: .episodeCount) ?? 0
        self.totalEpisodes = try c.decodeIfPresent(Int.self, forKey: .totalEpisodes) ?? 0
        self.episodes = try c.decodeIfPresent([Episode].self, forKey: .episodes) ?? []
        self.counts = try c.decodeIfPresent(DetailCounts.self, forKey: .counts) ?? DetailCounts()
        self.author = try c.decodeIfPresent(DetailAuthor.self, forKey: .author) ?? DetailAuthor()
        self.interaction = try c.decodeIfPresent(Interaction.self, forKey: .interaction)
        self.videoUrl = try c.decodeIfPresent(String.self, forKey: .videoUrl) ?? ""
        self.videoUrlHevc = try c.decodeIfPresent(String.self, forKey: .videoUrlHevc) ?? ""
        self.isFinished = try c.decodeIfPresent(Bool.self, forKey: .isFinished) ?? false
        self.isOriginal = try c.decodeIfPresent(Int.self, forKey: .isOriginal) ?? 0
        self.videoType = try c.decodeIfPresent(Int.self, forKey: .videoType) ?? 0
        self.duration = try c.decodeIfPresent(Int.self, forKey: .duration) ?? 0
        self.createdAt = try c.decodeIfPresent(String.self, forKey: .createdAt) ?? ""
    }
}

/// DetailAuthor 是详情里的作者（TS 的 Author & {is_following?...}），单独具名。
public struct DetailAuthor: Codable, Hashable, Sendable {
    public var id: Int
    public var name: String
    public var avatar: String
    public var isFollowing: Bool
    public var following: Int
    public var likeCount: Int

    public init(id: Int = 0, name: String = "", avatar: String = "", isFollowing: Bool = false, following: Int = 0, likeCount: Int = 0) {
        self.id = id
        self.name = name
        self.avatar = avatar
        self.isFollowing = isFollowing
        self.following = following
        self.likeCount = likeCount
    }

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case avatar
        case isFollowing = "is_following"
        case following
        case likeCount = "like_count"
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decodeIfPresent(Int.self, forKey: .id) ?? 0
        self.name = try c.decodeIfPresent(String.self, forKey: .name) ?? ""
        self.avatar = try c.decodeIfPresent(String.self, forKey: .avatar) ?? ""
        self.isFollowing = try c.decodeIfPresent(Bool.self, forKey: .isFollowing) ?? false
        self.following = try c.decodeIfPresent(Int.self, forKey: .following) ?? 0
        self.likeCount = try c.decodeIfPresent(Int.self, forKey: .likeCount) ?? 0
    }
}

public struct DetailCounts: Codable, Hashable, Sendable {
    public var collect: Int
    public var comment: Int
    public var hot: Int
    public var like: Int

    public init(collect: Int = 0, comment: Int = 0, hot: Int = 0, like: Int = 0) {
        self.collect = collect
        self.comment = comment
        self.hot = hot
        self.like = like
    }

    enum CodingKeys: String, CodingKey {
        case collect
        case comment
        case hot
        case like
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.collect = try c.decodeIfPresent(Int.self, forKey: .collect) ?? 0
        self.comment = try c.decodeIfPresent(Int.self, forKey: .comment) ?? 0
        self.hot = try c.decodeIfPresent(Int.self, forKey: .hot) ?? 0
        self.like = try c.decodeIfPresent(Int.self, forKey: .like) ?? 0
    }
}

public struct Episode: Codable, Hashable, Sendable {
    public var epNum: Int
    public var title: String
    public var duration: Int
    public var status: Int

    public init(epNum: Int = 0, title: String = "", duration: Int = 0, status: Int = 0) {
        self.epNum = epNum
        self.title = title
        self.duration = duration
        self.status = status
    }

    enum CodingKeys: String, CodingKey {
        case epNum = "ep_num"
        case title
        case duration
        case status
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.epNum = try c.decodeIfPresent(Int.self, forKey: .epNum) ?? 0
        self.title = try c.decodeIfPresent(String.self, forKey: .title) ?? ""
        self.duration = try c.decodeIfPresent(Int.self, forKey: .duration) ?? 0
        self.status = try c.decodeIfPresent(Int.self, forKey: .status) ?? 0
    }
}

public struct Interaction: Codable, Hashable, Sendable {
    public var collected: Bool
    public var liked: Bool
    public var likedEp: Int

    public init(collected: Bool = false, liked: Bool = false, likedEp: Int = 0) {
        self.collected = collected
        self.liked = liked
        self.likedEp = likedEp
    }

    enum CodingKeys: String, CodingKey {
        case collected
        case liked
        case likedEp = "liked_ep"
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.collected = try c.decodeIfPresent(Bool.self, forKey: .collected) ?? false
        self.liked = try c.decodeIfPresent(Bool.self, forKey: .liked) ?? false
        self.likedEp = try c.decodeIfPresent(Int.self, forKey: .likedEp) ?? 0
    }
}

public struct LikeVideoEpisode: Codable, Hashable, Sendable {
    public var videoId: Int
    public var ep: Int

    public init(videoId: Int = 0, ep: Int = 0) {
        self.videoId = videoId
        self.ep = ep
    }

    enum CodingKeys: String, CodingKey {
        case videoId = "video_id"
        case ep
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.videoId = try c.decodeIfPresent(Int.self, forKey: .videoId) ?? 0
        self.ep = try c.decodeIfPresent(Int.self, forKey: .ep) ?? 0
    }
}

public struct Pagination: Codable, Hashable, Sendable {
    public var page: Int
    public var pages: Int
    public var size: Int
    public var total: Int

    public init(page: Int = 0, pages: Int = 0, size: Int = 0, total: Int = 0) {
        self.page = page
        self.pages = pages
        self.size = size
        self.total = total
    }

    enum CodingKeys: String, CodingKey {
        case page
        case pages
        case size
        case total
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.page = try c.decodeIfPresent(Int.self, forKey: .page) ?? 0
        self.pages = try c.decodeIfPresent(Int.self, forKey: .pages) ?? 0
        self.size = try c.decodeIfPresent(Int.self, forKey: .size) ?? 0
        self.total = try c.decodeIfPresent(Int.self, forKey: .total) ?? 0
    }
}

public struct PlayInfo: Codable, Hashable, Sendable {
    public var id: Int
    public var epNum: Int
    public var duration: Int
    public var title: String
    public var videoUrl: String
    public var videoUrlHevc: String
    public var streams: Streams?

    public init(id: Int = 0, epNum: Int = 0, duration: Int = 0, title: String = "", videoUrl: String = "", videoUrlHevc: String = "", streams: Streams? = nil) {
        self.id = id
        self.epNum = epNum
        self.duration = duration
        self.title = title
        self.videoUrl = videoUrl
        self.videoUrlHevc = videoUrlHevc
        self.streams = streams
    }

    enum CodingKeys: String, CodingKey {
        case id
        case epNum = "ep_num"
        case duration
        case title
        case videoUrl = "video_url"
        case videoUrlHevc = "video_url_hevc"
        case streams
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decodeIfPresent(Int.self, forKey: .id) ?? 0
        self.epNum = try c.decodeIfPresent(Int.self, forKey: .epNum) ?? 0
        self.duration = try c.decodeIfPresent(Int.self, forKey: .duration) ?? 0
        self.title = try c.decodeIfPresent(String.self, forKey: .title) ?? ""
        self.videoUrl = try c.decodeIfPresent(String.self, forKey: .videoUrl) ?? ""
        self.videoUrlHevc = try c.decodeIfPresent(String.self, forKey: .videoUrlHevc) ?? ""
        self.streams = try c.decodeIfPresent(Streams.self, forKey: .streams)
    }
}

public struct Post: Codable, Hashable, Sendable {
    public var id: Int
    public var title: String
    public var desc: String
    public var content: String
    public var cover: String
    public var tags: [String]
    public var hot: Int
    public var videoUrl: String
    public var videoId: Int
    public var author: PostAuthor?
    public var counts: PostCounts?
    public var createdAt: String
    public var publishTime: String

    public init(id: Int = 0, title: String = "", desc: String = "", content: String = "", cover: String = "", tags: [String] = [], hot: Int = 0, videoUrl: String = "", videoId: Int = 0, author: PostAuthor? = nil, counts: PostCounts? = nil, createdAt: String = "", publishTime: String = "") {
        self.id = id
        self.title = title
        self.desc = desc
        self.content = content
        self.cover = cover
        self.tags = tags
        self.hot = hot
        self.videoUrl = videoUrl
        self.videoId = videoId
        self.author = author
        self.counts = counts
        self.createdAt = createdAt
        self.publishTime = publishTime
    }

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case desc
        case content
        case cover
        case tags
        case hot
        case videoUrl = "video_url"
        case videoId = "video_id"
        case author
        case counts
        case createdAt = "created_at"
        case publishTime = "publish_time"
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decodeIfPresent(Int.self, forKey: .id) ?? 0
        self.title = try c.decodeIfPresent(String.self, forKey: .title) ?? ""
        self.desc = try c.decodeIfPresent(String.self, forKey: .desc) ?? ""
        self.content = try c.decodeIfPresent(String.self, forKey: .content) ?? ""
        self.cover = try c.decodeIfPresent(String.self, forKey: .cover) ?? ""
        self.tags = try c.decodeIfPresent([String].self, forKey: .tags) ?? []
        self.hot = try c.decodeIfPresent(Int.self, forKey: .hot) ?? 0
        self.videoUrl = try c.decodeIfPresent(String.self, forKey: .videoUrl) ?? ""
        self.videoId = try c.decodeIfPresent(Int.self, forKey: .videoId) ?? 0
        self.author = try c.decodeIfPresent(PostAuthor.self, forKey: .author)
        self.counts = try c.decodeIfPresent(PostCounts.self, forKey: .counts)
        self.createdAt = try c.decodeIfPresent(String.self, forKey: .createdAt) ?? ""
        self.publishTime = try c.decodeIfPresent(String.self, forKey: .publishTime) ?? ""
    }
}

public struct PostAuthor: Codable, Hashable, Sendable {
    public var id: Int
    public var name: String
    public var avatar: String
    public var isFollowing: Bool

    public init(id: Int = 0, name: String = "", avatar: String = "", isFollowing: Bool = false) {
        self.id = id
        self.name = name
        self.avatar = avatar
        self.isFollowing = isFollowing
    }

    enum CodingKeys: String, CodingKey {
        case id
        case name
        case avatar
        case isFollowing = "is_following"
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decodeIfPresent(Int.self, forKey: .id) ?? 0
        self.name = try c.decodeIfPresent(String.self, forKey: .name) ?? ""
        self.avatar = try c.decodeIfPresent(String.self, forKey: .avatar) ?? ""
        self.isFollowing = try c.decodeIfPresent(Bool.self, forKey: .isFollowing) ?? false
    }
}

public struct PostCounts: Codable, Hashable, Sendable {
    public var like: Int
    public var comment: Int
    public var collect: Int
    public var share: Int
    public var hot: Int

    public init(like: Int = 0, comment: Int = 0, collect: Int = 0, share: Int = 0, hot: Int = 0) {
        self.like = like
        self.comment = comment
        self.collect = collect
        self.share = share
        self.hot = hot
    }

    enum CodingKeys: String, CodingKey {
        case like
        case comment
        case collect
        case share
        case hot
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.like = try c.decodeIfPresent(Int.self, forKey: .like) ?? 0
        self.comment = try c.decodeIfPresent(Int.self, forKey: .comment) ?? 0
        self.collect = try c.decodeIfPresent(Int.self, forKey: .collect) ?? 0
        self.share = try c.decodeIfPresent(Int.self, forKey: .share) ?? 0
        self.hot = try c.decodeIfPresent(Int.self, forKey: .hot) ?? 0
    }
}

public struct PostPage: Codable, Hashable, Sendable {
    public var items: [Post]
    public var pagination: Pagination

    public init(items: [Post] = [], pagination: Pagination = Pagination()) {
        self.items = items
        self.pagination = pagination
    }

    enum CodingKeys: String, CodingKey {
        case items
        case pagination
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.items = try c.decodeIfPresent([Post].self, forKey: .items) ?? []
        self.pagination = try c.decodeIfPresent(Pagination.self, forKey: .pagination) ?? Pagination()
    }
}

public struct RankItem: Codable, Hashable, Sendable {
    public var rank: Int
    public var videoId: Int
    public var title: String
    public var cover: String
    public var tags: [String]
    public var score: Double
    public var cateName: String
    public var episodeCount: Int
    public var isFinished: Bool
    public var metricLabel: String
    public var metricValue: Int
    public var metrics: RankMetrics

    public init(rank: Int = 0, videoId: Int = 0, title: String = "", cover: String = "", tags: [String] = [], score: Double = 0, cateName: String = "", episodeCount: Int = 0, isFinished: Bool = false, metricLabel: String = "", metricValue: Int = 0, metrics: RankMetrics = RankMetrics()) {
        self.rank = rank
        self.videoId = videoId
        self.title = title
        self.cover = cover
        self.tags = tags
        self.score = score
        self.cateName = cateName
        self.episodeCount = episodeCount
        self.isFinished = isFinished
        self.metricLabel = metricLabel
        self.metricValue = metricValue
        self.metrics = metrics
    }

    enum CodingKeys: String, CodingKey {
        case rank
        case videoId = "video_id"
        case title
        case cover
        case tags
        case score
        case cateName = "cate_name"
        case episodeCount = "episode_count"
        case isFinished = "is_finished"
        case metricLabel = "metric_label"
        case metricValue = "metric_value"
        case metrics
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.rank = try c.decodeIfPresent(Int.self, forKey: .rank) ?? 0
        self.videoId = try c.decodeIfPresent(Int.self, forKey: .videoId) ?? 0
        self.title = try c.decodeIfPresent(String.self, forKey: .title) ?? ""
        self.cover = try c.decodeIfPresent(String.self, forKey: .cover) ?? ""
        self.tags = try c.decodeIfPresent([String].self, forKey: .tags) ?? []
        self.score = try c.decodeIfPresent(Double.self, forKey: .score) ?? 0
        self.cateName = try c.decodeIfPresent(String.self, forKey: .cateName) ?? ""
        self.episodeCount = try c.decodeIfPresent(Int.self, forKey: .episodeCount) ?? 0
        self.isFinished = try c.decodeIfPresent(Bool.self, forKey: .isFinished) ?? false
        self.metricLabel = try c.decodeIfPresent(String.self, forKey: .metricLabel) ?? ""
        self.metricValue = try c.decodeIfPresent(Int.self, forKey: .metricValue) ?? 0
        self.metrics = try c.decodeIfPresent(RankMetrics.self, forKey: .metrics) ?? RankMetrics()
    }
}

public struct RankMetrics: Codable, Hashable, Sendable {
    public var collect: Int
    public var comment: Int
    public var like: Int
    public var play: Int

    public init(collect: Int = 0, comment: Int = 0, like: Int = 0, play: Int = 0) {
        self.collect = collect
        self.comment = comment
        self.like = like
        self.play = play
    }

    enum CodingKeys: String, CodingKey {
        case collect
        case comment
        case like
        case play
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.collect = try c.decodeIfPresent(Int.self, forKey: .collect) ?? 0
        self.comment = try c.decodeIfPresent(Int.self, forKey: .comment) ?? 0
        self.like = try c.decodeIfPresent(Int.self, forKey: .like) ?? 0
        self.play = try c.decodeIfPresent(Int.self, forKey: .play) ?? 0
    }
}

/// RankPage 是 /api/ranks/{kind} 的返回（对照 src/api.ts 的 {type, items}）。
public struct RankPage: Codable, Hashable, Sendable {
    public var type: String
    public var items: [RankItem]

    public init(type: String = "", items: [RankItem] = []) {
        self.type = type
        self.items = items
    }

    enum CodingKeys: String, CodingKey {
        case type
        case items
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.type = try c.decodeIfPresent(String.self, forKey: .type) ?? ""
        self.items = try c.decodeIfPresent([RankItem].self, forKey: .items) ?? []
    }
}

public struct SearchResult: Codable, Hashable, Sendable {
    public var items: [VideoCard]
    public var pagination: Pagination
    public var keyword: String

    public init(items: [VideoCard] = [], pagination: Pagination = Pagination(), keyword: String = "") {
        self.items = items
        self.pagination = pagination
        self.keyword = keyword
    }

    enum CodingKeys: String, CodingKey {
        case items
        case pagination
        case keyword
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.items = try c.decodeIfPresent([VideoCard].self, forKey: .items) ?? []
        self.pagination = try c.decodeIfPresent(Pagination.self, forKey: .pagination) ?? Pagination()
        self.keyword = try c.decodeIfPresent(String.self, forKey: .keyword) ?? ""
    }
}

public struct Streams: Codable, Hashable, Sendable {
    public var main: String
    public var hevc: String
    public var preview: String

    public init(main: String = "", hevc: String = "", preview: String = "") {
        self.main = main
        self.hevc = hevc
        self.preview = preview
    }

    enum CodingKeys: String, CodingKey {
        case main
        case hevc
        case preview
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.main = try c.decodeIfPresent(String.self, forKey: .main) ?? ""
        self.hevc = try c.decodeIfPresent(String.self, forKey: .hevc) ?? ""
        self.preview = try c.decodeIfPresent(String.self, forKey: .preview) ?? ""
    }
}

public struct TagLink: Codable, Hashable, Sendable {
    public var name: String
    public var url: String

    public init(name: String = "", url: String = "") {
        self.name = name
        self.url = url
    }

    enum CodingKeys: String, CodingKey {
        case name
        case url
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.name = try c.decodeIfPresent(String.self, forKey: .name) ?? ""
        self.url = try c.decodeIfPresent(String.self, forKey: .url) ?? ""
    }
}

public struct User: Codable, Hashable, Sendable {
    public var id: Int
    public var username: String
    public var nickname: String
    public var avatar: String
    public var avatarColor: String
    public var sign: String
    public var desc: String
    public var userType: String
    public var counts: [String: Int]
    public var interactions: UserInteractions?

    public init(id: Int = 0, username: String = "", nickname: String = "", avatar: String = "", avatarColor: String = "", sign: String = "", desc: String = "", userType: String = "", counts: [String: Int] = [:], interactions: UserInteractions? = nil) {
        self.id = id
        self.username = username
        self.nickname = nickname
        self.avatar = avatar
        self.avatarColor = avatarColor
        self.sign = sign
        self.desc = desc
        self.userType = userType
        self.counts = counts
        self.interactions = interactions
    }

    enum CodingKeys: String, CodingKey {
        case id
        case username
        case nickname
        case avatar
        case avatarColor = "avatar_color"
        case sign
        case desc
        case userType = "user_type"
        case counts
        case interactions
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decodeIfPresent(Int.self, forKey: .id) ?? 0
        self.username = try c.decodeIfPresent(String.self, forKey: .username) ?? ""
        self.nickname = try c.decodeIfPresent(String.self, forKey: .nickname) ?? ""
        self.avatar = try c.decodeIfPresent(String.self, forKey: .avatar) ?? ""
        self.avatarColor = try c.decodeIfPresent(String.self, forKey: .avatarColor) ?? ""
        self.sign = try c.decodeIfPresent(String.self, forKey: .sign) ?? ""
        self.desc = try c.decodeIfPresent(String.self, forKey: .desc) ?? ""
        self.userType = try c.decodeIfPresent(String.self, forKey: .userType) ?? ""
        self.counts = try c.decodeIfPresent([String: Int].self, forKey: .counts) ?? [:]
        self.interactions = try c.decodeIfPresent(UserInteractions.self, forKey: .interactions)
    }
}

public struct UserInteractions: Codable, Hashable, Sendable {
    public var likeVideoIds: [Int]
    public var favoriteVideoIds: [Int]
    public var followUserIds: [Int]
    public var likeVideoEpisodes: [LikeVideoEpisode]

    public init(likeVideoIds: [Int] = [], favoriteVideoIds: [Int] = [], followUserIds: [Int] = [], likeVideoEpisodes: [LikeVideoEpisode] = []) {
        self.likeVideoIds = likeVideoIds
        self.favoriteVideoIds = favoriteVideoIds
        self.followUserIds = followUserIds
        self.likeVideoEpisodes = likeVideoEpisodes
    }

    enum CodingKeys: String, CodingKey {
        case likeVideoIds = "like_video_ids"
        case favoriteVideoIds = "favorite_video_ids"
        case followUserIds = "follow_user_ids"
        case likeVideoEpisodes = "like_video_episodes"
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.likeVideoIds = try c.decodeIfPresent([Int].self, forKey: .likeVideoIds) ?? []
        self.favoriteVideoIds = try c.decodeIfPresent([Int].self, forKey: .favoriteVideoIds) ?? []
        self.followUserIds = try c.decodeIfPresent([Int].self, forKey: .followUserIds) ?? []
        self.likeVideoEpisodes = try c.decodeIfPresent([LikeVideoEpisode].self, forKey: .likeVideoEpisodes) ?? []
    }
}

public struct VideoCard: Codable, Hashable, Sendable {
    public var id: Int
    public var title: String
    public var cover: String
    public var description: String
    public var tags: [String]
    public var score: Double
    public var episodeCount: Int
    public var totalEpisodes: Int
    public var duration: String
    public var hot: Int
    public var likeCount: Int
    public var collectCount: Int
    public var commentCount: Int
    public var isFinished: Bool
    public var isOriginal: Int
    public var createdAt: String
    public var lastEpAt: Int?
    public var newAt: Int?
    public var author: Author?

    public init(id: Int = 0, title: String = "", cover: String = "", description: String = "", tags: [String] = [], score: Double = 0, episodeCount: Int = 0, totalEpisodes: Int = 0, duration: String = "", hot: Int = 0, likeCount: Int = 0, collectCount: Int = 0, commentCount: Int = 0, isFinished: Bool = false, isOriginal: Int = 0, createdAt: String = "", lastEpAt: Int? = nil, newAt: Int? = nil, author: Author? = nil) {
        self.id = id
        self.title = title
        self.cover = cover
        self.description = description
        self.tags = tags
        self.score = score
        self.episodeCount = episodeCount
        self.totalEpisodes = totalEpisodes
        self.duration = duration
        self.hot = hot
        self.likeCount = likeCount
        self.collectCount = collectCount
        self.commentCount = commentCount
        self.isFinished = isFinished
        self.isOriginal = isOriginal
        self.createdAt = createdAt
        self.lastEpAt = lastEpAt
        self.newAt = newAt
        self.author = author
    }

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case cover
        case description
        case tags
        case score
        case episodeCount = "episode_count"
        case totalEpisodes = "total_episodes"
        case duration
        case hot
        case likeCount = "like_count"
        case collectCount = "collect_count"
        case commentCount = "comment_count"
        case isFinished = "is_finished"
        case isOriginal = "is_original"
        case createdAt = "created_at"
        case lastEpAt = "last_ep_at"
        case newAt = "new_at"
        case author
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decodeIfPresent(Int.self, forKey: .id) ?? 0
        self.title = try c.decodeIfPresent(String.self, forKey: .title) ?? ""
        self.cover = try c.decodeIfPresent(String.self, forKey: .cover) ?? ""
        self.description = try c.decodeIfPresent(String.self, forKey: .description) ?? ""
        self.tags = try c.decodeIfPresent([String].self, forKey: .tags) ?? []
        self.score = try c.decodeIfPresent(Double.self, forKey: .score) ?? 0
        self.episodeCount = try c.decodeIfPresent(Int.self, forKey: .episodeCount) ?? 0
        self.totalEpisodes = try c.decodeIfPresent(Int.self, forKey: .totalEpisodes) ?? 0
        self.duration = try c.decodeIfPresent(String.self, forKey: .duration) ?? ""
        self.hot = try c.decodeIfPresent(Int.self, forKey: .hot) ?? 0
        self.likeCount = try c.decodeIfPresent(Int.self, forKey: .likeCount) ?? 0
        self.collectCount = try c.decodeIfPresent(Int.self, forKey: .collectCount) ?? 0
        self.commentCount = try c.decodeIfPresent(Int.self, forKey: .commentCount) ?? 0
        self.isFinished = try c.decodeIfPresent(Bool.self, forKey: .isFinished) ?? false
        self.isOriginal = try c.decodeIfPresent(Int.self, forKey: .isOriginal) ?? 0
        self.createdAt = try c.decodeIfPresent(String.self, forKey: .createdAt) ?? ""
        self.lastEpAt = try c.decodeIfPresent(Int.self, forKey: .lastEpAt)
        self.newAt = try c.decodeIfPresent(Int.self, forKey: .newAt)
        self.author = try c.decodeIfPresent(Author.self, forKey: .author)
    }
}

/// 分页具名实例（避免泛型，供 schema AST 解析）。
public struct VideoCardPage: Codable, Hashable, Sendable {
    public var items: [VideoCard]
    public var pagination: Pagination

    public init(items: [VideoCard] = [], pagination: Pagination = Pagination()) {
        self.items = items
        self.pagination = pagination
    }

    enum CodingKeys: String, CodingKey {
        case items
        case pagination
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.items = try c.decodeIfPresent([VideoCard].self, forKey: .items) ?? []
        self.pagination = try c.decodeIfPresent(Pagination.self, forKey: .pagination) ?? Pagination()
    }
}
