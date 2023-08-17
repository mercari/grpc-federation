package generator

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Imports []string      `yaml:"imports"`
	Src     []string      `yaml:"src"`
	Out     string        `yaml:"out"`
	Option  *PluginOption `yaml:"opt"`
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
