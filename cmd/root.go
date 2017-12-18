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
	"log"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/docker/docker/pkg/term"
	"github.com/spf13/cobra"
)

const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{ wrappedFlagUsages . | trimRightSpace}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

var cfgFile string

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ver string, rev string) {
	version = ver
	revision = rev

	rootCmd := setupRootCommand()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupRootCommand() *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "zap",
		Short: "Listen or publish to a MQTT broker",
		Long: `zap - what happens when technology meets mosquito

	zap is a little utility for publishing or subscribing to events for the
	MQTT message bus`,
	}

	cobra.AddTemplateFunc("wrappedFlagUsages", wrappedFlagUsages)

	rootCmd.AddCommand(
		newSubscribeCommand(),
		newVersionCommand(),
		newPublishCommand(),
		newStatsCommand(),
	)
	rootCmd.SetUsageTemplate(usageTemplate)

	// Uncomment these to turn on debugging from within the mqtt library.
	MQTT.ERROR = log.New(os.Stdout,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	MQTT.CRITICAL = log.New(os.Stdout,
		"CRITICAL: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	// MQTT.WARN = log.New(os.Stdout,
	//        "WARN: ",
	//        log.Ldate|log.Ltime|log.Lshortfile)
	// MQTT.DEBUG = log.New(os.Stdout,
	//        "DEBUG: ",
	//        log.Ldate|log.Ltime|log.Lshortfile)

	return rootCmd
}

func wrappedFlagUsages(cmd *cobra.Command) string {
	width := 80
	if ws, err := term.GetWinsize(0); err == nil {
		width = int(ws.Width)
	}
	return cmd.Flags().FlagUsagesWrapped(width - 1)
}
