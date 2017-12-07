
BINARY=zap
VERSION := $(shell cat VERSION)
COMMIT=$(shell git rev-parse HEAD)

SOURCES := $(shell find main.go cmd viewstats -name '*.go')

LDFLAGS := -ldflags "-X main.VERSION=$(VERSION) -X main.COMMIT=${COMMIT}"
PLATFORMS=darwin linux windows
ARCHITECTURES=386 amd64
RELEASE_ROOT=release
REPORTS=reports

$(BINARY): $(SOURCES)
	@mkdir -p $(REPORTS)
	$(shell export GORACE=log_path=$(REPORTS)/race.log; go build ${LDFLAGS} -race -o ${BINARY} main.go)

$(RELEASE_ROOT)/darwin-amd64/$(BINARY): $(SOURCES)
	$(shell export GOOS=darwin; export GOARCH=amd64; go build -v ${LDFLAGS} -o $(RELEASE_ROOT)/darwin-amd64/$(BINARY))

$(RELEASE_ROOT)/linux-amd64/$(BINARY): $(SOURCES)
	$(shell export GOOS=linux; export GOARCH=amd64; go build -v ${LDFLAGS} -o $(RELEASE_ROOT)/linux-amd64/$(BINARY))

$(RELEASE_ROOT)/windows-amd64/$(BINARY).exe: $(SOURCES)
	$(shell export GOOS=windows; export GOARCH=amd64; go build -v ${LDFLAGS} -o $(RELEASE_ROOT)/windows-amd64/$(BINARY).exe)

.PHONY: setup
setup:  ## Creates vendor directory with all dependencies
	@dep ensure

.PHONY: build
build: $(BINARY)  ## Build the source

.PHONY: install
install: build. ## Builds and installs zap into your go/bin
	go install ${LDFLAGS}

.PHONY: build_all
build_all: $(RELEASE_ROOT)/darwin-amd64/$(BINARY) $(RELEASE_ROOT)/linux-amd64/$(BINARY) $(RELEASE_ROOT)/windows-amd64/$(BINARY).exe

.PHONY: clean
clean:  ## Clean up any generated files
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	@if [ -f reports ] ; then rm reports ; fi
	@if [ -f release ] ; then rm release ; fi

.PHONY: lint
lint:  ## Run golint and go fmt on source base
	@go fmt $(shell go list ./... | grep -v /vendor/)
	@golint $(shell go list ./... | grep -v /vendor/)

.PHONY: dep_graph
dep_graph:  ## Generate a dependency graph from dep and graphvis
	@mkdir -p $(REPORTS)
	@dep status -dot | dot -T png > $(REPORTS)/dependancy_graph.png

.PHONY: help
help:   ## Display this help message
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: todo
todo:   ## Greps for any TODO comments in the source code
	@grep "// TODO" $(SOURCES)

.PHONY: version
version:  ## Show the version the Makefile will build
	@echo $(VERSION)
