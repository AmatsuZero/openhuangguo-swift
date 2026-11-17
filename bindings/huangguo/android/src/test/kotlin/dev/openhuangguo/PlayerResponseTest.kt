package dev.openhuangguo

import org.junit.Assert.assertArrayEquals
import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Test
import java.io.IOException

class PlayerResponseTest {
    @Test
    fun partialResponseDoesNotApplyAbsoluteOffsetTwice() {
        val bytes = PlaybackRequestPolicy.responseBytes(
            byteArrayOf(10, 11, 12), 206, mapOf("content-range" to "bytes 10-12/100"), 10, 3,
        )
        assertArrayEquals(byteArrayOf(10, 11, 12), bytes)
    }

    @Test
    fun ignoredRangeIsSlicedAndBoundedLocally() {
        val bytes = PlaybackRequestPolicy.responseBytes(byteArrayOf(0, 1, 2, 3, 4, 5), 200, emptyMap(), 2, 3)
        assertArrayEquals(byteArrayOf(2, 3, 4), bytes)
    }

    @Test
    fun malformedPartialResponseAndHttpErrorAreRejected() {
        assertThrows(IOException::class.java) {
            PlaybackRequestPolicy.responseBytes(byteArrayOf(10, 11, 12), 206, mapOf("Content-Range" to "bytes 0-2/100"), 10, 3)
        }
        assertThrows(IOException::class.java) {
            PlaybackRequestPolicy.responseBytes("denied".toByteArray(), 403, emptyMap(), 0, -1)
        }
    }

    @Test
    fun signedRangeCannotOverridePlayerRange() {
        val headers = PlaybackRequestPolicy.headers(
            "https://cdn.example/root", "https://cdn.example/key.bin", emptyMap(),
            mapOf("range" to "bytes=0-1", "X-Ticket" to "signed"), 10, 3,
        )
        assertEquals(mapOf("Range" to "bytes=10-12", "X-Ticket" to "signed"), headers)
    }
}
