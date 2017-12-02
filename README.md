# Zap

A command-line utility for working with MQTT

This project is very much in progress.  Check back soon.

### Building the source

This project depends on the following tools:
* golang
* dep
* graphviz

```bash
brew install graphviz
brew install dep
```

I have a Makefile that manages most things.

Run this to get an understanding of what the Makefile does:
```bash
$ make help
setup           Creates vendor directory with all dependencies
clean           Clean up any generated files
lint            Run golint and go fmt on source base
dep_graph       Generate a dependency graph from dep and graphvis
help            Display this help message
todo            Greps for any TODO comments in the source code
version         Show the version the Makefile will build
```

To build everything you should just need to do:
```bash
$ make setup
$ make build
```

NOTE: I still need to build support for cross compilation and releases
