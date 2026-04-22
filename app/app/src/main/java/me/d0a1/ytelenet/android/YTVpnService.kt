package me.d0a1.ytelenet.android

import android.annotation.SuppressLint
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Intent
import android.content.pm.ServiceInfo
import android.net.VpnService
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.ServiceCompat
import me.d0a1.ytelenet.ConnectionState
import me.d0a1.ytelenet.MainActivity
import me.d0a1.ytelenet.R
import me.d0a1.ytelenet.VpnManager
import org.koin.android.ext.android.inject
import vpn.Vpn
import kotlin.io.encoding.Base64

@SuppressLint("VpnServicePolicy")
class YTVpnService : VpnService() {
	private val vpnManager: VpnManager by inject()

	companion object {
		private const val NOTIFICATION_CHANNEL_ID = "vpn_channel"
		private const val NOTIFICATION_ID = 1
	}

	override fun onCreate() {
		super.onCreate()
		createNotificationChannel()
	}

	override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
		if (intent?.action == "DISCONNECT") {
			Vpn.stop()
			return START_NOT_STICKY
		}

		ServiceCompat.startForeground(
			this,
			NOTIFICATION_ID,
			createNotification(),
			if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.UPSIDE_DOWN_CAKE)
				ServiceInfo.FOREGROUND_SERVICE_TYPE_SYSTEM_EXEMPTED // Android 14+
			else 0
		)

		val token = intent?.getStringExtra("token") ?: return START_NOT_STICKY

		val parts = Base64.decode(token).decodeToString().split(";")
		val pcNum = parts[2].toInt()

		val pfd = Builder()
			.setMtu(1186)
			.setSession("YTelenet")
			.addAddress("42.42.42.${pcNum}", 32)
			.addRoute("0.0.0.0", 0)
			.addDisallowedApplication(packageName)
//			.addDnsServer("8.8.8.8")
			.establish() ?: run {
			stopSelf()
			return START_NOT_STICKY
		}

		Vpn.start(
			pfd.detachFd().toLong(),
			token,
			{ vpnManager.log(it) },
			{ vpnManager.updateConnState(ConnectionState.Connected) }) { errMsg ->
			vpnManager.updateConnState(ConnectionState.Disconnected)

			if (errMsg.isEmpty()) {
				vpnManager.log("info", "Vpn stopped without errors")
			} else {
				vpnManager.log("error", "Vpn stopped with error: $errMsg")
			}

			pfd.close()
			ServiceCompat.stopForeground(this, ServiceCompat.STOP_FOREGROUND_REMOVE)
			stopSelf()
		}

		return START_STICKY
	}

	private fun createNotificationChannel() {
		val channel = NotificationChannel(
			NOTIFICATION_CHANNEL_ID,
			getString(R.string.status_chan_name),
			NotificationManager.IMPORTANCE_LOW
		).apply {
			description = getString(R.string.status_chan_description)
		}

		val manager = getSystemService(NotificationManager::class.java)
		manager.createNotificationChannel(channel)
	}

	private fun createNotification(): Notification {
		val openAppIntent = Intent(this, MainActivity::class.java).apply {
			flags = Intent.FLAG_ACTIVITY_SINGLE_TOP or Intent.FLAG_ACTIVITY_CLEAR_TOP
		}
		val openAppPendingIntent = PendingIntent.getActivity(
			this, 0, openAppIntent, PendingIntent.FLAG_IMMUTABLE
		)

		val disconnectIntent = Intent(this, YTVpnService::class.java).apply {
			action = "DISCONNECT"
		}
		val disconnectPendingIntent = PendingIntent.getService(
			this, 1, disconnectIntent, PendingIntent.FLAG_IMMUTABLE
		)

		return NotificationCompat.Builder(this, NOTIFICATION_CHANNEL_ID)
			.setContentTitle(getString(R.string.app_name))
			.setContentText(getString(R.string.status_notif_description))
			.setSmallIcon(R.drawable.ic_stat_name)
			.setContentIntent(openAppPendingIntent)
			.setOngoing(true)
			.addAction(
				R.drawable.mode_off_on_48px,
				getString(R.string.disconnect),
				disconnectPendingIntent
			)
			.build()
	}

	override fun onRevoke() =
		vpnManager.log("debug", "revoking vpn service")

	override fun onDestroy() =
		Vpn.stop()
}