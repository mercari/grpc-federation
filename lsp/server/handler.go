package server

import (
	"context"
	"io"
	"log/slog"
	"sync"

	"go.lsp.dev/protocol"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/validator"
)

var _ protocol.Server = &Handler{}

type Handler struct {
	logger           *slog.Logger
	importPaths      []string
	client           protocol.Client
	compiler         *compiler.Compiler
	validator        *validator.Validator
	completer        *Completer
	fileContentMu    sync.Mutex
	fileToContentMap map[string][]byte
	tokenTypeMap     map[string]uint32
	tokenModifierMap map[string]uint32
}

func NewHandler(client protocol.Client, w io.Writer, importPaths []string) *Handler {
	logger := slog.New(slog.NewJSONHandler(w, nil))
	c := compiler.New()
	return &Handler{
		logger:           logger,
		importPaths:      importPaths,
		client:           client,
		compiler:         c,
		completer:        NewCompleter(c, logger),
		validator:        validator.New(),
		fileToContentMap: map[string][]byte{},
		tokenTypeMap:     make(map[string]uint32),
		tokenModifierMap: make(map[string]uint32),
	}
}

func (h *Handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	h.logger.Info("Initialize", slog.Any("params", params))
	return h.initialize(params)
}

func (h *Handler) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	h.logger.Info("Initialized", slog.Any("params", params))
	return nil
}

func (h *Handler) Shutdown(ctx context.Context) error {
	h.logger.Info("Shutdown")
	return nil
}

func (h *Handler) Exit(ctx context.Context) error {
	h.logger.Info("Exit")
	return nil
}

func (h *Handler) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
	h.logger.Info("WorkDoneProgressCancel", slog.Any("params", params))
	return nil
}

func (h *Handler) LogTrace(ctx context.Context, params *protocol.LogTraceParams) error {
	h.logger.Info("LogTrace", slog.Any("params", params))
	return nil
}

func (h *Handler) SetTrace(ctx context.Context, params *protocol.SetTraceParams) error {
	h.logger.Info("SetTrace", slog.Any("params", params))
	return nil
}

func (h *Handler) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	h.logger.Info("CodeAction", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	h.logger.Info("CodeLens", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) CodeLensResolve(ctx context.Context, params *protocol.CodeLens) (*protocol.CodeLens, error) {
	h.logger.Info("CodeLensResolve", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) ColorPresentation(ctx context.Context, params *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
	h.logger.Info("ColorPresentation", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	h.logger.Info("Completion", slog.Any("params", params))
	defer func() {
		if err := recover(); err != nil {
			h.logger.Error("recovered", slog.Any("error", err))
		}
	}()
	return h.completion(ctx, params)
}

func (h *Handler) CompletionResolve(ctx context.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	h.logger.Info("CompletionResolve", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Declaration(ctx context.Context, params *protocol.DeclarationParams) ([]protocol.Location, error) {
	h.logger.Info("Declaration", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	h.logger.Info("Definition", slog.Any("params", params))
	defer func() {
		if err := recover(); err != nil {
			h.logger.Error("recovered", slog.Any("error", err))
		}
	}()
	return h.definition(ctx, params)
}

func (h *Handler) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	h.logger.Info("DidChange", slog.Any("params", params))
	defer func() {
		if err := recover(); err != nil {
			h.logger.Error("recovered", slog.Any("error", err))
		}
	}()
	return h.didChange(ctx, params)
}

func (h *Handler) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) error {
	h.logger.Info("DidChangeConfiguration", slog.Any("params", params))
	return nil
}

func (h *Handler) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	h.logger.Info("DidChangeWatchedFiles", slog.Any("params", params))
	return nil
}

func (h *Handler) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
	h.logger.Info("DidChangeWorkspaceFolders", slog.Any("params", params))
	return nil
}

func (h *Handler) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	h.logger.Info("DidClose", slog.Any("params", params))
	return nil
}

func (h *Handler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	h.logger.Info("DidOpen", slog.Any("params", params))
	return nil
}

func (h *Handler) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	h.logger.Info("DidSave", slog.Any("params", params))
	return nil
}

func (h *Handler) DocumentColor(ctx context.Context, params *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	h.logger.Info("DocumentColor", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DocumentHighlight(ctx context.Context, params *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	h.logger.Info("DocumentHighlight", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DocumentLink(ctx context.Context, params *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	h.logger.Info("DocumentLink", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DocumentLinkResolve(ctx context.Context, params *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	h.logger.Info("DocumentLinkResolve", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	h.logger.Info("DocumentSymbol", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	h.logger.Info("ExecuteCommand", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) FoldingRanges(ctx context.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	h.logger.Info("FoldingRanges", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	h.logger.Info("Formatting", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	h.logger.Info("Hover", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Implementation(ctx context.Context, params *protocol.ImplementationParams) ([]protocol.Location, error) {
	h.logger.Info("Implementation", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	h.logger.Info("OnTypeFormatting", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (*protocol.Range, error) {
	h.logger.Info("PrepareRename", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) RangeFormatting(ctx context.Context, params *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	h.logger.Info("RangeFormatting", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) References(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	h.logger.Info("References", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Rename(ctx context.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Info("Rename", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	h.logger.Info("SignatureHelp", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Symbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	h.logger.Info("Symbols", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) TypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	h.logger.Info("TypeDefinition", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	h.logger.Info("WillSave", slog.Any("params", params))
	return nil
}

func (h *Handler) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	h.logger.Info("WillSaveWaitUntil", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (*protocol.ShowDocumentResult, error) {
	h.logger.Info("ShowDocument", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) WillCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Info("WillCreateFiles", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) error {
	h.logger.Info("DidCreateFiles", slog.Any("params", params))
	return nil
}

func (h *Handler) WillRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Info("WillRenameFiles", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DidRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) error {
	h.logger.Info("DidRenameFiles", slog.Any("params", params))
	return nil
}

func (h *Handler) WillDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Info("WillDeleteFiles", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) error {
	h.logger.Info("DidDeleteFiles", slog.Any("params", params))
	return nil
}

func (h *Handler) CodeLensRefresh(ctx context.Context) error {
	h.logger.Info("CodeLensRefresh")
	return nil
}

func (h *Handler) PrepareCallHierarchy(ctx context.Context, params *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	h.logger.Info("PrepareCallHierarchy", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) IncomingCalls(ctx context.Context, params *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	h.logger.Info("IncomingCalls", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) OutgoingCalls(ctx context.Context, params *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	h.logger.Info("OutgoingCalls", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	h.logger.Info("SemanticTokensFull", slog.Any("params", params))
	defer func() {
		if err := recover(); err != nil {
			h.logger.Error("recovered", slog.Any("error", err))
		}
	}()
	return h.semanticTokensFull(params)
}

func (h *Handler) SemanticTokensFullDelta(ctx context.Context, params *protocol.SemanticTokensDeltaParams) (interface{}, error) {
	h.logger.Info("SemanticTokensFullDelta", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	h.logger.Info("SemanticTokensRange", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) SemanticTokensRefresh(ctx context.Context) error {
	h.logger.Info("SemanticTokensRefresh")
	return nil
}

func (h *Handler) LinkedEditingRange(ctx context.Context, params *protocol.LinkedEditingRangeParams) (*protocol.LinkedEditingRanges, error) {
	h.logger.Info("LinkedEditingRange", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Moniker(ctx context.Context, params *protocol.MonikerParams) ([]protocol.Moniker, error) {
	h.logger.Info("Moniker", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Request(ctx context.Context, method string, params interface{}) (interface{}, error) {
	h.logger.Info("Request", slog.Any("params", params))
	return nil, nil
}
