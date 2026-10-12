package me.d0a1.ytelenet

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map

interface SettingsRepository {
	val savedToken: Flow<String>
	suspend fun saveToken(token: String)
}

class DataStoreSettingsRepository(private val dataStore: DataStore<Preferences>) : SettingsRepository {
	private val TOKEN_KEY = stringPreferencesKey("token")

	override val savedToken = dataStore.data.map { prefs -> prefs[TOKEN_KEY] ?: "" }

	override suspend fun saveToken(token: String) {
		dataStore.edit { prefs ->
			prefs[TOKEN_KEY] = token
		}
	}
}
