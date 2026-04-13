package me.d0a1.ytelenet

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.shrinkVertically
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.input.TextFieldState
import androidx.compose.foundation.text.input.TextObfuscationMode
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.List
import androidx.compose.material.icons.automirrored.outlined.List
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.material.icons.filled.VpnKey
import androidx.compose.material.icons.outlined.VpnKey
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilledIconButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.IconButtonDefaults
import androidx.compose.material3.IconButtonShapes
import androidx.compose.material3.LoadingIndicator
import androidx.compose.material3.OutlinedIconButton
import androidx.compose.material3.OutlinedSecureTextField
import androidx.compose.material3.PlainTooltip
import androidx.compose.material3.Scaffold
import androidx.compose.material3.ShortNavigationBar
import androidx.compose.material3.ShortNavigationBarItem
import androidx.compose.material3.Text
import androidx.compose.material3.TooltipAnchorPosition
import androidx.compose.material3.TooltipBox
import androidx.compose.material3.TooltipDefaults
import androidx.compose.material3.rememberTooltipState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Rect
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Matrix
import androidx.compose.ui.graphics.Outline
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.graphics.asComposePath
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.Density
import androidx.compose.ui.unit.LayoutDirection
import androidx.compose.ui.unit.dp
import androidx.graphics.shapes.RoundedPolygon
import androidx.graphics.shapes.toPath
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.NavController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import kotlinx.coroutines.FlowPreview
import me.d0a1.ytelenet.ui.theme.YTelenetTheme
import kotlin.math.max
import vpn.Log as LogEntry

class MainActivity : ComponentActivity() {
	override fun onCreate(savedInstanceState: Bundle?) {
		super.onCreate(savedInstanceState)
		enableEdgeToEdge()
		setContent {
			App()
		}
	}
}

@OptIn(ExperimentalMaterial3Api::class)
@Preview(
	device = "spec:width=1080px,height=1920px,dpi=440,isRound=true,cutout=punch_hole", apiLevel = 34
)
@Composable
fun App() {
	val vm = viewModel<VpnViewModel>()
	val navController = rememberNavController()
	val connState by vm.connState.collectAsStateWithLifecycle()
	val logEntries by vm.logEntries.collectAsStateWithLifecycle()

	YTelenetTheme {
		Scaffold(modifier = Modifier.fillMaxSize(), topBar = {
			CenterAlignedTopAppBar(
				{ Text("YTelenet") })
		}, bottomBar = {
			NavBar(navController) {
				navController.navigate(it) {
					popUpTo("main")
					launchSingleTop = true
				}
			}
		}) { innerPadding ->
			NavHost(
				navController,
				startDestination = "main",
				enterTransition = { slideInHorizontally { it } + fadeIn() },
				exitTransition = { slideOutHorizontally { -it } + fadeOut() },
				popEnterTransition = { slideInHorizontally { -it } + fadeIn() },
				popExitTransition = { slideOutHorizontally { it } + fadeOut() },
				modifier = Modifier.padding(innerPadding)
			) {
				composable("main") {
					MainScreen(connState, vm)
				}
				composable("logs") {
					LogsScreen(logEntries)
				}
			}
		}
	}
}

@Composable
fun MainScreen(connState: ConnectionState, vm: VpnViewModel) {
	Column(
		Modifier
			.fillMaxSize()
			.padding(horizontal = 32.dp)
			.padding(bottom = 32.dp),
		horizontalAlignment = Alignment.CenterHorizontally
	) {
		TokenField(connState, vm.token)
		Column(
			Modifier
				.fillMaxSize()
				.padding(top = 32.dp),
			verticalArrangement = Arrangement.Center,
			horizontalAlignment = Alignment.CenterHorizontally
		) {
			VpnButton(isTokenValid(vm.token.text.toString()), connState, vm::toggleConnection)
		}
	}
}

@Composable
fun LogsScreen(logEntries: List<LogEntry>) {
	LazyColumn(
		Modifier
			.fillMaxSize()
			.padding(horizontal = 32.dp)
			.padding(bottom = 32.dp)
	) {
		items(logEntries) {
			Text("${it.level}[${it.time}]: ${it.msg}")
		}
	}
}

