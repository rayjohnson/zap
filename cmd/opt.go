package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
	"github.com/spf13/pflag"
)

type connectionOptions struct {
	insecure     bool
	keepAlive    int
	server       string
	username     string
	password     string
	clientID     string
	clientPrefix string
	certFile     string
	keyFile      string
	caFile       string
}

type zapOptions struct {
	configFile string
	broker     string
	verbose    bool
	conOpts    *connectionOptions
	clientOpts *MQTT.ClientOptions
	pubOpts    *publishOptions
	subOpts    *subscribeOptions

	configTree *toml.Tree
}

// these are really global flags - but the struct will hold pointers to all other types
func buildZapFlags(fs *pflag.FlagSet) *zapOptions {
	zapOpts := &zapOptions{}

	fs.StringVar(&zapOpts.configFile, "config", "", "Config file path (default is $HOME/.zap.toml)")
	fs.BoolVar(&zapOpts.verbose, "verbose", false, "Give more verbose information")
	fs.StringVarP(&zapOpts.broker, "broker", "b", "", "Specifies a section of the config file to use")

	return zapOpts
}

// addConnectionFlags adds to a pFlag set the options related to connecting to a broker
func addConnectionFlags(fs *pflag.FlagSet) *connectionOptions {
	conOpts := &connectionOptions{}

	fs.StringVar(&conOpts.server, "server", "tcp://127.0.0.1:1883", "Url of MQTT server")
	fs.StringVar(&conOpts.username, "username", "", "Username for accessing MQTT")
	fs.StringVar(&conOpts.password, "password", "", "Password for accessing MQTT")
	fs.StringVarP(&conOpts.clientID, "id", "i", "", "ID to use for this client (default is generated from client-prefix)")
	fs.StringVar(&conOpts.clientPrefix, "client-prefix", "zap_", "Prefix to use to generate a client id if none is specified")
	fs.IntVarP(&conOpts.keepAlive, "keepalive", "k", 60, "The number of seconds after which a PING is sent to the broker")
	fs.StringVar(&conOpts.caFile, "tls-cacert", "", "Trust certs signed only by this CA")
	fs.StringVar(&conOpts.certFile, "tls-cert", "", "Path to TLS certificate file")
	fs.StringVar(&conOpts.keyFile, "tls-key", "", "Path to TLS key file")
	fs.BoolVar(&conOpts.insecure, "tls-skip-verify", false, "Skips verification for TLS")

	return conOpts
}

func loadConfigFile(zapOpts *zapOptions) error {
	var configFile string
	if zapOpts.configFile != "" {
		if _, err := os.Stat(zapOpts.configFile); os.IsNotExist(err) {
			// return err because user passed in the config file to user
			return fmt.Errorf("path from --config option does not exist: %s", zapOpts.configFile)
		}

		configFile = zapOpts.configFile
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return err
		}
		configFile = path.Join(home, ".zap.toml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			// you do not have to have a config file
			return nil
		}
	}
	configTree, err := toml.LoadFile(configFile)
	if err != nil {
		return fmt.Errorf("error loading config file: %s", err.Error())
	}

	if zapOpts.broker != "" {
		if configTree.Has(zapOpts.broker) {
			configTree = configTree.Get(zapOpts.broker).(*toml.Tree)
		} else {
			return fmt.Errorf("broker \"%s\" does not exist in config file: %s", zapOpts.broker, zapOpts.configFile)
		}
	}

	zapOpts.configTree = configTree
	return nil
}

func getValueFromConfig(fs *pflag.FlagSet, configTree *toml.Tree, key string, def interface{}) interface{} {
	if fs.Lookup(key).Changed {
		return def // value was passed as command line option no change needed
	}
	if configTree.Has(key) {
		return configTree.Get(key)
	}

	// If not set by command line option this will be default value
	return def
}

