package dev.openhuangguo

import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.CoroutineStart
import kotlinx.coroutines.async
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withTimeout
import org.json.JSONArray
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Test
import java.io.IOException

/** 用假 Engine / TokenBridge 代替 Go core 库，检查参数编码、结果解码、错误映射、取消与 token 持久化。 */
class ClientTest {
    private class FakeEngine(
        val reply: (method: String, args: JSONArray) -> String,
        var onCancel: (() -> Unit)? = null,
    ) : Engine {
        val calls = mutableListOf<Pair<String, String>>()
        override fun start(method: String, args: String, done: (String) -> Unit): () -> Unit {
            calls += method to args
            val out = reply(method, JSONArray(args))
            if (out.isNotEmpty()) done(out) // 空串表示"不立即回调"，用于测试取消
            return { onCancel?.invoke() }
        }
    }

    private class FakeBridge(var token: String = """{"origin":"https://huangguoai.com","token":"","generation":0}""") : TokenBridge {
        override fun setTokenJSON(json: String) { token = json }
        override fun getTokenJSON(): String? = token
    }

    private class MemoryStore : TokenStore {
        var saved: String? = null
        override fun load() = saved
        override fun save(json: String) { saved = json }
        override fun clear() { saved = null }
    }

    private fun ok(result: Any?) = JSONObject().put("result", result ?: JSONObject.NULL).toString()

    @Test
    fun feedEncodesArgumentsAndDecodesPage() = runBlocking {
        val engine = FakeEngine(reply = { method, args ->
            assertEquals("feed", method)
            assertEquals("[1,20]", args.toString())
            ok(JSONObject().put("items", JSONArray()).put("pagination", JSONObject().put("total", 100)))
        })
        val page = OpenHuangguoClient(engine, null, null).feed(1, 20)
        assertEquals(100, page.pagination.total)
    }

    @Test
    fun errorsBecomeHuangguoException() = runBlocking {
        val engine = FakeEngine(reply = { _, _ ->
            """{"error":{"code":"AUTHENTICATION_REQUIRED","message":"login","retryable":false,"retryAfterMs":0}}"""
        })
        try {
            OpenHuangguoClient(engine, null, null).me()
            fail("no exception")
        } catch (e: HuangguoException) {
            assertEquals(ErrorCode.AUTHENTICATION_REQUIRED, e.code)
            assertEquals("login", e.message)
        }
    }

    @Test
    fun tokenPersistenceOnlyOnSuccess() = runBlocking {
        val store = MemoryStore()
        val bridge = FakeBridge("""{"origin":"https://huangguoai.com","token":"tok","generation":1}""")
        val engine = FakeEngine(reply = { _, _ -> ok(JSONObject().put("user", JSONObject())) })
        val client = OpenHuangguoClient(engine, bridge, store)
        client.login("u", "p")
        assertTrue(store.saved!!.contains("token"))
    }

    @Test
    fun cancellingTheCoroutineCancelsTheCall() = runBlocking {
        val cancelled = CompletableDeferred<Unit>()
        // reply 返回空串 → FakeEngine 不调用 done，让协程挂起在 callRaw；
        // 这样 job.cancel() 才会触发 invokeOnCancellation → onCancel。
        val engine = FakeEngine(reply = { _, _ -> "" }, onCancel = { cancelled.complete(Unit) })
        val job = async(start = CoroutineStart.UNDISPATCHED) { OpenHuangguoClient(engine, null, null).rank("hot") }
        job.cancel()
        withTimeout(5_000) { cancelled.await() }
    }

    @Test
    fun segmentResultRetainsFinalResponseUrl() = runBlocking {
        val engine = FakeEngine(reply = { method, _ ->
            assertEquals("fetchSegment", method)
            ok(JSONObject()
                .put("status", 206)
                .put("headers", JSONObject().put("Content-Range", "bytes 10-12/100"))
                .put("body", "AQID")
                .put("url", "https://cdn.example/final.ts"))
        })

        val result = OpenHuangguoClient(engine, null, null)
            .fetchSegment("https://cdn.example/original.ts", emptyMap())

        assertEquals("https://cdn.example/final.ts", result.url)
    }

    @Test
    fun playbackRequestPolicyScopesSignedHeadersAndBuildsOpenEndedRanges() {
        val signed = mapOf("X-Cdn-Auth" to "signed", "X-Cdn-App" to "player")
        val initial = mapOf("X-Cdn-Auth" to "stale", "User-Agent" to "client")

        assertEquals(
            mapOf("User-Agent" to "client", "X-Cdn-Auth" to "signed", "X-Cdn-App" to "player", "Range" to "bytes=10-"),
            PlaybackRequestPolicy.headers(
                rootUrl = "HTTPS://CDN.example:443/master.m3u8",
                requestUrl = "https://cdn.EXAMPLE/part.ts",
                requestHeaders = initial,
                signedHeaders = signed,
                position = 10,
                length = -1,
            ),
        )
        assertEquals(
            mapOf("User-Agent" to "client", "Range" to "bytes=10-14"),
            PlaybackRequestPolicy.headers(
                rootUrl = "https://cdn.example/master.m3u8",
                requestUrl = "https://other.example/part.ts",
                requestHeaders = initial,
                signedHeaders = signed,
                position = 10,
                length = 5,
            ),
        )
    }

    @Test
    fun playbackRequestPolicyRejectsUnsupportedManifestMethod() {
        try {
            PlaybackRequestPolicy.requireSupportedManifestMethod("POST")
            fail("expected unsupported manifest method")
        } catch (error: IOException) {
            assertTrue(error.message!!.contains("POST"))
        }
    }
}
