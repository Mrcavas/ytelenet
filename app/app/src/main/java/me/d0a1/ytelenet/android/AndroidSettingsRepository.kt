package me.d0a1.ytelenet.android

import android.content.Context
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.map
import me.d0a1.ytelenet.SettingsRepository

class AndroidSettingsRepository(private val context: Context) : SettingsRepository {
	val Context.dataStore by preferencesDataStore("settings")
	private val TOKEN_KEY = stringPreferencesKey("token")
//	private val SAVED_TOKENS_KEY = stringPreferencesKey("token")

	override val savedToken = context.dataStore.data.map { prefs -> prefs[TOKEN_KEY] ?: "" }

	override suspend fun saveToken(token: String) {
		context.dataStore.edit { prefs ->
			prefs[TOKEN_KEY] = token
		}
	}
}