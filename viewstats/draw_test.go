package viewstats

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFixedLenNum(t *testing.T) {
	var str, expected string

	table := [][]string{
		{"n/a", "   n/a"},
		{"dog", "   dog"},
		{"funny dog", "  #err"},
		{"4", "     4"},
		{"0.25", "  0.25"},
		{"12345.6", " 12346"},
		{"12345.3", " 12345"},
		{"1234.56", "1234.6"},
		{"1234.567", "1234.6"},
		{"123456.7", "123457"},
		{"999999.1", "999999"},
		{"999999.9", "1.00 M"},
		{"1234567", "1.23 M"},
		{"12345678", "12.3 M"},
		{"123456789", " 123 M"},
		{"1234567896", "1235 M"},
		{"12345678968", "12.35 B"},
	}

	for i := 0; i < len(table); i++ {
		str = fixedLenNum(table[i][0])
		expected = table[i][1]
		if str != expected {
			t.Errorf("parseload: expected \"%s\", actual \"%s\"", expected, str)
		}
	}
}

func TestDrawLoadTitle(t *testing.T) {
	var output string
	var f termWriter

	f = func(x, y int, str string) { output = str }

	drawLoadTitle(f, 0, 0, 25)
	assert.Equal(t, "Load                       1 min  5 min 15 min", output, "error message not right")
	drawLoadTitle(f, 0, 0, 5)
	assert.Equal(t, "Load   1 min  5 min 15 min", output, "error message not right")
}

func TestDrawUptime(t *testing.T) {
	var output string
	var f termWriter

	f = func(x, y int, str string) { output = str }

	drawUptime(f, 0, 0, 10, "Broker Uptime", "10 seconds")
	assert.Equal(t, "Broker Uptime : 10s", output, "error message not right")

	drawUptime(f, 0, 0, 10, "Broker Uptime", "100 seconds")
	assert.Equal(t, "Broker Uptime : 1m40s", output, "error message not right")

	drawUptime(f, 0, 0, 10, "Broker Uptime", "1000 seconds")
	assert.Equal(t, "Broker Uptime : 16m40s", output, "error message not right")

	drawUptime(f, 0, 0, 10, "Broker Uptime", "10000 seconds")
	assert.Equal(t, "Broker Uptime : 2h46m40s", output, "error message not right")

	drawUptime(f, 0, 0, 10, "Broker Uptime", "100000 seconds")
	assert.Equal(t, "Broker Uptime : 27h46m40s", output, "error message not right")
}
