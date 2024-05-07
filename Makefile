PROJ_NAME = bhlnames

VERSION = $(shell git describe --tags)
VER = $(shell git describe --tags --abbrev=0)
DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S%Z')

NO_C = CGO_ENABLED=0
FLAGS_SHARED = CGO_ENABLED=0 GOARCH=amd64
FLAGS_LINUX = $(FLAGS_SHARED) GOOS=linux
FLAGS_MAC = $(FLAGS_SHARED) GOOS=darwin
FLAGS_WIN = $(FLAGS_SHARED) GOOS=windows

FLAGS_LD = -ldflags "-X github.com/gnames/$(PROJ_NAME)/pkg.Build=${DATE} \
                  -X github.com/gnames/$(PROJ_NAME)/pkg.Version=${VERSION}"
FLAGS_REL = -trimpath -ldflags "-s -w -X github.com/gnames/$(PROJ_NAME)/pkg.Build=$(DATE)"
RELEASE_DIR = /tmp
TEST_OPTS =  -p 1 -shuffle=on  ./internal/ent/input ./internal/ent/score ./internal/io/dictio ./pkg ./pkg/config


GOCMD = go
GOTEST = $(GOCMD) test
GOVET = $(GOCMD) vet
GOBUILD = $(GOCMD) build $(FLAGS_LD)
GORELEASE = $(GOCMD) build $(FLAGS_REL)
GOINSTALL = $(GOCMD) install $(FLAGS_LD)
GOCLEAN = $(GOCMD) clean
GOGET = $(GOCMD) get

.PHONY: help build test deps tools release install openapi

all: install

## Test:
test: ## Run the tests of the project
	$(GOTEST) $(TEST_OPTS);
	@echo "Also start restful service and run tests with 'make testrest'"

testrest:
	$(GOTEST) -p 1 -shuffle=on ./internal/io/restio

## Dependencies
deps: ## Download dependencies
	$(GOCMD) mod download;

## Tools
tools: deps ## Install tools
	@echo Installing tools from tools.go
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

## Build:
build: openapi ## Build binary
	$(NO_C) $(GOBUILD) \
		-o $(PROJ_NAME) \
		.

## Build Release
buildrel: openapi ## Build binary without debug info and with hardcoded version
	$(NO_C) $(GORELEASE) \
		-o $(PROJ_NAME) \
		.

## Release
release: openapi dockerhub ## Build and package binaries for release
	$(GOCLEAN); \
	$(FLAGS_LINUX) $(GORELEASE); \
	tar zcvf /tmp/bhlnames-${VER}-linux.tar.gz bhlnames; \
	$(GOCLEAN); \
	$(FLAGS_MAC) $(GORELEASE);  \
	tar zcvf /tmp/bhlnames-${VER}-mac.tar.gz bhlnames; \
	$(GOCLEAN); \
	$(FLAGS_WIN) $(GORELEASE); \
	zip -9 /tmp/bhlnames-$(VER)-win-64.zip bhlnames.exe; \
	$(GOCLEAN);

## Install
install: openapi ## Build and install binary
	$(FLAGS_SHARED) $(GOINSTALL);
	
## OpenAPI generation
openapi: ## Generate documentation for OpenAPI
	swag init -g restio.go -d ./internal/io/restio  --parseDependency --parseInternal

## Build docker image
dc: build
	docker-compose build; \

docker: buildrel
	docker buildx build -t gnames/$(PROJ_NAME):latest -t gnames/$(PROJ_NAME):$(VERSION) .; \
	$(GOCLEAN);

dockerhub: docker
	docker push gnames/$(PROJ_NAME); \
	docker push gnames/$(PROJ_NAME):$(VERSION)

## Help:
help: ## Show this help
	@echo ''
	@echo 'Usage:'
	@echo '  $(YELLOW)make${RESET} $(GREEN)<target>$(RESET)'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[0-9a-zA-Z_-]+:.*?##.*$$/) \
		  {printf "    $(YELLOW)%-20s$(GREEN)%s$(RESET)\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  $(CYAN)%s$(RESET)\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

## Version
version: ## Display current version
	@echo $(VERSION)
