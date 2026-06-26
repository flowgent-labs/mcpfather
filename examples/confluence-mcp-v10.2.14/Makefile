.PHONY: build build-all run clean test

BINARY_NAME := confluence-mcp-v10.2.14
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS := -v -trimpath

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

BIN := bin/$(BINARY_NAME)-$(GOOS)-$(GOARCH)-$(VERSION)$(if $(filter windows,$(GOOS)),.exe,)

build: go.sum
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -o $(BIN) .
	@ln -sf $(notdir $(BIN)) bin/$(BINARY_NAME)

build-all: go.sum
	GOOS=linux   GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-linux-amd64-$(VERSION)   .
	GOOS=linux   GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-linux-arm64-$(VERSION)   .
	GOOS=darwin  GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-amd64-$(VERSION)  .
	GOOS=darwin  GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-arm64-$(VERSION)  .
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-windows-amd64-$(VERSION).exe .
	GOOS=windows GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-windows-arm64-$(VERSION).exe .

go.sum: go.mod
	go mod tidy

run: build
	@bin/$(BINARY_NAME)

clean:
	@rm -f bin/$(BINARY_NAME)*

test:
	@go test ./...
