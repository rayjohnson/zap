package viewstats

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

type termWriter func(x, y int, str string)

func stdWriter(x, y int, str string) {
	for i, c := range str {
		termbox.SetCell(x+i, y, c, coldef, coldef)
	}
}

func underlineWriter(x, y int, str string) {
	for i, c := range str {
		termbox.SetCell(x+i, y, c, coldef+termbox.AttrUnderline, coldef)
	}
}

func redrawAll() {
	termbox.Clear(coldef, coldef)
	w, h = termbox.Size()
	half := w / 2

	curY := 0
	drawHeader(stdWriter, 0, curY, conData)

	curY = 2
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
	y = drawTitle(underlineWriter, x, y, mid+8, "Broker")
	y = drawOne(stdWriter, x, y, mid, "Broker Version", mqttData.get("Broker Version"))
	y = drawOne(stdWriter, x, y, mid, "Broker Time", mqttData.get("Broker Time"))
	y = drawUptime(stdWriter, x, y, mid, "Broker Uptime", mqttData.get("Broker Uptime"))
	y = drawOne(stdWriter, x, y, mid, "Subscriptions Count", mqttData.get("Subscriptions Count"))
	y = drawOne(stdWriter, x, y, mid, "Total Bytes Sent", mqttData.get("Bytes Sent"))
	y = drawOne(stdWriter, x, y, mid, "Total Bytes Received", mqttData.get("Bytes Received"))

	return y
}

func drawLoad(x, y int) int {
	mid := 14
	y = drawLoadTitle(underlineWriter, x, y, mid+2)
	y = drawThree(stdWriter, x, y, mid, "Sockets", mqttData["LoadSockets1min"], mqttData["LoadSockets5min"], mqttData["LoadSockets15min"])
	y = drawThree(stdWriter, x, y, mid, "Connections", mqttData["LoadConnections1min"], mqttData["LoadConnections5min"], mqttData["LoadConnections15min"])
	y = drawThree(stdWriter, x, y, mid, "Msg Received", mqttData["LoadMessagesReceived1min"], mqttData["LoadMessagesReceived5min"], mqttData["LoadMessagesReceived15min"])
	y = drawThree(stdWriter, x, y, mid, "Msg Sent", mqttData["LoadMessagesSent1min"], mqttData["LoadMessagesSent5min"], mqttData["LoadMessagesSent15min"])

	y = drawThree(stdWriter, x, y, mid, "Bytes Received", mqttData["LoadBytesReceived1min"], mqttData["LoadBytesReceived5min"], mqttData["LoadBytesReceived15min"])
	y = drawThree(stdWriter, x, y, mid, "Bytes Sent", mqttData["LoadBytesSent1min"], mqttData["LoadBytesSent5min"], mqttData["LoadBytesSent15min"])

	y = drawThree(stdWriter, x, y, mid, "Pub Received", mqttData["LoadPublishReceived1min"], mqttData["LoadPublishReceived5min"], mqttData["LoadPublishReceived15min"])
	y = drawThree(stdWriter, x, y, mid, "Pub Sent", mqttData["LoadPublishSent1min"], mqttData["LoadPublishSent5min"], mqttData["LoadPublishSent15min"])
	y = drawThree(stdWriter, x, y, mid, "Pub Dropped", mqttData["LoadPublishDropped1min"], mqttData["LoadPublishDropped5min"], mqttData["LoadPublishDropped15min"])

	return y
}

func drawMessages(x, y int) int {
	mid := 18
	y = drawTitle(underlineWriter, x, y, mid+8, "Message Stats")
	y = drawOneFixedNum(stdWriter, x, y, mid, "Messages Received", mqttData.get("Messages Received"))
	y = drawOneFixedNum(stdWriter, x, y, mid, "Messages Sent", mqttData.get("Messages Sent"))
	y = drawOneFixedNum(stdWriter, x, y, mid, "Messages In-flight", mqttData.get("Messages Inflight"))
	y = drawOneFixedNum(stdWriter, x, y, mid, "Messages Stored", mqttData.get("Messages Stored"))
	y++
	y = drawOneFixedNum(stdWriter, x, y, mid, "Messages Publish Dropped", mqttData.get("Messages Publish Dropped"))
	y = drawOneFixedNum(stdWriter, x, y, mid, "Messages Publish Sent", mqttData.get("Messages Publish Sent"))
	y = drawOneFixedNum(stdWriter, x, y, mid, "Messages Publish Received", mqttData.get("Messages Publish Received"))
	y = drawOneFixedNum(stdWriter, x, y, mid, "Messages Retained Count", mqttData.get("Messages Retained Count"))

	return y
}

