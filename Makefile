
# Linting
GOLANGCI_LINT_VERSION=1.44.2

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
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_LINT_VERSION)
	golangci-lint run
