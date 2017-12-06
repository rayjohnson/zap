package viewstats

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

func redrawAll() {
	termbox.Clear(coldef, coldef)
	w, h = termbox.Size()
	half := w / 2

	drawCurrentTime(1, 0)

	curY := 2
	curY = drawBroker(0, curY)
	curY++
	curY = drawMessages(0, curY)

	curY = 2
	curY = drawLoad(half, curY)
	curY++
	curY = drawClient(half, curY)

	termbox.HideCursor()

	termbox.Flush()
}

func drawBroker(x, y int) int {
	mid := 19
	y = drawTitle(x, y, mid+8, "Broker")
	y = drawOne(x, y, mid, "Broker Version", mqttData.get("Broker Version"))
	y = drawOne(x, y, mid, "Broker Time", mqttData.get("Broker Time"))
	y = drawOne(x, y, mid, "Broker Uptime", mqttData.get("Broker Uptime"))
	y = drawOne(x, y, mid, "Subscriptions Count", mqttData.get("Subscriptions Count"))
	y = drawOne(x, y, mid, "Total Bytes Sent", mqttData.get("Bytes Sent"))
	y = drawOne(x, y, mid, "Total Bytes Received", mqttData.get("Bytes Received"))

	return y
}

func drawLoad(x, y int) int {
	mid := 14
	y = drawLoadTitle(x, y, mid+2)
	y = drawThree(x, y, mid, "Sockets", mqttData["LoadSockets1min"], mqttData["LoadSockets5min"], mqttData["LoadSockets15min"])
	y = drawThree(x, y, mid, "Connections", mqttData["LoadConnections1min"], mqttData["LoadConnections5min"], mqttData["LoadConnections15min"])
	y = drawThree(x, y, mid, "Msg Received", mqttData["LoadMessagesReceived1min"], mqttData["LoadMessagesReceived5min"], mqttData["LoadMessagesReceived15min"])
	y = drawThree(x, y, mid, "Msg Sent", mqttData["LoadMessagesSent1min"], mqttData["LoadMessagesSent5min"], mqttData["LoadMessagesSent15min"])

	y = drawThree(x, y, mid, "Bytes Received", mqttData["LoadBytesReceived1min"], mqttData["LoadBytesReceived5min"], mqttData["LoadBytesReceived15min"])
	y = drawThree(x, y, mid, "Bytes Sent", mqttData["LoadBytesSent1min"], mqttData["LoadBytesSent5min"], mqttData["LoadBytesSent15min"])

	y = drawThree(x, y, mid, "Pub Received", mqttData["LoadPublishReceived1min"], mqttData["LoadPublishReceived5min"], mqttData["LoadPublishReceived15min"])
	y = drawThree(x, y, mid, "Pub Sent", mqttData["LoadPublishSent1min"], mqttData["LoadPublishSent5min"], mqttData["LoadPublishSent15min"])
	y = drawThree(x, y, mid, "Pub Dropped", mqttData["LoadPublishDropped1min"], mqttData["LoadPublishDropped5min"], mqttData["LoadPublishDropped15min"])

	return y
}

func drawMessages(x, y int) int {
	mid := 18
	y = drawTitle(x, y, mid+8, "Message Stats")
	y = drawOne(x, y, mid, "Messages Received", mqttData.get("Messages Received"))
	y = drawOne(x, y, mid, "Messages Sent", mqttData.get("Messages Sent"))
	y = drawOne(x, y, mid, "Messages In-flight", mqttData.get("Messages Inflight"))
	y = drawOne(x, y, mid, "Messages Stored", mqttData.get("Messages Stored"))
	y++
	y = drawOne(x, y, mid, "Messages Publish Dropped", mqttData.get("Messages Publish Dropped"))
	y = drawOne(x, y, mid, "Messages Publish Sent", mqttData.get("Messages Publish Sent"))
	y = drawOne(x, y, mid, "Messages Publish Received", mqttData.get("Messages Publish Received"))
	y = drawOne(x, y, mid, "Messages Retained Count", mqttData.get("Messages Retained Count"))

	return y
}

func drawClient(x, y int) int {
	mid := 20
	y = drawTitle(x, y, mid+8, "Clients")
	y = drawOne(x, y, mid, "Clients Total", mqttData.get("Clients Total"))
	y = drawOne(x, y, mid, "Clients Connected", mqttData.get("Clients Connected"))
	y = drawOne(x, y, mid, "Clients Disconnected", mqttData.get("Clients Disconnected"))
	y = drawOne(x, y, mid, "Clients Expired", mqttData.get("Clients Expired"))
	y = drawOne(x, y, mid, "Clients Maximum", mqttData.get("Clients Maximum"))

	return y
}

func drawTitle(x int, y int, max int, title string) int {
	str := fmt.Sprintf("%-*s", max, title)
	for i, c := range str {
		termbox.SetCell(x+i, y, c, coldef+termbox.AttrUnderline, coldef)
	}

	return y + 1
}

func drawOne(x int, y int, mid int, label string, data string) int {
	s := fmt.Sprintf("%*s : %s", mid, label, data)
	for i, c := range s {
		termbox.SetCell(x+i, y, c, coldef, coldef)
	}

	return y + 1
}

func drawLoadTitle(x int, y int, max int) int {
	str := fmt.Sprintf("%-*s  %s  %s %s", max, "Load", "1 min", "5 min", "15 min")
	for i, c := range str {
		termbox.SetCell(x+i, y, c, coldef+termbox.AttrUnderline, coldef)
	}

	return y + 1
}

func parseLoad(num string) string {
	numWidth := 6
	if num == "n/a" || len(num) < 6 {
		return fmt.Sprintf("%*s", numWidth, num)
	}

	fNum, err := strconv.ParseFloat(num, 32)
	if err != nil {
		return "error"
	}
	return fmt.Sprintf("%5.f", fNum)
}

func drawThree(x int, y int, mid int, label string, oneMin string, fiveMin string, fifteenMin string) int {
	s := fmt.Sprintf("%*s : %6s %6s %6s", mid, label, parseLoad(oneMin), parseLoad(fiveMin), parseLoad(fifteenMin))
	for i, c := range s {
		termbox.SetCell(x+i, y, c, coldef, coldef)
	}

	return y + 1
}

func drawCurrentTime(x, y int) {
	now := time.Now()
	since := now.Sub(startTime)
	h := int(since.Hours())
	m := int(since.Minutes()) % 60
	s := int(since.Seconds()) % 60
	timeStr := fmt.Sprintf("Now:  %-24s  Watching:  %3d:%02d:%02d  [Q] to quit", now.Format(datePrint), h, m, s)
	for i, c := range timeStr {
		termbox.SetCell(x+i, y, c, coldef, coldef)
	}
}
