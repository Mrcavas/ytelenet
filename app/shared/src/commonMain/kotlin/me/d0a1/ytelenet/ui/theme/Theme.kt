package me.d0a1.ytelenet.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.ColorScheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable

private val DarkColorScheme = darkColorScheme(
	primary = Purple80,
	secondary = PurpleGrey80,
	tertiary = Pink80
)

private val LightColorScheme = lightColorScheme(
	primary = Purple40,
	secondary = PurpleGrey40,
	tertiary = Pink40

	/* Other default colors to override
    background = Color(0xFFFFFBFE),
    surface = Color(0xFFFFFBFE),
    onPrimary = Color.White,
    onSecondary = Color.White,
    onTertiary = Color.White,
    onBackground = Color(0xFF1C1B1F),
    onSurface = Color(0xFF1C1B1F),
    */
)

/** The system color scheme (dynamic colors on Android), or null if there's none */
@Composable
internal expect fun dynamicColorScheme(darkTheme: Boolean): ColorScheme?

@Composable
fun YTelenetTheme(
	darkTheme: Boolean = isSystemInDarkTheme(),
	dynamicColor: Boolean = true,
	content: @Composable () -> Unit
) {
	val colorScheme = (if (dynamicColor) dynamicColorScheme(darkTheme) else null)
		?: if (darkTheme) DarkColorScheme else LightColorScheme

	MaterialTheme(
		colorScheme = colorScheme,
		typography = Typography,
		content = content
	)
}
