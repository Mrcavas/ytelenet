package me.d0a1.ytelenet.android

import android.content.Context
import androidx.datastore.preferences.preferencesDataStore
import me.d0a1.ytelenet.DataStoreSettingsRepository
import me.d0a1.ytelenet.SettingsRepository

private val Context.dataStore by preferencesDataStore("settings")

class AndroidSettingsRepository(context: Context) :
	SettingsRepository by DataStoreSettingsRepository(context.dataStore)
