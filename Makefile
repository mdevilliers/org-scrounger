
# Linting
GOLANGCI_LINT_VERSION=1.45.0

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
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v$(GOLANGCI_LINT_VERSION)
	./bin/golangci-lint run
