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

var mqInbound = make(chan [2]string, 16)

func (d dataHash) get(key string) (result string) {
	if v, ok := d[key]; ok {
		return v
	}
	// d[key] = "n/a"
	return "n/a"
}

var mqttData dataHash

var (
	startTime time.Time

	w, h int
)

// ExitStatsViewer will be set to false by calling package - if set true viewer should end
var ExitStatsViewer bool

// PrepViewer is called to make sure our hash table exists before AddStat
func PrepViewer() {
	mqttData = make(dataHash)
	startTime = time.Now()
}

// StartStatsDisplay sets up the terminal UI to display
// data.  It will run in an infinite loop until Ctrl-C is hit
func StartStatsDisplay() {
	ExitStatsViewer = false

	err := termbox.Init()
	if err != nil {
		panic(err)
	}

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
	timer := time.Tick(time.Millisecond * 100)
	for {
		if ExitStatsViewer {
			termbox.Close()
			break
		}

		select {
		case <-timer:
			redrawAll()
		case inMsg := <-mqInbound:
			mqttData[inMsg[0]] = inMsg[1]
		}
	}
}

// AddStat parses the MQTT topics and puts the latest
// values in a hash that is displayed in the UI
func AddStat(topic string, data string) {
	var key string
	var msg [2]string

	switch topic {
	case "$SYS/broker/load/bytes/received":
		key = "Load Bytes Received"
	case "$SYS/broker/load/bytes/sent":
		key = "Load Bytes Sent"
	case "$SYS/broker/subscriptions/count":
		key = "Subscriptions Count"
	case "$SYS/broker/time":
		key = "Broker Time"

	case "$SYS/broker/uptime":
		key = "Broker Uptime"
	case "$SYS/broker/version":
		key = "Broker Version"

	case "$SYS/broker/clients/total":
		key = "Clients Total"
	case "$SYS/broker/clients/connected":
		key = "Clients Connected"
	case "$SYS/broker/clients/disconnected":
		key = "Clients Disconnected"
	case "$SYS/broker/clients/maximum":
		key = "Clients Maximum"
	case "$SYS/broker/clients/expired":
		key = "Clients Expired"

	case "$SYS/broker/heap/current size":
		key = "Heap Current Size"
	case "$SYS/broker/heap/maximum size":
		key = "Heap Maximum Size"

	case "$SYS/broker/messages/received":
		key = "Messages Received"
	case "$SYS/broker/messages/sent":
		key = "Messages Sent"
	case "$SYS/broker/messages/inflight":
		key = "Messages Inflight"
	case "$SYS/broker/messages/stored":
		key = "Messages Stored"

	case "$SYS/broker/publish/messages/dropped":
		key = "Messages Publish Dropped"
	case "$SYS/broker/messages/publish/sent":
		key = "Messages Publish Sent"
	case "$SYS/broker/messages/publish/received":
		key = "Messages Publish Received"
	case "$SYS/broker/messages/retained/count", "$SYS/broker/retained messages/count":
		key = "Messages Retained Count"

	case "$SYS/broker/load/messages/received/1min":
		key = "LoadMessagesReceived1min"
	case "$SYS/broker/load/messages/received/5min":
		key = "LoadMessagesReceived5min"
	case "$SYS/broker/load/messages/received/15min":
		key = "LoadMessagesReceived15min"

	case "$SYS/broker/load/messages/sent/1min":
		key = "LoadMessagesSent1min"
	case "$SYS/broker/load/messages/sent/5min":
		key = "LoadMessagesSent5min"
	case "$SYS/broker/load/messages/sent/15min":
		key = "LoadMessagesSent15min"

	case "$SYS/broker/load/bytes/sent/1min":
		key = "LoadBytesSent1min"
	case "$SYS/broker/load/bytes/sent/5min":
		key = "LoadBytesSent5min"
	case "$SYS/broker/load/bytes/sent/15min":
		key = "LoadBytesSent15min"

	case "$SYS/broker/load/bytes/received/1min":
		key = "LoadBytesReceived1min"
	case "$SYS/broker/load/bytes/received/5min":
		key = "LoadBytesReceived5min"
	case "$SYS/broker/load/bytes/received/15min":
		key = "LoadBytesReceived15min"

	case "$SYS/broker/load/sockets/1min":
		key = "LoadSockets1min"
	case "$SYS/broker/load/sockets/5min":
		key = "LoadSockets5min"
	case "$SYS/broker/load/sockets/15min":
		key = "LoadSockets15min"

	case "$SYS/broker/load/connections/1min":
		key = "LoadConnections1min"
	case "$SYS/broker/load/connections/5min":
		key = "LoadConnections5min"
	case "$SYS/broker/load/connections/15min":
		key = "LoadConnections15min"

	case "$SYS/broker/load/publish/received/1min":
		key = "LoadPublishReceived1min"
	case "$SYS/broker/load/publish/received/5min":
		key = "LoadPublishReceived5min"
	case "$SYS/broker/load/publish/received/15min":
		key = "LoadPublishReceived15min"

	case "$SYS/broker/load/publish/sent/1min":
		key = "LoadPublishSent1min"
	case "$SYS/broker/load/publish/sent/5min":
		key = "LoadPublishSent5min"
	case "$SYS/broker/load/publish/sent/15min":
		key = "LoadPublishSent15min"

	case "$SYS/broker/load/publish/dropped/1min":
		key = "LoadPublishDropped1min"
	case "$SYS/broker/load/publish/dropped/5min":
		key = "LoadPublishDropped5min"
	case "$SYS/broker/load/publish/dropped/15min":
		key = "LoadPublishDropped15min"
	}

	if key != "" {
		msg[0] = key
		msg[1] = data
		mqInbound <- msg
	}
}

func handleEvents(eventChan chan termbox.Event) {
	for {
		ev := <-eventChan
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {

			case termbox.KeyEsc, termbox.KeyCtrlQ, termbox.KeyCtrlC:
				ExitStatsViewer = true

			default:
				if ev.Ch == 'q' || ev.Ch == 'Q' {
					ExitStatsViewer = true
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}
