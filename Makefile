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

.PHONY: all build install clean test lint coverage vet fmt run cross-build dist-clean

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