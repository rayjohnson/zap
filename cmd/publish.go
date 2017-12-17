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
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type publishOptions struct {
	doStdinLine bool
	doStdinFile bool
	doNullMsg   bool
	message     string
	filePath    string
	retain      bool
	topic       string
	qos         int
}

func newPublishCommand() *cobra.Command {
	var conOpts *connectionOptions
	pubOpts := publishOptions{}

	cmd := &cobra.Command{
		Use:   "publish",
		Args:  cobra.NoArgs,
		Short: "Publish into MQTT",
		Long: `The publish command allows you to send data on an MQTT topic

Multiple options are available to send a single argument, a whole file, or
data coming from stdin.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPublish(cmd.Flags(), conOpts, pubOpts)
		},
	}
	cmd.SilenceUsage = true

	flags := cmd.Flags()
	flags.BoolVarP(&pubOpts.doStdinLine, "stdin-line", "l", false, "send stdin data as message with each newline is a new message")
	flags.BoolVarP(&pubOpts.doStdinFile, "stdin-file", "s", false, "read stdin until EOF and send all as one message")
	flags.StringVarP(&pubOpts.message, "message", "m", "", "send the argument to the topic and exit")
	flags.StringVarP(&pubOpts.filePath, "file", "f", "", "send contents of the file to the topic and exit")
	flags.BoolVarP(&pubOpts.retain, "retain", "r", false, "retain as the last good message")
	flags.BoolVarP(&pubOpts.doNullMsg, "null-message", "n", false, "send a null (zero length) message")
	flags.StringVar(&pubOpts.topic, "topic", "sample", "mqtt topic to post to")
	flags.IntVar(&pubOpts.qos, "qos", 0, "qos setting for outbound messages")

	conOpts = addConnectionFlags(flags)

	return cmd
}

func validatePublishOptions(pubOpts publishOptions) error {
	var count = 0

	if pubOpts.message != "" {
		count++
	}
	if pubOpts.doNullMsg {
		count++
	}
	if pubOpts.filePath != "" {
		if _, err := os.Stat(pubOpts.filePath); os.IsNotExist(err) {
			return err
		}

		count++
	}
	if pubOpts.doStdinLine {
		count++
	}
	if pubOpts.doStdinFile {
		count++
	}

	if count == 0 {
		return fmt.Errorf("must specify one of --message, --file, --stdin-line, --stdin-file, or --null-message to send any data")
	}

	if count > 1 {
		return fmt.Errorf("only one of --message, --file, --stdin-line, --stdin-file, or --null-message can be used")
	}

	if pubOpts.qos < 0 || pubOpts.qos > 2 {
		return fmt.Errorf("--qos value must or 0, 1 or 2")
	}

	return nil
}

func runPublish(flags *pflag.FlagSet, conOpts *connectionOptions, pubOpts publishOptions) error {
	clientOpts, err := ParseBrokerInfo(flags, conOpts)
	if err != nil {
		return err
	}
	clientOpts.CleanSession = true

	err = validatePublishOptions(pubOpts)
	if err != nil {
		return err
	}

	PrintConnectionInfo(conOpts, nil, &pubOpts)

	client := MQTT.NewClient(clientOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("could not connect: %s", token.Error())
	}
	defer client.Disconnect(250)

	if optVerbose {
		fmt.Printf("Connected to %s\n", clientOpts.Servers[0])
	}

	if pubOpts.message != "" {
		// send a single message
		client.Publish(pubOpts.topic, byte(pubOpts.qos), pubOpts.retain, pubOpts.message)
	}

	if pubOpts.doNullMsg {
		// send a null message (actually an empty string)
		client.Publish(pubOpts.topic, byte(pubOpts.qos), pubOpts.retain, "")
	}

	if pubOpts.filePath != "" {
		// send entire file as message
		buf, err := ioutil.ReadFile(pubOpts.filePath)
		if err != nil {
			return err
		}

		client.Publish(pubOpts.topic, byte(pubOpts.qos), pubOpts.retain, string(buf))
	}

	if pubOpts.doStdinLine {
		// read from stdin read by line - send one messages per line
		stdin := bufio.NewReader(os.Stdin)

		for {
			message, err := stdin.ReadString('\n')
			if err == io.EOF {
				break
			}
			client.Publish(pubOpts.topic, byte(pubOpts.qos), pubOpts.retain, message)
		}
	}

	if pubOpts.doStdinFile {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		client.Publish(pubOpts.topic, byte(pubOpts.qos), pubOpts.retain, data)
	}

	if optVerbose {
		fmt.Printf("message sent\n")
	}

	return nil
}
