package viewstats

import (
	// "fmt"
	"fmt"
	"time"

	"github.com/nsf/termbox-go"
)

const coldef = termbox.ColorDefault
const DATELAYOUT = "02/Jan/2006:15:04:05 -0700"
const DATEPRINT = "Jan 02, 2006 15:04:05"

type DataHash map[string]string

func (d DataHash) Get(key string) (result string) {
	if v, ok := d[key]; ok {
		return v
	} else {
		d[key] = "n/a"
		return d[key]
	}
}

var mqttData DataHash

var (
	startTime time.Time

	w, h   int
	doExit bool
)

func StartStatsDisplay(theChan chan [2]string) {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	mqttData = make(DataHash)

	termbox.SetInputMode(termbox.InputEsc)
	redraw_all()

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
			redraw_all()
		}
	}
}

func AddStat(topic string, data string) {
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

func redraw_all() {
	termbox.Clear(coldef, coldef)
	w, h = termbox.Size()
	half := w / 2

	drawCurrentTime(1, 0)

	cur_y := 2
	cur_y = drawBroker(0, cur_y)
	cur_y++
	cur_y = drawClient(0, cur_y)

	cur_y = 2
	cur_y = drawLoad(half, cur_y)
	cur_y++
	cur_y = drawMessages(half, cur_y)

	termbox.HideCursor()

	// tbprint(w-6, h-1, coldef, termbox.ColorBlue, "ʕ◔ϖ◔ʔ")
	termbox.Flush()
}

func drawBroker(x, y int) int {
	mid := 19
	drawTitle(x, y, mid+8, "Broker")
	y++
	drawOne(x, y, mid, "Broker Version", mqttData.Get("Broker Version"))
	y++
	drawOne(x, y, mid, "Broker Time", mqttData.Get("Broker Time"))
	y++
	drawOne(x, y, mid, "Broker Uptime", mqttData.Get("Broker Uptime"))
	y++
	drawOne(x, y, mid, "Subscriptions Count", mqttData.Get("Subscriptions Count"))
	y++
	drawOne(x, y, mid, "Total Bytes Sent", mqttData.Get("Bytes Sent"))
	y++
	drawOne(x, y, mid, "Total Bytes Received", mqttData.Get("Bytes Received"))
	y++

	return y
}

func drawLoad(x, y int) int {
	mid := 14
	drawTitle(x, y, mid+8, "Load")
	y++
	drawOne(x, y, mid, "Heap Current Size", mqttData.Get("Heap Current Size"))
	y++
	drawOne(x, y, mid, "Heap Maximum Size", mqttData.Get("Heap Maximum Size"))
	y++

	return y
}

func drawMessages(x, y int) int {
	mid := 14
	drawTitle(x, y, mid+8, "Message Stats")
	y++
	drawOne(x, y, mid, "Messages Received", mqttData.Get("Messages Received"))
	y++
	drawOne(x, y, mid, "Messages Sent", mqttData.Get("Messages Sent"))
	y++
	drawOne(x, y, mid, "Messages Inflight", mqttData.Get("Messages Inflight"))
	y++
	drawOne(x, y, mid, "Messages Stored", mqttData.Get("Messages Stored"))
	y++
	y++
	drawOne(x, y, mid, "Messages Publish Dropped", mqttData.Get("Messages Publish Dropped"))
	y++
	drawOne(x, y, mid, "Messages Publish Sent", mqttData.Get("Messages Publish Sent"))
	y++
	drawOne(x, y, mid, "Messages Publish Received", mqttData.Get("Messages Publish Received"))
	y++
	drawOne(x, y, mid, "Messages Retained Count", mqttData.Get("Messages Retained Count"))
	y++

	return y
}

func drawClient(x, y int) int {
	mid := 20
	drawTitle(x, y, mid+8, "Clients")
	y++
	drawOne(x, y, mid, "Clients Total", mqttData.Get("Clients Total"))
	y++
	drawOne(x, y, mid, "Clients Connected", mqttData.Get("Clients Connected"))
	y++
	drawOne(x, y, mid, "Clients Disconnected", mqttData.Get("Clients Disconnected"))
	y++
	drawOne(x, y, mid, "Clients Expired", mqttData.Get("Clients Expired"))
	y++
	drawOne(x, y, mid, "Clients Maximum", mqttData.Get("Clients Maximum"))
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
	timeStr := fmt.Sprintf("Now:  %-24s  Watching:  %3d:%02d:%02d", now.Format(DATEPRINT), h, m, s)
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
