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

type File interface {
	GetName() string
	GetPath() string
	GetContent() string
}

type Flux interface {
	Build() tea.Cmd
	Diff() tea.Cmd
}

type FocusType int

type Focus interface {
	NextFocus() FocusType
	PreviousFocus() FocusType
}

type FluxExecMsg struct {
	Output string
}

func FluxExecCmd(args []string) tea.Cmd {
	return func() tea.Msg {
		flux, err := exec.LookPath("flux")
		if err != nil {
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
				msg := "identified at least one change, exiting with non-zero exit code"
				if !strings.HasSuffix(err.Stderr, msg) {
					log.Debug("flux exec", "error", err)
					return ModelErrorMsg{Error: err}
				}
				out = err.Stdout
			default:
				log.Debug("flux exec", "error", err)
				return ModelErrorMsg{Error: err}
			}
		}

		log.Debug(args[0], "output", out)
		return FluxExecMsg{Output: out}
	}
}

type Focusable interface {
	Focus()
	Blur()
}

type Scalable interface {
	SetSize(w, h int) tea.Model
}

type FileMsg struct {
	File    File
	Ok      bool
	Content string
}

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

type ModelErrorMsg struct {
	Error error
}

func ModelErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ModelErrorMsg{Error: err}
	}
}

type ModelFatalMsg struct {
	Error error
}

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

type TabChangedMsg struct {
	NewTab TabType
}

func TabChangedCmd(msg TabType) tea.Cmd {
	return func() tea.Msg {
		return TabChangedMsg{
			NewTab: msg,
		}
	}
}

type KustomizationError struct {
	Name      string
	Namespace string
	Filepath  string
	error     error
}

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
