
BINARY=zap
VERSION := $(shell cat VERSION)
COMMIT=$(shell git rev-parse HEAD)

SOURCES := $(shell find main.go cmd viewstats -name '*.go')
PKGS := $(shell go list ./... | grep -v /vendor)

LDFLAGS := -ldflags "-X main.VERSION=$(VERSION) -X main.COMMIT=${COMMIT}"
RELEASE_ROOT=release
REPORTS=reports

$(BINARY): $(SOURCES)
	@mkdir -p $(REPORTS)
	$(shell export GORACE=log_path=$(REPORTS)/race.log; go build ${LDFLAGS} -race -o ${BINARY} main.go)

PLATFORMS=darwin linux windows
os = $(word 1, $@)

# Cross compile and build all platforms and add assets
.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p $(RELEASE_ROOT)/$(os)
	GOOS=$(os) GOARCH=amd64 go build -v ${LDFLAGS} -o $(RELEASE_ROOT)/$(os)/$(BINARY)
	cp -r examples $(RELEASE_ROOT)/$(os)/
	cp README.md $(RELEASE_ROOT)/$(os)/

.PHONY: setup
setup:  ## Creates vendor directory with all dependencies
	@dep ensure

.PHONY: build
build: $(BINARY)  ## Build the source

.PHONY: install
install: build  ## Builds and installs zap into your go/bin
	go install ${LDFLAGS}

.PHONY: release
release: windows linux darwin   ## Do cross platform build and package
	mv $(RELEASE_ROOT)/windows/zap $(RELEASE_ROOT)/windows/zap.exe
	#tar -cvzf blah.tar.gz $(RELEASE_ROOT)/linux/*

.PHONY: clean
clean:  ## Clean up any generated files
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	$(shell rm -rf release reports)

.PHONY: lint
lint:  ## Run golint and go fmt on source base
	@go fmt $(PKGS)
	@golint $(PKGS)

.PHONY: test
test:  ## Run test suite (go test)
	@go test $(PKGS)

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
