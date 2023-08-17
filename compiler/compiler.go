package compiler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/protoutil"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/source"
)

// Compiler provides a way to generate file descriptors from a Protocol Buffers file without relying on protoc command.
// This allows you to do things like validate proto files without relying on protoc.
type Compiler struct {
	importPaths []string
}

// Option represents compiler option.
type Option func(*Compiler)

// ImportPathOption used to add a path to reference a proto file.
// By default, only the directory where the starting file exists is added to the import path.
func ImportPathOption(path ...string) Option {
	return func(c *Compiler) {
		c.importPaths = append(c.importPaths, path...)
	}
}

// New creates compiler instance.
func New() *Compiler {
	return &Compiler{}
}

type errorReporter struct {
	errs []reporter.ErrorWithPos
}

func (r *errorReporter) Error(err reporter.ErrorWithPos) error {
	r.errs = append(r.errs, err)
	return nil
}

func (r *errorReporter) Warning(_ reporter.ErrorWithPos) {
}

// CompilerError has error with source position.
type CompilerError struct {
	Err        error
	ErrWithPos []reporter.ErrorWithPos
}

func (e *CompilerError) Error() string {
	if len(e.ErrWithPos) == 0 {
		return e.Err.Error()
	}
	var errs []string
	for _, err := range e.ErrWithPos {
		errs = append(errs, err.Error())
	}
	return fmt.Sprintf("%s\n%s", e.Err.Error(), strings.Join(errs, "\n"))
}

// Compile compile the target Protocol Buffers file and produces all file descriptors.
func (c *Compiler) Compile(ctx context.Context, file *source.File, opts ...Option) ([]*descriptorpb.FileDescriptorProto, error) {
	c.importPaths = c.importPaths[:]
	for _, opt := range opts {
		opt(c)
	}

	path := file.Path()
	dirName := filepath.Dir(path)
	fileName := filepath.Base(path)

	var r errorReporter

	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: append(c.importPaths, dirName),
			Accessor: func(p string) (io.ReadCloser, error) {
				if path == p {
					return io.NopCloser(bytes.NewBuffer(file.Content())), nil
				}
				return os.Open(p)
			},
		}),
		SourceInfoMode: protocompile.SourceInfoStandard,
		Reporter:       &r,
	}
	files := []string{fileName}
	files = append(files, file.Imports()...)
	linkedFiles, err := compiler.Compile(ctx, files...)
	if err != nil {
		return nil, &CompilerError{Err: err, ErrWithPos: r.errs}
	}
	protoFiles := c.getProtoFiles(linkedFiles)
	return protoFiles, nil
}

func (c *Compiler) getProtoFiles(linkedFiles []linker.File) []*descriptorpb.FileDescriptorProto {
	var allProtoFiles []*descriptorpb.FileDescriptorProto
	for _, linkedFile := range linkedFiles {
		allProtoFiles = append(allProtoFiles, c.getFileDescriptors(linkedFile)...)
	}
	protoFileUniqueMap := map[string]*descriptorpb.FileDescriptorProto{}
	for _, file := range allProtoFiles {
		protoFileUniqueMap[file.GetName()] = file
	}
	protoFiles := make([]*descriptorpb.FileDescriptorProto, 0, len(protoFileUniqueMap))
	for _, file := range protoFileUniqueMap {
		protoFiles = append(protoFiles, file)
	}
	sort.Slice(protoFiles, func(i, j int) bool {
		return protoFiles[i].GetName() < protoFiles[j].GetName()
	})
	return protoFiles
}

func (c *Compiler) getFileDescriptors(file protoreflect.FileDescriptor) []*descriptorpb.FileDescriptorProto {
	protoFiles := []*descriptorpb.FileDescriptorProto{protoutil.ProtoFromFileDescriptor(file)}
	fileImports := file.Imports()
	for i := 0; i < fileImports.Len(); i++ {
		protoFiles = append(protoFiles, c.getFileDescriptors(fileImports.Get(i))...)
	}
	return protoFiles
}
