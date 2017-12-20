// Copyright © 2017 Ray Johnson <ray.johnson@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rayjohnson/cobra-man/man"
	"github.com/spf13/cobra"
)

var version string
var revision string

type versionOptions struct {
	genAutoComplete bool
	genManPages     bool
	installDir      string
}

func newVersionCommand() *cobra.Command {
	verOpts := versionOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Args:  cobra.NoArgs,
		Short: "Shows version information about zap",
		Long: `Shows version information about zap

Run with the --generate-auto-complete option and a file named
zap.sh will be generated for use in autocomplete scripts.  Use the
--generate-man-pages option to generate man pages for the zap command.
Use the --directory option to specify the location for any generated files`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion(cmd, verOpts)
		},
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	flags := cmd.Flags()
	flags.BoolVar(&verOpts.genAutoComplete, "generate-auto-complete", false, "Generates a bash auto-complete script zap.sh")
	flags.BoolVar(&verOpts.genManPages, "generate-man-pages", false, "Generates the man pages for zap")
	flags.StringVar(&verOpts.installDir, "directory", ".", "Directory to install generated files")

	return cmd
}

func runVersion(cmd *cobra.Command, verOpts versionOptions) error {

	if verOpts.genAutoComplete || verOpts.genManPages {
		// Check the given directory
		stat, err := os.Stat(verOpts.installDir)

		if err != nil {
			return err
		}
		if !stat.IsDir() {
			return fmt.Errorf("--directory argument is not a directory")
		}
	}

	if verOpts.genAutoComplete {
		path := filepath.Join(verOpts.installDir, "zap.sh")
		cmd.Root().GenBashCompletionFile(path)
		fmt.Println("Generated auto-complete script here: " + path)
	}

	if verOpts.genManPages {
		manOpts := &man.GenerateManOptions{
			ProgramName: "Zap " + version,
			Author:      "Ray Johnson <ray.johnson+zap@gmail.com>",
			Directory:   verOpts.installDir,
			Bugs:        `Bugs related to zap can be filed at https://github.com/rjohnson/zap`,
		}
		err := man.GenerateManPages(cmd.Root(), manOpts)
		if err != nil {
			return err
		}
		fmt.Println("Generated man pages here: " + verOpts.installDir)
	}

	fmt.Println("zap version " + version + ", Revision: " + revision)
	return nil
}
