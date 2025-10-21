package resolver

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mercari/grpc-federation/grpc/federation"
)

func (f *File) PackageName() string {
	if f.Package == nil {
		return ""
	}
	return f.Package.Name
}

func (f *File) HasServiceWithRule() bool {
	for _, svc := range f.Services {
		if svc.Rule != nil {
			return true
		}
	}
	return false
}

func (f *File) Message(name string) *Message {
	for _, msg := range f.Messages {
		if msg.Name == name {
			return msg
		}
	}
	return nil
}

// AllEnums returns a list that includes the enums defined in the file itself.
func (f *File) AllEnums() []*Enum {
	enums := f.Enums
	for _, msg := range f.Messages {
		enums = append(enums, msg.AllEnums()...)
	}
	return enums
}

// AllEnumsIncludeDeps recursively searches for the imported file and returns a list of all enums.
func (f *File) AllEnumsIncludeDeps() []*Enum {
	return f.allEnumsIncludeDeps(make(map[string][]*Enum))
}

// AllUseMethods list of methods that the Services included in the file depend on.
func (f *File) AllUseMethods() []*Method {
	mtdMap := map[*Method]struct{}{}
	for _, svc := range f.Services {
		for _, mtd := range svc.UseMethods() {
			mtdMap[mtd] = struct{}{}
		}
	}
	mtds := make([]*Method, 0, len(mtdMap))
	for mtd := range mtdMap {
		mtds = append(mtds, mtd)
	}
	sort.Slice(mtds, func(i, j int) bool {
		return mtds[i].FQDN() < mtds[j].FQDN()
	})
	return mtds
}

func (f *File) AllCELPlugins() []*CELPlugin {
	pluginMap := make(map[string]*CELPlugin)
	for _, plugin := range f.CELPlugins {
		pluginMap[plugin.Name] = plugin
	}
	for _, file := range f.ImportFiles {
		for _, plugin := range file.AllCELPlugins() {
			pluginMap[plugin.Name] = plugin
		}
	}
	plugins := make([]*CELPlugin, 0, len(pluginMap))
	for _, plugin := range pluginMap {
		plugins = append(plugins, plugin)
	}
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})
	return plugins
}

func (f *File) allEnumsIncludeDeps(cacheMap map[string][]*Enum) []*Enum {
	if enums, exists := cacheMap[f.Name]; exists {
		return enums
	}
	enumMap := make(map[string]*Enum)
	for _, enum := range f.AllEnums() {
		enumMap[enum.FQDN()] = enum
	}
	for _, importFile := range f.ImportFiles {
		for _, enum := range importFile.allEnumsIncludeDeps(cacheMap) {
			enumMap[enum.FQDN()] = enum
		}
	}
	enums := make([]*Enum, 0, len(enumMap))
	for _, enum := range enumMap {
		enums = append(enums, enum)
	}
	sort.Slice(enums, func(i, j int) bool {
		return enums[i].FQDN() < enums[j].FQDN()
	})
	cacheMap[f.Name] = enums
	return enums
}

func (f *File) PrivatePackageName() string {
	return federation.PrivatePackageName + "." + f.PackageName()
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

func (r *OutputFilePathResolver) OutputDir(fileName string, gopkg *GoPackage) (string, error) {
	switch r.cfg.Mode {
	case ModulePrefixMode:
		return r.modulePrefixBasedOutputDir(gopkg)
	case SourceRelativeMode:
		return r.sourceRelativeBasedOutputDir(fileName)
	case ImportMode:
		return r.importBasedOutputDir(gopkg)
	}
	return "", fmt.Errorf("grpc-federation: unexpected output file mode: %d", r.cfg.Mode)
}

// OutputPath returns the path to the output file.
// Three output mode supported by protoc-gen-go are available.
// FYI: https://protobuf.dev/reference/go/go-generated.
func (r *OutputFilePathResolver) OutputPath(file *File) (string, error) {
	dir, err := r.OutputDir(file.Name, file.GoPackage)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, r.FileName(file)), nil
}

func (r *OutputFilePathResolver) importBasedOutputDir(gopkg *GoPackage) (string, error) {
	if gopkg == nil {
		return "", fmt.Errorf("grpc-federation: gopkg must be specified")
	}
	return gopkg.ImportPath, nil
}

// SourceRelativeBasedOutputPath returns the path to the output file when the `paths=source_relative` flag is specified.
// FYI: https://protobuf.dev/reference/go/go-generated.
func (r *OutputFilePathResolver) sourceRelativeBasedOutputDir(fileName string) (string, error) {
	filePath := fileName
	if r.cfg.FilePath != "" {
		filePath = r.cfg.FilePath
	}
	relativePath := r.relativePath(filePath)
	return filepath.Dir(relativePath), nil
}

// ModulePrefixBasedOutputPath returns the path to the output file when the `module=$PREFIX` flag is specified.
// FYI: https://protobuf.dev/reference/go/go-generated.
func (r *OutputFilePathResolver) modulePrefixBasedOutputDir(gopkg *GoPackage) (string, error) {
	if gopkg == nil {
		return "", fmt.Errorf("grpc-federation: gopkg must be specified")
	}
	var prefix string
	if r.cfg.Prefix != "" {
		prefix = r.cfg.Prefix
	}
	trimmedPrefix := strings.TrimPrefix(gopkg.ImportPath, prefix)
	trimmedSlash := strings.TrimPrefix(trimmedPrefix, "/")
	return trimmedSlash, nil
}

func IsGRPCFederationGeneratedFile(path string) bool {
	return strings.HasSuffix(path, "_grpc_federation.pb.go")
}

func (r *OutputFilePathResolver) FileName(file *File) string {
	baseNameWithExt := filepath.Base(file.Name)
	baseName := baseNameWithExt[:len(baseNameWithExt)-len(filepath.Ext(baseNameWithExt))]
	return fmt.Sprintf("%s_grpc_federation.pb.go", strings.ToLower(baseName))
}

func (r *OutputFilePathResolver) relativePath(filePath string) string {
	for _, importPath := range r.cfg.ImportPaths {
		rel, err := filepath.Rel(importPath, filePath)
		if err != nil {
			continue
		}
		if strings.HasPrefix(rel, "..") {
			continue
		}
		return rel
	}
	return filePath
}
