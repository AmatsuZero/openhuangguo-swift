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

/** 用假 Engine / TokenBridge 代替 Go core 库，检查参数编码、结果解码、错误映射、取消与 token 持久化。 */
class ClientTest {
    private class FakeEngine(
        val reply: (method: String, args: JSONArray) -> String,
        var onCancel: (() -> Unit)? = null,
    ) : Engine {
        val calls = mutableListOf<Pair<String, String>>()
        override fun start(method: String, args: String, done: (String) -> Unit): () -> Unit {
            calls += method to args
            done(reply(method, JSONArray(args)))
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
        val engine = FakeEngine { method, args ->
            assertEquals("feed", method)
            assertEquals("[1,20]", args.toString())
            ok(JSONObject().put("items", JSONArray()).put("pagination", JSONObject().put("total", 100)))
        }
        val page = OpenHuangguoClient(engine, null, null).feed(1, 20)
        assertEquals(100, page.pagination.total)
    }

    @Test
    fun errorsBecomeHuangguoException() = runBlocking {
        val engine = FakeEngine { _, _ ->
            """{"error":{"code":"AUTHENTICATION_REQUIRED","message":"login","retryable":false,"retryAfterMs":0}}"""
        }
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
        val engine = FakeEngine { _, _ -> ok(JSONObject().put("user", JSONObject())) }
        val client = OpenHuangguoClient(engine, bridge, store)
        client.login("u", "p")
        assertTrue(store.saved!!.contains("token"))
    }

    @Test
    fun cancellingTheCoroutineCancelsTheCall() = runBlocking {
        val cancelled = CompletableDeferred<Unit>()
        val engine = FakeEngine({ _, _ -> ok(JSONObject()) }, onCancel = { cancelled.complete(Unit) })
        val job = async(start = CoroutineStart.UNDISPATCHED) { OpenHuangguoClient(engine, null, null).rank("hot") }
        job.cancel()
        withTimeout(5_000) { cancelled.await() }
    }
}
