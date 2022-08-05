
# Linting
GOLANGCI_LINT_VERSION=1.48.0

# Build a binary
.PHONY: build
build: CMD = ./cmd/team-reporter
build:
	go build $(CMD)

# Run test suite
.PHONY: test
test:
	go test -v ./...

# The linting gods must be obeyed
.PHONY: lint
lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v$(GOLANGCI_LINT_VERSION)
	./bin/golangci-lint run

.PHONY: mocks
mocks:
	go generate ./...
