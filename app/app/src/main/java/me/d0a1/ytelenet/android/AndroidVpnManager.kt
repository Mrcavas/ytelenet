package me.d0a1.ytelenet.android

import android.content.Context
import android.content.Intent
import android.util.Log
import me.d0a1.ytelenet.ConnectionState
import me.d0a1.ytelenet.LogEntry
import me.d0a1.ytelenet.VpnManager

class AndroidVpnManager(
	private val context: Context
) : VpnManager() {
	override fun logNatively(entry: LogEntry) {
		Log.println(
			when (entry.level) {
				"error" -> Log.ERROR
				"warn" -> Log.WARN
				"info" -> Log.INFO
				else -> Log.DEBUG
			}, "YTelenet", "[${entry.time}]: ${entry.msg}"
		)
	}

	override fun connect(token: String) {
		log("debug", "Launching VPN service")
		updateConnState(ConnectionState.Loading)

		context.startForegroundService(
			Intent(context, YTVpnService::class.java)
				.putExtra("token", token)
		)
	}

	override fun disconnect() {
		log("debug", "Sending DISCONNECT command to VPN service")

		context.startService(Intent(context, YTVpnService::class.java).apply {
			action = "DISCONNECT"
		})
	}
}