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

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/rayjohnson/zap/viewstats"
)

const statsTopic = "$SYS/#"
const statsQos = 0

func newStatsCommand() *cobra.Command {
	var zapOpts *zapOptions

	cmd := &cobra.Command{
		Use:   "stats",
		Args:  cobra.NoArgs,
		Short: "Show stats reported by the MQTT broker",
		Long: `Show stats reported by the MQTT broker

The stats command subscribes to the brokers $SYS/# topics to get and
display statistics for how the broker is running.  Not all brokers show
the same information and you need to have permission to view those topics.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStats(cmd.Flags(), zapOpts)
		},
	}
	cmd.SilenceUsage = true

	flags := cmd.Flags()
	zapOpts = buildZapFlags(flags)
	zapOpts.conOpts = addConnectionFlags(flags)

	return cmd
}

func runStats(flags *pflag.FlagSet, zapOpts *zapOptions) error {
	if err := zapOpts.processOptions(flags); err != nil {
		return err
	}
	clientOpts := zapOpts.clientOpts

	// TODO: do I need to do this?  So what if it isn't clean?  I'll get data
	clientOpts.CleanSession = true

	client := MQTT.NewClient(clientOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("could not connect: %s", token.Error())
	}
	defer client.Disconnect(250)

	if zapOpts.verbose {
		fmt.Printf("Connected to %s\n", clientOpts.Servers[0])
	}

	viewstats.PrepViewer()
	if token := client.Subscribe(statsTopic, statsQos, statsHandler); token.Wait() && token.Error() != nil {
		return fmt.Errorf("could not subscribe: %s", token.Error())
	}
	defer client.Unsubscribe(statsTopic)

	viewstats.StartStatsDisplay()
	return nil
}

func statsHandler(client MQTT.Client, msg MQTT.Message) {
	viewstats.AddStat(msg.Topic(), string(msg.Payload()))
}
