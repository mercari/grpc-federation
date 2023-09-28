package server

import (
	"context"
	"io"
	"log"
	"sync"

	"github.com/k0kubun/pp/v3"
	"go.lsp.dev/protocol"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/validator"
)

var _ protocol.Server = &Handler{}

type Handler struct {
	pp               *pp.PrettyPrinter
	logger           *log.Logger
	importPaths      []string
	client           protocol.Client
	compiler         *compiler.Compiler
	validator        *validator.Validator
	completer        *Completer
	fileContentMu    sync.Mutex
	fileToContentMap map[string][]byte
	tokenTypeMap     map[string]uint32
}

func NewHandler(client protocol.Client, w io.Writer, importPaths []string) *Handler {
	logger := log.New(w, "", 0)
	printer := pp.New()
	printer.SetColoringEnabled(false)
	printer.SetOutput(w)
	c := compiler.New()
	return &Handler{
		pp:               printer,
		logger:           logger,
		importPaths:      importPaths,
		client:           client,
		compiler:         c,
		completer:        NewCompleter(c, logger, printer),
		validator:        validator.New(),
		fileToContentMap: map[string][]byte{},
		tokenTypeMap:     map[string]uint32{},
	}
}

func (h *Handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	h.logger.Println("Initialize", params)
	return h.initialize(params)
}

func (h *Handler) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	h.logger.Println("Initialized", params)
	return nil
}

func (h *Handler) Shutdown(ctx context.Context) error {
	h.logger.Println("Shutdown")
	return nil
}

func (h *Handler) Exit(ctx context.Context) error {
	h.logger.Println("Exit")
	return nil
}

func (h *Handler) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
	h.logger.Println("WorkDoneProgressCancel", params)
	return nil
}

func (h *Handler) LogTrace(ctx context.Context, params *protocol.LogTraceParams) error {
	h.logger.Println("LogTrace", params)
	return nil
}

func (h *Handler) SetTrace(ctx context.Context, params *protocol.SetTraceParams) error {
	h.logger.Println("SetTrace", params)
	return nil
}

func (h *Handler) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	h.logger.Println("CodeAction", params)
	return nil, nil
}

func (h *Handler) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	h.logger.Println("CodeLens", params)
	return nil, nil
}

func (h *Handler) CodeLensResolve(ctx context.Context, params *protocol.CodeLens) (*protocol.CodeLens, error) {
	h.logger.Println("CodeLensResolve", params)
	return nil, nil
}

func (h *Handler) ColorPresentation(ctx context.Context, params *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
	h.logger.Println("ColorPresentation", params)
	return nil, nil
}

func (h *Handler) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	h.logger.Println("Completion", params)
	return h.completion(ctx, params)
}

func (h *Handler) CompletionResolve(ctx context.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	h.logger.Println("CompletionResolve", params)
	return nil, nil
}

func (h *Handler) Declaration(ctx context.Context, params *protocol.DeclarationParams) ([]protocol.Location, error) {
	h.logger.Println("Declaration", params)
	return nil, nil
}

func (h *Handler) Definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	h.logger.Println("Definition", params)
	return h.definition(ctx, params)
}

func (h *Handler) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	h.logger.Println("didChange")
	return h.didChange(ctx, params)
}

func (h *Handler) DidChangeConfiguration(ctx context.Context, params *protocol.DidChangeConfigurationParams) error {
	h.logger.Println("DidChangeConfiguration", params)
	return nil
}

func (h *Handler) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	h.logger.Println("DidChangeWatchedFiles", params)
	return nil
}

func (h *Handler) DidChangeWorkspaceFolders(ctx context.Context, params *protocol.DidChangeWorkspaceFoldersParams) error {
	h.logger.Println("DidChangeWorkspaceFolders", params)
	return nil
}

func (h *Handler) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	h.logger.Println("DidClose", params)
	return nil
}

func (h *Handler) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	h.logger.Println("DidOpen", params)
	return nil
}

func (h *Handler) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	h.logger.Println("DidSave", params)
	return nil
}

func (h *Handler) DocumentColor(ctx context.Context, params *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	h.logger.Println("DocumentColor", params)
	return nil, nil
}

func (h *Handler) DocumentHighlight(ctx context.Context, params *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	h.logger.Println("DocumentHighlight", params)
	return nil, nil
}

func (h *Handler) DocumentLink(ctx context.Context, params *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	h.logger.Println("DocumentLink", params)
	return nil, nil
}

func (h *Handler) DocumentLinkResolve(ctx context.Context, params *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	h.logger.Println("DocumentLinkResolve", params)
	return nil, nil
}

func (h *Handler) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	h.logger.Println("DocumentSymbol", params)
	return nil, nil
}

