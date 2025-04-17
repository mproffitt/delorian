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

package queryinput

import (
	"io"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/delorian/pkg/theme"
	"gopkg.in/op/go-logging.v1"
)

const title = "yaml query"

type YqErrorMsg struct {
	Error error
}

type YqOutputMsg struct {
	Filter string
	Input  string
	Output string
}

func YqOutputCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return YqOutputMsg{
			Output: msg,
		}
	}
}

func YqErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return YqErrorMsg{
			Error: err,
		}
	}
}

type Model struct {
	decoder yqlib.Decoder
	encoder yqlib.Encoder
	filter  textinput.Model
	input   *string
	style   lipgloss.Style
}

func disableLogging() {
	backend := logging.NewLogBackend(io.Discard, "", 0)
	logging.SetBackend(backend)
}

func New(input *string, width int) *Model {
	disableLogging()
	prefs := yqlib.NewDefaultYamlPreferences()
	m := Model{
		decoder: yqlib.NewYamlDecoder(prefs),
		encoder: yqlib.NewYamlEncoder(prefs),
		filter:  textinput.New(),
		input:   input,
		style: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(theme.Colours.Green),
	}
	m.filter.TextStyle = m.filter.TextStyle.UnsetMargins()
	m.filter.Width = width
	return &m
}

// Blurs the textinput
func (m *Model) Blur() {
	m.filter.Blur()
}

// Passes the focus to the filter textinput
func (m *Model) Focus() {
	m.filter.Focus()
}

// Does the textinput currently have focus
func (m *Model) Focused() bool {
	return m.filter.Focused()
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) SetSize(width, height int) tea.Model {
	m.filter.Width = width
	return m
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		err error
		cmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		default:
			m.filter, _ = m.filter.Update(msg)
			filter := m.filter.Value()
			var output string
			{
				output, err = yqlib.NewStringEvaluator().
					Evaluate(filter, *m.input, m.encoder, m.decoder)
				log.Debug("query", "filter", filter, "input", m.input, "output", output, "error", err)
				cmd = YqOutputCmd(output)
				if err != nil {
					cmd = YqErrorCmd(err)
				}
			}
		}
	}
	return m, cmd
}

func (m *Model) View() string {
	colour := theme.Colours.Black
	titleColour := theme.Colours.Black
	if m.Focused() {
		colour = theme.Colours.Blue
		titleColour = theme.Colours.BrightYellow
	}
	content := m.style.
		BorderForeground(colour).
		Render(m.filter.View())
	return overlay.PlaceOverlay(2, 0,
		lipgloss.NewStyle().
			Foreground(titleColour).
			Render(title),
		content, false)
}
