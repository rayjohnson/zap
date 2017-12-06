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

package viewstats

import (
	"time"

	"github.com/nsf/termbox-go"
)

const coldef = termbox.ColorDefault
const datePrint = "Jan 02, 2006 15:04:05"

type dataHash map[string]string

func (d dataHash) get(key string) (result string) {
	if v, ok := d[key]; ok {
		return v
	}
	d[key] = "n/a"
	return d[key]
}

var mqttData dataHash

var (
	startTime time.Time

	w, h   int
	doExit bool
)

// StartStatsDisplay sets up the terminal UI to display
// data.  It will run in an infinite loop until Ctrl-C is hit
func StartStatsDisplay(theChan chan [2]string) {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	mqttData = make(dataHash)

	termbox.SetInputMode(termbox.InputEsc)
	redrawAll()

	//capture and process events from the CLI
	eventChan := make(chan termbox.Event, 16)
	go handleEvents(eventChan)
	go func() {
		for {
			ev := termbox.PollEvent()
			eventChan <- ev
		}
	}()

	// start update (redraw) ticker
	doExit = false
	timer := time.Tick(time.Millisecond * 100)
	for {
		if doExit {
			termbox.Close()

			exitStr := <-theChan
			exitStr[0] = "exit now"
			theChan <- exitStr
			break
		}

		select {
		case <-timer:
			redrawAll()
		}
	}
}

// AddStat parses the MQTT topics and puts the latest
// values in a hash that is displayed in the UI
func AddStat(topic string, data string) {
	switch topic {
	case "$SYS/broker/load/bytes/received":
		mqttData["Load Bytes Received"] = data
	case "$SYS/broker/load/bytes/sent":
		mqttData["Load Bytes Sent"] = data
	case "$SYS/broker/subscriptions/count":
		mqttData["Subscriptions Count"] = data
	case "$SYS/broker/time":
		mqttData["Broker Time"] = data

	case "$SYS/broker/uptime":
		mqttData["Broker Uptime"] = data
	case "$SYS/broker/version":
		mqttData["Broker Version"] = data

	case "$SYS/broker/clients/total":
		mqttData["Clients Total"] = data
	case "$SYS/broker/clients/connected":
		mqttData["Clients Connected"] = data
	case "$SYS/broker/clients/disconnected":
		mqttData["Clients Disconnected"] = data
	case "$SYS/broker/clients/maximum":
		mqttData["Clients Maximum"] = data
	case "$SYS/broker/clients/expired":
		mqttData["Clients Expired"] = data

	case "$SYS/broker/heap/current size":
		mqttData["Heap Current Size"] = data
	case "$SYS/broker/heap/maximum size":
		mqttData["Heap Maximum Size"] = data
	// TODO get these: $SYS/broker/load/connections/+
	// they come in sets of 4 or something

	case "$SYS/broker/messages/received":
		mqttData["Messages Received"] = data
	case "$SYS/broker/messages/sent":
		mqttData["Messages Sent"] = data
	case "$SYS/broker/messages/inflight":
		mqttData["Messages Inflight"] = data
	case "$SYS/broker/messages/stored":
		mqttData["Messages Stored"] = data

	case "$SYS/broker/publish/messages/dropped":
		mqttData["Messages Publish Dropped"] = data
	case "$SYS/broker/messages/publish/sent":
		mqttData["Messages Publish Sent"] = data
	case "$SYS/broker/messages/publish/received":
		mqttData["Messages Publish Received"] = data
	case "$SYS/broker/messages/retained/count", "$SYS/broker/retained messages/count":
		mqttData["Messages Retained Count"] = data

	case "$SYS/broker/load/messages/received/1min":
		mqttData["LoadMessagesReceived1min"] = data
	case "$SYS/broker/load/messages/received/5min":
		mqttData["LoadMessagesReceived5min"] = data
	case "$SYS/broker/load/messages/received/15min":
		mqttData["LoadMessagesReceived15min"] = data

	case "$SYS/broker/load/messages/sent/1min":
		mqttData["LoadMessagesSent1min"] = data
	case "$SYS/broker/load/messages/sent/5min":
		mqttData["LoadMessagesSent5min"] = data
	case "$SYS/broker/load/messages/sent/15min":
		mqttData["LoadMessagesSent15min"] = data

	case "$SYS/broker/load/bytes/sent/1min":
		mqttData["LoadBytesSent1min"] = data
	case "$SYS/broker/load/bytes/sent/5min":
		mqttData["LoadBytesSent5min"] = data
	case "$SYS/broker/load/bytes/sent/15min":
		mqttData["LoadBytesSent15min"] = data

	case "$SYS/broker/load/bytes/received/1min":
		mqttData["LoadBytesReceived1min"] = data
	case "$SYS/broker/load/bytes/received/5min":
		mqttData["LoadBytesReceived5min"] = data
	case "$SYS/broker/load/bytes/received/15min":
		mqttData["LoadBytesReceived15min"] = data

	case "$SYS/broker/load/sockets/1min":
		mqttData["LoadSockets1min"] = data
	case "$SYS/broker/load/sockets/5min":
		mqttData["LoadSockets5min"] = data
	case "$SYS/broker/load/sockets/15min":
		mqttData["LoadSockets15min"] = data

	case "$SYS/broker/load/connections/1min":
		mqttData["LoadConnections1min"] = data
	case "$SYS/broker/load/connections/5min":
		mqttData["LoadConnections5min"] = data
	case "$SYS/broker/load/connections/15min":
		mqttData["LoadConnections15min"] = data

	case "$SYS/broker/load/publish/received/1min":
		mqttData["LoadPublishReceived1min"] = data
	case "$SYS/broker/load/publish/received/5min":
		mqttData["LoadPublishReceived5min"] = data
	case "$SYS/broker/load/publish/received/15min":
		mqttData["LoadPublishReceived15min"] = data

	case "$SYS/broker/load/publish/sent/1min":
		mqttData["LoadPublishSent1min"] = data
	case "$SYS/broker/load/publish/sent/5min":
		mqttData["LoadPublishSent5min"] = data
	case "$SYS/broker/load/publish/sent/15min":
		mqttData["LoadPublishSent15min"] = data

	case "$SYS/broker/load/publish/dropped/1min":
		mqttData["LoadPublishDropped1min"] = data
	case "$SYS/broker/load/publish/dropped/5min":
		mqttData["LoadPublishDropped5min"] = data
	case "$SYS/broker/load/publish/dropped/15min":
		mqttData["LoadPublishDropped15min"] = data

	}
}

func init() {
	startTime = time.Now()
}

func handleEvents(eventChan chan termbox.Event) {
	for {
		ev := <-eventChan
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {

			case termbox.KeyEsc, termbox.KeyCtrlQ, termbox.KeyCtrlC:
				doExit = true
			case 'q', 'Q':
				doExit = true

			default:
				if ev.Ch != 0 {
					// edit_box.InsertRune(ev.Ch)
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}
