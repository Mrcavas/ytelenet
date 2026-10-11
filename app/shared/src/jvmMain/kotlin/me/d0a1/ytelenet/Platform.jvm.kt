package me.d0a1.ytelenet

import androidx.compose.runtime.Composable

// the desktop backend asks for admin rights itself when it's launched
@Composable
actual fun rememberVpnPermission(onError: () -> Unit, onGranted: () -> Unit): () -> Unit = onGranted
