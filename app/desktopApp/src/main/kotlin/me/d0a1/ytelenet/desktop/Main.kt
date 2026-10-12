package me.d0a1.ytelenet.desktop

import androidx.compose.ui.graphics.painter.BitmapPainter
import androidx.compose.ui.graphics.toComposeImageBitmap
import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import androidx.compose.ui.window.rememberWindowState
import me.d0a1.ytelenet.App
import me.d0a1.ytelenet.SettingsRepository
import me.d0a1.ytelenet.VpnManager
import me.d0a1.ytelenet.sharedModule
import org.koin.core.context.startKoin
import org.koin.dsl.bind
import org.koin.dsl.module
import org.koin.plugin.module.dsl.single
import org.jetbrains.skia.Image
import java.awt.Dimension

val desktopModule = module {
	single<DesktopSettingsRepository>() bind SettingsRepository::class
	single<DesktopVpnManager>() bind VpnManager::class
}

private val appIcon by lazy {
	val bytes = object {}.javaClass.getResourceAsStream("/ytelenet.png")!!.use { it.readBytes() }
	BitmapPainter(Image.makeFromEncoded(bytes).toComposeImageBitmap())
}

fun main() {
	startKoin {
		modules(sharedModule, desktopModule)
	}

	application {
		Window(
			onCloseRequest = ::exitApplication,
			title = "YTelenet",
			icon = appIcon,
			state = rememberWindowState(size = DpSize(420.dp, 720.dp)),
		) {
			window.minimumSize = Dimension(360, 480)
			App()
		}
	}
}
