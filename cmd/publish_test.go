package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePublishOptions(t *testing.T) {
	pubOpts := publishOptions{
		message:     "some data",
		doNullMsg:   true,
		doStdinLine: true,
		doStdinFile: true,
		filePath:    "bad/path",
	}

	err := pubOpts.validateOptions()
	assert.Equal(t, "stat bad/path: no such file or directory", err.Error(), "error message not right")

	// The test runs in the cmd dir so the path below exists
	pubOpts.filePath = "publish_test.go"
	err = pubOpts.validateOptions()
	assert.Equal(t, "only one of --message, --file, --stdin-line, --stdin-file, or --null-message can be used", err.Error(), "error message not right")

	pubOpts = publishOptions{}
	err = pubOpts.validateOptions()
	assert.Equal(t, "must specify one of --message, --file, --stdin-line, --stdin-file, or --null-message to send any data", err.Error(), "error message not right")

	pubOpts = publishOptions{
		doNullMsg: true,
		qos:       7,
	}
	err = pubOpts.validateOptions()
	assert.Equal(t, "--qos value must or 0, 1 or 2", err.Error(), "error message not right")

	pubOpts.qos = 2
	err = pubOpts.validateOptions()
	assert.Nil(t, err)
}

func TestNewPublishCommand(t *testing.T) {
	cmd := newPublishCommand()
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("message"))
	assert.NotNil(t, flags.Lookup("file"))
	assert.NotNil(t, flags.Lookup("stdin-line"))
	assert.NotNil(t, flags.Lookup("stdin-file"))
	assert.NotNil(t, flags.Lookup("null-message"))
	assert.Nil(t, flags.Lookup("not-an-option"))
}
