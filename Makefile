
BINARY=zap
VERSION := $(shell cat VERSION)

SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

LDFLAGS := -ldflags "-X main.VERSION=$(VERSION)"

$(BINARY): $(SOURCES)
	@go build ${LDFLAGS} -o ${BINARY} main.go

.PHONY: setup
setup:  ## Creates vendor directory with all dependencies
	@dep ensure

.PHONY: build
build: $(BINARY)

.PHONY: clean
clean:  ## Clean up any generated files
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	@if [ -f reports ] ; then rm reports ; fi

.PHONY: fmt
fmt:  ## Run go fmt on source base
	@go fmt ./...

.PHONY: dep_graph
dep_graph:  ## Generate a dependency graph from dep and graphvis
	@mkdir -p reports
	@dep status -dot | dot -T png > reports/dependancy_graph.png

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



