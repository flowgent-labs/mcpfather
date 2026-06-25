.PHONY: build build-all test install clean

BINARY_NAME := mcpgen
CMD_PATH := ./cmd/mcpgen
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.versionStr=$(VERSION)"
BUILD_FLAGS := -v -trimpath

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) $(CMD_PATH)

build-all:
	GOOS=linux   GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64   $(CMD_PATH)
	GOOS=linux   GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64   $(CMD_PATH)
	GOOS=darwin  GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64  $(CMD_PATH)
	GOOS=darwin  GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64  $(CMD_PATH)
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	GOOS=windows GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-arm64.exe $(CMD_PATH)

test:
	go test ./...

install:
	go install $(BUILD_FLAGS) $(LDFLAGS) $(CMD_PATH)

clean:
	rm -rf bin/
