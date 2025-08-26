package com.github.mercari.grpcfederation.lsp

import com.github.mercari.grpcfederation.settings.projectSettings
import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.platform.lsp.api.LspServerSupportProvider
import com.intellij.platform.lsp.api.ProjectWideLspServerDescriptor

internal class GrpcFederationLspServerSupportProvider : LspServerSupportProvider {
    override fun fileOpened(project: Project, file: VirtualFile, serverStarter: LspServerSupportProvider.LspServerStarter) {
        if (file.extension == "proto") {
            serverStarter.ensureServerStarted(GrpcFederationLspServerDescriptor(project))
        }
    }
}

private class GrpcFederationLspServerDescriptor(project: Project) : ProjectWideLspServerDescriptor(project, "gRPC Federation") {
    override fun isSupportedFile(file: VirtualFile) = file.extension == "proto"
    override fun createCommandLine(): GeneralCommandLine {
        val state = project.projectSettings.currentState
        val cmd = mutableListOf<String>()
        cmd.add("grpc-federation-language-server")
        
        // Use new format if available, otherwise fall back to old format
        val paths = if (state.importPathEntries.isNotEmpty()) {
            state.importPathEntries.filter { it.enabled }.map { it.path }
        } else {
            state.importPaths
        }
        
        for (p in paths) {
            if (p.isNotEmpty()) {
                cmd.add("-I")
                cmd.add(expandPath(project, p))
            }
        }
        return GeneralCommandLine(cmd)
    }
    
    private fun expandPath(project: Project, path: String): String {
        var expanded = path
        if (path.startsWith("~/")) {
            expanded = System.getProperty("user.home") + path.substring(1)
        }
        if (path.contains("\${PROJECT_ROOT}")) {
            expanded = expanded.replace("\${PROJECT_ROOT}", project.basePath ?: "")
        }
        return expanded
    }
}