.PHONY: all build test lint wasm serve

BIN_DIR   := bin
WEB_DIR   := web
WASM_EXEC := $(shell go env GOROOT)/lib/wasm/wasm_exec.js

all: build wasm

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/serve ./cmd/serve

test:
	go test ./...

lint:
	golangci-lint run ./...

wasm:
	@mkdir -p $(WEB_DIR)
	cp "$(WASM_EXEC)" $(WEB_DIR)/wasm_exec.js
	GOOS=js GOARCH=wasm go build -o $(WEB_DIR)/main.wasm ./cmd/webdemo

serve:
	go run ./cmd/serve
