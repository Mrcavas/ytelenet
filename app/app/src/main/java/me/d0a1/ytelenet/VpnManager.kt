package me.d0a1.ytelenet

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.getAndUpdate
import kotlinx.coroutines.flow.update
import kotlin.time.Clock
import kotlin.time.Instant

enum class ConnectionState {
	Connected, Loading, Disconnected
}

data class LogEntry(val key: Int, var level: String, var time: String, var msg: String)

abstract class VpnManager {
	private val _connState = MutableStateFlow(ConnectionState.Disconnected)
	val connState = _connState.asStateFlow()

	private val logIdx = MutableStateFlow(0)
	private val _logEntries = MutableStateFlow<List<LogEntry>>(emptyList())
	val logEntries = _logEntries.asStateFlow()

	fun updateConnState(state: ConnectionState) = _connState.update { state }

	fun log(entry: vpn.Log) {
		val entryObj = LogEntry(
			logIdx.getAndUpdate { it + 1 },
			entry.level,
			entry.time.replace("Z", "").split("T").getOrNull(1) ?: "",
			entry.msg.trim()
		)

		logNatively(entryObj)
		_logEntries.update {
			if (it.size < 256) it + entryObj else it.drop(1) + entryObj
		}
	}

	fun log(level: String, msg: String) =
		log(vpn.Log().apply {
			this.level = level
			this.time = Instant.fromEpochSeconds(Clock.System.now().epochSeconds).toString()
			this.msg = msg
		})

	abstract fun logNatively(entry: LogEntry)
	abstract fun connect(token: String)
	abstract fun disconnect()
}