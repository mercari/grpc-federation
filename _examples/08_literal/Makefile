MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOBIN := $(MAKEFILE_DIR)/../../bin
PATH := $(GOBIN):$(PATH)

.PHONY: generate
generate:
	$(GOBIN)/buf generate

.PHONY: lint
lint:
	@$(GOBIN)/grpc-federation-linter -Iproto -Iproto_deps ./proto/federation/federation.proto

.PHONY: test
test:
	go test -race ./ -count=1
