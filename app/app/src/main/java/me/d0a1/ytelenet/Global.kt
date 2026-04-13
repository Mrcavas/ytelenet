package me.d0a1.ytelenet

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.update
import kotlin.time.Clock
import vpn.Log as LogEntry

object Global {
	val connState = MutableStateFlow(ConnectionState.Disconnected)
	val logEntries = MutableStateFlow(listOf<LogEntry>())

	fun log(entry: LogEntry) {
		AndroidController.log(entry)
		logEntries.update { it + entry }
	}

	fun log(level: String, msg: String) {
		log(LogEntry().apply {
			this.level = level
			this.time = Clock.System.now().toString()
			this.msg = msg
		})
	}
}