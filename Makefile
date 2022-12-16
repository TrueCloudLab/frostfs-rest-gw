#!/usr/bin/make -f

REPO ?= "$(shell go list -m)"
VERSION ?= "$(shell git describe --tags --match "v*" --dirty --always --abbrev=8 2>/dev/null || cat VERSION 2>/dev/null || echo "develop")"

BUILD_OS ?= linux
BUILD_ARCH ?= amd64
GO_VERSION ?= 1.19
LINT_VERSION ?= v1.49.0

HUB_IMAGE ?= truecloudlab/frostfs-rest-gw
HUB_TAG ?= "$(shell echo ${VERSION} | sed 's/^v//')"

SWAGGER_VERSION ?= v0.29.0

UNAME = "$(shell uname)/$(shell uname -m)"
SWAGGER_ARCH = linux_amd64

ifeq ($(UNAME), "Darwin/arm64")
	SWAGGER_ARCH = darwin_arm64
	BUILD_OS=darwin
	BUILD_ARCH=arm64
endif
ifeq ($(UNAME), "Darwin/x86_64")
	SWAGGER_ARCH = darwin_amd64
	BUILD_OS=darwin
endif

SWAGGER_URL ?= "https://github.com/go-swagger/go-swagger/releases/download/$(SWAGGER_VERSION)/swagger_$(SWAGGER_ARCH)"

# List of binaries to build. For now just one.
BINDIR = bin
DIRS = "$(BINDIR)"
BINS = "$(BINDIR)/frostfs-rest-gw"

.PHONY: help all dep clean format test cover lint docker/lint

# Make all binaries
all: generate-server $(BINS)

$(BINS): $(DIRS) dep
	@echo "⇒ Build $@"
	CGO_ENABLED=0 \
	GOOS=$(BUILD_OS) \
	GOARCH=$(BUILD_ARCH) \
	go build -v -trimpath \
	-ldflags "-X main.Version=$(VERSION)" \
	-o $@ ./cmd/frostfs-rest-gw

$(DIRS):
	@echo "⇒ Ensure dir: $@"
	@mkdir -p $@

# Pull go dependencies
dep:
	@printf "⇒ Download requirements: "
	@CGO_ENABLED=0 \
	go mod download && echo OK
	@printf "⇒ Tidy requirements: "
	@CGO_ENABLED=0 \
	go mod tidy -v && echo OK

# Install swagger
swagger:
ifeq (,$(wildcard ./bin/swagger))
	curl --create-dirs -o ./bin/swagger -L'#' $(SWAGGER_URL)
	chmod +x ./bin/swagger
endif

# Generate server by swagger spec
generate-server: swagger
	./bin/swagger generate server -t gen -f ./spec/rest.yaml --exclude-main \
		-A frostfs-rest-gw -P models.Principal \
		-C templates/server-config.yaml --template-dir templates

# Run tests
test:
	@go test ./... -cover

# Run tests with race detection and produce coverage output
cover:
	@go test -v -race ./... -coverprofile=coverage.txt -covermode=atomic
	@go tool cover -html=coverage.txt -o coverage.html

# Reformat code
format:
	@echo "⇒ Processing gofmt check"
	@gofmt -s -w ./
	@echo "⇒ Processing goimports check"
	@goimports -w ./

# Build clean Docker image
image:
	@echo "⇒ Build FrostFS REST Gateway docker image "
	@docker build \
		--build-arg REPO=$(REPO) \
		--build-arg VERSION=$(VERSION) \
		--rm \
		-f Dockerfile \
		-t $(HUB_IMAGE):$(HUB_TAG) .

# Push Docker image to the hub
image-push:
	@echo "⇒ Publish image"
	@docker push $(HUB_IMAGE):$(HUB_TAG)

# Build dirty Docker image
image-dirty:
	@echo "⇒ Build FrostFS REST Gateway dirty docker image "
	@docker build \
		--build-arg REPO=$(REPO) \
		--build-arg VERSION=$(VERSION) \
		--rm \
		-f Dockerfile.dirty \
		-t $(HUB_IMAGE)-dirty:$(HUB_TAG) .

# Run linters
lint:
	@golangci-lint --timeout=5m run

# Make all binaries in clean docker environment
docker/all:
	@echo "=> Running 'make all' in clean Docker environment" && \
	docker run --rm -t \
		-v `pwd`:/src \
		-w /src \
		-u `stat -c "%u:%g" .` \
		--env HOME=/src \
		--env BUILD_OS=$(BUILD_OS) \
		--env BUILD_ARCH=$(BUILD_ARCH) \
		golang:$(GO_VERSION) make all

# Generate server by swagger spec using swagger docker image
docker/generate-server:
	@docker run --rm -t \
		-v `pwd`:/src \
		-w /src \
		-u `stat -c "%u:%g" .` \
		--env HOME=/src \
		quay.io/goswagger/swagger:$(SWAGGER_VERSION) generate server \
			-t gen -f ./spec/rest.yaml --exclude-main \
			-A frostfs-rest-gw -P models.Principal \
			-C templates/server-config.yaml --template-dir templates

# Run linters in Docker
docker/lint:
	docker run --rm -it \
	-v `pwd`:/src \
	-u `stat -c "%u:%g" .` \
	--env HOME=/src \
	golangci/golangci-lint:$(LINT_VERSION) bash -c 'cd /src/ && make lint'

# Print version
version:
	@echo $(VERSION)

# Show this help prompt
help:
	@echo '  Usage:'
	@echo ''
	@echo '    make <target>'
	@echo ''
	@echo '  Targets:'
	@echo ''
	@awk '/^#/{ comment = substr($$0,3) } comment && /^[a-zA-Z][a-zA-Z0-9_-]+ ?:/{ print "   ", $$1, comment }' $(MAKEFILE_LIST) | column -t -s ':' | grep -v 'IGNORE' | sort -u

# Clean up
clean:
	rm -rf .cache
	rm -rf $(BINDIR)
