package cmd

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCmd(t *testing.T) {
	cmd := newVersionCommand()
	flags := cmd.Flags()
	assert.NotNil(t, flags.Lookup("generate-auto-complete"))
	assert.NotNil(t, flags.Lookup("generate-man-pages"))
	assert.NotNil(t, flags.Lookup("directory"))
	assert.Nil(t, flags.Lookup("not-an-option"))

	cmd.SetOutput(ioutil.Discard)

	// With no args should just print and no err
	assert.Nil(t, cmd.Execute())

	cmd.SetArgs([]string{"--directory", "bad_dir", "--generate-auto-complete"})
	err := cmd.Execute()
	assert.Equal(t, "stat bad_dir: no such file or directory", err.Error())

	// Create tmp dir for generated files
	// dir, err := ioutil.TempDir()

	// err = os.Remove(tempDirPath)
	//     if err != nil {
	//        log.Fatal(err)
	//    }

}
