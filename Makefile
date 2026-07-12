BINARY      := asn1x
CMD         := ./cmd/asn1x
OUT_DIR     := bin
VERSION_PKG := github.com/en-vee/asn1x/cmd/asn1x/internal/cli

GO          := go
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GIT_COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)

LDFLAGS := -s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).GitCommit=$(GIT_COMMIT)

.PHONY: all build install test clean version help \
	darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 build-all

.DEFAULT_GOAL := build

all: build

build:
	@mkdir -p $(OUT_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY) $(CMD)

install:
	$(GO) install -ldflags "$(LDFLAGS)" $(CMD)

test:
	$(GO) test ./...

clean:
	rm -rf $(OUT_DIR)

version: build
	./$(OUT_DIR)/$(BINARY) version

darwin-amd64:
	@mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-darwin-amd64 $(CMD)

darwin-arm64:
	@mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-darwin-arm64 $(CMD)

linux-amd64:
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-linux-amd64 $(CMD)

linux-arm64:
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-linux-arm64 $(CMD)

build-all: darwin-amd64 darwin-arm64 linux-amd64 linux-arm64

help:
	@echo "Targets:"
	@echo "  build         Build $(BINARY) into $(OUT_DIR)/ (default)"
	@echo "  install       Install $(BINARY) into \$$GOBIN"
	@echo "  test          Run all tests"
	@echo "  version       Build and print version information"
	@echo "  clean         Remove $(OUT_DIR)/"
	@echo "  build-all     Cross-compile for darwin/linux amd64/arm64"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  GIT_COMMIT=$(GIT_COMMIT)"
