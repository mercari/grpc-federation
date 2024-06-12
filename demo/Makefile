export

GOBIN := $(PWD)/bin
PATH := $(GOBIN):$(PATH)

KO_DOCKER_REPO := ko.local
KO_DEFAULTPLATFORMS := linux/amd64

SERVICES := swapi film person planet species starship vehicle

.PHONY: tools
tools:
	go install github.com/bufbuild/buf/cmd/buf@v1.32.2
	go install github.com/google/ko@v0.15.4
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@v1.9.1

.PHONY: build
build:
	$(foreach svc,$(SERVICES),$(MAKE) build/$(svc))

build/%:
	ko build ./cmd/$* -B --sbom=none

.PHONY: up
up:
	docker compose up

.PHONY: generate
generate:
	$(GOBIN)/buf generate

.PHONY: test
test:
	go test -race ./... -count=1
