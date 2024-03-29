GO_PKG_NAME=pocket-pick
TARGET=bin/pocket-pick
SRC=$(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "*_test.go")

GIT_COMMIT ?= $(shell git rev-parse HEAD)
GIT_SHA ?= $(shell git rev-parse --short HEAD)
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
GIT_TAG ?= $(shell git describe --tags --always)
GIT_DIRTY ?= $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")
VER_BUILD_TIME ?= $(shell date +%Y-%m-%dT%H:%M:%S%z)

LDFLAGS = -s -w
LDFLAGS += -X ${GO_PKG_NAME}.GitCommit=${GIT_COMMIT}
LDFLAGS += -X ${GO_PKG_NAME}.GitSHA=${GIT_SHA}
LDFLAGS += -X ${GO_PKG_NAME}.GitBranch=${GIT_BRANCH}
LDFLAGS += -X ${GO_PKG_NAME}.GitTag=${GIT_TAG}
LDFLAGS += -X ${GO_PKG_NAME}.GitDirty=${GIT_DIRTY}
LDFLAGS += -X ${GO_PKG_NAME}.BuildTime=${VER_BUILD_TIME}

BUILD_FLAGS?=-v -ldflags="${LDFLAGS}"

.PHONY: clean test dep tidy

all: build
build: $(TARGET)

$(TARGET): $(SRC)
	@mkdir -p bin
	go build -o bin/ ${BUILD_FLAGS} ./cmd/...

clean:
	rm -f ${TARGET}

test:
	go test

# update modules & tidy
dep:
	@rm -f go.mod go.sum
	@go mod init ${GO_PKG_NAME}
	@$(MAKE) tidy

tidy:
	@go mod tidy -v

swag:
	swag init -d .  -g app.go
