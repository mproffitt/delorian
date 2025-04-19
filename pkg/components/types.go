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

package components

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	bmx "github.com/mproffitt/bmx/pkg/exec"
)

// File interface is implemented by objects which can be
// displayed as a file in one of the viewports such as
// yamlview
type File interface {
	// GetName should return the name associated with this file
	GetName() string

	// GetPath should return the absolute path to the file.
	//
	// If content is rendered, this would normally return an
	// empty string
	GetPath() string

	// GetContent gets the content of the file
	//
	// Depending on the implementation, this may return
	// rendered GetContent. If this is the case, then
	// GetPath should be made to return empty
	GetContent() string
}

// FileMsg is returned by a call from FileCmd
// and contains the underlying file, whether that
// file is Ok and the content of that file discovered
// by a call to GetContent
type FileMsg struct {
	File    File
	Ok      bool
	Content string
}

// FileCmd should be returned by objects which
// have read a file and are required to return the
// contents of that file to a different viewport
//
// This command accepts two arguments, the file, and
// whether that file is OK (e.g. does the file exist)
func FileCmd(msg File, ok bool) tea.Cmd {
	content := msg.GetContent()
	return func() tea.Msg {
		return FileMsg{
			File:    msg,
			Ok:      ok,
			Content: content,
		}
	}
}

// Flux interface defines the methods used to run
// flux commands.
type Flux interface {
	Build() tea.Cmd
	Diff() tea.Cmd
}

// FocusType is used by multi-part components to
// define which part has the current focus, or
// none when focused should be returned
type FocusType int

// Focus interface is used by components to switch between
// elements in their own area that can receive input Focus
//
// Normally these areas would then be highlighted to
// identify visually that a given element has focus.
type Focus interface {
	NextFocus() FocusType
	PreviousFocus() FocusType
}

// Focusable is the interface that defines if a component
// can accept input focus
type Focusable interface {
	Focus()
	Blur()
}

// Scalable is the interface that defines if a component
// can be resized directly.
type Scalable interface {
	SetSize(w, h int) tea.Model
}

// FluxExecMsg is the message sent after the
// execution of a FluxExecCmd
type FluxExecMsg struct {
	Output string
}

// FluxExecCmd executes flux and captures the output
//
// This command should be returned by any object that
// depends on flux execution, and as part of its Update
// function should handle a `FluxExecMsg`
func FluxExecCmd(args []string) tea.Cmd {
	return func() tea.Msg {
		// TODO: This check should occur at program start and be
		// handled in the same way as checking if this is a git repo.
		// It shouldn't wait until the program is already running to
		// know if flux is installed.
		flux, err := exec.LookPath("flux")
		if err != nil {
			log.Error("unable to find flux in path. is this installed?")
			err = &bmx.BmxExecError{
				Command: fmt.Sprintf("%s %s", flux, strings.Join(args, " ")),
				Stdout:  "",
				Stderr:  err.Error(),
			}
			return ModelErrorMsg{Error: err}
		}

		out, _, err := bmx.Exec(flux, args)
		if err != nil {
			switch err := err.(type) {
			case *bmx.BmxExecError:
				// I almost certainly want the option to identify other error
				// strings at this point as some errors contain large blocks of
				// text which may be better displayed in a different manner.
				msg := "identified at least one change, exiting with non-zero exit code"
				if !strings.HasSuffix(err.Stderr, msg) {
					log.Error("flux exec", "error", err)
					return ModelErrorMsg{Error: err}
				}
				out = err.Stdout
			default:
				log.Error("flux exec", "error", err)
				return ModelErrorMsg{Error: err}
			}
		}

		log.Debug(args[0], "output", out)
		return FluxExecMsg{Output: out}
	}
}

// ModelErrorMsg is returned when the UI should enter an error state
type ModelErrorMsg struct {
	Error error
}

// ModelErrorCmd is returned by components that have detected
// an error and wish to alert the calling object
func ModelErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ModelErrorMsg{Error: err}
	}
}

// ModelFatalMsg is raised when the model has entered
// a fatal state from which it cannot recover
//
// If this message is recieved by the component it should
// try to manage it by shutting down or cleaning up its
// properties
type ModelFatalMsg struct {
	Error error
}

// ModelFatalCmd is the command that triggers the ModelFatalMsg
//
// Only use this if the component enters an unrecoverable state
// otherwise use ModelErrorCmd
func ModelFatalCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ModelFatalMsg{Error: err}
	}
}

type TabType string

const (
	TabKustomize TabType = "Kustomization"
	TabSource    TabType = "Source"
	TabFluxBuild TabType = "Flux Build"
	TabFluxDiff  TabType = "Flux Diff"
	TabGraph     TabType = "Graph"
)

// TabChangedMsg is returned when the tabs change on the
// primary view - this helps the program understand what
// information it should be returning
type TabChangedMsg struct {
	NewTab TabType
}

// TabChangedCmd is triggered when a tab change is made
// by the user. This is sent by the `tabview.Model`
func TabChangedCmd(msg TabType) tea.Cmd {
	return func() tea.Msg {
		return TabChangedMsg{
			NewTab: msg,
		}
	}
}

// KustomizationError is an error type raised when
// an error is detected in a kustomization.
type KustomizationError struct {
	Name      string
	Namespace string
	Filepath  string
	error     error
}

// PrettyPrint the error for display
func (k *KustomizationError) Error() {
	var builder strings.Builder

	builder.WriteString("metadata:\n")
	if k.Name != "" {
		builder.WriteString(fmt.Sprintf("  name: %s\n", k.Name))
	}
	if k.Namespace != "" {
		builder.WriteString(fmt.Sprintf("  namespace: %s\n", k.Namespace))
	}
	builder.WriteString(fmt.Sprintf("filepath: %s", k.Filepath))
	builder.WriteString(fmt.Sprintf("error: %s", k.error.Error()))
}
