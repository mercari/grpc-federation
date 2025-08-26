package com.github.mercari.grpcfederation.settings

import com.intellij.openapi.project.Project
import com.intellij.util.xmlb.annotations.Tag
import com.intellij.util.xmlb.annotations.XCollection

interface ProjectSettingsService {
    @Tag("ImportPathEntry")
    data class ImportPathEntry(
            var path: String = "",
            var enabled: Boolean = true
    )

    @Tag("GrpcFederationSettings")
    data class State(
            @get:XCollection(style = XCollection.Style.v2)
            var importPaths: MutableList<String> = mutableListOf(),
            @get:XCollection(style = XCollection.Style.v2)
            var importPathEntries: MutableList<ImportPathEntry> = mutableListOf()
    )

    var currentState: State
}

val Project.projectSettings: ProjectSettingsService
    get() = getService(ProjectSettingsService::class.java)
            ?: error("Failed to get ProjectSettingsService for $this")