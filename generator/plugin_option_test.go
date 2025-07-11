package generator

import (
	"net/url"
	"path/filepath"
	"testing"
)

func TestParsePluginURI(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantScheme      string
		wantPath        string
		wantHash        string
		wantOriginalURL string
		wantErr         bool
	}{
		{
			name:            "file scheme without hash",
			input:           "file:///path/to/plugin.wasm",
			wantScheme:      "file",
			wantPath:        "/path/to/plugin.wasm",
			wantHash:        "",
			wantOriginalURL: "file:///path/to/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "file scheme without leading slash",
			input:           "file://path/to/plugin.wasm",
			wantScheme:      "file",
			wantPath:        "/to/plugin.wasm",
			wantHash:        "",
			wantOriginalURL: "file://path/to/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "file scheme with hash",
			input:           "file:///path/to/plugin.wasm:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantScheme:      "file",
			wantPath:        "/path/to/plugin.wasm",
			wantHash:        "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantOriginalURL: "file:///path/to/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "file scheme with Windows path without hash",
			input:           "file:///C:/path/to/plugin.wasm",
			wantScheme:      "file",
			wantPath:        "/C:/path/to/plugin.wasm",
			wantHash:        "",
			wantOriginalURL: "file:///C:/path/to/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "file scheme with Windows path with hash",
			input:           "file:///C:/path/to/plugin.wasm:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			wantScheme:      "file",
			wantPath:        "/C:/path/to/plugin.wasm",
			wantHash:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			wantOriginalURL: "file:///C:/path/to/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "http scheme without hash",
			input:           "http://example.com/plugin.wasm",
			wantScheme:      "http",
			wantPath:        "/plugin.wasm",
			wantHash:        "",
			wantOriginalURL: "http://example.com/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "http scheme with hash",
			input:           "http://example.com/plugin.wasm:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantScheme:      "http",
			wantPath:        "/plugin.wasm",
			wantHash:        "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantOriginalURL: "http://example.com/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "http scheme with port without hash",
			input:           "http://example.com:8080/plugin.wasm",
			wantScheme:      "http",
			wantPath:        "/plugin.wasm",
			wantHash:        "",
			wantOriginalURL: "http://example.com:8080/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "http scheme with port and hash",
			input:           "http://example.com:8080/plugin.wasm:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantScheme:      "http",
			wantPath:        "/plugin.wasm",
			wantHash:        "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantOriginalURL: "http://example.com:8080/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "https scheme without hash",
			input:           "https://example.com/plugin.wasm",
			wantScheme:      "https",
			wantPath:        "/plugin.wasm",
			wantHash:        "",
			wantOriginalURL: "https://example.com/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "https scheme with hash",
			input:           "https://example.com/plugin.wasm:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			wantScheme:      "https",
			wantPath:        "/plugin.wasm",
			wantHash:        "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			wantOriginalURL: "https://example.com/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "https scheme with port and path without hash",
			input:           "https://example.com:443/path/to/plugin.wasm",
			wantScheme:      "https",
			wantPath:        "/path/to/plugin.wasm",
			wantHash:        "",
			wantOriginalURL: "https://example.com:443/path/to/plugin.wasm",
			wantErr:         false,
		},
		{
			name:            "invalid hash (not 64 chars)",
			input:           "file:///path/to/plugin.wasm:shorthhash",
			wantScheme:      "file",
			wantPath:        "/path/to/plugin.wasm:shorthhash",
			wantHash:        "",
			wantOriginalURL: "file:///path/to/plugin.wasm:shorthhash",
			wantErr:         false,
		},
		{
			name:            "invalid hash (non-hex chars) results in invalid path",
			input:           "file:///path/to/plugin.wasm:zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			wantScheme:      "file",
			wantPath:        "/path/to/plugin.wasm:zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			wantHash:        "",
			wantOriginalURL: "file:///path/to/plugin.wasm:zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			wantErr:         false, // parsePluginURI itself doesn't error, but parseFileSchemeOption will
		},
		{
			name:            "http with port and invalid hash length",
			input:           "http://example.com:8080/plugin.wasm:shortinvalidhash",
			wantScheme:      "http",
			wantPath:        "/plugin.wasm:shortinvalidhash",
			wantHash:        "",
			wantOriginalURL: "http://example.com:8080/plugin.wasm:shortinvalidhash",
			wantErr:         false, // parsePluginURI treats the whole string as URL when hash is invalid
		},
		{
			name:            "https with port and invalid hash characters",
			input:           "https://example.com:443/plugin.wasm:invalidnonhexhash",
			wantScheme:      "https",
			wantPath:        "/plugin.wasm:invalidnonhexhash",
			wantHash:        "",
			wantOriginalURL: "https://example.com:443/plugin.wasm:invalidnonhexhash",
			wantErr:         false, // parsePluginURI treats the whole string as URL when hash is invalid
		},
		{
			name:            "url with query parameters",
			input:           "https://example.com/plugin.wasm?version=1.0",
			wantScheme:      "https",
			wantPath:        "/plugin.wasm",
			wantHash:        "",
			wantOriginalURL: "https://example.com/plugin.wasm?version=1.0",
			wantErr:         false,
		},
		{
			name:            "url with fragment",
			input:           "https://example.com/plugin.wasm#latest",
			wantScheme:      "https",
			wantPath:        "/plugin.wasm",
			wantHash:        "",
			wantOriginalURL: "https://example.com/plugin.wasm#latest",
			wantErr:         false,
		},
		{
			name:    "invalid url",
			input:   "://invalid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePluginURI(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePluginURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.parsedURL.Scheme != tt.wantScheme {
				t.Errorf("parsePluginURI() scheme = %v, want %v", got.parsedURL.Scheme, tt.wantScheme)
			}
			if got.parsedURL.Path != tt.wantPath {
				t.Errorf("parsePluginURI() path = %v, want %v", got.parsedURL.Path, tt.wantPath)
			}
			if got.hash != tt.wantHash {
				t.Errorf("parsePluginURI() hash = %v, want %v", got.hash, tt.wantHash)
			}
			if got.originalURL != tt.wantOriginalURL {
				t.Errorf("parsePluginURI() originalURL = %v, want %v", got.originalURL, tt.wantOriginalURL)
			}
		})
	}
}

