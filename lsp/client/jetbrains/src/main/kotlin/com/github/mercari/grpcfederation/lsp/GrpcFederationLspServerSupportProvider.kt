package com.github.mercari.grpcfederation.lsp

import com.github.mercari.grpcfederation.settings.PathUtils
import com.github.mercari.grpcfederation.settings.projectSettings
import com.intellij.execution.configurations.GeneralCommandLine
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import com.intellij.openapi.editor.DefaultLanguageHighlighterColors
import com.intellij.openapi.editor.colors.TextAttributesKey
import com.intellij.platform.lsp.api.LspServerDescriptor
import com.intellij.platform.lsp.api.LspServerSupportProvider
import com.intellij.platform.lsp.api.customization.LspSemanticTokensSupport
import com.intellij.platform.lsp.api.customization.LspDiagnosticsSupport
import com.intellij.psi.PsiFile
import com.intellij.openapi.util.text.StringUtil
import org.eclipse.lsp4j.Diagnostic

internal class GrpcFederationLspServerSupportProvider : LspServerSupportProvider {
    override fun fileOpened(project: Project, file: VirtualFile, serverStarter: LspServerSupportProvider.LspServerStarter) {
        if (file.extension == "proto") {
            serverStarter.ensureServerStarted(GrpcFederationLspServerDescriptor(project))
        }
    }
}

private class GrpcFederationLspServerDescriptor(project: Project) : LspServerDescriptor(project, "gRPC Federation") {
    override fun isSupportedFile(file: VirtualFile) = file.extension == "proto"
    
    // Enable diagnostics support with HTML escaping
    override val lspDiagnosticsSupport: LspDiagnosticsSupport = object : LspDiagnosticsSupport() {
        override fun getTooltip(diagnostic: Diagnostic): String {
            // Escape HTML/XML entities in the diagnostic message to prevent UI issues
            // The tooltip is rendered as HTML, so we need to escape special characters
            // Always use <pre> tag to preserve formatting and ensure consistent display
            val escapedMessage = StringUtil.escapeXmlEntities(diagnostic.message)
            return "<pre>$escapedMessage</pre>"
        }
    }
    
    // Enable semantic tokens support for gRPC Federation syntax highlighting
    override val lspSemanticTokensSupport: LspSemanticTokensSupport = object : LspSemanticTokensSupport() {
        override fun shouldAskServerForSemanticTokens(psiFile: PsiFile): Boolean {
            // Always request semantic tokens for proto files
            return psiFile.virtualFile?.extension == "proto"
        }
        
        override fun getTextAttributesKey(
            tokenType: String,
            modifiers: List<String>
        ): TextAttributesKey? {
            return when (tokenType) {
                // Types
                "type" -> DefaultLanguageHighlighterColors.CLASS_NAME
                "class" -> DefaultLanguageHighlighterColors.CLASS_NAME
                "enum" -> DefaultLanguageHighlighterColors.CLASS_NAME
                "interface" -> DefaultLanguageHighlighterColors.INTERFACE_NAME
                "struct" -> DefaultLanguageHighlighterColors.CLASS_NAME
                "typeParameter" -> DefaultLanguageHighlighterColors.CLASS_NAME
                
                // Variables and properties
                "variable" -> when {
                    modifiers.contains("readonly") -> DefaultLanguageHighlighterColors.CONSTANT
                    modifiers.contains("static") -> DefaultLanguageHighlighterColors.STATIC_FIELD
                    else -> DefaultLanguageHighlighterColors.LOCAL_VARIABLE
                }
                "property" -> when {
                    modifiers.contains("static") -> DefaultLanguageHighlighterColors.STATIC_FIELD
                    else -> DefaultLanguageHighlighterColors.INSTANCE_FIELD
                }
                "parameter" -> DefaultLanguageHighlighterColors.PARAMETER
                
                // Functions and methods
                "function" -> when {
                    modifiers.contains("declaration") -> DefaultLanguageHighlighterColors.FUNCTION_DECLARATION
                    else -> DefaultLanguageHighlighterColors.FUNCTION_CALL
                }
                "method" -> when {
                    modifiers.contains("static") -> DefaultLanguageHighlighterColors.STATIC_METHOD
                    else -> DefaultLanguageHighlighterColors.INSTANCE_METHOD
                }
                
                // Other
                "namespace" -> DefaultLanguageHighlighterColors.CLASS_NAME
                "enumMember" -> DefaultLanguageHighlighterColors.CONSTANT
                "keyword" -> DefaultLanguageHighlighterColors.KEYWORD
                "string" -> DefaultLanguageHighlighterColors.STRING
                "number" -> DefaultLanguageHighlighterColors.NUMBER
                "comment" -> DefaultLanguageHighlighterColors.LINE_COMMENT
                "operator" -> DefaultLanguageHighlighterColors.OPERATION_SIGN
                
                // gRPC Federation specific
                "decorator" -> DefaultLanguageHighlighterColors.METADATA  // for annotations like grpc.federation
                
                else -> null
            }
        }
    }
    
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