@Composable
fun NavBar(navController: NavController, onSelect: (String) -> Unit) {
	val currentDestination by navController.currentBackStackEntryAsState()
	val selectedTab = currentDestination?.destination?.route

	ShortNavigationBar {
		ShortNavigationBarItem(
			selected = selectedTab == "main",
			icon = {
				Icon(
					if (selectedTab == "main") Icons.Filled.VpnKey
					else Icons.Outlined.VpnKey, contentDescription = null
				)
			},
			onClick = { if (selectedTab != "main") onSelect("main") },
			label = { Text("VPN") },
		)
		ShortNavigationBarItem(
			selected = selectedTab == "logs",
			icon = {
				Icon(
					if (selectedTab == "logs") Icons.AutoMirrored.Filled.List
					else Icons.AutoMirrored.Outlined.List, contentDescription = null
				)
			},
			onClick = { if (selectedTab != "logs") onSelect("logs") },
			label = { Text("Logs") },
		)
	}
}

@OptIn(FlowPreview::class)
@Composable
fun TokenField(connState: ConnectionState, tokenState: TextFieldState) {
	var passwordHidden by rememberSaveable { mutableStateOf(true) }
	var isValid by rememberSaveable { mutableStateOf(isTokenValid(tokenState.text.toString())) }

	LaunchedEffect(Unit) {
		snapshotFlow { tokenState.text }.collect {
			isValid = isTokenValid(it.toString())
		}
	}

	AnimatedVisibility(
		connState == ConnectionState.Disconnected,
		enter = expandVertically() + fadeIn(),
		exit = shrinkVertically() + fadeOut()
	) {
		OutlinedSecureTextField(
			supportingText = {
				Text(
					when {
						tokenState.text.isEmpty() -> "Enter token"
						!isValid -> "Invalid token"
						else -> ""
					},
					Modifier.clearAndSetSemantics {})
			},
			isError = !isValid,
			enabled = connState == ConnectionState.Disconnected,
			state = tokenState,
			label = { Text("Token") },
			modifier = Modifier.fillMaxWidth(),
			shape = RoundedCornerShape(16.dp),
			textObfuscationMode = if (passwordHidden) TextObfuscationMode.Hidden
			else TextObfuscationMode.Visible,
			trailingIcon = {
				val description = if (passwordHidden) "Show password" else "Hide password"
				TooltipBox(
					positionProvider = TooltipDefaults.rememberTooltipPositionProvider(TooltipAnchorPosition.Above),
					tooltip = { PlainTooltip { Text(description) } },
					state = rememberTooltipState(),
				) {
					IconButton(onClick = { passwordHidden = !passwordHidden }) {
						val visibilityIcon = if (passwordHidden) {
							Icons.Filled.Visibility
						} else Icons.Filled.VisibilityOff
						Icon(imageVector = visibilityIcon, contentDescription = description)
					}
				}
			},
		)
	}
}

@Composable
fun VpnButton(enabled: Boolean, state: ConnectionState, onToggle: () -> Unit) {
	AnimatedContent(
		state, Modifier.size(IconButtonDefaults.extraLargeContainerSize())
	) {
		when (it) {
			ConnectionState.Disconnected -> OutlinedIconButton(
				enabled = enabled, onClick = onToggle, shapes = IconButtonShapes(
					IconButtonDefaults.extraLargeSquareShape, IconButtonDefaults.extraLargePressedShape
				)
			) {
				Icon(painterResource(R.drawable.mode_off_on_48px), "Turn on")
			}

			ConnectionState.Loading -> LoadingIndicator(
				modifier = Modifier
					.clip(RoundedCornerShape(16.dp))
					.clickable { onToggle() })

			ConnectionState.Connected -> FilledIconButton(
				onClick = onToggle, shapes = IconButtonShapes(
					IconButtonDefaults.extraLargeSquareShape, IconButtonDefaults.extraLargePressedShape
				)
			) {
				Icon(painterResource(R.drawable.mode_off_on_48px), "Turn off")
			}
		}
	}
}

fun RoundedPolygon.toShape() = RoundedPolygonShape(polygon = this)
fun RoundedPolygon.getBounds() = calculateBounds().let { Rect(it[0], it[1], it[2], it[3]) }

class RoundedPolygonShape(
	private val polygon: RoundedPolygon, private var matrix: Matrix = Matrix()
) : Shape {
	private var path = Path()
	override fun createOutline(
		size: Size, layoutDirection: LayoutDirection, density: Density
	): Outline {
		path.rewind()
		path = polygon.toPath().asComposePath()
		matrix.reset()
		val bounds = polygon.getBounds()
		val maxDimension = max(bounds.width, bounds.height)
		matrix.scale(size.width / maxDimension, size.height / maxDimension)
		matrix.translate(-bounds.left, -bounds.top)

		path.transform(matrix)
		return Outline.Generic(path)
	}
}