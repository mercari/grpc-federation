package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/pkg/xcontext"
	"go.lsp.dev/protocol"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type Server struct {
	conn    jsonrpc2.Conn
	logFile *os.File
	paths   []string
	handler *Handler
}

type readWriteCloser struct {
	readCloser  io.ReadCloser
	writeCloser io.WriteCloser
}

func (r *readWriteCloser) Read(b []byte) (int, error) {
	return r.readCloser.Read(b)
}

func (r *readWriteCloser) Write(b []byte) (int, error) {
	return r.writeCloser.Write(b)
}

func (r *readWriteCloser) Close() error {
	return multierr.Append(r.readCloser.Close(), r.writeCloser.Close())
}

type ServerOption func(*Server)

func LogFileOption(f *os.File) ServerOption {
	return func(s *Server) {
		s.logFile = f
	}
}

func ImportPathsOption(paths []string) ServerOption {
	return func(s *Server) {
		s.paths = paths
	}
}

func New(opts ...ServerOption) *Server {
	conn := jsonrpc2.NewConn(
		jsonrpc2.NewStream(
			&readWriteCloser{
				readCloser:  os.Stdin,
				writeCloser: os.Stdout,
			},
		),
	)
	client := protocol.ClientDispatcher(conn, zap.NewNop())

	server := &Server{conn: conn}
	for _, opt := range opts {
		opt(server)
	}
	writer := os.Stderr
	if server.logFile != nil {
		writer = server.logFile
	}
	server.handler = NewHandler(client, writer, server.paths)
	return server
}

func (s *Server) customServerHandler() jsonrpc2.Handler {
	orgHandler := protocol.ServerHandler(s.handler, nil)

	return func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		defer func() {
			if err := recover(); err != nil {
				s.handler.logger.Error(
					"recovered",
					slog.Any("error", err),
					slog.String("stack_trace", string(debug.Stack())),
				)
			}
		}()
		if ctx.Err() != nil {
			xctx := xcontext.Detach(ctx)
			return reply(xctx, nil, protocol.ErrRequestCancelled)
		}

		switch req.Method() {
		case protocol.MethodTextDocumentDefinition:
			if s.handler.supportedDefinitionLinkClient {
				var params protocol.DefinitionParams
				if err := json.NewDecoder(bytes.NewReader(req.Params())).Decode(&params); err != nil {
					return reply(ctx, nil, fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err))
				}
				resp, err := s.handler.DefinitionWithLink(ctx, &params)
				return reply(ctx, resp, err)
			}
		}
		return orgHandler(ctx, reply, req)
	}
}

func (s *Server) Run(ctx context.Context) {
	s.conn.Go(ctx, s.customServerHandler())
	<-s.conn.Done()
}
