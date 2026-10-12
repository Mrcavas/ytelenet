package me.d0a1.ytelenet.android

import android.annotation.SuppressLint
import android.app.PendingIntent
import android.content.Intent
import android.net.VpnService
import android.os.Build
import android.service.quicksettings.Tile
import android.service.quicksettings.TileService
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.launch
import me.d0a1.ytelenet.ConnectionState
import me.d0a1.ytelenet.MainActivity
import me.d0a1.ytelenet.R
import me.d0a1.ytelenet.SettingsRepository
import me.d0a1.ytelenet.VpnManager
import me.d0a1.ytelenet.resources.Res
import me.d0a1.ytelenet.resources.connected
import me.d0a1.ytelenet.resources.connecting
import me.d0a1.ytelenet.resources.disconnected
import org.jetbrains.compose.resources.getString
import org.koin.android.ext.android.inject

class YTVpnTileService : TileService() {
	private val vpnManager: VpnManager by inject()
	private val settingsRepo: SettingsRepository by inject()
	private var job: Job? = null

	override fun onStartListening() {
		super.onStartListening()

		job = CoroutineScope(Dispatchers.Main).launch {
			vpnManager.connState.collect { updateTileUi(it, getStateLabel(it)) }
		}
	}

	override fun onStopListening() {
		super.onStopListening()
		job?.cancel()
	}

	override fun onClick() {
		super.onClick()

		val currentState = vpnManager.connState.value

		if (currentState == ConnectionState.Connected) {
			vpnManager.disconnect()
		}
		else if (currentState == ConnectionState.Disconnected) {
			if (VpnService.prepare(this) != null) {
				openApp()
				return
			}

			CoroutineScope(Dispatchers.IO).launch {
				val token = settingsRepo.savedToken.firstOrNull()
				if (!token.isNullOrEmpty()) {
					vpnManager.connect(token)
				} else {
					openApp() // No token saved, open the app so they can type it in
				}
			}
		}
	}

	private suspend fun getStateLabel(state: ConnectionState) = getString(
		when (state) {
			ConnectionState.Connected -> Res.string.connected
			ConnectionState.Loading -> Res.string.connecting
			ConnectionState.Disconnected -> Res.string.disconnected
		}
	)

	private fun updateTileUi(state: ConnectionState, subtitle: String) {
		val tile = qsTile ?: return

		tile.state = when (state) {
			ConnectionState.Connected -> Tile.STATE_ACTIVE
			ConnectionState.Loading -> Tile.STATE_UNAVAILABLE
			ConnectionState.Disconnected -> Tile.STATE_INACTIVE
		}
		tile.label = getString(R.string.app_name)
		tile.subtitle = subtitle
		tile.updateTile()
	}

	@SuppressLint("StartActivityAndCollapseDeprecated")
	private fun openApp() {
		val intent = Intent(this, MainActivity::class.java).apply {
			flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
		}

		if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.UPSIDE_DOWN_CAKE) {
			startActivityAndCollapse(PendingIntent.getActivity(this, 0, intent, PendingIntent.FLAG_IMMUTABLE))
		} else {
			startActivityAndCollapse(intent)
		}
	}
}