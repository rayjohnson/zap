package cmd

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func parseConOpts(args []string) (*connectionOptions, error) {
	flags := pflag.NewFlagSet("con", pflag.ContinueOnError)
	flags.SetOutput(ioutil.Discard)
	flags.Usage = nil
	conOpts := addConnectionFlags(flags)
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	// TODO - this is not interesting yet
	// need to refactor a little more to do proper arg checking
	return conOpts, nil
}

func parseMustError(t *testing.T, args string) {
	_, err := parseConOpts(strings.Split(args, " "))
	assert.Error(t, err, args)
}

func mustParse(t *testing.T, args string) *connectionOptions {
	conOpts, err := parseConOpts(strings.Split(args, " "))
	assert.NoError(t, err)
	return conOpts
}

func TestBadOptions(t *testing.T) {
	parseMustError(t, "--bad-option foo")
	parseMustError(t, "--server foo  --bad-option")
}

func TestValidOptions(t *testing.T) {
	mustParse(t, "--server foo") // TODO: need better arg checking
	conOpts := mustParse(t, "--cert someCert --key someKey --cafile someCa")
	assert.Equal(t, conOpts.certFile, "someCert", "they should be equal")
	assert.Equal(t, conOpts.caFile, "someCa", "they should be equal")
	assert.Equal(t, conOpts.keyFile, "someKey", "they should be equal")
}
