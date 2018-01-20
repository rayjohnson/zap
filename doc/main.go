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
	"github.com/rayjohnson/cobraman"
	"github.com/rayjohnson/zap/cmd"
	"os"
)

var (
	// VERSION will be set at build time from the VERSION file
	VERSION string
	// COMMIT will come from the Makefile and contains the git rev
	COMMIT string
)

func main() {
	// Get the root cobra command for the zap application
	appCmds := cmd.SetupRootCommand(VERSION, COMMIT)

	docGenerator := cobraman.CreateDocGenCmdLineTool(appCmds)
	docGenerator.AddBashCompletionGenerator("zap.sh")

	manOpts := &cobraman.CobraManOptions{
		LeftFooter:   "Zap " + VERSION,
		CenterHeader: "Zap Manual",
		Author:       "Ray Johnson <ray.johnson+zap@gmail.com>",
		Bugs:         `Bugs related to zap can be filed at https://github.com/rayjohnson/zap`,
	}
	docGenerator.AddDocGenerator(manOpts, "mdoc")
	docGenerator.AddDocGenerator(manOpts, "troff")
	docGenerator.AddDocGenerator(manOpts, "markdown")

	if err := docGenerator.Execute(); err != nil {
		os.Exit(1)
	}
}