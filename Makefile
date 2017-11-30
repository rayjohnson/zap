
BINARY=zap
VERSION=0.1.0

SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

LDFLAGS := -ldflags "-X main.VERSION=$(VERSION)"

$(BINARY): $(SOURCES)
	go build ${LDFLAGS} -o ${BINARY} main.go

build: $(BINARY)

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

fmt:
	go fmt

version:
	@echo $(VERSION)

.PHONY: build clean version

