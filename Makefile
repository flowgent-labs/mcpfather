.PHONY: build build-all test install clean gen-config-dsl-schema

BINARY_NAME := mcpgen
CMD_PATH := ./cmd/mcpgen
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.versionStr=$(VERSION)"
BUILD_FLAGS := -v -trimpath

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

BIN := bin/$(BINARY_NAME)-$(GOOS)-$(GOARCH)-$(VERSION)$(if $(filter windows,$(GOOS)),.exe,)

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BIN) $(CMD_PATH)
	@ln -sf $(notdir $(BIN)) bin/$(BINARY_NAME)

build-all:
	GOOS=linux   GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64-$(VERSION)   $(CMD_PATH)
	GOOS=linux   GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64-$(VERSION)   $(CMD_PATH)
	GOOS=darwin  GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64-$(VERSION)  $(CMD_PATH)
	GOOS=darwin  GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64-$(VERSION)  $(CMD_PATH)
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64-$(VERSION).exe $(CMD_PATH)
	GOOS=windows GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-arm64-$(VERSION).exe $(CMD_PATH)

test:
	go test -v -count=1 ./...

install:
	go install $(BUILD_FLAGS) $(LDFLAGS) $(CMD_PATH)

clean:
	rm -rf bin/

# gen-config-dsl-schema regenerates the JSON Schema for aggregated tool configuration
# from the Go struct definitions in internal/generator/mcpaggregator/.
# Output is written to the skill resources directory for use by the aggregate-tool-creator skill.
gen-config-dsl-schema:
	@mkdir -p bin
	@go build -o bin/gen-config-dsl-schema ./cmd/gen-config-dsl-schema/
	@./bin/gen-config-dsl-schema --output .agents/skills/aggregate-tool-creator/resources/dsl-schema.json
	@echo "==> Schema updated: .agents/skills/aggregate-tool-creator/resources/dsl-schema.json"
