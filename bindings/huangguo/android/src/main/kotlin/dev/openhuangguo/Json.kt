package dev.openhuangguo

import org.json.JSONArray
import org.json.JSONObject

// Types.kt 解码用的辅助函数：值缺省、为 null 或类型不符时取零值。

internal fun Any?.asString(): String = this as? String ?: ""

internal fun Any?.asBoolean(): Boolean = this as? Boolean ?: false

internal fun Any?.asInt(): Int = (this as? Number)?.toInt() ?: 0

internal fun Any?.asLong(): Long = (this as? Number)?.toLong() ?: 0L

internal fun Any?.asDouble(): Double = (this as? Number)?.toDouble() ?: 0.0

internal fun Any?.asObject(): JSONObject = this as? JSONObject ?: JSONObject()

internal inline fun <T> Any?.asList(f: (Any?) -> T): List<T> {
    val a = this as? JSONArray ?: return emptyList()
    return List(a.length()) { f(a.opt(it)) }
}

internal inline fun <T> Any?.asMap(f: (Any?) -> T): Map<String, T> {
    val o = this as? JSONObject ?: return emptyMap()
    return o.keys().asSequence().associateWith { f(o.opt(it)) }
}

/** 把 Core JSON 的 result 值解码成 T：基本类型直取，其余走生成的 fromJson。 */
internal inline fun <reified T> decodeResult(value: Any?): T {
    when (T::class) {
        String::class -> return (value as? String ?: "") as T
        Int::class -> return ((value as? Number)?.toInt() ?: 0) as T
        Long::class -> return ((value as? Number)?.toLong() ?: 0L) as T
        Boolean::class -> return (value as? Boolean ?: false) as T
        Double::class -> return ((value as? Number)?.toDouble() ?: 0.0) as T
        SegmentResult::class -> {
            val o = value as? JSONObject ?: JSONObject()
            return SegmentResult(
                status = o.optInt("status"),
                headers = o.optJSONObject("headers")?.let { h ->
                    h.keys().asSequence().associateWith { k -> h.optString(k) }
                } ?: emptyMap(),
                body = o.optString("body"),
                url = o.optString("url"),
            ) as T
        }
        AuthResult::class -> {
            val o = value as? JSONObject ?: JSONObject()
            return AuthResult(user = o.opt("user")?.toString() ?: "") as T
        }
    }
    val o = value as? JSONObject ?: JSONObject()
    // Types.kt 把 fromJson 生成为 companion object 的实例方法（无 @JvmStatic），
    // 需先从类的 Companion 静态字段取出实例，再调用其方法。
    val companion = T::class.java.getDeclaredField("Companion").apply { isAccessible = true }.get(null)
    val m = companion.javaClass.getMethod("fromJson", JSONObject::class.java)
    @Suppress("UNCHECKED_CAST")
    return m.invoke(companion, o) as T
}
