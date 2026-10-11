package me.d0a1.ytelenet

import org.koin.dsl.module
import org.koin.plugin.module.dsl.viewModel

val sharedModule = module {
	viewModel<VpnViewModel>()
}
