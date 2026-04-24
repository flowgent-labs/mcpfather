.PHONY: build test install clean

BINARY_NAME := mcpgen
CMD_PATH := ./cmd/mcpgen

build:
	go build -o bin/$(BINARY_NAME) $(CMD_PATH)

test:
	go test ./...

install:
	go install $(CMD_PATH)

clean:
	rm -rf bin/
