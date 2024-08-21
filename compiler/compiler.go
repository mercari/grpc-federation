package compiler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/protoutil"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/mercari/grpc-federation/proto/grpc/federation"
	"github.com/mercari/grpc-federation/proto_deps/google/rpc"
	"github.com/mercari/grpc-federation/source"

	_ "embed"
)

// Compiler provides a way to generate file descriptors from a Protocol Buffers file without relying on protoc command.
// This allows you to do things like validate proto files without relying on protoc.
type Compiler struct {
	importPaths  []string
	manualImport bool
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

// ManualImportOption stops importing `grpc/federation/federation.proto` file and its dependencies
// By default if `grpc/federation/federation.proto` file and its dependencies do not exist, automatically imports it.
// The version of the proto file is the same as the version when the compiler is built.
func ManualImportOption() Option {
	return func(c *Compiler) {
		c.manualImport = true
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

const (
	grpcFederationFilePath        = "grpc/federation/federation.proto"
	grpcFederationPrivateFilePath = "grpc/federation/private.proto"
	grpcFederationTimeFilePath    = "grpc/federation/time.proto"
	grpcFederationPluginFilePath  = "grpc/federation/plugin.proto"
	googleRPCCodeFilePath         = "google/rpc/code.proto"
	googleRPCErrorDetailsFilePath = "google/rpc/error_details.proto"
)

func RelativePathFromImportPaths(protoPath string, importPaths []string) (string, error) {
	if len(importPaths) == 0 {
		return protoPath, nil
	}

	absImportPaths := make([]string, 0, len(importPaths))
	for _, path := range importPaths {
		if filepath.IsAbs(path) {
			absImportPaths = append(absImportPaths, path)
		} else {
			abs, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("failed to get absolute path from %s: %w", path, err)
			}
			absImportPaths = append(absImportPaths, abs)
		}
	}

	absProtoPath := protoPath
	if !filepath.IsAbs(protoPath) {
		path, err := filepath.Abs(protoPath)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path from %s: %w", protoPath, err)
		}
		absProtoPath = path
	}

	for _, importPath := range absImportPaths {
		if strings.HasPrefix(absProtoPath, importPath) {
			relPath, err := filepath.Rel(importPath, absProtoPath)
			if err != nil {
				return "", fmt.Errorf("failed to get relative path from %s and %s: %w", importPath, absProtoPath, err)
			}
			return relPath, nil
		}
	}
	return protoPath, nil
}

type dependentProtoFileSet struct {
	path string
	data []byte
}

// Compile compile the target Protocol Buffers file and produces all file descriptors.
func (c *Compiler) Compile(ctx context.Context, file *source.File, opts ...Option) ([]*descriptorpb.FileDescriptorProto, error) {
	c.importPaths = []string{}

	for _, opt := range opts {
		opt(c)
	}

	relPath, err := RelativePathFromImportPaths(file.Path(), c.importPaths)
	if err != nil {
		return nil, err
	}

	var r errorReporter

	compiler := protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: append(c.importPaths, filepath.Dir(relPath), ""),
			Accessor: func(p string) (io.ReadCloser, error) {
				if p == file.Path() {
					return io.NopCloser(bytes.NewBuffer(file.Content())), nil
				}
				f, err := os.Open(p)
				if err != nil {
					for _, set := range []*dependentProtoFileSet{
						{path: grpcFederationFilePath, data: federation.ProtoFile},
						{path: grpcFederationPrivateFilePath, data: federation.PrivateProtoFile},
						{path: grpcFederationTimeFilePath, data: federation.TimeProtoFile},
						{path: grpcFederationPluginFilePath, data: federation.PluginProtoFile},
						{path: googleRPCCodeFilePath, data: rpc.GoogleRPCCodeProtoFile},
						{path: googleRPCErrorDetailsFilePath, data: rpc.GoogleRPCErrorDetailsProtoFile},
					} {
						if !c.manualImport && strings.HasSuffix(p, set.path) {
							return io.NopCloser(bytes.NewBuffer(set.data)), nil
						}
					}
					return nil, err
				}
				return f, nil
			},
		}),
		SourceInfoMode: protocompile.SourceInfoStandard,
		Reporter:       &r,
	}
	files := []string{relPath}
	files = append(files, file.Imports()...)
	linkedFiles, err := compiler.Compile(ctx, files...)
	if err != nil {
		return nil, &CompilerError{Err: err, ErrWithPos: r.errs}
	}
	protoFiles := c.getProtoFiles(linkedFiles)
	return protoFiles, nil
}

func (c *Compiler) getProtoFiles(linkedFiles []linker.File) []*descriptorpb.FileDescriptorProto {
	var (
		protos         []*descriptorpb.FileDescriptorProto
		protoUniqueMap = make(map[string]struct{})
	)
	for _, linkedFile := range linkedFiles {
		for _, proto := range c.getFileDescriptors(linkedFile) {
			if _, exists := protoUniqueMap[proto.GetName()]; exists {
				continue
			}
			protos = append(protos, proto)
			protoUniqueMap[proto.GetName()] = struct{}{}
		}
	}
	return protos
}

func (c *Compiler) getFileDescriptors(file protoreflect.FileDescriptor) []*descriptorpb.FileDescriptorProto {
	var protoFiles []*descriptorpb.FileDescriptorProto
	fileImports := file.Imports()
	for i := 0; i < fileImports.Len(); i++ {
		protoFiles = append(protoFiles, c.getFileDescriptors(fileImports.Get(i))...)
	}
	protoFiles = append(protoFiles, protoutil.ProtoFromFileDescriptor(file))
	return protoFiles
}
