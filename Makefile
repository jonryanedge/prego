BINARY     := prego
CMD        := ./cmd/$(BINARY)
VERSION    := 0.1.0
BUILD      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
LD_FLAGS   := -ldflags "-s -w -X github.com/jonryanedge/prego/internal/cmd.version=$(VERSION)"
GO         := go
GOFLAGS    :=
DESTDIR    ?= /usr/local/bin

LINT_BIN   := $(shell command -v golangci-lint 2>/dev/null)
COVERAGE   := coverage.out

.PHONY: all build install clean test lint coverage vet fmt run cross-build dist-clean help

help:
	@echo "$(BINARY) v$(VERSION) ($(BUILD))"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  all          lint, test, and build (default)"
	@echo "  build        compile binary to bin/"
	@echo "  install      build and copy to DESTDIR (default: /usr/local/bin)"
	@echo "  run          build and run the binary"
	@echo "  test         run all tests with race detector"
	@echo "  test-short   run tests in short mode"
	@echo "  coverage     generate HTML coverage report"
	@echo "  vet          run go vet"
	@echo "  lint         run go vet and golangci-lint"
	@echo "  fmt          run gofmt and goimports"
	@echo "  tidy         run go mod tidy"
	@echo "  cross-build  cross-compile for darwin/linux amd64/arm64"
	@echo "  version      print version info"
	@echo "  clean        remove bin/ and coverage files"
	@echo "  dist-clean   clean + purge Go build/test caches"
	@echo "  help          show this help menu"

all: lint test build

build:
	$(GO) build $(GOFLAGS) $(LD_FLAGS) -o bin/$(BINARY) $(CMD)

install: build
	cp bin/$(BINARY) $(DESTDIR)/$(BINARY)

clean:
	rm -rf bin/ $(COVERAGE) coverage.html

dist-clean: clean
	$(GO) clean -cache -testcache

test:
	$(GO) test ./... -v -count=1 -race

test-short:
	$(GO) test ./... -short -count=1

coverage:
	$(GO) test ./... -coverprofile=$(COVERAGE) -covermode=atomic -count=1
	$(GO) tool cover -html=$(COVERAGE) -o coverage.html
	@echo "Coverage report: coverage.html"

vet:
	$(GO) vet ./...

fmt:
	gofmt -w -s .
	goimports -w .

lint: vet
ifdef LINT_BIN
	golangci-lint run ./...
else
	@echo "golangci-lint not found, skipping"
endif

run: build
	./bin/$(BINARY)

cross-build:
	@mkdir -p bin
	GOOS=darwin GOARCH=amd64 $(GO) build $(LD_FLAGS) -o bin/$(BINARY)-darwin-amd64 $(CMD)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LD_FLAGS) -o bin/$(BINARY)-darwin-arm64 $(CMD)
	GOOS=linux GOARCH=amd64 $(GO) build $(LD_FLAGS) -o bin/$(BINARY)-linux-amd64 $(CMD)
	GOOS=linux GOARCH=arm64 $(GO) build $(LD_FLAGS) -o bin/$(BINARY)-linux-arm64 $(CMD)
	@echo "Cross-compiled binaries in bin/"

tidy:
	$(GO) mod tidy

version:
	@echo "$(BINARY) v$(VERSION) ($(BUILD))"