func (h *Handler) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	h.logger.Println("ExecuteCommand", params)
	return nil, nil
}

func (h *Handler) FoldingRanges(ctx context.Context, params *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	h.logger.Println("FoldingRanges", params)
	return nil, nil
}

func (h *Handler) Formatting(ctx context.Context, params *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	h.logger.Println("Formatting", params)
	return nil, nil
}

func (h *Handler) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	h.logger.Println("Hover", params)
	return nil, nil
}

func (h *Handler) Implementation(ctx context.Context, params *protocol.ImplementationParams) ([]protocol.Location, error) {
	h.logger.Println("Implementation", params)
	return nil, nil
}

func (h *Handler) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	h.logger.Println("OnTypeFormatting", params)
	return nil, nil
}

func (h *Handler) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (*protocol.Range, error) {
	h.logger.Println("PrepareRename", params)
	return nil, nil
}

func (h *Handler) RangeFormatting(ctx context.Context, params *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	h.logger.Println("RangeFormatting", params)
	return nil, nil
}

func (h *Handler) References(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	h.logger.Println("References", params)
	return nil, nil
}

func (h *Handler) Rename(ctx context.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Println("Rename", params)
	return nil, nil
}

func (h *Handler) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	h.logger.Println("SignatureHelp", params)
	return nil, nil
}

func (h *Handler) Symbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	h.logger.Println("Symbols", params)
	return nil, nil
}

func (h *Handler) TypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	h.logger.Println("TypeDefinition", params)
	return nil, nil
}

func (h *Handler) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	h.logger.Println("WillSave", params)
	return nil
}

func (h *Handler) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	h.logger.Println("WillSaveWaitUntil", params)
	return nil, nil
}

func (h *Handler) ShowDocument(ctx context.Context, params *protocol.ShowDocumentParams) (*protocol.ShowDocumentResult, error) {
	h.logger.Println("ShowDocument", params)
	return nil, nil
}

func (h *Handler) WillCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Println("WillCreateFiles", params)
	return nil, nil
}

func (h *Handler) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) error {
	h.logger.Println("DidCreateFiles", params)
	return nil
}

func (h *Handler) WillRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Println("WillRenameFiles", params)
	return nil, nil
}

func (h *Handler) DidRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) error {
	h.logger.Println("DidRenameFiles", params)
	return nil
}

func (h *Handler) WillDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (*protocol.WorkspaceEdit, error) {
	h.logger.Println("WillDeleteFiles", params)
	return nil, nil
}

func (h *Handler) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) error {
	h.logger.Println("DidDeleteFiles", params)
	return nil
}

func (h *Handler) CodeLensRefresh(ctx context.Context) error {
	h.logger.Println("CodeLensRefresh")
	return nil
}

func (h *Handler) PrepareCallHierarchy(ctx context.Context, params *protocol.CallHierarchyPrepareParams) ([]protocol.CallHierarchyItem, error) {
	h.logger.Println("PrepareCallHierarchy", params)
	return nil, nil
}

func (h *Handler) IncomingCalls(ctx context.Context, params *protocol.CallHierarchyIncomingCallsParams) ([]protocol.CallHierarchyIncomingCall, error) {
	h.logger.Println("IncomingCalls", params)
	return nil, nil
}

func (h *Handler) OutgoingCalls(ctx context.Context, params *protocol.CallHierarchyOutgoingCallsParams) ([]protocol.CallHierarchyOutgoingCall, error) {
	h.logger.Println("OutgoingCalls", params)
	return nil, nil
}

func (h *Handler) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	h.logger.Println("SemanticTokensFull", params)
	return h.semanticTokensFull(params)
}

func (h *Handler) SemanticTokensFullDelta(ctx context.Context, params *protocol.SemanticTokensDeltaParams) (interface{}, error) {
	h.logger.Println("SemanticTokensFullDelta", params)
	return nil, nil
}

func (h *Handler) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	h.logger.Println("SemanticTokensRange", params)
	return nil, nil
}

func (h *Handler) SemanticTokensRefresh(ctx context.Context) error {
	h.logger.Println("SemanticTokensRefresh")
	return nil
}

func (h *Handler) LinkedEditingRange(ctx context.Context, params *protocol.LinkedEditingRangeParams) (*protocol.LinkedEditingRanges, error) {
	h.logger.Println("LinkedEditingRange", params)
	return nil, nil
}

func (h *Handler) Moniker(ctx context.Context, params *protocol.MonikerParams) ([]protocol.Moniker, error) {
	h.logger.Println("Moniker", params)
	return nil, nil
}

func (h *Handler) Request(ctx context.Context, method string, params interface{}) (interface{}, error) {
	h.logger.Println("Request", params)
	return nil, nil
}
