package me.d0a1.ytelenet.desktop

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import me.d0a1.ytelenet.DataStoreSettingsRepository
import me.d0a1.ytelenet.SettingsRepository
import java.io.File

class DesktopSettingsRepository : SettingsRepository by DataStoreSettingsRepository(
	PreferenceDataStoreFactory.create { File(configDir(), "settings.preferences_pb") }
)

private fun configDir(): File {
	val home = System.getProperty("user.home")
	val os = System.getProperty("os.name").lowercase()

	return when {
		os.contains("windows") ->
			File(System.getenv("APPDATA") ?: "$home\\AppData\\Roaming", "YTelenet")

		os.contains("mac") -> File(home, "Library/Application Support/YTelenet")
		else -> File(System.getenv("XDG_CONFIG_HOME") ?: "$home/.config", "ytelenet")
	}
}
