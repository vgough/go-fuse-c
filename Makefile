PROJECT_NAME := "go-fuse-c"
PKG := "github.com/vgough/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

.PHONY: all dep build clean test lint rust-docker-test

all: build

lint: ## Lint the files
	revive -config revive.toml ./...

test: ## Run unittests
	@go test -short ${PKG_LIST}

race: dep ## Run data race detector
	@go test -race -short ${PKG_LIST}

msan: dep ## Run memory sanitizer
	@go test -msan -short ${PKG_LIST}

dep: ## Get the dependencies
	@go get -v -d ./...

build: dep ## Build the binary file
	@go build -i -v $(PKG)

rust-docker-test: ## Run the Rust hello smoke test in a Linux Docker container
	@bash rust/tests/linux/run-in-docker.sh

  
