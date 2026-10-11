package me.d0a1.ytelenet

import androidx.compose.runtime.Composable

/**
 * Returns an action that asks for whatever the platform needs to bring up a
 * VPN, and calls [onGranted] once it's allowed
 */
@Composable
expect fun rememberVpnPermission(onError: () -> Unit, onGranted: () -> Unit): () -> Unit
