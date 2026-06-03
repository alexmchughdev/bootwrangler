VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X github.com/alexmchughdev/bootwrangler/internal/version.Value=$(VERSION)
BUILD_DIR := dist

.PHONY: all build build-desktop build-cli test lint fmt clean

all: build

## build: build the CLI tool
build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/bootwrangler ./cmd/bootwrangler

## build-desktop: build the Wails desktop application
build-desktop:
	./scripts/build-desktop.sh -ldflags "$(LDFLAGS)"

## build-cli: alias for build
build-cli: build

## test: run Go and frontend tests
test:
	go test ./...
	cd frontend && npm test -- --run

## lint: run gofmt and go vet
lint:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then echo "Unformatted:"; echo "$$unformatted"; exit 1; fi
	go vet ./...

## fmt: format all Go source files
fmt:
	gofmt -w .

## clean: remove build artefacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf frontend/dist
