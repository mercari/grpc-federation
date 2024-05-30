MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOBIN := $(MAKEFILE_DIR)/../../bin
PATH := $(GOBIN):$(PATH)

JAEGER_IMAGE := jaegertracing/all-in-one:latest

.PHONY: generate
generate: generate/config
	$(GOBIN)/buf generate

generate/config: build/wasm
	go run ./cmd/config/main.go

.PHONY: lint
lint:
	@$(GOBIN)/grpc-federation-linter -Iproto -Iproto_deps ./proto/federation/federation.proto

.PHONY: test
test:
	go test -race ./ -count=1

.PHONY: grpc-federation/generate
grpc-federation/generate: generate/config
	@$(GOBIN)/grpc-federation-generator ./proto/federation/federation.proto

.PHONY: grpc-federation/watch
grpc-federation/watch:
	@$(GOBIN)/grpc-federation-generator -w

.PHONY: jaeger/start
jaeger/start:
	@docker run \
		-e COLLECTOR_OTLP_ENABLED=true \
		-p 16686:16686 \
		-p 4317:4317 \
		-p 4318:4318 \
		-d \
		$(JAEGER_IMAGE)

.PHONY: jaeger/stop
jaeger/stop:
	@docker stop $(shell docker ps -q  --filter ancestor=$(JAEGER_IMAGE))

build/wasm:
	GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm ./cmd/plugin
