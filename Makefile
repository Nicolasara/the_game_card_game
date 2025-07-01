# Makefile for The Game

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
GOMOD=$(GOCMD) mod
GOTIDY=$(GOMOD) tidy

# Protoc parameters
PROTOC=protoc
PROTOC_GEN_GO_PATH=$(shell go env GOMODCACHE)/github.com/protocolbuffers/protobuf-go@v1.34.2
PROTOC_GEN_GO_GRPC_PATH=$(shell go env GOMODCACHE)/google.golang.org/grpc@v1.65.0

.PHONY: all test clean deps proto mocks generate

all: server client

# Build the server and client
server:
	$(GOBUILD) -o bin/server ./cmd/server

client:
	$(GOBUILD) -o bin/client ./cmd/client

# Run tests
test:
	$(GOTEST) -v ./...

# Clean up
clean:
	$(GOCLEAN)
	rm -rf bin

# Install dependencies
deps:
	$(GOINSTALL) google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOINSTALL) google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GOINSTALL) github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	$(GOINSTALL) github.com/vektra/mockery/v2@v2.43.2

# Generate all files
generate: proto mocks

# Generate proto files
proto:
	PATH="$(shell go env GOPATH)/bin:$(PATH)" $(PROTOC) --proto_path=proto \
	 --go_out=proto --go_opt=paths=source_relative \
     --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
     --grpc-gateway_out=proto --grpc-gateway_opt=paths=source_relative \
     -I=third_party/googleapis \
	 proto/game.proto

# Generate mocks
mocks:
	$(GOCMD) run github.com/vektra/mockery/v2 --name Storer --dir pkg/storage --output pkg/storage/mocks

# Tidy dependencies
tidy:
	$(GOTIDY)

.DEFAULT_GOAL := all 