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

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/rayjohnson/zap/viewstats"
)

const statsTopic = "$SYS/#"

func newStatsCommand() *cobra.Command {
	var conOpts *connectionOptions

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show stats reported by the MQTT broker",
		Long: `Show stats reported by the MQTT broker

The stats command subscribes to the brokers $SYS/# topics to get and
display statistics for how the broker is running.  Not all brokers show
the same information and you need to have permission to view those topics.`,
		Run: func(cmd *cobra.Command, args []string) {
			runStats(cmd.Flags(), conOpts)
		},
	}

	flags := cmd.Flags()
	conOpts = addConnectionFlags(flags)

	return cmd
}

func runStats(flags *pflag.FlagSet, conOpts *connectionOptions) {
	clientOpts := ParseBrokerInfo(flags, conOpts)
	clientOpts.CleanSession = true

	PrintConnectionInfo(conOpts)

	exitWithError := false
	defer func() {
		if exitWithError {
			os.Exit(1)
		}
	}()

	client := MQTT.NewClient(clientOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Could not connect: %s\n", token.Error())
		exitWithError = true
		return
	}
	defer client.Disconnect(250)

	if optVerbose {
		fmt.Printf("Connected to %s\n", clientOpts.Servers[0])
	}

	viewstats.PrepViewer()
	if token := client.Subscribe(statsTopic, byte(optQos), statsHandler); token.Wait() && token.Error() != nil {
		exitWithError = true
		fmt.Printf("Could not subscribe: %s\n", token.Error())
		return
	}
	defer client.Unsubscribe(statsTopic)

	viewstats.StartStatsDisplay()
}

func statsHandler(client MQTT.Client, msg MQTT.Message) {
	viewstats.AddStat(msg.Topic(), string(msg.Payload()))
}
