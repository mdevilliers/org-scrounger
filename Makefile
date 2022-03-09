OS := $(shell uname)
BIN_OUTDIR ?= ./build/bin

# Linting
GOLANGCI_LINT_VERSION=1.44.2
ifeq ($(OS),Darwin)
	GOLANGCI_LINT_ARCHIVE=golangci-lint-$(GOLANGCI_LINT_VERSION)-darwin-amd64.tar.gz
else
	GOLANGCI_LINT_ARCHIVE=golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz
endif

# Build a binary
.PHONY: build
build: CMD = ./cmd/team-reporter
build:
	go build $(CMD)

.PHONY: test
# Run test suite
test:
	go test -v ./...

# the linting gods must be obeyed
.PHONY: lint
lint: $(BIN_OUTDIR)/golangci-lint/golangci-lint
	$(BIN_OUTDIR)/golangci-lint/golangci-lint run

$(BIN_OUTDIR)/golangci-lint/golangci-lint:
	curl -OL https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/$(GOLANGCI_LINT_ARCHIVE)
	mkdir -p $(BIN_OUTDIR)/golangci-lint/
	tar -xf $(GOLANGCI_LINT_ARCHIVE) --strip-components=1 -C $(BIN_OUTDIR)/golangci-lint/
	chmod +x $(BIN_OUTDIR)/golangci-lint
	rm -f $(GOLANGCI_LINT_ARCHIVE)
