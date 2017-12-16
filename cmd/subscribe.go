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
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const builtinTemplate = "Received message on topic: {{.Topic}}\nMessage: {{.Message}}\n"

var cleanSession bool

var optTemplate string
var stdoutTemplate *template.Template

// MqttMessage is the struct passed to the template engine
type MqttMessage struct {
	Topic   string
	Message string
}

func newSubscribeCommand() *cobra.Command {
	var conOpts *connectionOptions

	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Listen to an MQTT server on a topic",
		Long:  `Subscribe to a topic on the MQTT server`,
		// TODO: put in long description for subscribe
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubscribe(cmd.Flags(), conOpts)
		},
	}
	cmd.SilenceUsage = true

	flags := cmd.Flags()
	flags.BoolVar(&cleanSession, "clean-session", true, "set to false and will send queued up messages if mqtt has persistence - be sure to set client id")
	flags.StringVar(&optTemplate, "template", builtinTemplate, "template to use for output to stdout")
	flags.StringVar(&optTopic, "topic", "#", "mqtt topic to listen to")
	// TODO: add -C, --count option - after count of messages disconnect and exit
	// TODO: -R option - not even sure about this one
	// TODO: -T, --filter-out. - use regexp for this maybe?  (this is for the topic but what about the message?)

	conOpts = addConnectionFlags(flags)

	return cmd
}

func runSubscribe(flags *pflag.FlagSet, conOpts *connectionOptions) error {
	clientOpts, err := ParseBrokerInfo(flags, conOpts)
	if err != nil {
		return err
	}

	PrintConnectionInfo(conOpts)
	stdoutTemplate, err = getTemplate(flags, conOpts)
	if err != nil {
		return err
	}

	quit := make(chan bool)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("signal received, exiting")
		quit <- true
	}()

	client := MQTT.NewClient(clientOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("could not connect: %s", token.Error())
	}
	defer client.Disconnect(250)

	if optVerbose {
		fmt.Printf("Connected to %s\n", clientOpts.Servers[0])
	}

	if token := client.Subscribe(optTopic, byte(optQos), subscriptionHandler); token.Wait() && token.Error() != nil {
		return fmt.Errorf("could not subscribe: %s", token.Error())
	}
	defer client.Unsubscribe(optTopic)

loop:
	for {
		time.Sleep(time.Millisecond * 10)
		select {
		case <-quit:
			break loop
		}
	}

	return nil
}

func subscriptionHandler(client MQTT.Client, msg MQTT.Message) {
	data := MqttMessage{Topic: msg.Topic(), Message: string(msg.Payload())}

	var buf bytes.Buffer

	err := stdoutTemplate.Execute(&buf, data)
	if err != nil {
		fmt.Printf("error using template: %s", err)
	}
	fmt.Printf("%s", buf.String())
}

func getTemplate(flags *pflag.FlagSet, conOpts *connectionOptions) (*template.Template, error) {
	if !flags.Lookup("template").Changed {
		if key := getCorrectConfigKey(conOpts.broker, "template"); key != "" {
			optTemplate = viper.GetString(key)
		}
	}

	return template.New("stdout").Parse(optTemplate)
}
