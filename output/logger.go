package output

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

var (
	// VERBOSE will print to stdout if --verbose flag is set
	VERBOSE *log.Logger

	// STDOUTPUT will print anhything to stdout
	STDOUTPUT *log.Logger
)

func init() {
	// This is an undocumented feature just for debugging of the zap tool.
	// Mainly used to debug any issues with MQTT library.
	i, _ := strconv.ParseInt(os.Getenv("ZAP_DEBUG_LEVEL"), 10, 64)
	if i > 0 {
		f, _ := os.OpenFile("zap.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer f.Close()
		setupDebugLog(i, f)
	}

	VERBOSE = log.New(ioutil.Discard, "", 0)
	STDOUTPUT = log.New(os.Stdout, "", 0)
}

func setupDebugLog(level int64, f io.Writer) {
	if level >= 1 {
		MQTT.ERROR = log.New(f,
			"ERROR: ",
			log.Ldate|log.Ltime|log.Lshortfile)
	}
	if level >= 2 {
		MQTT.CRITICAL = log.New(f,
			"CRITICAL: ",
			log.Ldate|log.Ltime|log.Lshortfile)
	}
	if level >= 3 {
		MQTT.WARN = log.New(f,
			"WARN: ",
			log.Ldate|log.Ltime|log.Lshortfile)
	}
	if level >= 4 {
		MQTT.DEBUG = log.New(f,
			"DEBUG: ",
			log.Ldate|log.Ltime|log.Lshortfile)
	}
}
