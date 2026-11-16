package dev.openhuangguo

import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException
import kotlinx.coroutines.suspendCancellableCoroutine
import org.json.JSONArray
import org.json.JSONObject

/** 原生 token 持久化接口；实现必须线程安全。 */
interface TokenStore {
    fun load(): String?
    fun save(json: String)
    fun clear()
}

/** 客户端设置。 */
data class ClientOptions(
    val base: String? = null,
    val proxy: String? = null,
    val credTrustOrigins: Map<String, Boolean>? = null,
    val tokenStore: TokenStore? = null,
    val logger: ((String) -> Unit)? = null,
) {
    internal fun toJson(): String = JSONObject().apply {
        base?.let { put("base", it) }
        proxy?.let { put("proxy", it) }
        credTrustOrigins?.let { m ->
            put("credTrustOrigins", JSONObject().apply { m.forEach { (k, v) -> put(k, v) } })
        }
    }.toString()
}

/** 登录/注册结果：token 已写入原生 store，仅回传 user（原始 JSON 字符串）。 */
data class AuthResult(val user: String = "")

/** 一次按名字的调用；done 收到结果 JSON，返回的函数用来取消。 */
internal fun interface Engine {
    fun start(method: String, args: String, done: (String) -> Unit): () -> Unit
}

/** token JSON 桥接：原生层读写 Keychain/Keystore 的入口。 */
internal interface TokenBridge {
    fun setTokenJSON(json: String)
    fun getTokenJSON(): String?
}

private class GoBridge(private val client: dev.openhuangguo.core.Client) : TokenBridge {
    override fun setTokenJSON(json: String) { client.setTokenJSON(json) }
    override fun getTokenJSON(): String? = client.getTokenJSON()
}

private class GoEngine(val client: dev.openhuangguo.core.Client) : Engine {
    override fun start(method: String, args: String, done: (String) -> Unit): () -> Unit {
        val call = client.start(method, args, dev.openhuangguo.core.Callback { done(it) })
        return { call.cancel() }
    }
}

/** 黄果客户端；所有方法为挂起函数，失败抛 HuangguoException。 */
class OpenHuangguoClient internal constructor(
    private val engine: Engine,
    private val bridge: TokenBridge?,
    private val tokenStore: TokenStore?,
) {
    companion object {
        @JvmStatic
        @JvmOverloads
        fun create(options: ClientOptions = ClientOptions()): OpenHuangguoClient {
            val logger = options.logger?.let { dev.openhuangguo.core.Logger { line -> it(line) } }
            // gomobile 把包级函数 NewClient 生成为 Core 类的静态方法 newClient（首字母小写）。
            val client = try {
                dev.openhuangguo.core.Core.newClient(options.toJson(), logger)
            } catch (e: Exception) {
                throw HuangguoException.parse(e.message)
            }
            val engine = GoEngine(client)
            val bridge = GoBridge(client)
            options.tokenStore?.load()?.let { bridge.setTokenJSON(it) }
            return OpenHuangguoClient(engine, bridge, options.tokenStore)
        }
    }

    // 账号
    suspend fun login(username: String, password: String): AuthResult = callPersist("login", username, password)
    suspend fun register(username: String, password: String, email: String = ""): AuthResult =
        callPersist("register", username, password, email)
    suspend fun logout() { callVoidPersist("logout") }
    suspend fun me(): User = call("me")

    // 列表 / 详情
    suspend fun feed(page: Int = 1, size: Int = 20): VideoCardPage = call("feed", page, size)
    suspend fun category(cate: String, page: Int = 1, size: Int = 20): VideoCardPage = call("category", cate, page, size)
    suspend fun rank(kind: String = "hot"): RankPage = call("rank", kind)
    suspend fun search(q: String, page: Int = 1, size: Int = 20): SearchResult = call("search", q, page, size)
    suspend fun posts(page: Int = 1, size: Int = 20): PostPage = call("posts", page, size)
    suspend fun post(id: Long): Post = call("post", id)
    suspend fun video(id: Long): Detail = call("video", id)
    suspend fun play(id: Long, ep: Int): PlayInfo = call("play", id, ep)

    // 评论 / 互动
    suspend fun comments(id: Long, ep: Int, page: Int = 1, size: Int = 20): CommentPage =
        call("comments", id, ep, page, size)
    suspend fun commentSend(id: Long, ep: Int, text: String, replyId: Long = 0): String =
        call("commentSend", id, ep, text, replyId)
    suspend fun like(id: Long, ep: Int, on: Boolean) { callVoid("like", id, ep, on) }
    suspend fun collect(id: Long, on: Boolean) { callVoid("collect", id, on) }
    suspend fun follow(uid: Long, on: Boolean) { callVoid("follow", uid, on) }
    suspend fun postLike(id: Long, on: Boolean) { callVoid("postLike", id, on) }
    suspend fun postCollect(id: Long, on: Boolean) { callVoid("postCollect", id, on) }
    suspend fun postComments(id: Long, page: Int = 1, size: Int = 20): CommentPage =
        call("postComments", id, page, size)

    /** 切线路：成功后清 token 并持久化。 */
    suspend fun setBase(base: String) { callVoidPersist("setBase", base) }

    // 媒体纯函数
    suspend fun decryptCover(data: ByteArray): ByteArray {
        val b64 = call<String>("decryptCover", android.util.Base64.encodeToString(data, android.util.Base64.DEFAULT))
        return android.util.Base64.decode(b64, android.util.Base64.DEFAULT)
    }
    suspend fun rewritePlaylist(upstream: String, scheme: String, host: String, text: String): String =
        call("rewritePlaylist", upstream, scheme, host, text)
    suspend fun fetchSegment(url: String, headers: Map<String, String>): SegmentResult =
        call("fetchSegment", url, JSONObject(headers))

    private suspend inline fun <reified T> call(method: String, vararg args: Any): T =
        decode(callRaw(method, args))

    private suspend fun callVoid(method: String, vararg args: Any) {
        callRaw(method, args)
    }

    private suspend inline fun <reified T> callPersist(method: String, vararg args: Any): T {
        val json = callRaw(method, args)
        val result = decode<T>(json)
        persistIfSuccess(json)
        return result
    }

    private suspend fun callVoidPersist(method: String, vararg args: Any) {
        val json = callRaw(method, args)
        persistIfSuccess(json)
    }

    private fun persistIfSuccess(json: String) {
        val o = JSONObject(json)
        if (o.has("result")) {
            val store = tokenStore ?: return
            val b = bridge ?: return
            store.save(b.getTokenJSON() ?: "")
        }
    }

    private suspend fun callRaw(method: String, args: Array<out Any>): String {
        val arr = JSONArray()
        args.forEach { arr.put(it) }
        return suspendCancellableCoroutine { cont ->
            val cancel = engine.start(method, arr.toString()) { json -> cont.resume(json) }
            cont.invokeOnCancellation { cancel() }
        }
    }

    private inline fun <reified T> decode(json: String): T {
        val o = JSONObject(json)
        if (o.has("error")) {
            val e = HuangguoException.fromJson(o.getJSONObject("error"))
            throw e
        }
        val result = o.opt("result") ?: throw HuangguoException(ErrorCode.UNKNOWN, "missing result")
        return decodeResult<T>(result)
    }
}

/** 媒体分片结果（body 为 base64 字符串）。 */
data class SegmentResult(
    val status: Int = 0,
    val headers: Map<String, String> = emptyMap(),
    val body: String = "",
)
