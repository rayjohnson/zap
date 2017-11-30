
BINARY=zap
VERSION=0.1.0

SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

$(BINARY): $(SOURCES)
	go build ${LDFLAGS} -o ${BINARY} main.go

build: $(BINARY)

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

fmt:
	go fmt

.PHONY: build clean

