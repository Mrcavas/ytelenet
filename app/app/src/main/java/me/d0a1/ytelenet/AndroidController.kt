package me.d0a1.ytelenet

import android.content.Context
import android.content.Intent
import android.util.Log
import vpn.Log as LogEntry

class AndroidController(private val context: Context) {
	fun log(entry: LogEntry) = Log.println(
		when (entry.level) {
			"error" -> Log.ERROR
			"warn" -> Log.WARN
			"info" -> Log.INFO
			else -> Log.DEBUG
		}, "YTelenet", "[${entry.time}]: ${entry.msg}"
	)

	fun connect(token: String) {
		val intent = Intent(context, YTVpnService::class.java)
			.putExtra("token", token)
		context.startForegroundService(intent)
	}

	fun disconnect() {
		context.stopService(Intent(context, YTVpnService::class.java))
	}
}