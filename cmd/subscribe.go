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
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const builtinTemplate = "Received message on topic: {{.Topic}}\nMessage: {{.Message}}\n"

var cleanSession bool
var mqInbound = make(chan [2]string)
var done = false

var optTemplate string
var stdoutTemplate *template.Template

// MqttMessage is the struct passed to the template engine
type MqttMessage struct {
	Topic   string
	Message string
}

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
	stdoutTemplate = getTemplate(cmd)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("signal received, exiting")
		done = true
	}()

	exitWithError := false
	defer func() {
		if exitWithError {
			os.Exit(1)
		}
	}()

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Could not connect: %s\n", token.Error())
		exitWithError = true
		return
	}
	defer client.Disconnect(250)

	if optVerbose {
		fmt.Printf("Connected to %s\n", connOpts.Servers[0])
	}

	if token := client.Subscribe(optTopic, byte(optQos), subscriptionHandler); token.Wait() && token.Error() != nil {
		exitWithError = true
		fmt.Printf("Could not subscribe: %s\n", token.Error())
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
	data := MqttMessage{Topic: msg.Topic(), Message: string(msg.Payload())}

	var buf bytes.Buffer

	err := stdoutTemplate.Execute(&buf, data)
	if err != nil {
		fmt.Printf("error using template: %s\n", err)
	}
	fmt.Printf("%s", buf.String())
}

func getTemplate(cmd *cobra.Command) *template.Template {
	if !cmd.Flags().Lookup("template").Changed {
		if key := getCorrectConfigKey(optBroker, "template"); key != "" {
			optTemplate = viper.GetString(key)
		}
	}

	theTemplate, err := template.New("stdout").Parse(optTemplate)
	if err != nil {
		fmt.Printf("error in template: %s\n", err)
		os.Exit(1)
	}

	return theTemplate
}

func init() {
	rootCmd.AddCommand(subscribeCmd)

	// TODO: add -C, --count option - after count of messages disconnect and exit
	// TODO: -R option - not even sure about this one
	// TODO: -T, --filter-out. - use regexp for this maybe?  (this is for the topic but what about the message?)
	subscribeCmd.Flags().BoolVar(&cleanSession, "clean-session", true, "set to false and will send queued up messages if mqtt has persistence - be sure to set client id")
	subscribeCmd.Flags().StringVar(&optTemplate, "template", builtinTemplate, "template to use for output to stdout")
}
