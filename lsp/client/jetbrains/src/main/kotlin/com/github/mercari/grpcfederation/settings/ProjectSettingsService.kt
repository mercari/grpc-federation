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
            /**
             * Legacy field for backward compatibility.
             * Contains only enabled paths for compatibility with older plugin versions.
             * 
             * @deprecated This field will be removed in version 0.3.0.
             * Migration plan:
             * - v0.2.0 (current): Both fields are maintained, new format is preferred
             *   Automatic migration from old to new format on first load
             * - v0.3.0: Remove this field completely, only use importPathEntries
             * 
             * TODO: Remove in v0.3.0
             */
            @get:XCollection(style = XCollection.Style.v2)
            var importPaths: MutableList<String> = mutableListOf(),
            
            /**
             * New format that supports enable/disable functionality for each path.
             * This is the primary storage format from v0.2.0 onwards.
             */
            @get:XCollection(style = XCollection.Style.v2)
            var importPathEntries: MutableList<ImportPathEntry> = mutableListOf()
    )

    var currentState: State
}

val Project.projectSettings: ProjectSettingsService
    get() = getService(ProjectSettingsService::class.java)
            ?: error("Failed to get ProjectSettingsService for $this")