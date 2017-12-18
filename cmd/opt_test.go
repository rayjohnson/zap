package cmd

import (
	"crypto/tls"
	"io/ioutil"
	"os"
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
	zapOpts := buildZapFlags(flags)
	zapOpts.conOpts = addConnectionFlags(flags)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	if err := zapOpts.processOptions(flags); err != nil {
		return nil, err
	}

	return zapOpts.clientOpts, nil
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

func TestLoadConfig(t *testing.T) {
	err := parseMustError(t, "--config bad_config_file")
	assert.Equal(t, "path from --config option does not exist: bad_config_file", err.Error(), "error message not right")

	err = parseMustError(t, "--config ../examples/example.zap.toml -b bad_broker")
	assert.Equal(t, "broker \"bad_broker\" does not exist in config file: ../examples/example.zap.toml", err.Error(), "error message not right")

	f, _ := os.Create("bad.toml")
	f.WriteString(`[broke
bad := toml_syntax
`)
	f.Close()
	defer os.Remove("bad.toml")

	err = parseMustError(t, "--config bad.toml")
	assert.Equal(t, "error loading config file: (1, 2): unexpected token unclosed table key, was expecting a table key", err.Error(), "error message not right")
}

func TestBadOptions(t *testing.T) {
	var err error

	parseMustError(t, "--bad-option foo")
	parseMustError(t, "--server foo  --bad-option")
	err = parseMustError(t, "--keepalive badArg --server foo")
	assert.Equal(t, "invalid argument \"badArg\" for \"-k, --keepalive\" flag: strconv.ParseInt: parsing \"badArg\": invalid syntax", err.Error(), "error message not right")

	// err = parseMustError(t, "badArg badArg badArg badArg") //TODO: why does this pass?
}

func TestClientId(t *testing.T) {
	clientOpts := mustParse(t, "")
	assert.Regexp(t, "^zap_[0-9]+$", clientOpts.ClientID)

	clientOpts = mustParse(t, "--client-prefix foo-")
	assert.Regexp(t, "^foo-[0-9]+$", clientOpts.ClientID)

	clientOpts = mustParse(t, "--id myid --client-prefix foo-")
	assert.Equal(t, "myid", clientOpts.ClientID)
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

	err := parseMustError(t, "--tls-cert someCert")
	assert.Equal(t, "for tls: both --tls-key and --tls-cert options must be set", err.Error())

	err = parseMustError(t, "--tls-cert noFile --tls-key noFile")
	assert.Equal(t, "open noFile: no such file or directory", err.Error())

	err = parseMustError(t, "--tls-cacert noCaFile")
	assert.Equal(t, "open noCaFile: no such file or directory", err.Error())
}
