// 由 go run ./bindings/huangguo/gen-huangguo 生成，请勿手工修改。

package dev.openhuangguo

import org.json.JSONArray
import org.json.JSONObject

data class Author(
    val id: Long = 0L,
    val name: String = "",
    val avatar: String = "",
    val followers: Int = 0,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("id", id)
        put("name", name)
        put("avatar", avatar)
        put("followers", followers)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): Author = Author(
            id = o.opt("id").asLong(),
            name = o.opt("name").asString(),
            avatar = o.opt("avatar").asString(),
            followers = o.opt("followers").asInt(),
        )
    }
}

data class Comment(
    val id: Long = 0L,
    val userId: Long = 0L,
    val nickname: String = "",
    val avatar: String = "",
    val avatarColor: String = "",
    val text: String = "",
    val time: String = "",
    val createdAt: String = "",
    val likeCount: Long = 0L,
    val liked: Boolean = false,
    val replyCount: Int = 0,
    val replyId: Long = 0L,
    val replyList: List<Comment> = emptyList(),
    val isPinned: Boolean = false,
    val ipRegion: String = "",
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("id", id)
        put("user_id", userId)
        put("nickname", nickname)
        put("avatar", avatar)
        put("avatar_color", avatarColor)
        put("text", text)
        put("time", time)
        put("created_at", createdAt)
        put("like_count", likeCount)
        put("liked", liked)
        put("reply_count", replyCount)
        put("reply_id", replyId)
        put("reply_list", JSONArray(replyList.map { it.toJson() }))
        put("is_pinned", isPinned)
        put("ip_region", ipRegion)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): Comment = Comment(
            id = o.opt("id").asLong(),
            userId = o.opt("user_id").asLong(),
            nickname = o.opt("nickname").asString(),
            avatar = o.opt("avatar").asString(),
            avatarColor = o.opt("avatar_color").asString(),
            text = o.opt("text").asString(),
            time = o.opt("time").asString(),
            createdAt = o.opt("created_at").asString(),
            likeCount = o.opt("like_count").asLong(),
            liked = o.opt("liked").asBoolean(),
            replyCount = o.opt("reply_count").asInt(),
            replyId = o.opt("reply_id").asLong(),
            replyList = o.opt("reply_list").asList { Comment.fromJson(it.asObject()) },
            isPinned = o.opt("is_pinned").asBoolean(),
            ipRegion = o.opt("ip_region").asString(),
        )
    }
}

data class CommentPage(
    val ep: Int = 0,
    val revision: String = "",
    val reset: Boolean = false,
    val items: List<Comment> = emptyList(),
    val pagination: Pagination = Pagination(),
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("ep", ep)
        put("revision", revision)
        put("reset", reset)
        put("items", JSONArray(items.map { it.toJson() }))
        put("pagination", pagination.toJson())
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): CommentPage = CommentPage(
            ep = o.opt("ep").asInt(),
            revision = o.opt("revision").asString(),
            reset = o.opt("reset").asBoolean(),
            items = o.opt("items").asList { Comment.fromJson(it.asObject()) },
            pagination = Pagination.fromJson(o.opt("pagination").asObject()),
        )
    }
}

