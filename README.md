# Zap

A command-line utility for working with MQTT

This project is very much in progress.  Check back soon.

Zap can be used to publish to or subscribe from an mqtt message broker.  It was modeled after the command-line tools [mosquitto_pub](https://mosquitto.org/man/mosquitto_pub-1.html) and [mosquitto_sub](https://mosquitto.org/man/mosquitto_sub-1.html).  In addition, Zap supports a configuration file that makes it easier to work with multiple mqtt brokers.  It also has a stats command that provides a real-time dashboard to the metrics published on the $SYS/# topics.

## Installation

To install from source, do:
```bash
$ make install
```

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

In this example, the default using of zap would connect to localhost and use a client of something like me-9735.  However, if you specify the broker *zap -b production* then the server would be set to mqtt.production.com and the client id will be "prod-viewer" and the username and password would be used for connecting to the sever.

Of course, _production_ is just a label.  You can name it whatever you want and have as many sections like that as you want.  It is ver useful if you have multiple mqtt servers you deal with.

Note that the global settings are still in effect when specifying a broker.  It is just that the broker will override any global config settings.  Also, any command-line options will override any options set in the config file.

### Global configs
```
  -b, --broker string          broker configuration
      --client-prefix string   prefix to use to generate a client id if none is specified (default "zap_")
      --config string          config file (default is $HOME/.zap.toml)
  -i, --id string              id to use for this client (default is generated from client-prefix)
  -k, --keepalive int          the number of seconds after which a PING is sent to the broker (default 60)
      --password string        password for accessing MQTT
      --qos int                qos setting
      --server string          location of MQTT server (default "tcp://127.0.0.1:1883")
      --topic string           mqtt topic (default "#")
      --username string        username for accessing MQTT
      --verbose                give more verbose information
```

### Publish command

The *zap publish* command allows you to publish to a server on a given topic.

There are several options for how you can pass the data to zap.  Only one of these
options can be used at a time:

| Option         |   Description   |
|---------------:|-----------------|
| --file         |  Takes a file name and sends the entire contents of the file as a single message |
| --message      |  Takes an argument that is the data sent to the broker. |
| --null-message |  Just sends an empty string as a message.  |
| --stdin-file   |  This takes no argument - it reads from stdin until it reaches EOF and sends the entire contents as one message.  |
| --stdin-line   |  This also takes no argument and reads from stdin.  Each new-line sends a new message on the topic.  |

So, for example, the following will send one message to the topic of test/my_test with the contents of Hello World!

```
zap publish --server tcp://test.mosquitto.org:1883 --topic test/my_test --message "Hello World!"
```

### Subscribe command

The *zap subscribe* command allows you to subscribe to topics from an mqtt broker.

Let's test by connecting to a public mqtt server:
zap subscribe --server tcp://test.mosquitto.org:1883 --topic \#

You will see output like this (but the topics and messages will be very different).
```
Received message on topic: iot-2/type/niagara/id/456/evt/event8528/fmt/txt
Message: iot-2/type/niagara/id/456/evt/event8528/fmt/txt
Received message on topic: iot-2/type/niagara/id/456/evt/event8529/fmt/txt
Message: iot-2/type/niagara/id/456/evt/event8529/fmt/txt
Received message on topic: iot-2/type/niagara/id/456/evt/event8530/fmt/txt
Message: iot-2/type/niagara/id/456/evt/event8530/fmt/txt
Received message on topic: iot-2/type/niagara/id/456/evt/event8531/fmt/txt
Message: iot-2/type/niagara/id/456/evt/event8531/fmt/txt
```

You can change how the output looks by using the **--template** flag.  Zap uses the [Go lang template](https://golang.org/pkg/text/template/) language to specify how the output looks.  The default is ```Received message on topic: {{.Topic}}\nMessage: {{.Message}}\n")``` which generates the output above.

So, for exmaple, if you wanted to generate a CSV file of -- topic, message -- you could specify a template like this:
```"{{.Topic}},{{.Message}}\n"```

Note: It can be a pain to specify the \n on the command line.  You just have to hit enter and make it a multi-line command.

### Stats command

The stats command is a fun little tool monitors listens to $SYS/# messages from the broker and displays a real-time textual monitor to what is going on with the broker.  Unfortunately, what the documentation says
about what might be in those messages from actual brokers can both vary or not exist.  Indeed, I'm not even
sure what all of them actually mean.  However, it does work with the open source mqtt server and provides
some interesting insight to the load on the server.

The stats command takes over your terminal and you just hit Q to quit.  Here is an example of what looks like.  (The items saying n/a are values not sent by the broker.)

```
 Now:  Dec 06, 2017 22:38:52     Watching:    0:00:07  [Q] to quit

Broker                                  Load              1 min  5 min 15 min
     Broker Version : n/a                      Sockets :    645    670    689
        Broker Time : n/a                  Connections :    647    668    686
      Broker Uptime : 2971685 seconds     Msg Received :   4364   4510   4643
Subscriptions Count : 2151                    Msg Sent :   7650   8370   9316
   Total Bytes Sent : n/a               Bytes Received : 474966 458255 455120
Total Bytes Received : n/a                  Bytes Sent : 932692 978414 1081784
                                          Pub Received :   1792   1877   1959
Message Stats                                 Pub Sent :   5037   5699   6602
 Messages Received : 314874148             Pub Dropped :   2552   3371   3673
     Messages Sent : 736594949
Messages In-flight : n/a                Clients  
   Messages Stored : 622703                    Clients Total : 1740
                                           Clients Connected : 806
Messages Publish Dropped : 852400923    Clients Disconnected : 934
Messages Publish Sent : n/a                  Clients Expired : 29699
Messages Publish Received : n/a              Clients Maximum : n/a
Messages Retained Count : n/a
```

If you have any ideas on how this could be made more useful please let me know!

## Building the source

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
build           Build the source
install         Builds and installs zap into your go/bin
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