func (zapOpts *zapOptions) processOptions(fs *pflag.FlagSet) error {
	var err error
	if err = loadConfigFile(zapOpts); err != nil {
		return err
	}

	// get values from config file if they exist and are not overridden by a flag
	conOpts := zapOpts.conOpts
	if zapOpts.configTree != nil {
		conOpts.server = getValueFromConfig(fs, zapOpts.configTree, "server", conOpts.server).(string)
		conOpts.username = getValueFromConfig(fs, zapOpts.configTree, "username", conOpts.username).(string)
		conOpts.password = getValueFromConfig(fs, zapOpts.configTree, "password", conOpts.password).(string)
		conOpts.clientID = getValueFromConfig(fs, zapOpts.configTree, "id", conOpts.clientID).(string)
		conOpts.clientPrefix = getValueFromConfig(fs, zapOpts.configTree, "client-prefix", conOpts.clientPrefix).(string)
		conOpts.keepAlive = getValueFromConfig(fs, zapOpts.configTree, "keepalive", conOpts.keepAlive).(int)
		conOpts.caFile = getValueFromConfig(fs, zapOpts.configTree, "tls-cacert", conOpts.caFile).(string)
		conOpts.certFile = getValueFromConfig(fs, zapOpts.configTree, "tls-cert", conOpts.certFile).(string)
		conOpts.keyFile = getValueFromConfig(fs, zapOpts.configTree, "tls-key", conOpts.keyFile).(string)
		conOpts.insecure = getValueFromConfig(fs, zapOpts.configTree, "tls-skip-verify", conOpts.insecure).(bool)

		if zapOpts.subOpts != nil {
			subOpts := zapOpts.subOpts
			subOpts.cleanSession = getValueFromConfig(fs, zapOpts.configTree, "clean-session", subOpts.cleanSession).(bool)
			subOpts.templateString = getValueFromConfig(fs, zapOpts.configTree, "template", subOpts.templateString).(string)
			subOpts.count = getValueFromConfig(fs, zapOpts.configTree, "count", subOpts.count).(int)
			subOpts.qos = getValueFromConfig(fs, zapOpts.configTree, "qos", subOpts.qos).(int)
			subOpts.skipRetained = getValueFromConfig(fs, zapOpts.configTree, "skip-retained", subOpts.skipRetained).(bool)

			// subscribe and publish share the same --topic flag but have different defaults
			// so in the config file this requires you to specify subscribe-topic as the value for --topic
			// TODO - need to figure out how to get topic to be different in config
			subOpts.topic = getValueFromConfig(fs, zapOpts.configTree, "topic", subOpts.topic).(string)
		}
		if zapOpts.pubOpts != nil {
			pubOpts := zapOpts.pubOpts
			// Flags like --message, etc. I do not think make sense to specify in config file - skip them
			pubOpts.qos = getValueFromConfig(fs, zapOpts.configTree, "qos", pubOpts.qos).(int)
			pubOpts.topic = getValueFromConfig(fs, zapOpts.configTree, "topic", pubOpts.topic).(string)
		}
	}

	zapOpts.clientOpts, err = conOpts.buildAndValidateClientOpts()
	if err != nil {
		return err
	}

	if zapOpts.pubOpts != nil {
		err := zapOpts.pubOpts.validateOptions()
		if err != nil {
			return err
		}
	}

	if zapOpts.subOpts != nil {
		err := zapOpts.subOpts.validateOptions()
		if err != nil {
			return err
		}

		// CleanSession is only a subscribe option - but must be set on clientOptions
		zapOpts.clientOpts.CleanSession = zapOpts.subOpts.cleanSession
	}

	zapOpts.PrintConnectionInfo()

	return nil
}

func (conOpts *connectionOptions) buildAndValidateClientOpts() (*MQTT.ClientOptions, error) {
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
func (zapOpts *zapOptions) PrintConnectionInfo() {
	if zapOpts.verbose {
		fmt.Println("Connecting to server with following parameters")
		if zapOpts.broker != "" {
			fmt.Println("  From broker config: ", zapOpts.broker)
		}
		fmt.Println("  Server: ", zapOpts.conOpts.server)
		if zapOpts.conOpts.keyFile != "" {
			fmt.Println("  TLS Key path: ", zapOpts.conOpts.keyFile)
			fmt.Println("  TLS Cert path: ", zapOpts.conOpts.certFile)
		}
		if zapOpts.conOpts.caFile != "" {
			fmt.Println("  TLS CA path: ", zapOpts.conOpts.caFile)
		}
		if zapOpts.conOpts.insecure {
			fmt.Println("  TLS Skip Verify: ", zapOpts.conOpts.insecure)
		}
		fmt.Println("  ClientId: ", zapOpts.conOpts.clientID)
		fmt.Println("  Username: ", zapOpts.conOpts.username)
		fmt.Println("  Password: ", zapOpts.conOpts.password)
		if zapOpts.subOpts != nil {
			fmt.Println("  QOS: ", zapOpts.subOpts.qos)
			fmt.Println("  Topic: ", zapOpts.subOpts.topic)
		}
		if zapOpts.pubOpts != nil {
			fmt.Println("  QOS: ", zapOpts.pubOpts.qos)
			fmt.Println("  Topic: ", zapOpts.pubOpts.topic)
		}
	}
}