data class Detail(
    val id: Long = 0L,
    val title: String = "",
    val description: String = "",
    val cover: String = "",
    val tags: List<String> = emptyList(),
    val tagLinks: List<TagLink> = emptyList(),
    val score: Double = 0.0,
    val episodeCount: Int = 0,
    val totalEpisodes: Int = 0,
    val episodes: List<Episode> = emptyList(),
    val counts: DetailCounts = DetailCounts(),
    val author: DetailAuthor = DetailAuthor(),
    val interaction: Interaction? = null,
    val videoUrl: String = "",
    val videoUrlHevc: String = "",
    val isFinished: Boolean = false,
    val isOriginal: Int = 0,
    val videoType: Int = 0,
    val duration: Int = 0,
    val createdAt: String = "",
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("id", id)
        put("title", title)
        put("description", description)
        put("cover", cover)
        put("tags", JSONArray(tags))
        put("tag_links", JSONArray(tagLinks.map { it.toJson() }))
        put("score", score)
        put("episode_count", episodeCount)
        put("total_episodes", totalEpisodes)
        put("episodes", JSONArray(episodes.map { it.toJson() }))
        put("counts", counts.toJson())
        put("author", author.toJson())
        put("interaction", interaction?.toJson())
        put("video_url", videoUrl)
        put("video_url_hevc", videoUrlHevc)
        put("is_finished", isFinished)
        put("is_original", isOriginal)
        put("video_type", videoType)
        put("duration", duration)
        put("created_at", createdAt)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): Detail = Detail(
            id = o.opt("id").asLong(),
            title = o.opt("title").asString(),
            description = o.opt("description").asString(),
            cover = o.opt("cover").asString(),
            tags = o.opt("tags").asList { it.asString() },
            tagLinks = o.opt("tag_links").asList { TagLink.fromJson(it.asObject()) },
            score = o.opt("score").asDouble(),
            episodeCount = o.opt("episode_count").asInt(),
            totalEpisodes = o.opt("total_episodes").asInt(),
            episodes = o.opt("episodes").asList { Episode.fromJson(it.asObject()) },
            counts = DetailCounts.fromJson(o.opt("counts").asObject()),
            author = DetailAuthor.fromJson(o.opt("author").asObject()),
            interaction = (o.opt("interaction") as? JSONObject)?.let(Interaction::fromJson),
            videoUrl = o.opt("video_url").asString(),
            videoUrlHevc = o.opt("video_url_hevc").asString(),
            isFinished = o.opt("is_finished").asBoolean(),
            isOriginal = o.opt("is_original").asInt(),
            videoType = o.opt("video_type").asInt(),
            duration = o.opt("duration").asInt(),
            createdAt = o.opt("created_at").asString(),
        )
    }
}

/** DetailAuthor 是详情里的作者（TS 的 Author & {is_following?...}），单独具名。 */
data class DetailAuthor(
    val id: Long = 0L,
    val name: String = "",
    val avatar: String = "",
    val isFollowing: Boolean = false,
    val following: Int = 0,
    val likeCount: Int = 0,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("id", id)
        put("name", name)
        put("avatar", avatar)
        put("is_following", isFollowing)
        put("following", following)
        put("like_count", likeCount)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): DetailAuthor = DetailAuthor(
            id = o.opt("id").asLong(),
            name = o.opt("name").asString(),
            avatar = o.opt("avatar").asString(),
            isFollowing = o.opt("is_following").asBoolean(),
            following = o.opt("following").asInt(),
            likeCount = o.opt("like_count").asInt(),
        )
    }
}

data class DetailCounts(
    val collect: Long = 0L,
    val comment: Long = 0L,
    val hot: Long = 0L,
    val like: Long = 0L,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("collect", collect)
        put("comment", comment)
        put("hot", hot)
        put("like", like)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): DetailCounts = DetailCounts(
            collect = o.opt("collect").asLong(),
            comment = o.opt("comment").asLong(),
            hot = o.opt("hot").asLong(),
            like = o.opt("like").asLong(),
        )
    }
}

data class Episode(
    val epNum: Int = 0,
    val title: String = "",
    val duration: Int = 0,
    val status: Int = 0,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("ep_num", epNum)
        put("title", title)
        put("duration", duration)
        put("status", status)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): Episode = Episode(
            epNum = o.opt("ep_num").asInt(),
            title = o.opt("title").asString(),
            duration = o.opt("duration").asInt(),
            status = o.opt("status").asInt(),
        )
    }
}

data class Interaction(
    val collected: Boolean = false,
    val liked: Boolean = false,
    val likedEp: Int = 0,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("collected", collected)
        put("liked", liked)
        put("liked_ep", likedEp)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): Interaction = Interaction(
            collected = o.opt("collected").asBoolean(),
            liked = o.opt("liked").asBoolean(),
            likedEp = o.opt("liked_ep").asInt(),
        )
    }
}

data class LikeVideoEpisode(
    val videoId: Long = 0L,
    val ep: Int = 0,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("video_id", videoId)
        put("ep", ep)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): LikeVideoEpisode = LikeVideoEpisode(
            videoId = o.opt("video_id").asLong(),
            ep = o.opt("ep").asInt(),
        )
    }
}

data class Pagination(
    val page: Int = 0,
    val pages: Int = 0,
    val size: Int = 0,
    val total: Int = 0,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("page", page)
        put("pages", pages)
        put("size", size)
        put("total", total)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): Pagination = Pagination(
            page = o.opt("page").asInt(),
            pages = o.opt("pages").asInt(),
            size = o.opt("size").asInt(),
            total = o.opt("total").asInt(),
        )
    }
}

