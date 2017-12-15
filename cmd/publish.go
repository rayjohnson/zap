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

var optMessage string
var optFilePath string
var optRetain bool

func newPublishCommand() *cobra.Command {
	var conOpts *connectionOptions

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish into MQTT",
		Long: `The publish command allows you to send data on an MQTT topic

Multiple options are available to send a single argument, a whole file, or
data coming from stdin.`,
		Run: func(cmd *cobra.Command, args []string) {
			runPublish(cmd.Flags(), conOpts)
		},
	}

	flags := cmd.Flags()
	flags.BoolP("stdin-line", "l", false, "send stdin data as message with each newline is a new message")
	flags.BoolP("stdin-file", "s", false, "read stdin until EOF and send all as one message")
	flags.StringVarP(&optMessage, "message", "m", "", "send the argument to the topic and exit")
	flags.StringVarP(&optFilePath, "file", "f", "", "send contents of the file to the topic and exit")
	flags.BoolVarP(&optRetain, "retain", "r", false, "retain as the last good message")
	flags.BoolP("null-message", "n", false, "send a null (zero length) message")
	flags.StringVar(&optTopic, "topic", "sample", "mqtt topic to post to")

	conOpts = addConnectionFlags(flags)

	return cmd
}

func validatePublishOptions(flags *pflag.FlagSet) {
	var count = 0

	if flags.Lookup("message").Changed {
		count++
	}
	if flags.Lookup("null-message").Changed {
		count++
	}
	if flags.Lookup("file").Changed {
		count++
	}
	if flags.Lookup("stdin-line").Changed {
		count++
	}
	if flags.Lookup("stdin-file").Changed {
		count++
	}

	if count == 0 {
		fmt.Println("must specify one of --message, --file, --stdin-line, --stdin-file, or --null-message to send any data")
		os.Exit(1)
	}

	if count > 1 {
		fmt.Println("only one of --message, --file, --stdin-line, --stdin-file, or --null-message can be used")
		os.Exit(1)
	}

}

func runPublish(flags *pflag.FlagSet, conOpts *connectionOptions) {
	clientOpts := ParseBrokerInfo(flags, conOpts)
	clientOpts.CleanSession = true

	validatePublishOptions(flags)

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

	if flags.Lookup("message").Changed {
		// send a single message
		client.Publish(optTopic, byte(optQos), optRetain, optMessage)
	}

	if flags.Lookup("null-message").Changed {
		// send a null message (actually an empty string)
		client.Publish(optTopic, byte(optQos), optRetain, "")
	}

	if flags.Lookup("file").Changed {
		// send entire file as message
		if _, err := os.Stat(optFilePath); !os.IsNotExist(err) {
			buf, err := ioutil.ReadFile(optFilePath) // just pass the file name
			if err != nil {
				fmt.Print("error reading file: \"%s\"\n", err)
				exitWithError = true
				return
			}

			client.Publish(optTopic, byte(optQos), optRetain, string(buf))
		} else {
			fmt.Printf("the file \"%s\" does not exist\n", optFilePath)
			exitWithError = true
			return
		}
	}

	if flags.Lookup("stdin-line").Changed {
		// read from stdin read by line - send one messages per line
		stdin := bufio.NewReader(os.Stdin)

		for {
			message, err := stdin.ReadString('\n')
			if err == io.EOF {
				fmt.Printf("message sent or EOF\n")
				break
			}
			client.Publish(optTopic, byte(optQos), optRetain, message)
		}
	}

	if flags.Lookup("stdin-file").Changed {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("error reading file: %s\n", err)
			exitWithError = true
			return
		}
		client.Publish(optTopic, byte(optQos), optRetain, data)
	}

	fmt.Printf("message sent\n")
}
