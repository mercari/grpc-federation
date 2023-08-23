package generator

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	// Imports specify list of import path. This is the same as the list of paths specified by protoc's '-I' option.
	Imports []string `yaml:"imports"`
	// Src specifies the directory to be monitored when watch mode is enabled.
	Src []string `yaml:"src"`
	// Out specify the output destination for automatically generated code.
	Out string `yaml:"out"`
	// Option specify options to be passed gRPC Federation code generator.
	Option *PluginOption `yaml:"opt"`
	// AutoProtocGenGo automatically run protoc-gen-go at the time of editing proto. default is true.
	AutoProtocGenGo *bool `yaml:"auto-protoc-gen-go"`
	// AutoProtocGenGoGRPC automatically run protoc-gen-go-grpc at the time of editing proto. default is true.
	AutoProtocGenGoGRPC *bool `yaml:"auto-protoc-gen-go-grpc"`
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

type PluginOption struct {
	Paths  PathMode `yaml:"paths"`
	Module string   `yaml:"module"`
}

type PathMode string

const (
	SourceRelativeMode PathMode = "source_relative"
	ImportMode         PathMode = "import"
)

func LoadConfig(path string) (Config, error) {
	var cfg Config
	content, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
