package server

import (
	"context"
	"io"
	"log/slog"
	"sync"

	"go.lsp.dev/protocol"

	"github.com/mercari/grpc-federation/validator"
)

var _ protocol.Server = &Handler{}

type Handler struct {
	logger                        *slog.Logger
	importPaths                   []string
	client                        protocol.Client
	validator                     *validator.Validator
	completer                     *Completer
	fileCacheMu                   sync.RWMutex
	fileCacheMap                  map[string]*File
	tokenTypeMap                  map[string]uint32
	tokenModifierMap              map[string]uint32
	supportedDefinitionLinkClient bool
}

func NewHandler(client protocol.Client, w io.Writer, importPaths []string) *Handler {
	logger := slog.New(slog.NewJSONHandler(w, nil))
	return &Handler{
		logger:           logger,
		importPaths:      importPaths,
		client:           client,
		completer:        NewCompleter(logger),
		validator:        validator.New(),
		fileCacheMap:     make(map[string]*File),
		tokenTypeMap:     make(map[string]uint32),
		tokenModifierMap: make(map[string]uint32),
	}
}

func (h *Handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	h.logger.Debug("Initialize", slog.Any("params", params))
	return h.initialize(params)
}

func (h *Handler) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	h.logger.Debug("Initialized", slog.Any("params", params))
	return nil
}

func (h *Handler) Shutdown(ctx context.Context) error {
	h.logger.Debug("Shutdown")
	return nil
}

func (h *Handler) Exit(ctx context.Context) error {
	h.logger.Debug("Exit")
	return nil
}

func (h *Handler) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
	h.logger.Debug("WorkDoneProgressCancel", slog.Any("params", params))
	return nil
}

func (h *Handler) LogTrace(ctx context.Context, params *protocol.LogTraceParams) error {
	h.logger.Debug("LogTrace", slog.Any("params", params))
	return nil
}

func (h *Handler) SetTrace(ctx context.Context, params *protocol.SetTraceParams) error {
	h.logger.Debug("SetTrace", slog.Any("params", params))
	return nil
}

func (h *Handler) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	h.logger.Debug("CodeAction", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	h.logger.Debug("CodeLens", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) CodeLensResolve(ctx context.Context, params *protocol.CodeLens) (*protocol.CodeLens, error) {
	h.logger.Debug("CodeLensResolve", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) ColorPresentation(ctx context.Context, params *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
	h.logger.Debug("ColorPresentation", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	h.logger.Debug("Completion", slog.Any("params", params))
	defer func() {
		if err := recover(); err != nil {
			h.logger.Error("recovered", slog.Any("error", err))
		}
	}()
	return h.completion(ctx, params)
}

func (h *Handler) CompletionResolve(ctx context.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	h.logger.Debug("CompletionResolve", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Declaration(ctx context.Context, params *protocol.DeclarationParams) ([]protocol.Location, error) {
	h.logger.Debug("Declaration", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	h.logger.Debug("Definition", slog.Any("params", params))
	defer func() {
		if err := recover(); err != nil {
			h.logger.Error("recovered", slog.Any("error", err))
		}
	}()
	return h.definition(ctx, params)
}

func (h *Handler) DefinitionWithLink(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.LocationLink, error) {
	h.logger.Debug("DefinitionWithLink", slog.Any("params", params))
	return h.definitionWithLink(ctx, params)
}

func (h *Handler) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	h.logger.Debug("DidChange", slog.Any("params", params))
	defer func() {
		if err := recover(); err != nil {
			h.logger.Error("recovered", slog.Any("error", err))
		}
	}()
	return h.didChange(ctx, params)
}

func (h *Handler) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) error {
	h.logger.Debug("DidChangeConfiguration", slog.Any("params", params))
	return nil
}

func (h *Handler) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	h.logger.Debug("DidChangeWatchedFiles", slog.Any("params", params))
	return nil
}

func (h *Handler) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
	h.logger.Debug("DidChangeWorkspaceFolders", slog.Any("params", params))
	return nil
}

func (h *Handler) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	h.logger.Debug("DidClose", slog.Any("params", params))
	return nil
}

func (h *Handler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	h.logger.Debug("DidOpen", slog.Any("params", params))
	return nil
}

func (h *Handler) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	h.logger.Debug("DidSave", slog.Any("params", params))
	return nil
}

