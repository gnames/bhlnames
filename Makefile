VERSION = $(shell git describe --tags)
VER = $(shell git describe --tags --abbrev=0)
DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S%Z')

FLAGS_LINUX = $(FLAGS_SHARED) GOOS=linux
FLAGS_MAC = $(FLAGS_SHARED) GOOS=darwin
FLAGS_WIN = $(FLAGS_SHARED) GOOS=windows
NO_C = CGO_ENABLED=0

FLAGS_SHARED = CGO_ENABLED=0 GOARCH=amd64
FLAGS_LD = -ldflags "-X github.com/gnames/bhlnames/pkg.Build=${DATE} \
                  -X github.com/gnames/bhlnames/pkg.Version=${VERSION}"
GOCMD = go
GOINSTALL = $(GOCMD) install $(FLAGS_LD)
GOBUILD = $(GOCMD) build $(FLAGS_LD)
GOCLEAN = $(GOCMD) clean
GOGET = $(GOCMD) get

all: install

test: deps install
	go test -shuffle=on -race -coverprofile=coverage.txt -covermode=atomic ./...

deps:
	$(GOCMD) mod download;

tools: deps
	@echo Installing tools from tools.go
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

build:
	$(GOCLEAN); \
	$(FLAGS_SHARED) GOOS=linux $(GOBUILD);

release:
	$(GOCLEAN); \
	$(FLAGS_SHARED) GOOS=linux $(GOBUILD); \
	tar zcvf /tmp/bhlnames-${VER}-linux.tar.gz bhlnames; \
	$(GOCLEAN); \
	$(FLAGS_SHARED) GOOS=darwin $(GOBUILD); \
	tar zcvf /tmp/bhlnames-${VER}-mac.tar.gz bhlnames; \
	$(GOCLEAN); \
	$(FLAGS_WIN) $(NO_C) $(GOBUILD); \
	zip -9 /tmp/bhlnames-$(VER)-win-64.zip bhlnames.exe; \
	$(GOCLEAN);

install:
	$(FLAGS_SHARED) $(GOINSTALL);

dc: build
	docker-compose build; \
