package com.github.mercari.grpcfederation.lsp

import com.github.mercari.grpcfederation.settings.PathUtils
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
        
        // Use new format (importPathEntries) if available, otherwise fall back to legacy format
        // TODO: Remove legacy format support in v0.3.0
        val paths = if (state.importPathEntries.isNotEmpty()) {
            state.importPathEntries.filter { it.enabled }.map { it.path }
        } else {
            // Legacy format: all paths are considered enabled
            state.importPaths
        }
        
        for (p in paths) {
            if (p.isNotEmpty()) {
                cmd.add("-I")
                cmd.add(PathUtils.expandPath(p, project))
            }
        }
        return GeneralCommandLine(cmd)
    }
}