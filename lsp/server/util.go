package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"go.lsp.dev/uri"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/compiler"
	"github.com/mercari/grpc-federation/source"
)

type File struct {
	path     string
	src      *source.File
	descs    []*descriptorpb.FileDescriptorProto
	compiled chan struct{}
}

func newFile(path string, src *source.File, importPaths []string) *File {
	ret := &File{
		path:     path,
		src:      src,
		compiled: make(chan struct{}),
	}
	go func() {
		defer close(ret.compiled)
		descs, err := compiler.New().Compile(
			context.Background(),
			src,
			compiler.ImportPathOption(importPaths...),
			compiler.ImportRuleOption(),
		)
		if err != nil {
			return
		}
		ret.descs = descs
	}()
	return ret
}

func (f *File) getPath() string {
	return f.path
}

func (f *File) getSource() *source.File {
	return f.src
}

func (f *File) getCompiledProtos() []*descriptorpb.FileDescriptorProto {
	<-f.compiled
	return f.descs
}

func (h *Handler) getFile(docURI uri.URI) (*File, error) {
	parsedURI, err := url.ParseRequestURI(string(docURI))
	if err != nil {
		return nil, err
	}
	if parsedURI.Scheme != uri.FileScheme {
		return nil, fmt.Errorf("text document uri must specify the %s scheme, got %v", uri.FileScheme, parsedURI.Scheme)
	}
	path := parsedURI.Path
	return h.getFileByPath(path)
}

func (h *Handler) getFileByPath(path string) (*File, error) {
	if file := h.getCachedFile(path); file != nil {
		return file, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return h.setFileCache(path, content)
}

func (h *Handler) getCachedFile(path string) *File {
	h.fileCacheMu.RLock()
	defer h.fileCacheMu.RUnlock()

	return h.fileCacheMap[path]
}

func (h *Handler) setFileCache(path string, content []byte) (*File, error) {
	h.fileCacheMu.Lock()
	defer h.fileCacheMu.Unlock()

	h.logger.Debug("set file cache", slog.String("path", path), slog.Any("importPaths", h.importPaths))

	srcFile, err := source.NewFile(path, content)
	if err != nil {
		return nil, err
	}
	ret := newFile(path, srcFile, h.importPaths)
	h.fileCacheMap[path] = ret
	return ret, nil
}
