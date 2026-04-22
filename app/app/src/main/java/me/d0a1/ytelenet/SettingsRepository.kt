package me.d0a1.ytelenet

import kotlinx.coroutines.flow.Flow

interface SettingsRepository {
	val savedToken: Flow<String>
	suspend fun saveToken(token: String)
}