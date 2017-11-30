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
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/spf13/viper"
	"github.com/spf13/cobra"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish into MQTT",
	Long: `The publish command allows you to send data on an MQTT topic

Multiple options are available to send a single argument, a whole file, or 
data coming from stdin.`,
	RunE: publish,
}

var Message string
var FilePath string
var Retain bool
var NullMessage bool

func publish(cmd *cobra.Command, args []string) error {

	// TODO support -I, --id-prefix
	if ClientId == "" {
		hostname, _ := os.Hostname()
		ClientId = hostname + "_" + strconv.Itoa(time.Now().Second())
	}

	ParseBrokerInfo(cmd, args)
	
	// TODO: maybe put this behind a --verbose flag
	altServer := viper.GetString("server") // retrieve values from viper instead of pflag
	fmt.Println("Alt Server: ", altServer)

	fmt.Println("Starting to publish with following parameters")
	fmt.Println("Server: ", Server)
	fmt.Println("ClientId: ", ClientId)
	fmt.Println("Username: ", Username)
	fmt.Println("Password: ", Password)
	fmt.Println("QOS: ", Qos)
	fmt.Println("Retain: ", Retain)
	os.Exit(1)

	connOpts := &MQTT.ClientOptions{
		ClientID:             ClientId,
		CleanSession:         true,
		Username:             Username,
		Password:             Password,
		MaxReconnectInterval: 1 * time.Second,
		KeepAlive:            KeepAlive,
		TLSConfig:            tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert},
	}
	connOpts.AddBroker(Server)

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("Could not connect: %s\n", token.Error())
		os.Exit(1)
	} else {
		// TODO: put behind verbose
		fmt.Printf("Connected to %s\n", Server)
	}

	// TODO: only one message type option can be set - need to figure out
	// how check if args were set

	// TODO: support -s, --stdin-file
	if Message != "" {
		// send a single message
		client.Publish(Topic, byte(Qos), Retain, Message)
	} else if NullMessage {
		// send a null message (actually an empty string)
		client.Publish(Topic, byte(Qos), Retain, "")
	} else if FilePath != "" {
		// send entire file as message
		if _, err := os.Stat(FilePath); !os.IsNotExist(err) {
    		buf, err := ioutil.ReadFile(FilePath) // just pass the file name
		    if err != nil {
		        fmt.Print("error reading file: \"%s\"\n", err)
		        os.Exit(1)
		    }

			client.Publish(Topic, byte(Qos), Retain, string(buf))
		} else {
			fmt.Printf("the file \"%s\" does not exist\n", FilePath)
			os.Exit(1)
		}

	} else {
		// stdin case
		// TODO: this should be -l, --stdin-line option
		stdin := bufio.NewReader(os.Stdin)

		for {
			message, err := stdin.ReadString('\n')
			if err == io.EOF {
				os.Exit(0)
			}
			client.Publish(Topic, byte(Qos), Retain, message)
		}
	}

	fmt.Printf("message sent\n")
	return nil
}

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringVarP(&Message, "message", "m", "", "send the argument to the topic and exit")
	publishCmd.Flags().StringVarP(&FilePath, "file", "f", "", "send contents of the file to the topic and exit")
	publishCmd.Flags().BoolVarP(&Retain, "retain", "r", false, "retain as the last good message")
	publishCmd.Flags().BoolVarP(&NullMessage, "null-message", "n", false, "send a null (zero length) message")
}
