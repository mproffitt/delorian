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

package diffview

import (
	"slices"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/exec"
	"github.com/mproffitt/delorian/pkg/components"
	"github.com/mproffitt/delorian/pkg/components/filter"
	"github.com/mproffitt/delorian/pkg/components/splash"
	"github.com/mproffitt/delorian/pkg/theme"
)

type Model struct {
	border     bool
	entries    []DiffEntry
	filter     tea.Model
	focus      components.FocusType
	height     int
	showFilter bool
	style      lipgloss.Style
	viewport   viewport.Model
	width      int
	splash     *splash.Model
	error      error
}

func New(w, h int, showFilter bool) *Model {
	m := Model{
		border:     false,
		entries:    []DiffEntry{},
		focus:      NoFocus,
		showFilter: showFilter,
		style: lipgloss.NewStyle().
			BorderForeground(theme.Colours.Blue),
		viewport: viewport.New(w, h),
		splash:   splash.New("Waiting for Kustomization diffing..."),
	}

	return &m
}

func (m *Model) Init() tea.Cmd {
	return m.splash.Init()
}

func (m *Model) NextFocus() components.FocusType {
	switch m.focus {
	case NoFocus:
		m.focus = ViewportFocus
		if m.showFilter && m.filter != nil {
			m.focus = FilterFocus
			m.filter.(components.Focusable).Focus()
		}
	case FilterFocus:
		m.focus = ViewportFocus
		if m.filter != nil {
			m.filter.(components.Focusable).Blur()
		}
	case ViewportFocus:
		m.focus = NoFocus
		if m.filter != nil {
			m.filter.(components.Focusable).Blur()
		}
	}
	return m.focus
}

func (m *Model) PreviousFocus() components.FocusType {
	switch m.focus {
	case NoFocus:
		m.focus = ViewportFocus
		if m.filter != nil {
			m.filter.(components.Focusable).Blur()
		}
	case FilterFocus:
		m.focus = NoFocus
		if m.filter != nil {
			m.filter.(components.Focusable).Blur()
		}
	case ViewportFocus:
		m.focus = NoFocus
		if m.showFilter && m.filter != nil {
			m.focus = FilterFocus
			m.filter.(components.Focusable).Focus()
		}
	}
	return m.focus
}

func (m *Model) SetSize(w, h int) tea.Model {
	m.width = w
	m.height = h
	m.viewport.Height = h
	m.viewport.Width = w
	if m.filter != nil {
		m.filter = m.filter.(*filter.Model).SetSize(w-(theme.Padding+1), h)
	}
	return m
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case components.TabChangedMsg:
		m.splash.SetVisible(true)
		cmd = splash.TickCmd()
	case components.FluxExecMsg:
		log.Debug("diffview", "update", msg)
		m.entries = m.parseFluxDiff(msg.Output)
		m.filter = m.getFilter()
		m.viewport.SetContent(m.print(m.entries))
		m.splash.SetVisible(false)
	case splash.TickMsg:
		m.splash, cmd = m.splash.Update(msg)
	case components.ModelErrorMsg:
		m.error = msg.Error
		m.splash.SetVisible(false)
	case tea.KeyMsg, tea.MouseMsg:
		switch m.focus {
		case FilterFocus:
			m.filter, cmd = m.filter.Update(msg)
			m.viewport.SetContent(m.print(m.entries))

		case ViewportFocus:
			m.viewport, cmd = m.viewport.Update(msg)
		}
	}
	return m, cmd
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
			msg = e.StyledError(m.width, config.ColourStyles(theme.Colours))
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

	if len(m.entries) == 0 {
		tick := lipgloss.NewStyle().
			Foreground(theme.Colours.BrightGreen).
			Render("✔")
		msg := lipgloss.NewStyle().
			Foreground(theme.Colours.Blue).
			MarginLeft(1).
			Render("No diff detected")
		msg = lipgloss.JoinHorizontal(lipgloss.Top, tick, msg)
		msg = lipgloss.Place(m.viewport.Width, m.viewport.Height,
			lipgloss.Center, lipgloss.Center, msg)
		m.viewport.SetContent(msg)
		return m.viewport.View()
	}

	m.viewport.Width = m.width
	m.viewport.Height = m.height - m.filter.(*filter.Model).GetHeight() - theme.Padding
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

	content := view
	if m.showFilter {
		content = lipgloss.JoinVertical(lipgloss.Left, m.filter.View(), view)
	}

	return lipgloss.NewStyle().
		Render(content)
}

func (m *Model) getFilter() tea.Model {
	options := []string{
		"metadata.generation",
	}
	selected := options

	for _, item := range m.entries {
		options = append(options, item.GetKind())
		for _, key := range item.Changes {
			options = append(options, key.Key)
		}
	}
	return filter.New(options, selected).
		SetSize(m.width-(theme.Padding+1), m.height)
}

func (m *Model) print(entries []DiffEntry) string {
	content := make([]string, 0)
	filters := m.filter.(*filter.Model).Values()
	log.Debug("printing entries", "filters", filters)
	for _, entry := range entries {
		if !slices.Contains(filters, entry.Kind) {
			content = append(content, entry.WithFilter(filters...).View())
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, content...)
}
