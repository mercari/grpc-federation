//nolint:gosec
package server

import (
	"context"
	"log/slog"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"github.com/mercari/grpc-federation/source"
	"github.com/mercari/grpc-federation/validator"
)

func (h *Handler) validateText(ctx context.Context, docURI uri.URI) {
	path, content, err := h.getFile(docURI)
	if err != nil {
		return
	}
	file, err := source.NewFile(path, content)
	if err != nil {
		return
	}
	errs := h.validator.Validate(ctx, file, validator.ImportPathOption(h.importPaths...))
	diagnostics := make([]protocol.Diagnostic, 0, len(errs))
	for _, err := range errs {
		start := protocol.Position{
			Line:      uint32(err.Start.Line) - 1,
			Character: uint32(err.Start.Col) - 1,
		}
		end := protocol.Position{
			Line:      uint32(err.End.Line) - 1,
			Character: uint32(err.End.Col) - 1,
		}
		var severity protocol.DiagnosticSeverity
		if err.IsWarning {
			severity = protocol.DiagnosticSeverityWarning
		} else {
			severity = protocol.DiagnosticSeverityError
		}
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Severity: severity,
			Range:    protocol.Range{Start: start, End: end},
			Message:  err.Message,
		})
	}
	if err := h.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
		URI:         docURI,
		Diagnostics: diagnostics,
	}); err != nil {
		h.logger.Error("failed to publish diagnostics", slog.String("error", err.Error()))
	}
}
