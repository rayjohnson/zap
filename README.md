# Zap

A command-line utility for working with MQTT

This project is very much in progress.  Check back soon.

Zap can be used to publish to or subscribe from an mqtt message broker.  It was modeled after the command-line tools [mosquitto_pub](https://mosquitto.org/man/mosquitto_pub-1.html) and [mosquitto_sub](https://mosquitto.org/man/mosquitto_sub-1.html).  In addition, Zap supports a configuration file that makes it easier to work with multiple mqtt brokers.  It also has a stats command that provides a real-time dashboard to the metrics published on the $SYS/# topics.

## Installation

TODO

## Configuration file

Zap supports having a configuration file that can set many of the common options related to connecting to a given mqtt broker.

You can create the configuration file at ~/.zap.toml.  (Though you can override the location with the --config option.). The file uses the [Toml(https://github.com/toml-lang/toml)] format.  

Here is an example of the options that can be specified in the configuration file:
```toml
# Sample configuration
server = "tcp://mqtt.local:1883"
client-prefix = "me-"
id = "me"     # This will override the client-prefix setting
username = "me"
password = "secret_password"
keepalive = 30
qos = 0
```

The configuration file also supports having multiple sets of "brokers" that you can then
use with the --broker option.  Take the following example:
```toml
server = "tcp://localhost:1883"
client-prefix = "me-"

[production]
server = "tcp://mqtt.production.com:1883"
id = "prod-viewer"
username = "me"
password = "secret_password"
```

In this example, the default using of zap would connect to localhost and use a client if of something like me-9735.  However, if you specify the broker *zap -b production* then the server would be set to mqtt.production.com and the client id will be "prod-viewer" and the username and password would be used for connecting to the sever.

Of course, _production_ is just a label.  You can name it whatever you want and have as many sections like that as you want.  It is ver useful if you have multiple mqtt servers you deal with.

Note that the global settings are still in effect when specifying a broker.  It is just that the broker will override any global config settings.  Also, any command-line options will override any options set in the config file.

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
