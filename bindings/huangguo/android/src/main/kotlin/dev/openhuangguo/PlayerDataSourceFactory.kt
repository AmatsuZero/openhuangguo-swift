package dev.openhuangguo

import android.net.Uri
import androidx.media3.common.util.UnstableApi
import androidx.media3.datasource.BaseDataSource
import androidx.media3.datasource.DataSource
import androidx.media3.datasource.DataSpec
import kotlinx.coroutines.runBlocking
import org.json.JSONObject

/**
 * Media3 DataSource.Factory：把 Core 的纯函数媒体工具接到播放器。
 *
 * RewritePlaylist 把 KEY / 分片 / 子清单改写成 `hg://localhost/seg?u=<pct>`；
 * 本工厂识别该 scheme，还原上游 URL，经 Core 的 fetchSegment / rewritePlaylist
 * 取回资源并把 Range / 206 / Content-Range 回传给播放器。其他 scheme 回退到默认工厂。
 */
@UnstableApi
class PlayerDataSourceFactory(
    private val client: OpenHuangguoClient,
    private val upstream: DataSource.Factory,
    private val scheme: String = "hg",
    private val host: String = "localhost",
) : DataSource.Factory {

    override fun createDataSource(): DataSource = HuangguoDataSource(client, upstream.createDataSource(), scheme, host)
}

@UnstableApi
private class HuangguoDataSource(
    private val client: OpenHuangguoClient,
    private val fallback: DataSource,
    private val scheme: String,
    private val host: String,
) : BaseDataSource(false) {

    private var result: SegmentResult? = null
    private var position = 0
    private var opened = false

    override fun open(dataSpec: DataSpec): Long {
        val uri = dataSpec.uri
        if (!isLocal(uri)) return fallback.open(dataSpec)

        val upstreamUrl = upstreamURL(uri) ?: throw IllegalArgumentException("invalid local media URL: $uri")
        val headers = mutableMapOf<String, String>()
        if (dataSpec.length != androidx.media3.common.C.LENGTH_UNSET) {
            headers["Range"] = "bytes=${dataSpec.position}-${dataSpec.position + dataSpec.length - 1}"
        }
        result = runBlocking { client.fetchSegment(upstreamUrl, headers) }
        val r = result!!
        if (looksLikePlaylist(upstreamUrl, r.headers, r.body)) {
            val bytes = android.util.Base64.decode(r.body, android.util.Base64.DEFAULT)
            val text = String(bytes, Charsets.UTF_8)
            val rewritten = runBlocking { client.rewritePlaylist(upstreamUrl, scheme, host, text) }
            result = r.copy(body = android.util.Base64.encodeToString(rewritten.toByteArray(), android.util.Base64.DEFAULT))
        }
        position = dataSpec.position.toInt()
        opened = true
        transferInitializing(dataSpec)
        transferStarted(dataSpec)
        val len = result!!.headers["Content-Length"]?.toLongOrNull() ?: androidx.media3.common.C.LENGTH_UNSET
        return if (len != androidx.media3.common.C.LENGTH_UNSET) len - dataSpec.position else len
    }

    override fun read(target: ByteArray, offset: Int, length: Int): Int {
        if (!opened) return androidx.media3.common.C.RESULT_END_OF_INPUT
        val r = result ?: return androidx.media3.common.C.RESULT_END_OF_INPUT
        val bytes = android.util.Base64.decode(r.body, android.util.Base64.DEFAULT)
        if (position >= bytes.size) return androidx.media3.common.C.RESULT_END_OF_INPUT
        val n = minOf(length, bytes.size - position)
        System.arraycopy(bytes, position, target, offset, n)
        position += n
        bytesTransferred(n)
        return n
    }

    override fun getUri(): Uri? = result?.let { Uri.parse(it.headers["Content-Range"] ?: "") } ?: fallback.uri

    override fun close() {
        if (!opened) { fallback.close(); return }
        opened = false
        transferEnded()
        result = null
    }

    private fun isLocal(uri: Uri) = uri.scheme?.equals(scheme, ignoreCase = true) == true &&
        uri.host?.equals(host, ignoreCase = true) == true

    private fun upstreamURL(uri: Uri): String? = uri.getQueryParameter("u")

    private fun looksLikePlaylist(url: String, headers: Map<String, String>, bodyB64: String): Boolean {
        if (url.substringBefore('?').endsWith(".m3u8", ignoreCase = true)) return true
        if (headers["Content-Type"]?.contains("mpegurl", ignoreCase = true) == true) return true
        val bytes = android.util.Base64.decode(bodyB64, android.util.Base64.DEFAULT)
        val prefix = "#EXTM3U".toByteArray()
        return bytes.size >= prefix.size && bytes.copyOfRange(0, prefix.size).contentEquals(prefix)
    }
}
