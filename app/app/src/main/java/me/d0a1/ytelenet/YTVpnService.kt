package me.d0a1.ytelenet

import android.content.Intent
import android.net.VpnService
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import vpn.Vpn
import kotlin.io.encoding.Base64

class YTVpnService : VpnService() {
	private var vpnThread: Job? = null

	override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
		val token = intent?.getStringExtra("token") ?: return START_NOT_STICKY

		val parts = Base64.decode(token).decodeToString().split(";")
		val pcNum = parts[2].toInt()

		// startForeground here

		val pfd = Builder()
			.setSession("YTelenet")
			.addAddress("42.42.42.${pcNum}", 32)
			.addRoute("0.0.0.0", 0)
			.addDnsServer("8.8.8.8")
			.establish() ?: run {
			stopSelf()
			return START_NOT_STICKY
		}

		vpnThread = CoroutineScope(Dispatchers.IO).launch {
			Vpn.start(
				pfd.detachFd().toLong(),
				token,
				{ Global.log(it) },
				{ Global.connState.update { ConnectionState.Connected } }) { errMsg ->
				Global.connState.update { ConnectionState.Disconnected }

				if (errMsg.isEmpty()) {
					Global.log("info","Vpn stopped without errors")
				} else {
					Global.log("error", "Vpn stopped with error: $errMsg")
				}
			}
		}

		return START_STICKY
	}

	override fun onRevoke() {
		vpnThread?.cancel()
		stopSelf()
	}

	override fun onDestroy() {
		vpnThread?.cancel()
	}
}