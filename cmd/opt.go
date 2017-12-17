package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type connectionOptions struct {
	insecure     bool
	cfgFile      string
	keepAlive    int
	broker       string
	server       string
	username     string
	password     string
	clientID     string
	clientPrefix string
	certFile     string
	keyFile      string
	caFile       string
}

// addConnectionFlags adds to a pFlag set the options related to connecting to a broker
func addConnectionFlags(fs *pflag.FlagSet) *connectionOptions {
	conOpts := &connectionOptions{}

	fs.StringVar(&cfgFile, "config", "", "Config file path (default is $HOME/.zap.toml)")
	fs.BoolVar(&optVerbose, "verbose", false, "Give more verbose information")

	fs.StringVar(&conOpts.server, "server", "tcp://127.0.0.1:1883", "Url of MQTT server")
	fs.StringVar(&conOpts.username, "username", "", "Username for accessing MQTT")
	fs.StringVar(&conOpts.password, "password", "", "Password for accessing MQTT")
	fs.StringVarP(&conOpts.clientID, "id", "i", "", "ID to use for this client (default is generated from client-prefix)")
	fs.StringVar(&conOpts.clientPrefix, "client-prefix", "zap_", "Prefix to use to generate a client id if none is specified")
	fs.IntVarP(&conOpts.keepAlive, "keepalive", "k", 60, "The number of seconds after which a PING is sent to the broker")
	fs.StringVarP(&conOpts.broker, "broker", "b", "", "Broker configuration")
	fs.StringVar(&conOpts.caFile, "tls-cacert", "", "Trust certs signed only by this CA")
	fs.StringVar(&conOpts.certFile, "tls-cert", "", "Path to TLS certificate file")
	fs.StringVar(&conOpts.keyFile, "tls-key", "", "Path to TLS key file")
	fs.BoolVar(&conOpts.insecure, "tls-skip-verify", false, "Skips verification for TLS")

	return conOpts
}

func getCorrectConfigKey(broker string, key string) string {
	section := ""
	if broker != "" {
		section = (broker + ".")
	}

	if viper.IsSet(section + key) {
		return (section + key)
	} else if viper.IsSet(key) {
		return key
	} else {
		return ""
	}
}

