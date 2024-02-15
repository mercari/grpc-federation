package server

import (
	"bufio"
	"context"
	"io"
	"os"

	"go.lsp.dev/jsonrpc2"
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

type readerWrapper struct {
	readCloser io.ReadCloser
	reader     *bufio.Reader
}

func newReaderWrapper(readCloser io.ReadCloser) *readerWrapper {
	return &readerWrapper{
		readCloser: readCloser,
		reader:     bufio.NewReaderSize(readCloser, 8192),
	}
}

func (r *readerWrapper) Read(b []byte) (int, error) {
	return r.reader.Read(b)
}

func (r *readerWrapper) Close() error {
	return r.readCloser.Close()
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
				readCloser:  newReaderWrapper(os.Stdin),
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

func (s *Server) Run(ctx context.Context) {
	s.conn.Go(
		ctx,
		protocol.ServerHandler(
			s.handler,
			nil,
		),
	)
	<-s.conn.Done()
}