func TestParsePluginsOption(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(*testing.T, *CodeGeneratorOption)
	}{
		{
			name:    "file scheme",
			input:   "file:///path/to/plugin.wasm",
			wantErr: false,
			validate: func(t *testing.T, opt *CodeGeneratorOption) {
				if len(opt.Plugins) != 1 {
					t.Errorf("expected 1 plugin, got %d", len(opt.Plugins))
					return
				}
				if opt.Plugins[0].Path != "/path/to/plugin.wasm" {
					t.Errorf("expected path '/path/to/plugin.wasm', got %s", opt.Plugins[0].Path)
				}
				if opt.Plugins[0].Sha256 != "" {
					t.Errorf("expected empty hash, got %s", opt.Plugins[0].Sha256)
				}
			},
		},
		{
			name:    "file scheme with hash",
			input:   "file:///path/to/plugin.wasm:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr: false,
			validate: func(t *testing.T, opt *CodeGeneratorOption) {
				if len(opt.Plugins) != 1 {
					t.Errorf("expected 1 plugin, got %d", len(opt.Plugins))
					return
				}
				if opt.Plugins[0].Path != "/path/to/plugin.wasm" {
					t.Errorf("expected path '/path/to/plugin.wasm', got %s", opt.Plugins[0].Path)
				}
				if opt.Plugins[0].Sha256 != "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" {
					t.Errorf("expected hash '1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef', got %s", opt.Plugins[0].Sha256)
				}
			},
		},
		{
			name:    "unsupported scheme",
			input:   "ftp://example.com/plugin.wasm",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			input:   "://invalid",
			wantErr: true,
		},
		{
			name:    "file with invalid path (invalid hash treated as path)",
			input:   "file:///path/to/plugin.wasm:zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &CodeGeneratorOption{}
			err := parsePluginsOption(opt, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePluginsOption() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, opt)
			}
		})
	}
}

