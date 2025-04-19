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

package yamlview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/token"
	"github.com/mproffitt/bmx/pkg/exec"
	"github.com/mproffitt/delorian/pkg/components"
	"github.com/mproffitt/delorian/pkg/components/queryinput"
	"github.com/mproffitt/delorian/pkg/components/splash"
	"github.com/mproffitt/delorian/pkg/theme"
	wrap "github.com/muesli/reflow/wrap"
)

const (
	NoFocus components.FocusType = iota
	QueryFocus
	ViewportFocus
)

type Model struct {
	border           bool
	current          components.File
	error            error
	focus            components.FocusType
	filename         string
	height           int
	input            string
	ok               bool
	output           string
	query            tea.Model
	showQuery        bool
	splash           *splash.Model
	style            lipgloss.Style
	viewport         viewport.Model
	width            int
	LineNumber       bool
	LineNumberFormat func(num int) string
}

func (m *Model) defaultLineNumberFormat(num int) string {
	number := fmt.Sprintf("%4d │ ", num)
	if m.focus == ViewportFocus {
		return lipgloss.NewStyle().Foreground(theme.Colours.BrightBlack).Render(number)
	}
	return lipgloss.NewStyle().Foreground(theme.Colours.Black).Render(number)
}

func New(w, h int, query bool) *Model {
	m := Model{
		border: false,
		style: lipgloss.NewStyle().
			BorderForeground(theme.Colours.Blue),
		focus:      NoFocus,
		splash:     splash.New("loading kustomizations..."),
		showQuery:  query,
		input:      "",
		viewport:   viewport.New(w, h),
		LineNumber: true,
	}
	m.query = queryinput.New(&m.input, w)

	return &m
}

func (m *Model) Init() tea.Cmd {
	return m.splash.Init()
}

func (m *Model) NextFocus() components.FocusType {
	switch m.focus {
	case NoFocus:
		m.focus = ViewportFocus
		if m.showQuery {
			m.focus = QueryFocus
			m.query.(components.Focusable).Focus()
		}
	case QueryFocus:
		m.focus = ViewportFocus
		m.query.(components.Focusable).Blur()
	case ViewportFocus:
		m.focus = NoFocus
	}
	return m.focus
}

func (m *Model) PreviousFocus() components.FocusType {
	switch m.focus {
	case NoFocus:
		m.focus = ViewportFocus
	case QueryFocus:
		m.focus = NoFocus
		m.query.(components.Focusable).Blur()
	case ViewportFocus:
		m.focus = NoFocus
		if m.showQuery {
			m.focus = QueryFocus
			m.query.(components.Focusable).Focus()
		}
	}
	return m.focus
}

func (m *Model) formatFilename() int {
	if !m.ok {
		return 0
	}

	title := "Filename: "
	padding := len(title)
	title = lipgloss.NewStyle().Foreground(theme.Colours.BrightRed).Render(title)

	filename := wrap.String(m.current.GetPath(), m.width-padding)
	lines := make([]string, 0)

	style := lipgloss.NewStyle().Foreground(theme.Colours.Purple)
	for i, line := range strings.Split(filename, "\n") {
		l := style.PaddingLeft(padding).Render(line)
		if i == 0 {
			l = style.Render(line)
		}
		lines = append(lines, l)
	}
	m.filename = title + strings.Join(lines, "\n")
	return len(lines)
}

func (m *Model) SetSize(w, h int) tea.Model {
	m.width = w
	m.height = h
	l := m.formatFilename()
	subtract := (2 * theme.Padding) + 1
	m.query.(components.Scalable).SetSize(w-subtract, 0)
	m.viewport.Height = h - l
	m.viewport.Width = w // + 1) - subtract
	return m
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case components.TabChangedMsg:
		m.splash.SetVisible(true)
		cmd = splash.TickCmd()
	case queryinput.YqErrorMsg:
		m.output = msg.Error.Error()
	case components.ModelErrorMsg:
		m.error = msg.Error
		m.splash.SetVisible(false)
	case queryinput.YqOutputMsg:
		m.output = msg.Output
	case components.FileMsg:
		m.current = msg.File
		m.SetSize(m.width, m.height)
		m.ok = msg.Ok
		m.error = fmt.Errorf("no content")
		if m.ok {
			m.error = nil
			m.input = msg.Content
			m.output = m.input
		}
		m.splash.SetVisible(false)
	case components.FluxExecMsg:
		m.error = nil
		m.input = msg.Output
		m.output = m.input
		m.splash.SetVisible(false)
	case tea.KeyMsg:
		switch m.focus {
		case QueryFocus:
			m.query, cmd = m.query.Update(msg)
		case ViewportFocus:
			m.viewport, cmd = m.viewport.Update(msg)
		}
	}
	return m, cmd
}

