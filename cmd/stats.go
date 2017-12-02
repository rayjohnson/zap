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

	"github.com/rayjohnson/zap/viewstats"
)

const statsTopic = "$SYS/#"

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show stats reported by the MQTT broker",
	Long: `Show stats reported by the MQTT broker

TODO: a little more documentation about what the values mean`,
	Run: stats,
}

func stats(cmd *cobra.Command, args []string) {
	connOpts := ParseBrokerInfo(cmd, args)
	connOpts.CleanSession = true

	mqInbound := make(chan [2]string)

	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(statsTopic, byte(Qos), func(client MQTT.Client, msg MQTT.Message) {
			mqInbound <- [2]string{msg.Topic(), string(msg.Payload())}
		}); token.Wait() && token.Error() != nil {
			fmt.Printf("Could not subscribe: %s\n", token.Error())
			os.Exit(1)
		}
	}

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Could not connect: %s\n", token.Error())
		os.Exit(1)
	} else {
		fmt.Printf("Connected to %s\n", connOpts.Servers[0])
	}

	defer client.Disconnect(250)
	go viewstats.StartStatsDisplay(mqInbound)

	for {
		incoming := <-mqInbound
		viewstats.AddStat(incoming[0], incoming[1])
		if incoming[0] == "exit now" {
			break
		}

	}

}

func init() {
	rootCmd.AddCommand(statsCmd)
}
