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

package main

import (
	"github.com/rayjohnson/cobra-man/man"
	"github.com/rayjohnson/zap/cmd"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	// VERSION will be set at build time from the VERSION file
	VERSION string
	// COMMIT will come from the Makefile and contains the git rev
	COMMIT string
)

var installDirectory string

// TODO: build different cmd line for this tool
func main() {
	docCmd := setupDocCommands()

	if err := docCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func generateManPages(myCmd *cobra.Command, args []string) error {
	appCmd := cmd.SetupRootCommand()

	manOpts := &man.GenerateManOptions{
		LeftFooter:   "Zap " + VERSION,
		CenterHeader: "Zap Manual",
		Author:       "Ray Johnson <ray.johnson+zap@gmail.com>",
		Directory:    installDirectory,
		Bugs:         `Bugs related to zap can be filed at https://github.com/rjohnson/zap`,
		UseTemplate:  man.MdocManTemplate,
	}
	return man.GenerateManPages(appCmd, manOpts)
}

func generateAutoComplete(myCmd *cobra.Command, args []string) error {
	appCmd := cmd.SetupRootCommand()

	path := filepath.Join(installDirectory, "zap.sh")
	return appCmd.GenBashCompletionFile(path)
}

func setupDocCommands() *cobra.Command {
	var docCmd = &cobra.Command{
		Use:   "doc",
		Args:  cobra.NoArgs,
		Short: "Generate documentation, etc.",
	}
	docCmd.PersistentFlags().StringVar(&installDirectory, "directory", ".", "Directory to install generated files")

	manCmd := &cobra.Command{
		Use:   "generate-man-pages",
		Args:  cobra.NoArgs,
		Short: "Generate man pages",
		RunE:  generateManPages,
	}

	completeCmd := &cobra.Command{
		Use:   "generate-auto-complete",
		Args:  cobra.NoArgs,
		Short: "Generate bash auto complete script",
		RunE:  generateAutoComplete,
	}

	docCmd.AddCommand(manCmd, completeCmd)

	return docCmd
}