func drawClient(x, y int) int {
	mid := 20
	y = drawTitle(underlineWriter, x, y, mid+8, "Clients")
	y = drawOne(stdWriter, x, y, mid, "Clients Total", mqttData.get("Clients Total"))
	y = drawOne(stdWriter, x, y, mid, "Clients Connected", mqttData.get("Clients Connected"))
	y = drawOne(stdWriter, x, y, mid, "Clients Disconnected", mqttData.get("Clients Disconnected"))
	y = drawOne(stdWriter, x, y, mid, "Clients Expired", mqttData.get("Clients Expired"))
	y = drawOne(stdWriter, x, y, mid, "Clients Maximum", mqttData.get("Clients Maximum"))

	return y
}

func drawTitle(w termWriter, x int, y int, max int, title string) int {
	str := fmt.Sprintf("%-*s", max, title)
	w(x, y, str)

	return y + 1
}

func drawOne(w termWriter, x int, y int, mid int, label string, data string) int {
	str := fmt.Sprintf("%*s : %s", mid, label, data)
	w(x, y, str)

	return y + 1
}

func drawOneFixedNum(w termWriter, x int, y int, mid int, label string, data string) int {
	str := fmt.Sprintf("%*s : %s", mid, label, fixedLenNum(data))
	w(x, y, str)

	return y + 1
}

func drawLoadTitle(w termWriter, x int, y int, max int) int {
	str := fmt.Sprintf("%-*s  %s  %s %s", max, "Load", "1 min", "5 min", "15 min")
	w(x, y, str)

	return y + 1
}

// Print a number such that it always fits into 6 spaces
func fixedLenNum(num string) string {
	numWidth := 6
	if num == "n/a" || len(num) < 6 {
		return fmt.Sprintf("%*s", numWidth, num)
	}

	fNum, err := strconv.ParseFloat(num, 32)
	if err != nil {
		return fmt.Sprintf("%*s", numWidth, "#err")
	}

	for i := 2; i >= 0; i-- {
		str := fmt.Sprintf("%*.*f", numWidth, i, fNum)
		if len(str) <= numWidth {
			return str
		}
	}

	numM := fNum / 1000000
	for i := 2; i >= 0; i-- {
		str := fmt.Sprintf("%*.*f M", numWidth-2, i, numM)
		if len(str) <= numWidth {
			return str
		}
	}

	numB := numM / 1000
	return fmt.Sprintf("%*.2f B", numWidth-2, numB)
}

func drawThree(w termWriter, x int, y int, mid int, label string, oneMin string, fiveMin string, fifteenMin string) int {
	str := fmt.Sprintf("%*s : %6s %6s %6s", mid, label, fixedLenNum(oneMin), fixedLenNum(fiveMin), fixedLenNum(fifteenMin))
	w(x, y, str)

	return y + 1
}

func drawUptime(w termWriter, x int, y int, mid int, label string, uptime string) int {
	var secs int
	var unit string
	_, err := fmt.Sscanf(uptime, "%d %s", &secs, &unit)
	if err == nil && unit == "seconds" && secs > 0 {
		uptime = fmt.Sprint(time.Duration(secs) * time.Second)
	}
	str := fmt.Sprintf("%*s : %s", mid, label, uptime)
	w(x, y, str)

	return y + 1
}

func drawHeader(w termWriter, x, y int, conData ConnectHandler) int {
	now := time.Now()
	since := now.Sub(startTime)
	h := int(since.Hours())
	m := int(since.Minutes()) % 60
	s := int(since.Seconds()) % 60
	if conData.IsConnected == true {
		headerStr := fmt.Sprintf("Now:  %-24s  Connected for: %3d:%02d:%02d  [Q] to quit", now.Format(datePrint), h, m, s)
		w(x, y, headerStr)

		y++
		return y
	}

	headerStr := fmt.Sprintf("Now:  %-24s  Disconnected!!!  [Q] to quit", now.Format(datePrint), h, m, s)
	w(x, y, headerStr)
	y++

	if conData.Err != nil {
		errString := fmt.Sprintf("    Error: %s", conData.Err.Error())
		w(x, y, errString)
		y++
	}
	return y
}
