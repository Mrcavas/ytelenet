package me.d0a1.ytelenet

import android.app.Application
import android.content.Context
import androidx.compose.foundation.text.input.TextFieldState
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.snapshotFlow
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.FlowPreview
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.debounce
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlin.time.Duration.Companion.milliseconds

val Context.dataStore by preferencesDataStore("settings")
private val TOKEN_KEY = stringPreferencesKey("token")

enum class ConnectionState {
	Connected, Loading, Disconnected
}

@OptIn(FlowPreview::class)
class VpnViewModel(app: Application) : AndroidViewModel(app) {
	private val dataStore = app.dataStore

	val token = TextFieldState()
	val savedToken = dataStore.data.map { prefs -> prefs[TOKEN_KEY] ?: "" }

	init {
		viewModelScope.launch {
			savedToken.firstOrNull()?.let {
				token.edit { replace(0, length, it) }
			}
			snapshotFlow { token.text.toString() }.debounce(250).collect {
				if (isTokenValid(it)) saveToken()
			}
		}
	}

	val connState = Global.connState.asStateFlow()
	val logEntries = Global.logEntries.asStateFlow()

	fun toggleConnection() {
		when (connState.value) {
			ConnectionState.Disconnected -> {
				Global.connState.update { ConnectionState.Loading }
				AndroidController.connect(token.text.toString())
			}
			else -> AndroidController.disconnect()
		}
	}

	fun saveToken() = viewModelScope.launch {
		dataStore.edit { prefs ->
			prefs[TOKEN_KEY] = token.text.toString()
		}
	}
}