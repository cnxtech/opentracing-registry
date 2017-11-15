# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

.PHONY: all binaries clean
.DEFAULT: all
all: binaries

# Go files
GOFILES=$(shell find . -type f -name '*.go')

${PREFIX}/bin/registry: $(GOFILES)
	@echo "+ $@"
	@go run cmd/buildkit-registry/main.go | buildctl build --local src=./cmd/registry

binaries: ${PREFIX}/bin/registry
	@echo "+ $@"

clean:
	@echo "+ $@"
	@rm -rf "${PREFIX}/bin/registry"

vendor:
	@echo "+ $@"
	@./hack/update-vendor

