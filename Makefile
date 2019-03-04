BINARY=slack-notify
BINARY_LINUX=slack-notify-linux

GOBUILD=go build
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%S%z)
BUILD_COMMIT=$(shell git rev-list -1 HEAD | cut -c1-6)
LDFLAGS="-s -w -X main.buildTime=$(BUILD_TIME) -X main.buildCommit=$(BUILD_COMMIT)"
UNAME_S := $(shell uname -s)

.PHONY: clean

all: $(BINARY)

$(BINARY): cli/*.go
	$(GOBUILD) -ldflags=$(LDFLAGS) -v -o $(BINARY) cli/*.go

$(BINARY_LINUX): cli/*.go
	GOOS=linux $(GOBUILD) -ldflags=$(LDFLAGS) -v -o $(BINARY_LINUX) cli/*.go
	upx $(BINARY_LINUX)

clean:
	/bin/rm -f $(BINARY)