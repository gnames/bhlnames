VERSION = $(shell git describe --tags)
VER = $(shell git describe --tags --abbrev=0)
DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S%Z')

FLAGS_LINUX = $(FLAGS_SHARED) GOOS=linux
FLAGS_MAC = $(FLAGS_SHARED) GOOS=darwin
FLAGS_WIN = $(FLAGS_SHARED) GOOS=windows
FLAG_MODULE = GO111MODULE=on
NO_C = CGO_ENABLED=0

FLAGS_SHARED = $(FLAG_MODULE) CGO_ENABLED=0 GOARCH=amd64
FLAGS_LD=-ldflags "-X github.com/gnames/bhlnames.Build=${DATE} \
                  -X github.com/gnames/bhlnames.Version=${VERSION}"
GOCMD=go
GOINSTALL=$(GOCMD) install $(FLAGS_LD)
GOBUILD=$(GOCMD) build $(FLAGS_LD)
GOCLEAN=$(GOCMD) clean
GOGET = $(GOCMD) get

all: install

test: deps install
	$(FLAG_MODULE) go test ./...

deps:
	$(FLAG_MODULE) $(GOGET) github.com/spf13/cobra/cobra@f2b07da; \
	$(FLAG_MODULE) $(GOGET) github.com/onsi/ginkgo/ginkgo@505cc35; \
	$(FLAG_MODULE) $(GOGET) github.com/onsi/gomega@ce690c5; \

build:
	cd bhlnames; \
	$(GOCLEAN); \
	$(FLAGS_SHARED) GOOS=linux $(GOBUILD);

release:
	cd bhlnames; \
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
	cd bhlnames; \
	$(FLAGS_SHARED) $(GOINSTALL);
