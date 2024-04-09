package com.github.mercari.grpcfederation.settings

import ai.grazie.nlp.utils.tokenizeByWhitespace
import com.github.mercari.grpcfederation.GrpcFederationBundle
import com.intellij.openapi.options.BoundConfigurable
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.DialogPanel
import com.intellij.ui.dsl.builder.*

class ProjectSettingsConfigurable(
        private val project: Project,
) : BoundConfigurable(GrpcFederationBundle.message("settings.grpcfederation.title")) {
    private val state: ProjectSettingsService.State = project.projectSettings.state.copy()

    override fun createPanel(): DialogPanel = panel {
        row(GrpcFederationBundle.message("settings.grpcfederation.importpaths")) {
            textField().bindText(
                    { state.importPaths.joinToString(separator = " ") },
                    { text -> state.importPaths = text.tokenizeByWhitespace() }
            ).columns(COLUMNS_LARGE)
                    .align(Align.FILL)
                    .comment("For example, /path/to/proto. Paths should be absolute. To specify multiple paths, separate with spaces.")
        }
    }

    override fun reset() {
        state.importPaths = project.projectSettings.state.importPaths
        super.reset()
    }

    override fun apply() {
        super.apply()
        project.projectSettings.state = state
    }

    override fun isModified(): Boolean {
        if (super.isModified()) return true
        return project.projectSettings.state != state
    }
}
