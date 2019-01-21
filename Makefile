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

.PHONY: startlocal
startlocal:
	docker run -d --rm --name tsheetsdevmariadb -p 3306:3306 -eMARIADB_ROOT_PASSWORD=Password1 -eMARIADB_DATABASE=tsheetsdev -eMARIADB_USER=tsheetsuser -eMARIADB_PASSWORD=changeme -v $(PWD)/install:/docker-entrypoint-initdb.d mariadb/server:10.3

.PHONY: stoplocal
stoplocal:
	docker stop tsheetsdevmariadb
