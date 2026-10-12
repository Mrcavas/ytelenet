package me.d0a1.ytelenet.desktop

import kotlinx.serialization.Serializable
import java.io.File

/** A line of the GUI <-> backend protocol, see vpn/backend.go */
@Serializable
data class BackendMessage(
	val type: String,

	// hello
	val key: String = "",

	// start
	val token: String = "",
	val dns: String = "",
	val noAutoRoute: Boolean = false,
	val debug: Boolean = false,

	// log
	val level: String = "",
	val msg: String = "",
	val time: String = "",

	// stopped
	val error: String = "",
)

class BackendException(message: String) : Exception(message)

object Backend {
	private val os = System.getProperty("os.name").lowercase()
	private val isWindows = os.contains("windows")

	/** Whether the launcher exits once the backend is started, instead of becoming it */
	val launcherDetaches = isWindows
	private val binaryName = if (isWindows) "ytelenet.exe" else "ytelenet"

	/**
	 * The Go binary: set with -Dytelenet.backend or YTELENET_BACKEND, otherwise
	 * the one bundled into the app resources
	 */
	fun findBinary(): File {
		val explicit = System.getProperty("ytelenet.backend") ?: System.getenv("YTELENET_BACKEND")
		if (explicit != null) return File(explicit)

		val resources = System.getProperty("compose.application.resources.dir")
			?: throw BackendException("Backend binary not found, set YTELENET_BACKEND")
		return File(resources, binaryName)
	}

	/**
	 * Starts the backend with admin rights, which it needs to create the tunnel.
	 * The returned process is the elevating launcher, the backend itself reports
	 * back over the socket.
	 */
	fun launch(binary: File, args: List<String>): Process {
		if (!binary.isFile) throw BackendException("Backend binary not found: $binary")

		val command = when {
			isWindows -> listOf(
				"powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command",
				"Start-Process -Verb RunAs -WindowStyle Hidden" +
						" -FilePath ${psQuote(binary.absolutePath)}" +
						" -ArgumentList ${psQuote(args.joinToString(" "))}"
			)

			isRoot() -> listOf(binary.absolutePath) + args
			hasCommand("pkexec") -> listOf("pkexec", binary.absolutePath) + args
			else -> throw BackendException("Root rights are needed: install polkit (pkexec) or run YTelenet as root")
		}

		return ProcessBuilder(command)
			.redirectErrorStream(true)
			.start()
	}

	private fun psQuote(s: String) = "'" + s.replace("'", "''") + "'"

	private fun isRoot() = runCatching {
		ProcessBuilder("id", "-u").start().inputStream.bufferedReader().readText().trim() == "0"
	}.getOrDefault(false)

	private fun hasCommand(name: String) =
		System.getenv("PATH").orEmpty().split(File.pathSeparator).any { File(it, name).canExecute() }
}
