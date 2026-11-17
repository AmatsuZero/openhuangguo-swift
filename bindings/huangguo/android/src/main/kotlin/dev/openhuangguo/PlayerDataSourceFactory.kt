package dev.openhuangguo

import android.net.Uri
import androidx.media3.common.C
import androidx.media3.common.PlaybackException
import androidx.media3.datasource.DataSourceException
import androidx.media3.common.util.UnstableApi
import androidx.media3.datasource.BaseDataSource
import androidx.media3.datasource.DataSource
import androidx.media3.datasource.DataSpec
import kotlinx.coroutines.runBlocking
import java.io.IOException
import java.net.URI

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
    private val playback: PlayInfo? = null,
    private val playbackUrl: String? = null,
) : DataSource.Factory {

    /**
     * Maps the selected playback root into this factory's local resource scheme.
     * Child playlist, segment, and key URLs are emitted by Core rewritePlaylist.
     */
    fun playbackUri(): Uri? {
        PlaybackRequestPolicy.requireSupportedManifestMethod(playback?.manifestMethod.orEmpty())
        val url = playbackUrl ?: playback?.videoUrl
        if (url.isNullOrBlank()) return null
        return Uri.Builder()
            .scheme(scheme)
            .authority(host)
            .appendPath("seg")
            .appendQueryParameter("u", url)
            .build()
    }

    override fun createDataSource(): DataSource = HuangguoDataSource(
        client = client,
        fallback = upstream.createDataSource(),
        scheme = scheme,
        host = host,
        playback = playback,
        playbackUrl = playbackUrl,
    )
}

@UnstableApi
private class HuangguoDataSource(
    private val client: OpenHuangguoClient,
    private val fallback: DataSource,
    private val scheme: String,
    private val host: String,
    private val playback: PlayInfo?,
    private val playbackUrl: String?,
) : BaseDataSource(false) {

    private var bytes = ByteArray(0)
    private var position = 0
    private var opened = false
    private var usingFallback = false
    private var resolvedUri: Uri? = null
    private var responseHeaders: Map<String, List<String>> = emptyMap()

    override fun open(dataSpec: DataSpec): Long {
        val uri = dataSpec.uri
        if (!isLocal(uri)) {
            usingFallback = true
            return fallback.open(dataSpec)
        }

        val upstreamUrl = upstreamURL(uri) ?: throw IllegalArgumentException("invalid local media URL: $uri")
        transferInitializing(dataSpec)
        val rootUrl = playbackUrl ?: playback?.videoUrl.orEmpty()
        PlaybackRequestPolicy.requireSupportedManifestMethod(playback?.manifestMethod.orEmpty())
        val headers = PlaybackRequestPolicy.headers(
            rootUrl = rootUrl,
            requestUrl = upstreamUrl,
            requestHeaders = dataSpec.httpRequestHeaders,
            signedHeaders = playback?.mediaHeaders.orEmpty(),
            position = dataSpec.position,
            length = dataSpec.length,
        )
        // Rewriting requires a complete playlist, rather than an upstream byte range.
        val isManifestUrl = URI(upstreamUrl).path.orEmpty().endsWith(".m3u8", ignoreCase = true)
        val fetchHeaders = if (isManifestUrl) headers.filterKeys { !it.equals("Range", ignoreCase = true) } else headers
        var result = runBlocking { client.fetchSegment(upstreamUrl, fetchHeaders) }
        var finalUrl = result.url.ifBlank { upstreamUrl }
        var body = android.util.Base64.decode(result.body, android.util.Base64.DEFAULT)
        val isPlaylist = looksLikePlaylist(finalUrl, result.headers, body)
        if (isPlaylist) {
            if (result.status == 206) {
                result = runBlocking { client.fetchSegment(upstreamUrl, headers.filterKeys { !it.equals("Range", ignoreCase = true) }) }
                finalUrl = result.url.ifBlank { upstreamUrl }
                body = android.util.Base64.decode(result.body, android.util.Base64.DEFAULT)
            }
            if (result.status != 200) throw IOException("incomplete media playlist: HTTP ${result.status}")
            body = runBlocking { client.rewritePlaylist(finalUrl, scheme, host, String(body, Charsets.UTF_8)) }.toByteArray()
            bytes = PlaybackRequestPolicy.responseBytes(body, 200, emptyMap(), dataSpec.position, dataSpec.length)
        } else {
            bytes = PlaybackRequestPolicy.responseBytes(body, result.status, result.headers, dataSpec.position, dataSpec.length)
        }
        position = 0
        resolvedUri = Uri.parse(finalUrl)
        val mappedHeaders = java.util.TreeMap<String, List<String>>(String.CASE_INSENSITIVE_ORDER)
        result.headers.forEach { (key, value) -> mappedHeaders[key] = listOf(value) }
        if (isPlaylist) {
            mappedHeaders.remove("Content-Range")
            mappedHeaders["Content-Length"] = listOf(body.size.toString())
        }
        responseHeaders = mappedHeaders
        opened = true
        transferStarted(dataSpec)
        return if (dataSpec.length == C.LENGTH_UNSET.toLong()) bytes.size.toLong() else dataSpec.length
    }

    override fun read(target: ByteArray, offset: Int, length: Int): Int {
        if (usingFallback) return fallback.read(target, offset, length)
        if (!opened) return androidx.media3.common.C.RESULT_END_OF_INPUT
        if (length == 0) return 0
        if (position >= bytes.size) return androidx.media3.common.C.RESULT_END_OF_INPUT
        val n = minOf(length, bytes.size - position)
        System.arraycopy(bytes, position, target, offset, n)
        position += n
        bytesTransferred(n)
        return n
    }

    override fun getUri(): Uri? = if (usingFallback) fallback.uri else resolvedUri

    override fun getResponseHeaders(): Map<String, List<String>> = if (usingFallback) fallback.responseHeaders else responseHeaders

    override fun close() {
        if (usingFallback) {
            usingFallback = false
            fallback.close()
            return
        }
        if (opened) transferEnded()
        opened = false
        bytes = ByteArray(0)
        position = 0
        resolvedUri = null
        responseHeaders = emptyMap()
    }

    private fun isLocal(uri: Uri) = uri.scheme?.equals(scheme, ignoreCase = true) == true &&
        uri.host?.equals(host, ignoreCase = true) == true && uri.path == "/seg"

    private fun upstreamURL(uri: Uri): String? = uri.getQueryParameter("u")

    private fun looksLikePlaylist(url: String, headers: Map<String, String>, body: ByteArray): Boolean {
        if (URI(url).path.orEmpty().endsWith(".m3u8", ignoreCase = true)) return true
        if (headers.entries.any { it.key.equals("Content-Type", ignoreCase = true) && it.value.contains("mpegurl", ignoreCase = true) }) return true
        return body.take(7).toByteArray().contentEquals("#EXTM3U".toByteArray())
    }

}

