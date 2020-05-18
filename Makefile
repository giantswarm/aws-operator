APPLICATION   ?= aws-operator
VERSION       ?= $(shell git describe --tags --always --dirty)
SOURCES       = $(shell find . -name '*.go')
GOPKGS        = $(shell go list ./...)
TAG           ?= $(VERSION)
BUILD_FLAGS   ?= -v

default: build

.PHONY: build
## build: builds a local binary
build: clean $(SOURCES)
	CGO_ENABLED=0 go build -o ${APPLICATION} $(BUILD_FLAGS) .

.PHONY: build-linux
## build-linux: builds binary for linux/amd64
build-linux: clean $(SOURCES)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) .

.PHONY: build-darwin
## build-darwin: builds binary for darwin/amd64
build-darwin: clean $(SOURCES)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) .

.PHONY: install
## install: install the application
install: $(SOURCES)
	go install .

.PHONY: run
## run: runs go run main.go
run: $(SOURCES)
	go run -race .

.PHONY: clean
## clean: cleans the binary
clean:
	go clean

.PHONY: lint
## lint: runs golangci-lint
lint:
	golangci-lint run --timeout=15m ./...

.PHONY: test
## test: runs go test with default values
test:
	go test -v -race -cover $(GOPKGS)

.PHONY: build-docker
## build-docker: builds docker image locally
build-docker: build-linux
	docker build -t ${APPLICATION}:${TAG} .

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
