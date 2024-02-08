package generator

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type Config struct {
	// Imports specify list of import path. This is the same as the list of paths specified by protoc's '-I' option.
	Imports []string `yaml:"imports"`
	// Src specifies the directory to be monitored when watch mode is enabled.
	Src []string `yaml:"src"`
	// Out specify the output destination for automatically generated code.
	Out string `yaml:"out"`
	// Plugins specify protoc's plugin configuration.
	Plugins []*PluginConfig `yaml:"plugins"`
	// AutoProtocGenGo automatically run protoc-gen-go at the time of editing proto. default is true.
	AutoProtocGenGo *bool `yaml:"autoProtocGenGo"`
	// AutoProtocGenGoGRPC automatically run protoc-gen-go-grpc at the time of editing proto. default is true.
	AutoProtocGenGoGRPC *bool `yaml:"autoProtocGenGoGrpc"`
}

type PluginConfig struct {
	// Plugin name of the protoc plugin.
	// If the name of the plugin is 'protoc-gen-go', write 'go'. ('protoc-gen-' prefix can be omitted).
	Plugin string `yaml:"plugin"`
	// Option specify options to be passed protoc plugin.
	Opt           *PluginOption `yaml:"opt"`
	installedPath string
}

type PluginOption struct {
	Opts []string
}

func (o *PluginOption) String() string {
	if o == nil {
		return ""
	}
	return strings.Join(o.Opts, ",")
}

func (o *PluginOption) MarshalYAML() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return []byte(o.String()), nil
}

func (o *PluginOption) UnmarshalYAML(b []byte) error {
	{
		var v []string
		if err := yaml.Unmarshal(b, &v); err == nil {
			o.Opts = append(o.Opts, v...)
			return nil
		}
	}
	var v string
	if err := yaml.Unmarshal(b, &v); err != nil {
		return err
	}
	o.Opts = append(o.Opts, v)
	return nil
}

func (c *Config) GetAutoProtocGenGo() bool {
	if c.AutoProtocGenGo == nil {
		return true
	}
	return *c.AutoProtocGenGo
}

func (c *Config) GetAutoProtocGenGoGRPC() bool {
	if c.AutoProtocGenGoGRPC == nil {
		return true
	}
	return *c.AutoProtocGenGoGRPC
}

func LoadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	return LoadConfigFromReader(f)
}

func LoadConfigFromReader(r io.Reader) (Config, error) {
	var cfg Config
	content, err := io.ReadAll(r)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return cfg, err
	}
	if err := setupConfig(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func setupConfig(cfg *Config) error {
	standardPluginMap := createStandardPluginMap(cfg)
	pluginNameMap := make(map[string]struct{})
	for _, plugin := range cfg.Plugins {
		name := plugin.Plugin
		if !strings.HasPrefix(name, "protoc-gen-") {
			name = "protoc-gen-" + name
		}
		path, err := lookupProtocPlugin(name, standardPluginMap)
		if err != nil {
			return fmt.Errorf(`failed to find protoc plugin "%s" on your system: %w`, name, err)
		}
		plugin.Plugin = name
		plugin.installedPath = path
		pluginNameMap[name] = struct{}{}
	}
	for standardPlugin, use := range standardPluginMap {
		if _, exists := pluginNameMap[standardPlugin]; !exists && use {
			addStandardPlugin(cfg, standardPlugin)
		}
	}
	if cfg.Out == "" {
		cfg.Out = "."
	}
	out, err := filepath.Abs(cfg.Out)
	if err != nil {
		return err
	}
	cfg.Out = out
	return nil
}

func createStandardPluginMap(cfg *Config) map[string]bool {
	return map[string]bool{
		protocGenGo:             cfg.GetAutoProtocGenGo(),
		protocGenGoGRPC:         cfg.GetAutoProtocGenGoGRPC(),
		protocGenGRPCFederation: true,
	}
}

func lookupProtocPlugin(name string, standardPluginMap map[string]bool) (string, error) {
	path, err := exec.LookPath(name)
	if err == nil {
		return path, nil
	}
	if _, exists := standardPluginMap[name]; exists {
		return "", nil
	}
	return "", fmt.Errorf(`failed to find protoc plugin "%s" on your system: %w`, name, err)
}

func addStandardPlugin(cfg *Config, name string) {
	path, _ := exec.LookPath(name)
	cfg.Plugins = append(cfg.Plugins, &PluginConfig{
		Plugin:        name,
		Opt:           &PluginOption{Opts: []string{"paths=source_relative"}},
		installedPath: path,
	})
}
