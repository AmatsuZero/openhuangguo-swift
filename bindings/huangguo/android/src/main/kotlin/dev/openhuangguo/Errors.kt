package dev.openhuangguo

import org.json.JSONObject

/** 错误码，对应 Go dispatch 的错误码。 */
enum class ErrorCode {
    CANCELLED,
    AUTHENTICATION_REQUIRED,
    API_ERROR,
    INVALID_BASE,
    RETRYABLE,
    TRANSPORT,
    UNKNOWN,
    ;

    internal companion object {
        fun of(name: String): ErrorCode = entries.firstOrNull { it.name == name } ?: UNKNOWN
    }
}

/** HuangguoException 是 SDK 返回的错误。 */
class HuangguoException(
    val code: ErrorCode,
    message: String,
    val retryable: Boolean = false,
    val retryAfterMs: Long = 0,
) : Exception(message) {
    internal companion object {
        fun fromJson(o: JSONObject) = HuangguoException(
            code = ErrorCode.of(o.optString("code")),
            message = o.optString("message"),
            retryable = o.optBoolean("retryable"),
            retryAfterMs = o.optLong("retryAfterMs"),
        )

        fun parse(message: String?): HuangguoException =
            try {
                fromJson(JSONObject(message ?: ""))
            } catch (e: Exception) {
                HuangguoException(ErrorCode.UNKNOWN, message ?: "unknown error")
            }
    }
}
