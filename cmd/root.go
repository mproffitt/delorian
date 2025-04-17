// Copyright (c) 2025 Martin Proffitt <mprooffitt@choclab.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cmd

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	zone "github.com/lrstanley/bubblezone"
	"github.com/mproffitt/delorian/pkg/manager"
	"github.com/spf13/cobra"
)

var logFile string

var rootCmd = &cobra.Command{
	Use:   "ff",
	Short: "Flux Build and Diff UI",
	Long: `Scans the current directory for kustomization files and offers
    intergrated and interactive build and search tooling for browsing
    rendered manifests`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if len(os.Getenv("DEBUG")) > 0 {
			log.SetLevel(log.DebugLevel)
			if logFile == "" {
				logFile = "debug.log"
			}
		}

		log.SetOutput(io.Discard)
		if logFile != "" {
			f, err := tea.LogToFile(logFile, "debug")
			if err != nil {
				fmt.Println("fatal:", err)
				os.Exit(1)
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Error("failed to close logfile", "file", logFile, "error", err)
				}
			}()
			log.SetOutput(f)
		}

		// Enable bubblezone mouse support
		zone.NewGlobal()
		zone.SetEnabled(true)
		// initialise the model and start the program
		model := manager.New()
		p := tea.NewProgram(model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion())
		if _, err := p.Run(); err != nil {
			log.Fatal("could not start program:", "error", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVarP(&logFile, "logfile", "l",
		"", "log filename to use (empty = no log, default)")
}
