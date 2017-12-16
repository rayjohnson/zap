package cmd

import (
	"crypto/tls"
	"io/ioutil"
	"strings"
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func parseConOpts(args []string) (*MQTT.ClientOptions, error) {
	flags := pflag.NewFlagSet("con", pflag.ContinueOnError)
	flags.SetOutput(ioutil.Discard)
	flags.Usage = nil
	conOpts := addConnectionFlags(flags)
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	return ParseBrokerInfo(flags, conOpts)
}

func parseMustError(t *testing.T, args string) error {
	_, err := parseConOpts(strings.Split(args, " "))
	assert.Error(t, err, args)
	return err
}

func mustParse(t *testing.T, args string) *MQTT.ClientOptions {
	clientOpts, err := parseConOpts(strings.Split(args, " "))
	assert.NoError(t, err)
	return clientOpts
}

func TestBadOptions(t *testing.T) {
	var err error

	parseMustError(t, "--bad-option foo")
	parseMustError(t, "--server foo  --bad-option")
	err = parseMustError(t, "--keepalive badArg --server foo")
	assert.Equal(t, err.Error(), "invalid argument \"badArg\" for \"-k, --keepalive\" flag: strconv.ParseInt: parsing \"badArg\": invalid syntax", "error message not right")
	// parseMustError(t, "--insecure badArg")  //TODO - this should fail missing noArgs
}

func TestClientId(t *testing.T) {
	clientOpts := mustParse(t, "")
	assert.Regexp(t, "^zap_[0-9]+$", clientOpts.ClientID)

	clientOpts = mustParse(t, "--client-prefix foo-")
	assert.Regexp(t, "^foo-[0-9]+$", clientOpts.ClientID)

	clientOpts = mustParse(t, "--id myid --client-prefix foo-")
	assert.Equal(t, clientOpts.ClientID, "myid")
}

func TestOptions(t *testing.T) {
	clientOpts := mustParse(t, "--server tcp://localhost:1883")
	assert.Equal(t, "tcp://localhost:1883", clientOpts.Servers[0].String(), "they should be equal")

	err := parseMustError(t, "--server foo")
	assert.Equal(t, "parse foo: invalid URI for request", err.Error())
}

func TestCertOptions(t *testing.T) {
	clientOpts := mustParse(t, "")
	assert.Equal(t, clientOpts.TLSConfig.ClientAuth, tls.NoClientCert)
	assert.Nil(t, clientOpts.TLSConfig.RootCAs)

	err := parseMustError(t, "--cert someCert")
	assert.Equal(t, err.Error(), "for tls: both --key and --cert options must be set")

	err = parseMustError(t, "--cert noFile --key noFile")
	assert.Equal(t, err.Error(), "open noFile: no such file or directory")

	err = parseMustError(t, "--cafile noCaFile")
	assert.Equal(t, err.Error(), "open noCaFile: no such file or directory")
}