func TestParseFileSchemeOption(t *testing.T) {
	tests := []struct {
		name     string
		uri      *pluginURI
		wantErr  bool
		validate func(*testing.T, *CodeGeneratorOption)
	}{
		{
			name: "valid file path",
			uri: &pluginURI{
				originalURL: "file:///path/to/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "file",
					Path:   "/path/to/plugin.wasm",
				},
				hash: "",
			},
			wantErr: false,
			validate: func(t *testing.T, opt *CodeGeneratorOption) {
				if len(opt.Plugins) != 1 {
					t.Errorf("expected 1 plugin, got %d", len(opt.Plugins))
					return
				}
				if opt.Plugins[0].Path != "/path/to/plugin.wasm" {
					t.Errorf("expected path '/path/to/plugin.wasm', got %s", opt.Plugins[0].Path)
				}
				if opt.Plugins[0].Sha256 != "" {
					t.Errorf("expected empty hash, got %s", opt.Plugins[0].Sha256)
				}
			},
		},
		{
			name: "relative path file",
			uri: &pluginURI{
				originalURL: "file://path/to/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "file",
					Host:   "path",
					Path:   "/to/plugin.wasm",
				},
				hash: "",
			},
			wantErr: false,
			validate: func(t *testing.T, opt *CodeGeneratorOption) {
				if len(opt.Plugins) != 1 {
					t.Errorf("expected 1 plugin, got %d", len(opt.Plugins))
					return
				}
				// Should be converted to absolute path
				if !filepath.IsAbs(opt.Plugins[0].Path) {
					t.Errorf("expected absolute path, got %s", opt.Plugins[0].Path)
				}
			},
		},
		{
			name: "valid file path with hash",
			uri: &pluginURI{
				originalURL: "file:///path/to/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "file",
					Path:   "/path/to/plugin.wasm",
				},
				hash: "abcdef1234567890",
			},
			wantErr: false,
			validate: func(t *testing.T, opt *CodeGeneratorOption) {
				if len(opt.Plugins) != 1 {
					t.Errorf("expected 1 plugin, got %d", len(opt.Plugins))
					return
				}
				if opt.Plugins[0].Sha256 != "abcdef1234567890" {
					t.Errorf("expected hash 'abcdef1234567890', got %s", opt.Plugins[0].Sha256)
				}
			},
		},
		{
			name: "empty path",
			uri: &pluginURI{
				originalURL: "file://",
				parsedURL: &url.URL{
					Scheme: "file",
					Path:   "",
				},
				hash: "",
			},
			wantErr: true,
		},
		{
			name: "invalid path with colon",
			uri: &pluginURI{
				originalURL: "file:///path/to/plugin.wasm:invalidhash",
				parsedURL: &url.URL{
					Scheme: "file",
					Path:   "/path/to/plugin.wasm:invalidhash",
				},
				hash: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &CodeGeneratorOption{}
			err := parseFileSchemeOption(opt, tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFileSchemeOption() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, opt)
			}
		})
	}
}

func TestParseHTTPSchemeOption(t *testing.T) {
	tests := []struct {
		name    string
		uri     *pluginURI
		wantErr bool
	}{
		{
			name: "invalid hash length",
			uri: &pluginURI{
				originalURL: "http://example.com/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/plugin.wasm",
				},
				hash: "invalidhash",
			},
			wantErr: true,
		},
		{
			name: "invalid hash with non-hex characters",
			uri: &pluginURI{
				originalURL: "http://example.com/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "http",
					Host:   "example.com",
					Path:   "/plugin.wasm",
				},
				hash: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			},
			wantErr: true,
		},
		{
			name: "url with port and invalid hash length",
			uri: &pluginURI{
				originalURL: "http://example.com:8080/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "http",
					Host:   "example.com:8080",
					Path:   "/plugin.wasm",
				},
				hash: "invalidhash",
			},
			wantErr: true,
		},
		{
			name: "url with port and invalid hash characters",
			uri: &pluginURI{
				originalURL: "http://example.com:3000/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "http",
					Host:   "example.com:3000",
					Path:   "/plugin.wasm",
				},
				hash: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &CodeGeneratorOption{}
			err := parseHTTPSchemeOption(opt, tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHTTPSchemeOption() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseHTTPSSchemeOption(t *testing.T) {
	tests := []struct {
		name    string
		uri     *pluginURI
		wantErr bool
	}{
		{
			name: "invalid hash length",
			uri: &pluginURI{
				originalURL: "https://example.com/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/plugin.wasm",
				},
				hash: "invalidhash",
			},
			wantErr: true,
		},
		{
			name: "invalid hash with non-hex characters",
			uri: &pluginURI{
				originalURL: "https://example.com/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/plugin.wasm",
				},
				hash: "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			},
			wantErr: true,
		},
		{
			name: "url with port and invalid hash length",
			uri: &pluginURI{
				originalURL: "https://example.com:443/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "https",
					Host:   "example.com:443",
					Path:   "/plugin.wasm",
				},
				hash: "shortinvalidhash",
			},
			wantErr: true,
		},
		{
			name: "url with port and invalid hash characters",
			uri: &pluginURI{
				originalURL: "https://example.com:9443/plugin.wasm",
				parsedURL: &url.URL{
					Scheme: "https",
					Host:   "example.com:9443",
					Path:   "/plugin.wasm",
				},
				hash: "gggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggggg",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := &CodeGeneratorOption{}
			err := parseHTTPSSchemeOption(opt, tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHTTPSSchemeOption() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
