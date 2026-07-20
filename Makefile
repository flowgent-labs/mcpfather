.PHONY: build build-all test test-unit test-integration install clean gen-config-dsl-schema \
	build-image push-image help

BINARY_NAME := mcpfather
CMD_PATH := ./cmd/mcpfather
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-s -w -X main.versionStr=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildDate=$(BUILD_DATE)"
BUILD_FLAGS := -v -trimpath

GOPROXY ?= https://goproxy.cn,direct

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

IMAGE_REPO ?= ghcr.io/flowgent-labs/$(BINARY_NAME)
IMAGE_TAG  ?= $(VERSION)

BIN := bin/$(BINARY_NAME)-$(GOOS)-$(GOARCH)-$(VERSION)$(if $(filter windows,$(GOOS)),.exe,)

help:
	@echo "Usage:"
	@echo "  make build                  Build $(BINARY_NAME) for current platform"
	@echo "  make build-all              Cross-compile for all platforms"
	@echo "  make build-image            Build Docker image"
	@echo "  make push-image             Build and push Docker image to ghcr.io"
	@echo "  make test                   Run unit tests"
	@echo "  make test-unit              Run unit tests"
	@echo "  make test-integration       Run integration tests"
	@echo "  make install                Install $(BINARY_NAME) to GOPATH/bin"
	@echo "  make clean                  Remove build artifacts"
	@echo "  make gen-config-dsl-schema  Regenerate JSON Schema for virtual tool DSL"
	@echo ""

build:
	GOPROXY=$(GOPROXY) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BIN) $(CMD_PATH)
	@ln -sf $(notdir $(BIN)) bin/$(BINARY_NAME)

build-all:
	GOPROXY=$(GOPROXY) GOOS=linux   GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64-$(VERSION)   $(CMD_PATH)
	GOPROXY=$(GOPROXY) GOOS=linux   GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64-$(VERSION)   $(CMD_PATH)
	GOPROXY=$(GOPROXY) GOOS=darwin  GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64-$(VERSION)  $(CMD_PATH)
	GOPROXY=$(GOPROXY) GOOS=darwin  GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64-$(VERSION)  $(CMD_PATH)
	GOPROXY=$(GOPROXY) GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64-$(VERSION).exe $(CMD_PATH)
	GOPROXY=$(GOPROXY) GOOS=windows GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-arm64-$(VERSION).exe $(CMD_PATH)

test: test-unit

test-unit:
	GOPROXY=$(GOPROXY) go test -v -count=1 -timeout 300s ./pkg/... ./cmd/...

test-integration:
	GOPROXY=$(GOPROXY) go test -v -count=1 -timeout 600s ./it/...

install:
	go install $(BUILD_FLAGS) $(LDFLAGS) $(CMD_PATH)

clean:
	rm -rf bin/

# gen-config-dsl-schema regenerates the JSON Schema for virtual tool configuration
# from the Go struct definitions in internal/generator/mcpvirtual/.
# Output is written to the skill resources directory for use by the virtual-tool-creator skill.
gen-config-dsl-schema:
	@mkdir -p bin
	@go build -o bin/gen-config-dsl-schema ./cmd/gen-config-dsl-schema/
	@./bin/gen-config-dsl-schema --output .agents/skills/virtual-tool-creator/resources/dsl-schema.json
	@echo "==> Schema updated: .agents/skills/virtual-tool-creator/resources/dsl-schema.json"

build-image: build
	docker build -t $(IMAGE_REPO):$(IMAGE_TAG) .

push-image: build-image
	docker push $(IMAGE_REPO):$(IMAGE_TAG)
