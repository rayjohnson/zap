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

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var version string
var revision string

var genAutoComplete bool
var genManPages bool
var installDir string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version information about zap",
	Long: `Shows version information about zap

Run with the --generate-auto-complete option and a file named
zap.sh will be generated for use in autocomplete scripts`,
	Run: func(cmd *cobra.Command, args []string) {

		if genAutoComplete || genManPages {
			// Check the given directory
			stat, err := os.Stat(installDir)

			if err != nil && !stat.IsDir() {
				fmt.Println("--directory argument does not exist or is not a directory")
				os.Exit(1)
			}
		}

		if genAutoComplete {
			path := filepath.Join(installDir, "zap.sh")
			cmd.Root().GenBashCompletionFile(path)
			return
		}

		if genManPages {
			header := &doc.GenManHeader{
				Title:   "ZAP",
				Section: "1",
			}
			err := doc.GenManTree(cmd.Root(), header, installDir)
			if err != nil {
				fmt.Printf("Error generating man pages: %s\n", err)
			}
		}

		fmt.Println("zap version " + version + ", Revision: " + revision)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVar(&genAutoComplete, "generate-auto-complete", false, "generates a bash autocomplete script zap.sh")
	versionCmd.Flags().BoolVar(&genManPages, "generate-man-pages", false, "generates the man pages for zap")
	versionCmd.Flags().StringVar(&installDir, "directory", ".", "directory to install generated files")
}
