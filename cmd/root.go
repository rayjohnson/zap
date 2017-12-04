// Copyright © 2017 Ray Johnson <ray.johnson@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var VERSION string
var COMMIT string

var optVerbose bool
var cfgFile string
var optBroker string
var optServer string
var optUsername string
var optPassword string
var optClientId string
var optClientPrefix string
var Qos int
var optKeepAlive int

// TODO: move topic to sub-command - each needs different defaults
var Topic string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "zap",
	Short: "Listen or publish to a MQTT broker",
	Long: `zap - what happens when technology meets mosquito

zap is a little utility for publishing or subscribing to events for the
MQTT message bus`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string, commit string) {
	VERSION = version
	COMMIT = commit

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
func ParseBrokerInfo(cmd *cobra.Command, args []string) *MQTT.ClientOptions {
	broker := optBroker

	if !cmd.Parent().PersistentFlags().Lookup("server").Changed {
		if key := getCorrectConfigKey(broker, "server"); key != "" {
			optServer = viper.GetString(key)
		}
	}

	if !cmd.Parent().PersistentFlags().Lookup("username").Changed {
		if key := getCorrectConfigKey(broker, "username"); key != "" {
			optUsername = viper.GetString(key)
		}
	}

	if !cmd.Parent().PersistentFlags().Lookup("password").Changed {
		if key := getCorrectConfigKey(broker, "password"); key != "" {
			optPassword = viper.GetString(key)
		}
	}

	if !cmd.Parent().PersistentFlags().Lookup("qos").Changed {
		if key := getCorrectConfigKey(broker, "qos"); key != "" {
			Qos = viper.GetInt(key)
		}
	}
	if qos < 0 || qos > 2 {
		fmt.Println("qos value must or 0, 1 or 2")
		os.Exit(1)
	}

	if !cmd.Parent().PersistentFlags().Lookup("keepalive").Changed {
		if key := getCorrectConfigKey(broker, "keepalive"); key != "" {
			optKeepAlive = viper.GetInt(key)
		}
	}

	if !cmd.Parent().PersistentFlags().Lookup("topic").Changed {
		if key := getCorrectConfigKey(broker, "topic"); key != "" {
			Topic = viper.GetString(key)
		}
	}

	if !cmd.Parent().PersistentFlags().Lookup("client-prefix").Changed {
		if key := getCorrectConfigKey(broker, "client-prefix"); key != "" {
			optClientPrefix = viper.GetString(key)
		}
	}

	if !cmd.Parent().PersistentFlags().Lookup("id").Changed {
		if key := getCorrectConfigKey(broker, "id"); key != "" {
			optClientId = viper.GetString(key)
		}
	}

	// If client id is not set we will generate one here
	if optClientId == "" {
		if optClientPrefix == "" {
			optClientPrefix = "zap_"
		}

		optClientId = fmt.Sprintf("%s%s", optClientPrefix, strconv.Itoa(os.Getpid()))
	}

	connOpts := &MQTT.ClientOptions{
		ClientID:             optClientId,
		CleanSession:         cleanSession,
		Username:             optUsername,
		Password:             optPassword,
		MaxReconnectInterval: 1 * time.Second,
		KeepAlive:            time.Duration(optKeepAlive),
		TLSConfig:            tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert},
	}
	connOpts.AddBroker(optServer)

	return connOpts
}

// PrintConnectionInfo will so all the args used if verbose is on
func PrintConnectionInfo() {
	if optVerbose {
		fmt.Println("Connecting to server with following parameters")
		if optBroker != "" {
			fmt.Println(". From broker config: ", optBroker)
		}
		fmt.Println("  Server: ", optServer)
		fmt.Println("  ClientId: ", optClientId)
		fmt.Println("  Username: ", optUsername)
		fmt.Println("  Password: ", optPassword)
		fmt.Println("  QOS: ", Qos)
		fmt.Println("  Retain: ", Retain)
		fmt.Println("  Topic: ", Topic)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set up flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.zap.toml)")

	rootCmd.PersistentFlags().StringVar(&optServer, "server", "tcp://127.0.0.1:1883", "location of MQTT server")
	rootCmd.PersistentFlags().StringVar(&optUsername, "username", "", "username for accessing MQTT")
	rootCmd.PersistentFlags().StringVar(&optPassword, "password", "", "password for accessing MQTT")
	rootCmd.PersistentFlags().StringVarP(&optClientId, "id", "i", "", "id to use for this client (default is generated from client-prefix)")
	rootCmd.PersistentFlags().StringVar(&optClientPrefix, "client-prefix", "zap_", "prefix to use to generate a client id if none is specified")
	rootCmd.PersistentFlags().IntVar(&Qos, "qos", 1, "qos setting")
	rootCmd.PersistentFlags().IntVarP(&optKeepAlive, "keepalive", "k", 60, "the number of seconds after which a PING is sent to the broker")
	rootCmd.PersistentFlags().StringVarP(&optBroker, "broker", "b", "", "broker configuration")
	rootCmd.PersistentFlags().BoolVar(&optVerbose, "verbose", false, "give more verbose information")

	// TODO: this should move to sub-command so it has different defaults
	rootCmd.PersistentFlags().StringVar(&Topic, "topic", "#", "mqtt topic")
}

// initConfig reads in config file if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".zap" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".zap")
	}

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
	}
}
