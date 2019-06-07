PREFIX  ?= /usr
DESTDIR ?=
BINDIR  ?= $(PREFIX)/bin

export GOPATH      ?= $(CURDIR)/.gopath
export GO111MODULE := on

all: timefind timefind-indexer

timefind: $(wildcard cmd/timefind/*.go) $(wildcard pkg/*/*.go)
	go build -v -o "$@" ./cmd/timefind

timefind-indexer: $(wildcard cmd/timefind-indexer/*.go) $(wildcard pkg/*/*.go)
	go build -v -o "$@" ./cmd/timefind-indexer

clean:
	$(RM) timefind timefind-indexer

.PHONY: all clean
