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
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var VERSION string

var cfgFile string
var Broker string
var Server string
var Username string
var Password string
var ClientId string
var ClientPrefix string
var Qos int16
var KeepAlive int64

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
func Execute(version string) {
	VERSION = version

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

// This should only be called by subcommands
func ParseBrokerInfo(cmd *cobra.Command, args []string) {
	if (! cmd.Parent().PersistentFlags().Lookup("server").Changed) {
		if key := getCorrectConfigKey(Broker, "server"); key != "" {
			Server = viper.GetString(key)
		}
	}

	if (! cmd.Parent().PersistentFlags().Lookup("username").Changed) {
		if key := getCorrectConfigKey(Broker, "username"); key != "" {
			Username = viper.GetString(key)
		}
	}

	if (! cmd.Parent().PersistentFlags().Lookup("password").Changed) {
		if key := getCorrectConfigKey(Broker, "password"); key != "" {
			Password = viper.GetString(key)
		}
	}

	if (! cmd.Parent().PersistentFlags().Lookup("qos").Changed) {
		if key := getCorrectConfigKey(Broker, "qos"); key != "" {
			Qos = int16(viper.GetInt(key))
		}
	}

	if (! cmd.Parent().PersistentFlags().Lookup("keepalive").Changed) {
		if key := getCorrectConfigKey(Broker, "keepalive"); key != "" {
			KeepAlive = viper.GetInt64(key)
		}
	}

	// TODO: need to implement --client-prefix here as well
	// TODO: should handle construction of dynamic default here too
	if (! cmd.Parent().PersistentFlags().Lookup("client-prefix").Changed) {
		if key := getCorrectConfigKey(Broker, "client-prefix"); key != "" {
			ClientPrefix = viper.GetString(key)
		}
	}

	if (! cmd.Parent().PersistentFlags().Lookup("id").Changed) {
		if key := getCorrectConfigKey(Broker, "id"); key != "" {
			ClientId = viper.GetString(key)
		}
	}

	// If client id is not set we will generate one here
	if ClientId == "" {
		if ClientPrefix == "" {
			ClientPrefix = "zap_"
		}

		ClientId = fmt.Sprintf("%S%S", ClientPrefix, os.Getpid())
	}
}

func init() { 
	cobra.OnInitialize(initConfig)

	// Set up flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.zap.toml)")

	rootCmd.PersistentFlags().StringVarP(&Server, "server", "s", "tcp://127.0.0.1:1883", "location of MQTT server")
	rootCmd.PersistentFlags().StringVar(&Username, "username", "", "username for accessing MQTT")
	rootCmd.PersistentFlags().StringVar(&Password, "password", "", "password for accessing MQTT")
	rootCmd.PersistentFlags().StringVarP(&ClientId, "id", "i", "", "id to use for this client (default is generated from client-prefix)")
	rootCmd.PersistentFlags().StringVar(&ClientPrefix, "client-prefix", "zap_", "prefix to use to generate a client id if none is specified")
	rootCmd.PersistentFlags().Int16Var(&Qos, "qos", 1, "qos setting")
	rootCmd.PersistentFlags().Int64VarP(&KeepAlive, "keepalive", "k", 60, "the number of seconds after which a PING is sent to the broker")
	rootCmd.PersistentFlags().StringVarP(&Broker, "broker", "b", "", "broker configuration")

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
