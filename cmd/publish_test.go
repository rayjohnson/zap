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
		filePath:    "some/path",
	}

	err := validatePublishOptions(pubOpts)
	assert.Equal(t, "only one of --message, --file, --stdin-line, --stdin-file, or --null-message can be used", err.Error(), "error message not right")

	pubOpts = publishOptions{}
	err = validatePublishOptions(pubOpts)
	assert.Equal(t, "must specify one of --message, --file, --stdin-line, --stdin-file, or --null-message to send any data", err.Error(), "error message not right")

	pubOpts = publishOptions{
		doNullMsg: true,
		qos:       7,
	}
	err = validatePublishOptions(pubOpts)
	assert.Equal(t, "--qos value must or 0, 1 or 2", err.Error(), "error message not right")

	pubOpts.qos = 2
	err = validatePublishOptions(pubOpts)
	assert.Nil(t, err)
}
