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
	"os/signal"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/spf13/cobra"
)

var cleanSession bool
var mqInbound = make(chan [2]string)
var done = false

// subscribeCmd represents the subscribe command
var subscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "Listen to an MQTT server on a topic",
	Long:  `Subscribe to a topic on the MQTT server`,
	Run:   subscribe,
}

func subscribe(cmd *cobra.Command, args []string) {

	connOpts := ParseBrokerInfo(cmd, args)

	PrintConnectionInfo()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("signal received, exiting")
		done = true
	}()

	var conErr error
	defer func() {
		if conErr != nil {
			os.Exit(1)
		}
	}()

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		conErr = token.Error()
		fmt.Printf("Could not connect: %s\n", conErr)
		return
	}
	defer client.Disconnect(250)

	if optVerbose {
		fmt.Printf("Connected to %s\n", connOpts.Servers[0])
	}

	if token := client.Subscribe(optTopic, byte(optQos), subscriptionHandler); token.Wait() && token.Error() != nil {
		conErr = token.Error()
		fmt.Printf("Could not subscribe: %s\n", conErr)
		return
	}
	defer client.Unsubscribe(optTopic)

	for {
		time.Sleep(time.Millisecond * 10)
		if done {
			break
		}
	}

}

func subscriptionHandler(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("Received message on topic: %s\nMessage: %s\n", msg.Topic(), msg.Payload())
}

func init() {
	rootCmd.AddCommand(subscribeCmd)

	// TODO: add -C, --count option - after count of messages disconnect and exit
	// TODO: -N option - do not print an extra newline at end of message
	// TODO: -R option - not even sure about this one
	// TODO: -T, --filter-out. - use regexp for this maybe?  (this is for the topic but what about the message?)
	publishCmd.Flags().BoolVar(&cleanSession, "clean-session", true, "set to false and will send queued up messages if mqtt has persistence - be sure to set client id")
}
