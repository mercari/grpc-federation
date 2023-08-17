package resolver

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (f *File) PackageName() string {
	if f.Package == nil {
		return ""
	}
	return f.Package.Name
}

func (f *File) Message(name string) *Message {
	for _, msg := range f.Messages {
		if msg.Name == name {
			return msg
		}
	}
	return nil
}

func (f Files) FindByPackageName(pkg string) Files {
	var ret Files
	for _, file := range f {
		if file.PackageName() == pkg {
			ret = append(ret, file)
		}
	}
	return ret
}

type OutputFilePathResolver struct {
	cfg *OutputFilePathConfig
}

func NewOutputFilePathResolver(cfg OutputFilePathConfig) *OutputFilePathResolver {
	return &OutputFilePathResolver{
		cfg: &cfg,
	}
}

type OutputFilePathMode int

const (
	ImportMode         OutputFilePathMode = 0
	ModulePrefixMode   OutputFilePathMode = 1
	SourceRelativeMode OutputFilePathMode = 2
)

type OutputFilePathConfig struct {
	// Mode for file output ( default: ImportMode ).
	Mode OutputFilePathMode
	// Prefix used in ModulePrefixMode.
	Prefix string
	// FilePath specify if you know the file path specified at compile time.
	FilePath string
	// ImportPaths list of import paths used during compile.
	ImportPaths []string
}

// OutputPath returns the path to the output file.
// Three output mode supported by protoc-gen-go are available.
// FYI: https://protobuf.dev/reference/go/go-generated.
func (r *OutputFilePathResolver) OutputPath(svc *Service) (string, error) {
	switch r.cfg.Mode {
	case ModulePrefixMode:
		return r.modulePrefixBasedOutputPath(svc)
	case SourceRelativeMode:
		return r.sourceRelativeBasedOutputPath(svc)
	case ImportMode:
		return r.importBasedOutputPath(svc)
	}
	return "", fmt.Errorf("grpc-federation: unexpected output file mode: %d", r.cfg.Mode)
}

func (r *OutputFilePathResolver) importBasedOutputPath(svc *Service) (string, error) {
	if svc.File.GoPackage == nil {
		return "", fmt.Errorf("grpc-federation: gopkg must be specified")
	}
	return filepath.Join(svc.File.GoPackage.ImportPath, r.fileName(svc)), nil
}

// SourceRelativeBasedOutputPath returns the path to the output file when the `paths=source_relative` flag is specified.
// FYI: https://protobuf.dev/reference/go/go-generated.
func (r *OutputFilePathResolver) sourceRelativeBasedOutputPath(svc *Service) (string, error) {
	filePath := svc.File.Name
	if r.cfg.FilePath != "" {
		filePath = r.cfg.FilePath
	}
	relativePath, err := r.relativePath(filePath)
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(relativePath), r.fileName(svc)), nil
}

// ModulePrefixBasedOutputPath returns the path to the output file when the `module=$PREFIX` flag is specified.
// FYI: https://protobuf.dev/reference/go/go-generated.
func (r *OutputFilePathResolver) modulePrefixBasedOutputPath(svc *Service) (string, error) {
	if svc.File.GoPackage == nil {
		return "", fmt.Errorf("grpc-federation: gopkg must be specified")
	}
	var prefix string
	if r.cfg.Prefix != "" {
		prefix = r.cfg.Prefix
	}
	trimmedPrefix := strings.TrimPrefix(svc.File.GoPackage.ImportPath, prefix)
	trimmedSlash := strings.TrimPrefix(trimmedPrefix, "/")
	return filepath.Join(trimmedSlash, r.fileName(svc)), nil
}

func (r *OutputFilePathResolver) fileName(svc *Service) string {
	return fmt.Sprintf("%s_grpc_federation.go", strings.ToLower(svc.Name))
}

func (r *OutputFilePathResolver) relativePath(filePath string) (string, error) {
	if len(r.cfg.ImportPaths) == 0 {
		return filePath, nil
	}
	for _, importPath := range r.cfg.ImportPaths {
		rel, err := filepath.Rel(importPath, filePath)
		if err != nil {
			continue
		}
		if strings.HasPrefix(rel, "..") {
			continue
		}
		return rel, nil
	}
	return "", fmt.Errorf("grpc-federation: failed to find relative path from %s", filePath)
}
