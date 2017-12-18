package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSubscribeOptions(t *testing.T) {
	subOpts := &subscribeOptions{
		qos: 10,
	}
	err := subOpts.validateOptions()
	assert.Equal(t, "--qos value must or 0, 1 or 2", err.Error(), "error message not right")

	subOpts = &subscribeOptions{
		templateString: "{{",
	}
	err = subOpts.validateOptions()
	assert.Equal(t, "template: stdout:1: unexpected unclosed action in command", err.Error(), "error message not right")
	assert.Nil(t, subOpts.stdoutTemplate)

	subOpts.templateString = ".Message"
	err = subOpts.validateOptions()
	assert.Nil(t, err)
	assert.NotNil(t, subOpts.stdoutTemplate)
}

func TestNewSubscribeCommand(t *testing.T) {
	cmd := newSubscribeCommand()
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("clean-session"))
	assert.NotNil(t, flags.Lookup("template"))
	assert.NotNil(t, flags.Lookup("topic"))
	assert.NotNil(t, flags.Lookup("count"))
	assert.NotNil(t, flags.Lookup("skip-retained"))
	assert.NotNil(t, flags.Lookup("qos"))
	assert.Nil(t, flags.Lookup("not-an-option"))
}
