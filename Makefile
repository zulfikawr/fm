BINARY_NAME=fm
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X 'github.com/zulfikawr/fm/internal/constants.AppVersion=$(VERSION)'

.PHONY: all build clean test install help run

all: build

## Build: Build the binary
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/fm

## Install: Install the binary to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/fm

## Run: Run the application
run:
	go run -ldflags "$(LDFLAGS)" ./cmd/fm

## Clean: Remove the binary
clean:
	go clean
	rm -f $(BINARY_NAME)

## Test: Run tests
test:
	go test ./...

## Lint: Run golangci-lint and go fmt
lint:
	go fmt ./...
	golangci-lint run

## Help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