data class PlayInfo(
    val id: Long = 0L,
    val epNum: Int = 0,
    val duration: Int = 0,
    val title: String = "",
    val videoUrl: String = "",
    val videoUrlHevc: String = "",
    val streams: Streams? = null,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("id", id)
        put("ep_num", epNum)
        put("duration", duration)
        put("title", title)
        put("video_url", videoUrl)
        put("video_url_hevc", videoUrlHevc)
        put("streams", streams?.toJson())
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): PlayInfo = PlayInfo(
            id = o.opt("id").asLong(),
            epNum = o.opt("ep_num").asInt(),
            duration = o.opt("duration").asInt(),
            title = o.opt("title").asString(),
            videoUrl = o.opt("video_url").asString(),
            videoUrlHevc = o.opt("video_url_hevc").asString(),
            streams = (o.opt("streams") as? JSONObject)?.let(Streams::fromJson),
        )
    }
}

data class Post(
    val id: Long = 0L,
    val title: String = "",
    val desc: String = "",
    val content: String = "",
    val cover: String = "",
    val tags: List<String> = emptyList(),
    val hot: Long = 0L,
    val videoUrl: String = "",
    val videoId: Long = 0L,
    val author: PostAuthor? = null,
    val counts: PostCounts? = null,
    val createdAt: String = "",
    val publishTime: String = "",
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("id", id)
        put("title", title)
        put("desc", desc)
        put("content", content)
        put("cover", cover)
        put("tags", JSONArray(tags))
        put("hot", hot)
        put("video_url", videoUrl)
        put("video_id", videoId)
        put("author", author?.toJson())
        put("counts", counts?.toJson())
        put("created_at", createdAt)
        put("publish_time", publishTime)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): Post = Post(
            id = o.opt("id").asLong(),
            title = o.opt("title").asString(),
            desc = o.opt("desc").asString(),
            content = o.opt("content").asString(),
            cover = o.opt("cover").asString(),
            tags = o.opt("tags").asList { it.asString() },
            hot = o.opt("hot").asLong(),
            videoUrl = o.opt("video_url").asString(),
            videoId = o.opt("video_id").asLong(),
            author = (o.opt("author") as? JSONObject)?.let(PostAuthor::fromJson),
            counts = (o.opt("counts") as? JSONObject)?.let(PostCounts::fromJson),
            createdAt = o.opt("created_at").asString(),
            publishTime = o.opt("publish_time").asString(),
        )
    }
}

data class PostAuthor(
    val id: Long = 0L,
    val name: String = "",
    val avatar: String = "",
    val isFollowing: Boolean = false,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("id", id)
        put("name", name)
        put("avatar", avatar)
        put("is_following", isFollowing)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): PostAuthor = PostAuthor(
            id = o.opt("id").asLong(),
            name = o.opt("name").asString(),
            avatar = o.opt("avatar").asString(),
            isFollowing = o.opt("is_following").asBoolean(),
        )
    }
}

data class PostCounts(
    val like: Long = 0L,
    val comment: Long = 0L,
    val collect: Long = 0L,
    val share: Long = 0L,
    val hot: Long = 0L,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("like", like)
        put("comment", comment)
        put("collect", collect)
        put("share", share)
        put("hot", hot)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): PostCounts = PostCounts(
            like = o.opt("like").asLong(),
            comment = o.opt("comment").asLong(),
            collect = o.opt("collect").asLong(),
            share = o.opt("share").asLong(),
            hot = o.opt("hot").asLong(),
        )
    }
}

data class PostPage(
    val items: List<Post> = emptyList(),
    val pagination: Pagination = Pagination(),
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("items", JSONArray(items.map { it.toJson() }))
        put("pagination", pagination.toJson())
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): PostPage = PostPage(
            items = o.opt("items").asList { Post.fromJson(it.asObject()) },
            pagination = Pagination.fromJson(o.opt("pagination").asObject()),
        )
    }
}

