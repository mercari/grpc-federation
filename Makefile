export

GOBIN := $(PWD)/bin
PATH := $(GOBIN):$(PATH)
SHELL := env PATH='$(PATH)' bash

PKG := github.com/mercari/grpc-federation

# retrieve all packages related to the test
PKGS := $(shell go list ./... | grep -v tools | grep -v cmd)

# remove $PKG prefix from package name and add a dot character to convert it to a relative path.
COVER_PKGS := $(foreach pkg,$(PKGS),$(subst $(PKG),.,$(pkg)))

COMMA := ,
EMPTY :=
SPACE := $(EMPTY) $(EMPTY)

# join package names for coverage with comma character.
COVERPKG_OPT := $(subst $(SPACE),$(COMMA),$(COVER_PKGS))

EXAMPLES := $(wildcard _examples/*)

GIT_REF := $(shell git rev-parse --short=7 HEAD)

VERSION ?= $(GIT_REF)

.PHONY: tools
tools:
	cd tools && GOFLAGS='-mod=readonly' go install \
		github.com/bufbuild/buf/cmd/buf \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/envoyproxy/protoc-gen-validate/cmd/protoc-gen-validate-go \
		github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: lint
lint: lint/examples lint/golangci-lint

lint/examples: $(foreach var,$(EXAMPLES),lint/examples/$(var))

lint/examples/%:
	$(MAKE) -C $* lint

lint/golangci-lint:
	$(GOBIN)/golangci-lint run $(args) ./...

.PHONY: generate
generate: generate/buf generate/examples

generate/buf:
	buf generate

generate/examples: $(foreach var,$(EXAMPLES),generate/examples/$(var))

generate/examples/%:
	$(MAKE) -C $* generate

.PHONY: build
build: build/protoc-gen-grpc-federation build/grpc-federation-linter build/grpc-federation-language-server build/grpc-federation-generator

build/protoc-gen-grpc-federation:
	go build -o $(GOBIN)/protoc-gen-grpc-federation ./cmd/protoc-gen-grpc-federation

build/grpc-federation-linter:
	go build -o $(GOBIN)/grpc-federation-linter ./cmd/grpc-federation-linter

build/grpc-federation-language-server:
	go build -o $(GOBIN)/grpc-federation-language-server ./cmd/grpc-federation-language-server

build/grpc-federation-generator:
	go build -o $(GOBIN)/grpc-federation-generator ./cmd/grpc-federation-generator

.PHONY: build/vscode-extension
build/vscode-extension: install/vscode-dependencies
	cd lsp/client/vscode && npx vsce package -o grpc-federation-$(VERSION).vsix

.PHONY: install/vscode-dependencies
install/vscode-dependencies:
	cd lsp/client/vscode && npm install

.PHONY: test
test: test/examples
	go test -race -cover `go list ./... | grep -v github.com/mercari/grpc-federation/tools`

test/examples: $(foreach var,$(EXAMPLES),test/examples/$(var))

test/examples/%:
	$(MAKE) -C $* test

.PHONY: cover-html
cover-html:
	go test -coverpkg=$(COVERPKG_OPT) -coverprofile=cover.out `go list ./... | grep -v github.com/mercari/grpc-federation/tools`
	go tool cover -html=cover.out
