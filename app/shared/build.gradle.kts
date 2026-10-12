import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
	alias(libs.plugins.kotlin.multiplatform)
	alias(libs.plugins.android.kmp.library)
	alias(libs.plugins.kotlin.compose)
	alias(libs.plugins.compose.multiplatform)
	alias(libs.plugins.koin.compiler)
}

kotlin {
	androidLibrary {
		namespace = "me.d0a1.ytelenet.shared"
		compileSdk = 36
		minSdk = 30

		androidResources {
			enable = true
		}

		compilerOptions {
			jvmTarget = JvmTarget.JVM_11
		}
	}

	jvm {
		compilerOptions {
			jvmTarget = JvmTarget.JVM_17
		}
	}

	sourceSets {
		commonMain.dependencies {
			api(libs.compose.runtime)
			api(libs.compose.foundation)
			api(libs.compose.animation)
			api(libs.compose.ui)
			api(libs.compose.material3)
			api(libs.compose.components.resources)
			implementation(libs.compose.ui.tooling.preview)
			implementation(libs.compose.material.icons)
			api(libs.jb.lifecycle.runtime.compose)
			api(libs.jb.lifecycle.viewmodel.compose)
			implementation(libs.jb.navigation.compose)
			api(libs.androidx.datastore.preferences.core)
			api(libs.kotlinx.coroutines.core)
			api(libs.koin.core)
			api(libs.koin.compose)
			api(libs.koin.compose.viewmodel)
		}
		androidMain.dependencies {
			implementation(libs.androidx.activity.compose)
			implementation(libs.androidx.compose.material3)
		}
	}
}

compose.resources {
	publicResClass = true
	packageOfResClass = "me.d0a1.ytelenet.resources"
}
