.PHONY: build test lint wasm serve

BIN_DIR  := bin
WASM_DIR := web

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/goscope ./cmd/goscope
	go build -o $(BIN_DIR)/webdemo ./cmd/webdemo

test:
	go test ./...

lint:
	golangci-lint run ./...

wasm:
	@mkdir -p $(WASM_DIR)
	cp "$(shell go env GOROOT)/lib/wasm/wasm_exec.js" $(WASM_DIR)/
	GOOS=js GOARCH=wasm go build -o $(WASM_DIR)/goscope.wasm ./cmd/goscope

serve:
	go run ./cmd/webdemo
