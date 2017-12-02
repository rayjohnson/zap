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
	"fmt"
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
	// TODO: make this a switch statement
	if topic == "$SYS/broker/load/bytes/received" {
		mqttData["Load Bytes Received"] = data
	}
	if topic == "$SYS/broker/load/bytes/sent" {
		mqttData["Load Bytes Sent"] = data
	}
	if topic == "$SYS/broker/subscriptions/count" {
		mqttData["Subscriptions Count"] = data
	}
	if topic == "$SYS/broker/time" {
		mqttData["Broker Time"] = data
	}
	if topic == "$SYS/broker/uptime" {
		mqttData["Broker Uptime"] = data
	}
	if topic == "$SYS/broker/version" {
		mqttData["Broker Version"] = data
	}

	if topic == "$SYS/broker/clients/total" {
		mqttData["Clients Total"] = data
	}
	if topic == "$SYS/broker/clients/connected" {
		mqttData["Clients Connected"] = data
	}
	if topic == "$SYS/broker/clients/disconnected" {
		mqttData["Clients Disconnected"] = data
	}
	if topic == "$SYS/broker/clients/maximum" {
		mqttData["Clients Maximum"] = data
	}
	if topic == "$SYS/broker/clients/expired" {
		mqttData["Clients Expired"] = data
	}

	if topic == "$SYS/broker/heap/current size" {
		mqttData["Heap Current Size"] = data
	}
	if topic == "$SYS/broker/heap/maximum size" {
		mqttData["Heap Maximum Size"] = data
	}
	// TODO get these: $SYS/broker/load/connections/+
	// they come in sets of 4 or something

	if topic == "$SYS/broker/messages/received" {
		mqttData["Messages Received"] = data
	}
	if topic == "$SYS/broker/messages/sent" {
		mqttData["Messages Sent"] = data
	}
	if topic == "$SYS/broker/messages/inflight" {
		mqttData["Messages Inflight"] = data
	}
	if topic == "$SYS/broker/messages/stored" {
		mqttData["Messages Stored"] = data
	}

	if topic == "$SYS/broker/publish/messages/dropped" {
		mqttData["Messages Publish Dropped"] = data
	}
	if topic == "$SYS/broker/messages/publish/sent" {
		mqttData["Messages Publish Sent"] = data
	}
	if topic == "$SYS/broker/messages/publish/received" {
		mqttData["Messages Publish Received"] = data
	}
	if topic == "$SYS/broker/messages/retained/count" {
		mqttData["Messages Retained Count"] = data
	}
	if topic == "$SYS/broker/retained messages/count" {
		mqttData["Messages Retained Count"] = data
	}
}

func init() {
	startTime = time.Now()
}

func redrawAll() {
	termbox.Clear(coldef, coldef)
	w, h = termbox.Size()
	half := w / 2

	drawCurrentTime(1, 0)

	curY := 2
	curY = drawBroker(0, curY)
	curY++
	curY = drawClient(0, curY)

	curY = 2
	curY = drawLoad(half, curY)
	curY++
	curY = drawMessages(half, curY)

	termbox.HideCursor()

	termbox.Flush()
}

func drawBroker(x, y int) int {
	mid := 19
	drawTitle(x, y, mid+8, "Broker")
	y++
	drawOne(x, y, mid, "Broker Version", mqttData.get("Broker Version"))
	y++
	drawOne(x, y, mid, "Broker Time", mqttData.get("Broker Time"))
	y++
	drawOne(x, y, mid, "Broker Uptime", mqttData.get("Broker Uptime"))
	y++
	drawOne(x, y, mid, "Subscriptions Count", mqttData.get("Subscriptions Count"))
	y++
	drawOne(x, y, mid, "Total Bytes Sent", mqttData.get("Bytes Sent"))
	y++
	drawOne(x, y, mid, "Total Bytes Received", mqttData.get("Bytes Received"))
	y++

	return y
}

func drawLoad(x, y int) int {
	mid := 14
	drawTitle(x, y, mid+8, "Load")
	y++
	drawOne(x, y, mid, "Heap Current Size", mqttData.get("Heap Current Size"))
	y++
	drawOne(x, y, mid, "Heap Maximum Size", mqttData.get("Heap Maximum Size"))
	y++

	return y
}

func drawMessages(x, y int) int {
	mid := 14
	drawTitle(x, y, mid+8, "Message Stats")
	y++
	drawOne(x, y, mid, "Messages Received", mqttData.get("Messages Received"))
	y++
	drawOne(x, y, mid, "Messages Sent", mqttData.get("Messages Sent"))
	y++
	drawOne(x, y, mid, "Messages In-flight", mqttData.get("Messages Inflight"))
	y++
	drawOne(x, y, mid, "Messages Stored", mqttData.get("Messages Stored"))
	y++
	y++
	drawOne(x, y, mid, "Messages Publish Dropped", mqttData.get("Messages Publish Dropped"))
	y++
	drawOne(x, y, mid, "Messages Publish Sent", mqttData.get("Messages Publish Sent"))
	y++
	drawOne(x, y, mid, "Messages Publish Received", mqttData.get("Messages Publish Received"))
	y++
	drawOne(x, y, mid, "Messages Retained Count", mqttData.get("Messages Retained Count"))
	y++

	return y
}

func drawClient(x, y int) int {
	mid := 20
	drawTitle(x, y, mid+8, "Clients")
	y++
	drawOne(x, y, mid, "Clients Total", mqttData.get("Clients Total"))
	y++
	drawOne(x, y, mid, "Clients Connected", mqttData.get("Clients Connected"))
	y++
	drawOne(x, y, mid, "Clients Disconnected", mqttData.get("Clients Disconnected"))
	y++
	drawOne(x, y, mid, "Clients Expired", mqttData.get("Clients Expired"))
	y++
	drawOne(x, y, mid, "Clients Maximum", mqttData.get("Clients Maximum"))
	y++

	return y
}

func drawTitle(x int, y int, max int, title string) {
	str := fmt.Sprintf("%-*s", max, title)
	for i, c := range str {
		termbox.SetCell(x+i, y, c, coldef+termbox.AttrUnderline, coldef)
	}
}

func drawOne(x int, y int, mid int, label string, data string) {
	s := fmt.Sprintf("%*s : %s", mid, label, data)
	for i, c := range s {
		termbox.SetCell(x+i, y, c, coldef, coldef)
	}
}

func drawCurrentTime(x, y int) {
	now := time.Now()
	since := now.Sub(startTime)
	h := int(since.Hours())
	m := int(since.Minutes()) % 60
	s := int(since.Seconds()) % 60
	timeStr := fmt.Sprintf("Now:  %-24s  Watching:  %3d:%02d:%02d", now.Format(datePrint), h, m, s)
	for i, c := range timeStr {
		termbox.SetCell(x+i, y, c, coldef, coldef)
	}
}

func handleEvents(eventChan chan termbox.Event) {
	for {
		ev := <-eventChan
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {

			case termbox.KeyEsc:
				goto endfunc
			case termbox.KeyCtrlQ:
				goto endfunc
			case termbox.KeyCtrlC:
				goto endfunc

			default:
				if ev.Ch != 0 {
					// edit_box.InsertRune(ev.Ch)
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
endfunc:
	doExit = true
}
