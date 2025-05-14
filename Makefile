export

GOBIN := $(PWD)/bin
PATH := $(GOBIN):$(PATH)
SHELL := env PATH='$(PATH)' bash

PKG := github.com/mercari/grpc-federation

# retrieve all packages related to the test
PKGS := $(shell go list ./... | grep -v cmd)

# remove $PKG prefix from package name and add a dot character to convert it to a relative path.
COVER_PKGS := $(foreach pkg,$(PKGS),$(subst $(PKG),.,$(pkg)))

COMMA := ,
EMPTY :=
SPACE := $(EMPTY) $(EMPTY)

# join package names for coverage with comma character.
COVERPKG_OPT := $(subst $(SPACE),$(COMMA),$(COVER_PKGS))

EXAMPLES := $(wildcard _examples/*)

GIT_REF := $(shell git rev-parse --short=7 HEAD)

VERSION ?= (devel)
LDFLAGS := -w -s -X=github.com/mercari/grpc-federation/grpc/federation.Version=$(VERSION)

.PHONY: tools
tools:
	go install github.com/bufbuild/buf/cmd/buf@v1.32.2
	go install github.com/envoyproxy/protoc-gen-validate/cmd/protoc-gen-validate-go@v1.0.4
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0

.PHONY: lint
lint: lint/examples lint/golangci-lint lint/gomod lint/buf

.PHONY: fmt
fmt: fmt/golangci-lint tidy fmt/buf

.PHONY: tidy
tidy: tidy/examples
	go mod tidy

tidy/examples: $(foreach var,$(EXAMPLES),tidy/examples/$(var))

tidy/examples/%:
	cd $* && go mod tidy

lint/examples: $(foreach var,$(EXAMPLES),lint/examples/$(var))

lint/examples/%:
	$(MAKE) -C $* lint

lint/golangci-lint:
	$(GOBIN)/golangci-lint run $(args) ./...

lint/gomod: tidy
	if git diff --quiet go.mod go.sum; then \
        exit 0; \
	else \
		echo "go mod tidy resulted in a change of files."; \
		echo "Run make tidy locally before pushing"; \
		exit 1; \
	fi

lint/buf: fmt/buf
	if git diff --quiet proto; then \
        exit 0; \
	else \
		echo "buf format resulted in a change of files."; \
		echo "Run make fmt/buf locally before pushing"; \
		exit 1; \
	fi

fmt/golangci-lint:
	$(GOBIN)/golangci-lint run --fix $(args) ./...

fmt/buf:
	buf format --write

.PHONY: generate
generate: build generate/buf generate/examples

generate/buf:
	buf generate

generate/examples: $(foreach var,$(EXAMPLES),generate/examples/$(var))

generate/examples/%:
	$(MAKE) -C $* generate

.PHONY: build
build: build/protoc-gen-grpc-federation build/grpc-federation-linter build/grpc-federation-language-server build/grpc-federation-generator

build/protoc-gen-grpc-federation:
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/protoc-gen-grpc-federation ./cmd/protoc-gen-grpc-federation

build/grpc-federation-linter:
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/grpc-federation-linter ./cmd/grpc-federation-linter

build/grpc-federation-language-server:
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/grpc-federation-language-server ./cmd/grpc-federation-language-server

build/grpc-federation-generator:
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/grpc-federation-generator ./cmd/grpc-federation-generator

.PHONY: build/vscode-extension
build/vscode-extension: versioning/vscode-extension install/vscode-dependencies
	cd lsp/client/vscode && npx vsce package -o grpc-federation-$(VERSION).vsix

.PHONY: install/vscode-dependencies
install/vscode-dependencies:
	cd lsp/client/vscode && npm install

.PHONY: versioning/vscode-extension
versioning/vscode-extension:
	cd lsp/client/vscode && npm version $(VERSION)

.PHONY: test
test: test/examples
	go test -race -coverpkg=$(COVERPKG_OPT) -covermode=atomic -coverprofile=cover.out.tmp `go list ./...`
	cat cover.out.tmp |grep -v "pb.go" > cover.out && rm cover.out.tmp

test/examples: $(foreach var,$(EXAMPLES),test/examples/$(var))

test/examples/%:
	$(MAKE) -C $* test

.PHONY: cover-html
cover-html: test
	go tool cover -html=cover.out