// ParseBrokerInfo is called by subcommands tp parse the global option
// values related to connecting to the mqtt broker.
func ParseBrokerInfo(fs *pflag.FlagSet, conOpts *connectionOptions) (*MQTT.ClientOptions, error) {

	// If --broker was set make sure the section is in the config file
	if fs.Lookup("broker").Changed {
		if conOpts.broker != "" {
			// TODO: maybe call InConfig
			table := viper.Sub(conOpts.broker)
			if table == nil {
				return nil, fmt.Errorf("broker \"%s\" could not be found in the config file", conOpts.broker)
			}
			list := table.AllKeys()
			if len(list) == 0 {
				return nil, fmt.Errorf("broker \"%s\" has no keys in the config file", conOpts.broker)
			}
		}
	}

	broker := conOpts.broker

	if !fs.Lookup("server").Changed {
		if key := getCorrectConfigKey(broker, "server"); key != "" {
			conOpts.server = viper.GetString(key)
		}
	}

	if !fs.Lookup("username").Changed {
		if key := getCorrectConfigKey(broker, "username"); key != "" {
			conOpts.username = viper.GetString(key)
		}
	}

	if !fs.Lookup("password").Changed {
		if key := getCorrectConfigKey(broker, "password"); key != "" {
			conOpts.password = viper.GetString(key)
		}
	}

	if !fs.Lookup("tls-cert").Changed {
		if key := getCorrectConfigKey(broker, "tls-cert"); key != "" {
			conOpts.certFile = viper.GetString(key)
		}
	}

	if !fs.Lookup("tls-key").Changed {
		if key := getCorrectConfigKey(broker, "tls-key"); key != "" {
			conOpts.keyFile = viper.GetString(key)
		}
	}

	if !fs.Lookup("tls-cacert").Changed {
		if key := getCorrectConfigKey(broker, "cafile"); key != "" {
			conOpts.caFile = viper.GetString(key)
		}
	}

	if !fs.Lookup("tls-skip-verify").Changed {
		if key := getCorrectConfigKey(broker, "tls-skip-verify"); key != "" {
			conOpts.insecure = viper.GetBool(key)
		}
	}

	if !fs.Lookup("keepalive").Changed {
		if key := getCorrectConfigKey(broker, "keepalive"); key != "" {
			conOpts.keepAlive = viper.GetInt(key)
		}
	}

	// if !fs.Lookup("topic").Changed {
	// 	if key := getCorrectConfigKey(broker, "topic"); key != "" {
	// 		optTopic = viper.GetString(key)
	// 	}
	// }

	if !fs.Lookup("client-prefix").Changed {
		if key := getCorrectConfigKey(broker, "client-prefix"); key != "" {
			conOpts.clientPrefix = viper.GetString(key)
		}
	}

	if !fs.Lookup("id").Changed {
		if key := getCorrectConfigKey(broker, "id"); key != "" {
			conOpts.clientID = viper.GetString(key)
		}
	}

	// If client id is not set we will generate one here
	if conOpts.clientID == "" {
		if conOpts.clientPrefix == "" {
			conOpts.clientPrefix = "zap_"
		}

		conOpts.clientID = fmt.Sprintf("%s%s", conOpts.clientPrefix, strconv.Itoa(os.Getpid()))
	}

	clientOpts := MQTT.NewClientOptions()
	clientOpts.SetClientID(conOpts.clientID)
	clientOpts.SetUsername(conOpts.username)
	clientOpts.SetPassword(conOpts.password)
	clientOpts.SetKeepAlive(time.Duration(conOpts.keepAlive) * time.Second)

	if _, err := url.ParseRequestURI(conOpts.server); err != nil {
		return nil, err
	}
	clientOpts.AddBroker(conOpts.server)

	// tls set up
	tlsConfig := tls.Config{InsecureSkipVerify: conOpts.insecure}
	// if either option is set
	if conOpts.certFile != "" || conOpts.keyFile != "" {
		// make sure both options are set
		if conOpts.certFile == "" || conOpts.keyFile == "" {
			return nil, fmt.Errorf("for tls: both --tls-key and --tls-cert options must be set")
		}
		cert, err := tls.LoadX509KeyPair(conOpts.certFile, conOpts.keyFile)
		if err != nil {
			return nil, err
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	} else {
		tlsConfig.ClientAuth = tls.NoClientCert
	}

	if conOpts.caFile != "" {
		// Load CA cert
		caCert, err := ioutil.ReadFile(conOpts.caFile)
		if err != nil {
			return nil, err
			// fmt.Printf("could not read cafile: %s\n", err). TODO: should error be more specific?
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	clientOpts.SetTLSConfig(&tlsConfig)

	return clientOpts, nil
}

// PrintConnectionInfo will so all the args used if verbose is on
func PrintConnectionInfo(conOpts *connectionOptions, subOpts *subscribeOptions, pubOpts *publishOptions) {
	if optVerbose {
		fmt.Println("Connecting to server with following parameters")
		if conOpts.broker != "" {
			fmt.Println("  From broker config: ", conOpts.broker)
		}
		fmt.Println("  Server: ", conOpts.server)
		if conOpts.keyFile != "" {
			fmt.Println("  TLS Key path: ", conOpts.keyFile)
			fmt.Println("  TLS Cert path: ", conOpts.certFile)
		}
		if conOpts.caFile != "" {
			fmt.Println("  TLS CA path: ", conOpts.caFile)
		}
		if conOpts.insecure {
			fmt.Println("  TLS Skip Verify: ", conOpts.insecure)
		}
		fmt.Println("  ClientId: ", conOpts.clientID)
		fmt.Println("  Username: ", conOpts.username)
		fmt.Println("  Password: ", conOpts.password)
		if subOpts != nil {
			fmt.Println("  QOS: ", subOpts.qos)
			fmt.Println("  Topic: ", subOpts.topic)
		}
		if pubOpts != nil {
			fmt.Println("  QOS: ", pubOpts.qos)
			fmt.Println("  Topic: ", pubOpts.topic)
		}
	}
}
