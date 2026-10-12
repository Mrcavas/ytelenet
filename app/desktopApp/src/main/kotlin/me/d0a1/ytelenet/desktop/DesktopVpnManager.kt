package me.d0a1.ytelenet.desktop

import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.currentCoroutineContext
import kotlinx.coroutines.ensureActive
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import me.d0a1.ytelenet.ConnectionState
import me.d0a1.ytelenet.LogEntry
import me.d0a1.ytelenet.VpnManager
import java.io.BufferedReader
import java.io.BufferedWriter
import java.net.InetAddress
import java.net.ServerSocket
import java.net.Socket
import java.net.SocketTimeoutException
import java.security.MessageDigest
import java.security.SecureRandom
import kotlin.time.Duration.Companion.minutes
import kotlin.time.TimeSource

class DesktopVpnManager : VpnManager() {
	private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
	private val json = Json { ignoreUnknownKeys = true }

	private var session: Job? = null

	@Volatile
	private var control: BufferedWriter? = null

	override fun logNatively(entry: LogEntry) {
		System.err.println("[${entry.time}] ${entry.level.uppercase()}: ${entry.msg}")
	}

	override fun connect(token: String) {
		if (session?.isActive == true) return

		log("debug", "Launching the VPN backend")
		updateConnState(ConnectionState.Loading)

		session = scope.launch {
			try {
				runBackend(token)
			} catch (e: CancellationException) {
				log("info", "Vpn stopped without errors")
				throw e
			} catch (e: Exception) {
				log("error", "Vpn stopped with error: ${e.message}")
			} finally {
				control = null
				updateConnState(ConnectionState.Disconnected)
			}
		}
	}

	override fun disconnect() {
		val control = control
		if (control == null) {
			// still waiting for the backend to start
			session?.cancel()
			return
		}

		log("debug", "Sending stop to the VPN backend")
		scope.launch {
			try {
				control.sendLine(BackendMessage(type = "stop"))
			} catch (e: Exception) {
				log("warn", "Failed to stop the backend: ${e.message}")
			}
		}
	}

	private suspend fun runBackend(token: String) {
		val key = ByteArray(32).also { SecureRandom().nextBytes(it) }.toHexString()

		ServerSocket(0, 1, InetAddress.getLoopbackAddress()).use { server ->
			val process = Backend.launch(
				Backend.findBinary(),
				listOf("gui-backend", "--connect", "127.0.0.1:${server.localPort}", "--key", key)
			)
			val output = StringBuffer()
			Thread { process.inputStream.bufferedReader().forEachLine { output.appendLine(it) } }
				.apply { isDaemon = true }
				.start()

			acceptBackend(server, process, key, output).use { talkToBackend(it, token) }
		}
	}

	/** Waits for the backend to connect, while the user deals with the password prompt */
	private suspend fun acceptBackend(
		server: ServerSocket, process: Process, key: String, output: StringBuffer
	): BackendConnection {
		val deadline = TimeSource.Monotonic.markNow() + 3.minutes
		server.soTimeout = 500

		while (true) {
			currentCoroutineContext().ensureActive()
			if (deadline.hasPassedNow()) throw BackendException("Timed out waiting for the backend")

			// On Windows the launcher exits right after starting the backend, on
			// linux pkexec becomes the backend itself
			if (!process.isAlive && (process.exitValue() != 0 || !Backend.launcherDetaches)) {
				Thread.sleep(100) // let the output reader catch up
				val details = output.toString().trim().ifEmpty { "exit code ${process.exitValue()}" }
				throw BackendException(
					if (!Backend.launcherDetaches && process.exitValue() in 126..127) "Authorization was cancelled"
					else "Failed to launch the backend: $details"
				)
			}

			val socket = try {
				server.accept()
			} catch (_: SocketTimeoutException) {
				continue
			}

			val conn = BackendConnection(socket)
			socket.soTimeout = 10_000
			val hello = runCatching { json.decodeFromString<BackendMessage>(conn.reader.readLine()!!) }.getOrNull()
			socket.soTimeout = 0

			if (hello?.type == "hello" && MessageDigest.isEqual(hello.key.toByteArray(), key.toByteArray())) {
				return conn
			}

			log("warn", "Rejected a connection that isn't the backend")
			socket.close()
		}
	}

	private fun talkToBackend(conn: BackendConnection, token: String) {
		conn.writer.sendLine(BackendMessage(type = "start", token = token, dns = "1.1.1.1"))
		control = conn.writer

		// runs until the backend stops, disconnect() asks it to. If the GUI quits,
		// the socket gets closed and the backend stops by itself.
		while (true) {
			val line = conn.reader.readLine() ?: throw BackendException("The backend quit unexpectedly")

			val msg = json.decodeFromString<BackendMessage>(line)
			when (msg.type) {
				"log" -> log(msg.level, msg.time, msg.msg)
				"connected" -> updateConnState(ConnectionState.Connected)
				"stopped" -> {
					if (msg.error.isNotEmpty()) throw BackendException(msg.error)
					log("info", "Vpn stopped without errors")
					return
				}
			}
		}
	}

	private class BackendConnection(val socket: Socket) : AutoCloseable {
		val reader: BufferedReader = socket.getInputStream().bufferedReader()
		val writer: BufferedWriter = socket.getOutputStream().bufferedWriter()

		override fun close() = socket.close()
	}

	private fun BufferedWriter.sendLine(msg: BackendMessage) = synchronized(this) {
		write(json.encodeToString(msg))
		newLine()
		flush()
	}
}
