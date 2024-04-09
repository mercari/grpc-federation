package com.github.mercari.grpcfederation.settings

import com.intellij.openapi.project.Project

interface ProjectSettingsService {
    data class State(
            var importPaths: List<String> = emptyList(),
    )

    var state: State
}

val Project.projectSettings: ProjectSettingsService
    get() = getService(ProjectSettingsService::class.java)
            ?: error("Failed to get ProjectSettingsService for $this")