func (h *Handler) DocumentColor(ctx context.Context, params *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	h.logger.Debug("DocumentColor", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DocumentHighlight(ctx context.Context, params *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	h.logger.Debug("DocumentHighlight", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DocumentLink(ctx context.Context, params *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	h.logger.Debug("DocumentLink", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DocumentLinkResolve(ctx context.Context, params *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	h.logger.Debug("DocumentLinkResolve", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	h.logger.Debug("DocumentSymbol", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	h.logger.Debug("ExecuteCommand", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) FoldingRanges(ctx context.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	h.logger.Debug("FoldingRanges", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	h.logger.Debug("Formatting", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	h.logger.Debug("Hover", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Implementation(ctx context.Context, params *protocol.ImplementationParams) ([]protocol.Location, error) {
	h.logger.Debug("Implementation", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	h.logger.Debug("OnTypeFormatting", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (*protocol.Range, error) {
	h.logger.Debug("PrepareRename", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) RangeFormatting(ctx context.Context, params *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	h.logger.Debug("RangeFormatting", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) References(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	h.logger.Debug("References", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Rename(ctx context.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Debug("Rename", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	h.logger.Debug("SignatureHelp", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Symbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	h.logger.Debug("Symbols", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) TypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	h.logger.Debug("TypeDefinition", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	h.logger.Debug("WillSave", slog.Any("params", params))
	return nil
}

func (h *Handler) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	h.logger.Debug("WillSaveWaitUntil", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (*protocol.ShowDocumentResult, error) {
	h.logger.Debug("ShowDocument", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) WillCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Debug("WillCreateFiles", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) error {
	h.logger.Debug("DidCreateFiles", slog.Any("params", params))
	return nil
}

func (h *Handler) WillRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Debug("WillRenameFiles", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DidRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) error {
	h.logger.Debug("DidRenameFiles", slog.Any("params", params))
	return nil
}

func (h *Handler) WillDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Debug("WillDeleteFiles", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) error {
	h.logger.Debug("DidDeleteFiles", slog.Any("params", params))
	return nil
}

func (h *Handler) CodeLensRefresh(ctx context.Context) error {
	h.logger.Debug("CodeLensRefresh")
	return nil
}

func (h *Handler) PrepareCallHierarchy(ctx context.Context, params *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	h.logger.Debug("PrepareCallHierarchy", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) IncomingCalls(ctx context.Context, params *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	h.logger.Debug("IncomingCalls", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) OutgoingCalls(ctx context.Context, params *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	h.logger.Debug("OutgoingCalls", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	h.logger.Debug("SemanticTokensFull", slog.Any("params", params))
	defer func() {
		if err := recover(); err != nil {
			h.logger.Error("recovered", slog.Any("error", err))
		}
	}()
	return h.semanticTokensFull(params)
}

func (h *Handler) SemanticTokensFullDelta(ctx context.Context, params *protocol.SemanticTokensDeltaParams) (interface{}, error) {
	h.logger.Debug("SemanticTokensFullDelta", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	h.logger.Debug("SemanticTokensRange", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) SemanticTokensRefresh(ctx context.Context) error {
	h.logger.Debug("SemanticTokensRefresh")
	return nil
}

func (h *Handler) LinkedEditingRange(ctx context.Context, params *protocol.LinkedEditingRangeParams) (*protocol.LinkedEditingRanges, error) {
	h.logger.Debug("LinkedEditingRange", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Moniker(ctx context.Context, params *protocol.MonikerParams) ([]protocol.Moniker, error) {
	h.logger.Debug("Moniker", slog.Any("params", params))
	return nil, nil
}

func (h *Handler) Request(ctx context.Context, method string, params interface{}) (interface{}, error) {
	h.logger.Debug("Request", slog.Any("params", params))
	return nil, nil
}
