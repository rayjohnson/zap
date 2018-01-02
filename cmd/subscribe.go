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
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const builtinTemplate = "Received message on topic: {{.Topic}}\nMessage: {{.Message}}\n"

//var stdoutTemplate *template.Template

// MqttMessage is the struct passed to the template engine
type MqttMessage struct {
	Topic   string
	Message string
	MsgJSON map[string]interface{}
}

type subscribeOptions struct {
	cleanSession   bool
	templateString string
	topic          string
	count          int
	skipRetained   bool
	qos            int
	stdoutTemplate *template.Template
}

type messageOptions struct {
	stdoutTemplate *template.Template
	quit           chan bool
	count          int
	numMsgs        int
	skipRetained   bool
}

func newSubscribeCommand() *cobra.Command {
	var zapOpts *zapOptions
	subOpts := &subscribeOptions{}

	cmd := &cobra.Command{
		Use:   "subscribe",
		Args:  cobra.NoArgs,
		Short: "Subscribe to an MQTT server on a topic",
		Long: `Subscribe to a topic on the MQTT server and print the contents
to stdout.  Use the --format flag to adjust the output.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubscribe(cmd.Flags(), zapOpts)
		},
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}
	annotations := make(map[string]string)
	annotations["man-files-section"] = filesManInfo
	cmd.Annotations = annotations

	flags := cmd.Flags()
	flags.BoolVar(&subOpts.cleanSession, "clean-session", true, "Set to false and mqtt will send queued up messages if service disconnects and restarts")
	flags.StringVar(&subOpts.templateString, "template", builtinTemplate, "Template to use for output to stdout")
	flags.StringVar(&subOpts.topic, "topic", "#", "The mqtt topic or topic filter to listen to")
	flags.IntVar(&subOpts.count, "count", -1, "After count of messages disconnect and exit")
	flags.BoolVar(&subOpts.skipRetained, "skip-retained", false, "Skip printing messages marked as retained from mqtt")
	flags.IntVar(&subOpts.qos, "qos", 0, "The qos setting for inbound messages")
	// TODO: -T, --filter-out. - use regexp for this maybe?  (this is for the topic but what about the message?)

	annotation := []string{"0|1|2"}
	flags.SetAnnotation("qos", "man-arg-hints", annotation)
	annotation = []string{"topic path"}
	flags.SetAnnotation("topic", "man-arg-hints", annotation)
	annotation = []string{"go template"}
	flags.SetAnnotation("template", "man-arg-hints", annotation)
	annotation = []string{"int"}
	flags.SetAnnotation("count", "man-arg-hints", annotation)

	zapOpts = buildZapFlags(flags)
	zapOpts.conOpts = addConnectionFlags(flags)
	zapOpts.subOpts = subOpts

	return cmd
}

func (subOpts *subscribeOptions) validateOptions() error {
	var err error

	if subOpts.qos < 0 || subOpts.qos > 2 {
		return fmt.Errorf("--qos value must or 0, 1 or 2")
	}

	subOpts.stdoutTemplate, err = template.New("stdout").Funcs(basicFunctions).Parse(subOpts.templateString)
	if err != nil {
		return err
	}

	return nil
}

func runSubscribe(flags *pflag.FlagSet, zapOpts *zapOptions) error {
	subOpts := zapOpts.subOpts

	if err := zapOpts.processOptions(flags); err != nil {
		return err
	}
	clientOpts := zapOpts.clientOpts

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

	output.VERBOSE.Printf("Connected to %s\n", clientOpts.Servers[0])

	msgOpts := messageOptions{}
	msgOpts.quit = quit
	msgOpts.count = subOpts.count
	msgOpts.skipRetained = subOpts.skipRetained
	msgOpts.stdoutTemplate = subOpts.stdoutTemplate

	if token := client.Subscribe(subOpts.topic, byte(subOpts.qos), func(client MQTT.Client, msg MQTT.Message) {
		subscriptionHandler(client, msg, &msgOpts)
	}); token.Wait() && token.Error() != nil {
		return fmt.Errorf("could not subscribe: %s", token.Error())
	}
	defer client.Unsubscribe(subOpts.topic)

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

func subscriptionHandler(client MQTT.Client, msg MQTT.Message, msgOpts *messageOptions) {
	doExit := false

	// skipping retained messages does not count toward --count value
	if msgOpts.skipRetained && msg.Retained() {
		return
	}

	// This handles the --count option
	if msgOpts.count > 0 {
		// TODO: this may be a race condition
		msgOpts.numMsgs++
		if msgOpts.numMsgs > msgOpts.count {
			// Skip displaying messages after we hit count
			return
		}
		if msgOpts.numMsgs == msgOpts.count {
			doExit = true
		}
	}

	var buf bytes.Buffer
	data := MqttMessage{Topic: msg.Topic(), Message: string(msg.Payload())}
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(data.Message), &m); err != nil {
		// TODO: Need flag to know if user wants json or not
		fmt.Printf("Can not parse as json: %s", err)
	}
	data.MsgJSON = m

	err := msgOpts.stdoutTemplate.Execute(&buf, data)
	if err != nil {
		fmt.Printf("error using template: %s", err)
		return
	}
	fmt.Printf("%s", buf.String())

	if doExit {
		msgOpts.quit <- true
	}
}

var basicFunctions = template.FuncMap{
	"json": func(v interface{}) string {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		enc.Encode(v)
		// Remove the trailing new line added by the encoder
		return strings.TrimSpace(buf.String())
	},
	"split":      strings.Split,
	"join":       strings.Join,
	"title":      strings.Title,
	"lower":      strings.ToLower,
	"upper":      strings.ToUpper,
	"pad":        padWithSpace,
	"truncate":   truncateWithLength,
	"prettyjson": prettyJSON,
}

func prettyJSON(source string) string {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, []byte(source), "", "    ")
	if error != nil {
		// TODO: spit something to stderr if verbose
		// fmt.Println("JSON parse error: ", error)
		return source
	}

	return prettyJSON.String()
}

// padWithSpace adds whitespace to the input if the input is non-empty
func padWithSpace(source string, prefix, suffix int) string {
	if source == "" {
		return source
	}
	return strings.Repeat(" ", prefix) + source + strings.Repeat(" ", suffix)
}

// truncateWithLength truncates the source string up to the length provided by the input
func truncateWithLength(source string, length int) string {
	if len(source) < length {
		return source
	}
	return source[:length]
}
