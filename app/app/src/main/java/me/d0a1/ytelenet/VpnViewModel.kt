package me.d0a1.ytelenet

import androidx.compose.foundation.text.input.TextFieldState
import androidx.compose.runtime.snapshotFlow
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.flow.debounce
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.launch

@OptIn(FlowPreview::class)
class VpnViewModel(
	private val vpnManager: VpnManager,
	private val settingsRepo: SettingsRepository
) : ViewModel() {
	val token = TextFieldState()

	init {
		viewModelScope.launch {
			settingsRepo.savedToken.firstOrNull()?.let {
				token.edit { replace(0, length, it) }
			}
			snapshotFlow { token.text.toString() }.debounce(250).collect {
				if (isTokenValid(it)) settingsRepo.saveToken(it)
			}
		}
	}

	val connState = vpnManager.connState
	val logEntries = vpnManager.logEntries

	fun connect() = vpnManager.connect(token.text.toString())
	fun disconnect() = vpnManager.disconnect()
}