package me.d0a1.ytelenet

import kotlin.io.encoding.Base64

fun isTokenValid(token: String): Boolean {
	try {
		val parts = Base64.decode(token).decodeToString().split(";")
		if (parts.size != 3) return false
		if (parts[2].toIntOrNull() == null) return false
	} catch (_: Exception) {
		return false
	}
	return true
}