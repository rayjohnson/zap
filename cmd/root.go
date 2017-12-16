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
	"log"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var optVerbose bool
var cfgFile string

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
func Execute(ver string, rev string) {
	version = ver
	revision = rev

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// TODO: maybe get rid of init and do this more explicitly
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(
		newSubscribeCommand(),
		newVersionCommand(),
		newPublishCommand(),
		newStatsCommand(),
	)
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
	viper.ReadInConfig()

	// Uncomment these to turn on debugging from within the mqtt library.
	MQTT.ERROR = log.New(os.Stdout,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	MQTT.CRITICAL = log.New(os.Stdout,
		"CRITICAL: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	// MQTT.WARN = log.New(os.Stdout,
	//        "WARN: ",
	//        log.Ldate|log.Ltime|log.Lshortfile)
	// MQTT.DEBUG = log.New(os.Stdout,
	//        "DEBUG: ",
	//        log.Ldate|log.Ltime|log.Lshortfile)

}
