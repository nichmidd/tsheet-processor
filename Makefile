.PHONY: all
all: test

BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

.PHONY: deps
deps: $(GOMETALINTER) 

.PHONY: lint
lint: deps
	gometalinter --deadline=120s --vendor

.PHONY: test
test:
	go test

BINARY := tsheet-processor
VERSION ?= vlatest

.PHONY: release
release:
	mkdir -p release
	go build -o release/$(BINARY)-$(VERSION)