data class RankItem(
    val rank: Long = 0L,
    val videoId: Long = 0L,
    val title: String = "",
    val cover: String = "",
    val tags: List<String> = emptyList(),
    val score: Double = 0.0,
    val cateName: String = "",
    val episodeCount: Int = 0,
    val isFinished: Boolean = false,
    val metricLabel: String = "",
    val metricValue: Long = 0L,
    val metrics: RankMetrics = RankMetrics(),
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("rank", rank)
        put("video_id", videoId)
        put("title", title)
        put("cover", cover)
        put("tags", JSONArray(tags))
        put("score", score)
        put("cate_name", cateName)
        put("episode_count", episodeCount)
        put("is_finished", isFinished)
        put("metric_label", metricLabel)
        put("metric_value", metricValue)
        put("metrics", metrics.toJson())
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): RankItem = RankItem(
            rank = o.opt("rank").asLong(),
            videoId = o.opt("video_id").asLong(),
            title = o.opt("title").asString(),
            cover = o.opt("cover").asString(),
            tags = o.opt("tags").asList { it.asString() },
            score = o.opt("score").asDouble(),
            cateName = o.opt("cate_name").asString(),
            episodeCount = o.opt("episode_count").asInt(),
            isFinished = o.opt("is_finished").asBoolean(),
            metricLabel = o.opt("metric_label").asString(),
            metricValue = o.opt("metric_value").asLong(),
            metrics = RankMetrics.fromJson(o.opt("metrics").asObject()),
        )
    }
}

data class RankMetrics(
    val collect: Long = 0L,
    val comment: Long = 0L,
    val like: Long = 0L,
    val play: Long = 0L,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("collect", collect)
        put("comment", comment)
        put("like", like)
        put("play", play)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): RankMetrics = RankMetrics(
            collect = o.opt("collect").asLong(),
            comment = o.opt("comment").asLong(),
            like = o.opt("like").asLong(),
            play = o.opt("play").asLong(),
        )
    }
}

/** RankPage 是 /api/ranks/{kind} 的返回（对照 src/api.ts 的 {type, items}）。 */
data class RankPage(
    val type: String = "",
    val items: List<RankItem> = emptyList(),
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("type", type)
        put("items", JSONArray(items.map { it.toJson() }))
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): RankPage = RankPage(
            type = o.opt("type").asString(),
            items = o.opt("items").asList { RankItem.fromJson(it.asObject()) },
        )
    }
}

data class SearchResult(
    val items: List<VideoCard> = emptyList(),
    val pagination: Pagination = Pagination(),
    val keyword: String = "",
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("items", JSONArray(items.map { it.toJson() }))
        put("pagination", pagination.toJson())
        put("keyword", keyword)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): SearchResult = SearchResult(
            items = o.opt("items").asList { VideoCard.fromJson(it.asObject()) },
            pagination = Pagination.fromJson(o.opt("pagination").asObject()),
            keyword = o.opt("keyword").asString(),
        )
    }
}

data class Streams(
    val main: String = "",
    val hevc: String = "",
    val preview: String = "",
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("main", main)
        put("hevc", hevc)
        put("preview", preview)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): Streams = Streams(
            main = o.opt("main").asString(),
            hevc = o.opt("hevc").asString(),
            preview = o.opt("preview").asString(),
        )
    }
}

data class TagLink(
    val name: String = "",
    val url: String = "",
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("name", name)
        put("url", url)
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): TagLink = TagLink(
            name = o.opt("name").asString(),
            url = o.opt("url").asString(),
        )
    }
}

data class User(
    val id: Long = 0L,
    val username: String = "",
    val nickname: String = "",
    val avatar: String = "",
    val avatarColor: String = "",
    val sign: String = "",
    val desc: String = "",
    val userType: String = "",
    val counts: Map<String, Long> = emptyMap(),
    val interactions: UserInteractions? = null,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("id", id)
        put("username", username)
        put("nickname", nickname)
        put("avatar", avatar)
        put("avatar_color", avatarColor)
        put("sign", sign)
        put("desc", desc)
        put("user_type", userType)
        put("counts", JSONObject(counts))
        put("interactions", interactions?.toJson())
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): User = User(
            id = o.opt("id").asLong(),
            username = o.opt("username").asString(),
            nickname = o.opt("nickname").asString(),
            avatar = o.opt("avatar").asString(),
            avatarColor = o.opt("avatar_color").asString(),
            sign = o.opt("sign").asString(),
            desc = o.opt("desc").asString(),
            userType = o.opt("user_type").asString(),
            counts = o.opt("counts").asMap { it.asLong() },
            interactions = (o.opt("interactions") as? JSONObject)?.let(UserInteractions::fromJson),
        )
    }
}

