import org.jetbrains.compose.desktop.application.dsl.TargetFormat
import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
	alias(libs.plugins.kotlin.jvm)
	alias(libs.plugins.kotlin.compose)
	alias(libs.plugins.kotlin.serialization)
	alias(libs.plugins.compose.multiplatform)
	alias(libs.plugins.koin.compiler)
}

kotlin {
	compilerOptions {
		jvmTarget = JvmTarget.JVM_17
	}
}

java {
	toolchain {
		languageVersion = JavaLanguageVersion.of(17)
	}
}

dependencies {
	implementation(projects.shared)
	implementation(compose.desktop.currentOs)
	implementation(libs.kotlinx.coroutines.swing)
	implementation(libs.kotlinx.serialization.json)
}

// The GUI runs the Go client (`ytelenet gui-backend`) elevated, so it's built
// for the host and bundled into the app resources. jpackage can only package
// for the OS it runs on anyway.
val hostOs: String = System.getProperty("os.name").lowercase()
val goos = when {
	hostOs.contains("windows") -> "windows"
	hostOs.contains("mac") -> "darwin"
	else -> "linux"
}
val goarch = when (System.getProperty("os.arch")) {
	"aarch64", "arm64" -> "arm64"
	else -> "amd64"
}
val resourcesPlatform = "${if (goos == "darwin") "macos" else goos}-${if (goarch == "arm64") "arm64" else "x64"}"
val backendName = if (goos == "windows") "ytelenet.exe" else "ytelenet"
val backendDir = layout.buildDirectory.dir("backend")

val buildBackend by tasks.registering(Exec::class) {
	group = "build"
	description = "Builds the Go backend for the host platform"

	val goModule = rootDir.parentFile
	val output = backendDir.map { it.file("$resourcesPlatform/$backendName") }

	inputs.files(fileTree(goModule) {
		include("go.mod", "go.sum", "main.go", "vpn/**", "ytnode/**")
		exclude("**/*_test.go")
	})
	outputs.file(output)

	workingDir = goModule
	environment("GOOS", goos)
	environment("GOARCH", goarch)
	environment("CGO_ENABLED", "0")
	// go.sum isn't committed
	environment("GOFLAGS", "-mod=mod")
	executable = providers.gradleProperty("goExecutable").getOrElse("go")
	args("build", "-trimpath", "-ldflags=-s -w", "-o", output.get().asFile.absolutePath, ".")
}

tasks.matching { it.name == "prepareAppResources" }.configureEach {
	dependsOn(buildBackend)
}

// jpackage drops the executable bit of the app resources, the packages are
// made from this app image
tasks.matching { it.name == "createDistributable" || it.name == "createReleaseDistributable" }.configureEach {
	doLast {
		outputs.files.asFileTree
			.matching { include("**/resources/$backendName") }
			.forEach { it.setExecutable(true, false) }
	}
}

val appVersion = "0.1.1"

compose.desktop {
	application {
		mainClass = "me.d0a1.ytelenet.desktop.MainKt"

		nativeDistributions {
			targetFormats(TargetFormat.Msi, TargetFormat.Deb, TargetFormat.Rpm)

			packageName = "YTelenet"
			packageVersion = appVersion
			description = "VPN over Yandex Telemost"
			vendor = "d0a1"
			licenseFile = rootProject.file("../LICENSE")
			appResourcesRootDir = backendDir

			modules("jdk.unsupported")

			windows {
				iconFile = project.file("icons/ytelenet.ico")
				menu = true
				shortcut = true
				upgradeUuid = "3d6c1c4e-8a0f-4f4b-9b8e-2f1f5e0b7a42"
			}
			macOS {
				// macOS doesn't allow a zero major version, it's only there for local builds anyway
				packageVersion = "1.0.0"
			}
			linux {
				iconFile = project.file("icons/ytelenet.png")
				packageName = "ytelenet"
				shortcut = true
				menuGroup = "Network"
			}
		}

		buildTypes.release.proguard {
			isEnabled = false
		}
	}
}

// jpackage can't make pacman packages, so the app image is packaged with
// makepkg. Runs on Arch only.
val archDir = layout.buildDirectory.dir("compose/binaries/main/arch")

val prepareArchPackage by tasks.registering(Sync::class) {
	dependsOn("createDistributable")
	from(layout.buildDirectory.dir("compose/binaries/main/app"))
	from("packaging/arch")
	from("icons/ytelenet.png")
	from(rootProject.file("../LICENSE"))
	into(archDir.map { it.dir("staging") })
}

val packageArch by tasks.registering(Exec::class) {
	group = "compose desktop"
	description = "Builds a pacman package for Arch Linux with makepkg"
	dependsOn(prepareArchPackage)

	workingDir(archDir.map { it.dir("staging") })
	environment("YTELENET_VERSION", appVersion)
	environment("PKGDEST", archDir.get().asFile.absolutePath)
	commandLine("makepkg", "--force", "--nodeps", "--noconfirm")

	doLast {
		archDir.get().asFile.listFiles { f -> f.name.endsWith(".pkg.tar.zst") }
			?.forEach { logger.lifecycle("The package is written to $it") }
	}
}
