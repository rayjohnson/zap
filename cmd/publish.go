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
	var zapOpts *zapOptions
	pubOpts := &publishOptions{}

	cmd := &cobra.Command{
		Use:   "publish",
		Args:  cobra.NoArgs,
		Short: "Publish into MQTT",
		Long: `The publish command allows you to send a message on an MQTT topic

Multiple options are available to send a single argument, a whole file, or
data coming from stdin.`,
		Example: `.nf
Publish to public mqtt server a test:
.RS
zap publish \-\-server tcp://test.mosquitto.org:1883
\-\-topic sample/hello \-m hello
.RE
Publish to the broker in config named mosquitto data from a file:
.RS
zap publish \-\-config examples/example.zap.toml \-b mosquitto
\-\-file examples/README.txt
.RE
.fi`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPublish(cmd.Flags(), zapOpts)
		},
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}
	annotations := make(map[string]string)
	annotations["man-files-section"] = filesManInfo
	cmd.Annotations = annotations

	flags := cmd.Flags()
	flags.BoolVarP(&pubOpts.doStdinLine, "stdin-line", "l", false, "Send each line of stdin as separate message until Ctrl-C")
	flags.BoolVarP(&pubOpts.doStdinFile, "stdin-file", "s", false, "Read stdin until EOF and send all as one message")
	flags.StringVarP(&pubOpts.message, "message", "m", "", "Send the argument to the topic and exit")
	flags.StringVarP(&pubOpts.filePath, "file", "f", "", "Send contents of the file to the topic and exit")
	flags.BoolVarP(&pubOpts.retain, "retain", "r", false, "Retain as the last good message")
	flags.BoolVarP(&pubOpts.doNullMsg, "null-message", "n", false, "Send a null (zero length) message")
	flags.StringVar(&pubOpts.topic, "topic", "sample", "Topic string for mqtt, should not use wild cards")
	flags.IntVar(&pubOpts.qos, "qos", 0, "The qos setting for outbound messages")

	// Flag annotations to help make docs more clear
	annotation := []string{"path"}
	flags.SetAnnotation("file", "man-arg-hints", annotation)
	annotation = []string{"data"}
	flags.SetAnnotation("message", "man-arg-hints", annotation)
	annotation = []string{"topic path"}
	flags.SetAnnotation("topic", "man-arg-hints", annotation)
	annotation = []string{"0|1|2"}
	flags.SetAnnotation("qos", "man-arg-hints", annotation)

	zapOpts = buildZapFlags(flags)
	zapOpts.conOpts = addConnectionFlags(flags)
	zapOpts.pubOpts = pubOpts

	return cmd
}

func (pubOpts *publishOptions) validateOptions() error {
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

func runPublish(flags *pflag.FlagSet, zapOpts *zapOptions) error {
	pubOpts := zapOpts.pubOpts

	if err := zapOpts.processOptions(flags); err != nil {
		return err
	}
	clientOpts := zapOpts.clientOpts

	client := MQTT.NewClient(clientOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("could not connect: %s", token.Error())
	}
	defer client.Disconnect(250)

	if zapOpts.verbose {
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

	if zapOpts.verbose {
		fmt.Printf("message sent\n")
	}

	return nil
}
