package me.d0a1.ytelenet.android

import android.app.Activity
import android.net.VpnService
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.runtime.Composable
import androidx.compose.ui.platform.LocalContext

@Composable
fun rememberVpnPermission(onError: () -> Unit, onGranted: () -> Unit): () -> Unit {
	val context = LocalContext.current

	val launcher = rememberLauncherForActivityResult(
		contract = ActivityResultContracts.StartActivityForResult()
	) { result ->
		if (result.resultCode == Activity.RESULT_OK)
			onGranted()
		else
			onError()
	}

	return {
		val intent = VpnService.prepare(context)
		if (intent != null)
			launcher.launch(intent)
		else
			onGranted()
	}
}