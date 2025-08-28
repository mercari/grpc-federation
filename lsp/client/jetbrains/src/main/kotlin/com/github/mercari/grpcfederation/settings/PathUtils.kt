package com.github.mercari.grpcfederation.settings

import com.intellij.openapi.project.Project

/**
 * Utility object for path-related operations.
 */
object PathUtils {
    /**
     * Expands a path string by replacing user home directory (~) and project root variables.
     *
     * @param path The path to expand
     * @param project The current project (used for ${PROJECT_ROOT} expansion)
     * @return The expanded path
     */
    fun expandPath(path: String, project: Project): String {
        var expanded = path
        
        // Expand user home directory
        if (path.startsWith("~/")) {
            expanded = System.getProperty("user.home") + path.substring(1)
        }
        
        // Expand project root variable
        if (path.contains("\${PROJECT_ROOT}")) {
            expanded = expanded.replace("\${PROJECT_ROOT}", project.basePath ?: "")
        }
        
        // Additional variable expansions can be added here in the future
        
        return expanded
    }
}