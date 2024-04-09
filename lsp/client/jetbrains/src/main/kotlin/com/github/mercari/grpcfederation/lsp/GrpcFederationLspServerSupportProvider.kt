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
        val cmd = buildList{
            add("grpc-federation-language-server")
            for (p in project.projectSettings.state.importPaths) {
                add("-I")
                add(p)
            }
        }
        return GeneralCommandLine(cmd)
    }
}