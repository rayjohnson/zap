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
	//"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/spf13/cobra"
)

var CleanSession bool

// subscribeCmd represents the subscribe command
var subscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "Listen to an MQTT server on a topic",
	Long:  `Subscribe to a topic on the MQTT server`,
	RunE:  subscribe,
}

func subscribe(cmd *cobra.Command, args []string) error {
	ParseBrokerInfo(cmd, args)

	// TODO: maybe put this behind a --verbose flag
	fmt.Println("Starting subscription with following parameters")
	fmt.Println("Server: ", Server)
	fmt.Println("ClientId: ", ClientId)
	fmt.Println("Username: ", Username)
	fmt.Println("Password: ", Password)
	fmt.Println("QOS: ", Qos)

	connOpts := &MQTT.ClientOptions{
		ClientID:             ClientId,
		CleanSession:         CleanSession,
		Username:             Username,
		Password:             Password,
		MaxReconnectInterval: 1 * time.Second,
		KeepAlive:            time.Duration(KeepAlive),
		TLSConfig:            tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert},
	}
	connOpts.AddBroker(Server)
	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(Topic, byte(Qos), onMessageReceived); token.Wait() && token.Error() != nil {
			fmt.Printf("Could not subscribe: %s\n", token.Error())
			os.Exit(1)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("signal received, exiting")
		os.Exit(0)
	}()

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Could not connect: %s\n", token.Error())
		os.Exit(1)
	} else {
		fmt.Printf("Connected to %s\n", Server)
	}

	for {
		time.Sleep(1 * time.Second)
	}

}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
}

func init() {
	rootCmd.AddCommand(subscribeCmd)

	// TODO: add -C, --count option - after count of messages disconnect and exit
	// TODO: -N option - do not print an extra newline at end of message
	// TODO: -R option - not even sure about this one
	// TODO: -T, --filter-out. - use regexp for this maybe?  (this is for the topic but what about the message?)
	publishCmd.Flags().BoolVar(&CleanSession, "disable-clean-session", false, "send queued up messages if mqtt has persistence - be sure to set client id")
}
