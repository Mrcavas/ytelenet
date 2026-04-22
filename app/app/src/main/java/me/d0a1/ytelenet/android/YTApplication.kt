package me.d0a1.ytelenet.android

import android.app.Application
import me.d0a1.ytelenet.SettingsRepository
import me.d0a1.ytelenet.VpnManager
import me.d0a1.ytelenet.VpnViewModel
import org.koin.android.ext.koin.androidContext
import org.koin.android.ext.koin.androidLogger
import org.koin.core.context.startKoin
import org.koin.dsl.bind
import org.koin.dsl.module
import org.koin.plugin.module.dsl.single
import org.koin.plugin.module.dsl.viewModel

val appModule = module {
	single<AndroidSettingsRepository>() bind SettingsRepository::class
	single<AndroidVpnManager>() bind VpnManager::class

	viewModel<VpnViewModel>()
}

class YTApplication : Application() {
	override fun onCreate() {
		super.onCreate()
		startKoin {
			androidLogger()
			androidContext(this@YTApplication)
			modules(appModule)
		}
	}
}