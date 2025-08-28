package com.github.mercari.grpcfederation.settings.impl

import com.github.mercari.grpcfederation.lsp.GrpcFederationLspServerSupportProvider
import com.github.mercari.grpcfederation.settings.ProjectSettingsService
import com.github.mercari.grpcfederation.settings.ProjectSettingsService.State
import com.intellij.codeInsight.daemon.DaemonCodeAnalyzer
import com.intellij.openapi.components.PersistentStateComponent
import com.intellij.openapi.components.Storage
import com.intellij.openapi.components.StoragePathMacros
import com.intellij.openapi.project.Project
import com.intellij.platform.lsp.api.LspServerManager
import com.intellij.util.xmlb.XmlSerializerUtil

private const val serviceName: String = "GRPCFederationProjectSettings"

@com.intellij.openapi.components.State(
        name = serviceName, storages = [Storage(StoragePathMacros.WORKSPACE_FILE)]
)
class ProjectSettingsServiceImpl(
        private val project: Project
) : PersistentStateComponent<State>, ProjectSettingsService {
    @Volatile
    private var myState: State = State()

    override var currentState: State
        get() = myState
        set(newState) {
            if (myState != newState) {
                XmlSerializerUtil.copyBean(newState, myState)
                // Restart LSP server with new settings
                LspServerManager.getInstance(project).stopAndRestartIfNeeded(GrpcFederationLspServerSupportProvider::class.java)
                DaemonCodeAnalyzer.getInstance(project).restart()
            }
        }

    override fun getState(): State {
        return myState
    }

    override fun loadState(state: State) {
        XmlSerializerUtil.copyBean(state, myState)
    }
}