data class UserInteractions(
    val likeVideoIds: List<Long> = emptyList(),
    val favoriteVideoIds: List<Long> = emptyList(),
    val followUserIds: List<Long> = emptyList(),
    val likeVideoEpisodes: List<LikeVideoEpisode> = emptyList(),
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("like_video_ids", JSONArray(likeVideoIds))
        put("favorite_video_ids", JSONArray(favoriteVideoIds))
        put("follow_user_ids", JSONArray(followUserIds))
        put("like_video_episodes", JSONArray(likeVideoEpisodes.map { it.toJson() }))
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): UserInteractions = UserInteractions(
            likeVideoIds = o.opt("like_video_ids").asList { it.asLong() },
            favoriteVideoIds = o.opt("favorite_video_ids").asList { it.asLong() },
            followUserIds = o.opt("follow_user_ids").asList { it.asLong() },
            likeVideoEpisodes = o.opt("like_video_episodes").asList { LikeVideoEpisode.fromJson(it.asObject()) },
        )
    }
}

data class VideoCard(
    val id: Long = 0L,
    val title: String = "",
    val cover: String = "",
    val description: String = "",
    val tags: List<String> = emptyList(),
    val score: Double = 0.0,
    val episodeCount: Int = 0,
    val totalEpisodes: Int = 0,
    val duration: String = "",
    val hot: Long = 0L,
    val likeCount: Long = 0L,
    val collectCount: Long = 0L,
    val commentCount: Long = 0L,
    val isFinished: Boolean = false,
    val isOriginal: Int = 0,
    val createdAt: String = "",
    val lastEpAt: Long? = null,
    val newAt: Long? = null,
    val author: Author? = null,
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("id", id)
        put("title", title)
        put("cover", cover)
        put("description", description)
        put("tags", JSONArray(tags))
        put("score", score)
        put("episode_count", episodeCount)
        put("total_episodes", totalEpisodes)
        put("duration", duration)
        put("hot", hot)
        put("like_count", likeCount)
        put("collect_count", collectCount)
        put("comment_count", commentCount)
        put("is_finished", isFinished)
        put("is_original", isOriginal)
        put("created_at", createdAt)
        put("last_ep_at", lastEpAt)
        put("new_at", newAt)
        put("author", author?.toJson())
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): VideoCard = VideoCard(
            id = o.opt("id").asLong(),
            title = o.opt("title").asString(),
            cover = o.opt("cover").asString(),
            description = o.opt("description").asString(),
            tags = o.opt("tags").asList { it.asString() },
            score = o.opt("score").asDouble(),
            episodeCount = o.opt("episode_count").asInt(),
            totalEpisodes = o.opt("total_episodes").asInt(),
            duration = o.opt("duration").asString(),
            hot = o.opt("hot").asLong(),
            likeCount = o.opt("like_count").asLong(),
            collectCount = o.opt("collect_count").asLong(),
            commentCount = o.opt("comment_count").asLong(),
            isFinished = o.opt("is_finished").asBoolean(),
            isOriginal = o.opt("is_original").asInt(),
            createdAt = o.opt("created_at").asString(),
            lastEpAt = o.opt("last_ep_at").asLong(),
            newAt = o.opt("new_at").asLong(),
            author = (o.opt("author") as? JSONObject)?.let(Author::fromJson),
        )
    }
}

/** 分页具名实例（避免泛型，供 schema AST 解析）。 */
data class VideoCardPage(
    val items: List<VideoCard> = emptyList(),
    val pagination: Pagination = Pagination(),
) {
    /** 编码成与 Go SDK 相同的 JSON。 */
    fun toJson(): JSONObject = JSONObject().apply {
        put("items", JSONArray(items.map { it.toJson() }))
        put("pagination", pagination.toJson())
    }

    companion object {
        /** 从 Go SDK 输出的 JSON 解码；缺省的字段取零值。 */
        fun fromJson(o: JSONObject): VideoCardPage = VideoCardPage(
            items = o.opt("items").asList { VideoCard.fromJson(it.asObject()) },
            pagination = Pagination.fromJson(o.opt("pagination").asObject()),
        )
    }
}