/** Request policy shared by the local playlist root and all rewritten resources. */
internal object PlaybackRequestPolicy {
    fun requireSupportedManifestMethod(value: String) {
        val method = value.trim()
        if (method.isNotEmpty() && !method.equals("GET", ignoreCase = true)) {
            throw IOException("unsupported manifest method: $method")
        }
    }

    fun headers(
        rootUrl: String,
        requestUrl: String,
        requestHeaders: Map<String, String>,
        signedHeaders: Map<String, String>,
        position: Long,
        length: Long,
    ): Map<String, String> {
        val signedNames = signedHeaders.keys.map { it.lowercase() }.toSet()
        val sameOrigin = sameOrigin(rootUrl, requestUrl)
        val result = linkedMapOf<String, String>()
        for ((key, value) in requestHeaders) {
            val normalized = key.lowercase()
            if (normalized == "range" || normalized in signedNames) continue
            if (!sameOrigin && normalized !in setOf("user-agent", "accept", "accept-encoding", "if-range", "if-none-match", "if-modified-since")) continue
            result[key] = value
        }
        if (sameOrigin) result.putAll(signedHeaders.filterKeys { !it.equals("Range", ignoreCase = true) })
        if (position < 0 || length == 0L || length < C.LENGTH_UNSET.toLong() || (length > 0 && position > Long.MAX_VALUE - length + 1)) throw IOException("invalid media range")
        result["Range"] = if (length == C.LENGTH_UNSET.toLong()) {
            "bytes=$position-"
        } else {
            "bytes=$position-${position + length - 1}"
        }
        return result
    }

    @androidx.annotation.OptIn(UnstableApi::class)
    fun responseBytes(body: ByteArray, status: Int, headers: Map<String, String>, position: Long, length: Long): ByteArray {
        if (status != 200 && status != 206) throw IOException("media HTTP $status")
        var start = 0L
        var total = body.size.toLong()
        if (status == 206) {
            val value = headers.entries.firstOrNull { it.key.equals("Content-Range", ignoreCase = true) }?.value.orEmpty()
            val match = Regex("bytes (\\d+)-(\\d+)/(\\d+)").matchEntire(value) ?: throw IOException("invalid media Content-Range")
            start = match.groupValues[1].toLongOrNull() ?: throw IOException("invalid media Content-Range")
            val end = match.groupValues[2].toLongOrNull() ?: throw IOException("invalid media Content-Range")
            total = match.groupValues[3].toLongOrNull() ?: throw IOException("invalid media Content-Range")
            if (start > end || end >= total || end - start != body.size.toLong() - 1) throw IOException("invalid media Content-Range")
        }
        if (position < start || position > total || position - start > body.size) {
            throw DataSourceException(PlaybackException.ERROR_CODE_IO_READ_POSITION_OUT_OF_RANGE)
        }
        val from = (position - start).toInt()
        val available = body.size - from
        val count = if (length == C.LENGTH_UNSET.toLong()) available else minOf(available.toLong(), length).toInt()
        return body.copyOfRange(from, from + count)
    }

    private fun sameOrigin(first: String, second: String): Boolean {
        val a = parseOrigin(first) ?: return false
        val b = parseOrigin(second) ?: return false
        return a.scheme.equals(b.scheme, ignoreCase = true) &&
            a.host.equals(b.host, ignoreCase = true) &&
            normalizedPort(a) == normalizedPort(b)
    }

    private fun parseOrigin(value: String): URI? = try {
        URI(value).takeIf { !it.scheme.isNullOrBlank() && !it.host.isNullOrBlank() }
    } catch (_: Exception) {
        null
    }

    private fun normalizedPort(uri: URI): Int = when {
        uri.port >= 0 -> uri.port
        uri.scheme.equals("https", ignoreCase = true) -> 443
        uri.scheme.equals("http", ignoreCase = true) -> 80
        else -> -1
    }
}