func (m *Model) UseBorder() tea.Model {
	m.border = true
	return m
}

func (m *Model) View() string {
	if m.splash.Visible() {
		splash := lipgloss.Place(
			m.viewport.Width,
			m.viewport.Height,
			lipgloss.Center,
			lipgloss.Center,
			m.splash.SetWidth(m.width).View(),
		)
		m.viewport.SetContent(splash)
		return m.viewport.View()
	}

	if m.error != nil {
		msg := m.error.Error()
		switch e := m.error.(type) {
		case *exec.BmxExecError:
			msg = e.StyledError(m.width)
		}
		msg = lipgloss.NewStyle().
			Foreground(theme.Colours.Red).
			MarginLeft(1).
			Render(msg)
		msg = lipgloss.Place(m.viewport.Width, m.viewport.Height,
			lipgloss.Center, lipgloss.Center, msg)
		m.viewport.SetContent(msg)
		return m.viewport.View()
	}

	m.viewport.SetContent(m.print(m.output))
	view := m.viewport.View()
	if m.border {
		m.style = m.style.Border(lipgloss.RoundedBorder(), true)
	}
	switch m.focus {
	case ViewportFocus:
		view = m.style.Render(view)
	default:
		view = m.style.BorderForeground(theme.Colours.Black).Render(view)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, view, m.filename)
	if m.showQuery {
		content = lipgloss.JoinVertical(
			lipgloss.Left, m.query.View(), view, m.filename)
	}
	return lipgloss.NewStyle().
		// MarginLeft(theme.Padding).
		Render(content)
}

func (m *Model) prop(col lipgloss.AdaptiveColor) func(...string) string {
	return lipgloss.NewStyle().Foreground(col).Render
}

func (m *Model) renderer(t *token.Token) func(...string) string {
	switch t.PreviousType() {
	case token.AnchorType:
		return m.prop(theme.Colours.Cyan)
	case token.AliasType:
		return m.prop(theme.Colours.Black)
	}
	switch t.NextType() {
	case token.MappingValueType:
		return m.prop(theme.Colours.Blue)
	}
	switch t.Type {
	case token.BoolType:
		return m.prop(theme.Colours.BrightRed)
	case token.AnchorType:
		return m.prop(theme.Colours.Cyan)
	case token.AliasType:
		return m.prop(theme.Colours.BrightCyan)
	case token.StringType, token.SingleQuoteType, token.DoubleQuoteType:
		return m.prop(theme.Colours.Green)
	case token.IntegerType, token.FloatType:
		return m.prop(theme.Colours.BrightYellow)
	case token.CommentType:
		return m.prop(theme.Colours.BrightBlack)
	}

	return m.prop(theme.Colours.Black)
}

func (m *Model) print(content string) string {
	tokens := lexer.Tokenize(content)
	if len(tokens) == 0 {
		return ""
	}

	if m.LineNumber {
		if m.LineNumberFormat == nil {
			m.LineNumberFormat = m.defaultLineNumberFormat
		}
	}

	texts := []string{}
	lineNumber := tokens[0].Position.Line
	for _, tk := range tokens {
		lines := strings.Split(tk.Origin, "\n")
		render := m.renderer(tk)
		header := ""
		if m.LineNumber {
			header = m.LineNumberFormat(lineNumber)
		}
		if len(lines) == 1 {
			line := render(lines[0])
			if len(texts) == 0 {
				texts = append(texts, header+line)
				lineNumber++
			} else {
				text := texts[len(texts)-1]
				texts[len(texts)-1] = text + line
			}
		} else {
			for idx, src := range lines {
				if m.LineNumber {
					header = m.LineNumberFormat(lineNumber)
				}
				line := render(src)
				if idx == 0 {
					if len(texts) == 0 {
						texts = append(texts, header+line)
						lineNumber++
					} else {
						text := texts[len(texts)-1]
						texts[len(texts)-1] = text + line
					}
				} else {
					texts = append(texts, fmt.Sprintf("%s%s", header, line))
					lineNumber++
				}
			}
		}
	}
	for _, line := range texts {
		m.viewport.Width = max(m.viewport.Width, len(line))
		// texts[i] = truncate.String(line, uint(m.viewport.Width))
	}
	return strings.Join(texts, "\n")
}
