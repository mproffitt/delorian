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

package tabview

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/mproffitt/delorian/pkg/components"
	"github.com/mproffitt/delorian/pkg/components/diffview"
	"github.com/mproffitt/delorian/pkg/components/splash"
	"github.com/mproffitt/delorian/pkg/components/yamlview"
	"github.com/mproffitt/delorian/pkg/theme"
)

type Model struct {
	id         string
	activeTab  int
	height     int
	focus      bool
	tabs       []components.TabType
	tabContent map[components.TabType]tea.Model
	styles     styles
	width      int
}

type styles struct {
	docStyle         lipgloss.Style
	windowStyle      lipgloss.Style
	activeTabStyle   lipgloss.Style
	inactiveTabStyle lipgloss.Style
	tabGap           lipgloss.Style
}

func New() *Model {
	id := zone.NewPrefix()
	m := Model{
		id: id,
		tabs: []components.TabType{
			components.TabKustomize,
			components.TabSource,
			components.TabFluxBuild,
			components.TabFluxDiff,

			/*components.TabGraph,*/
		},
		tabContent: map[components.TabType]tea.Model{
			components.TabKustomize: yamlview.New(0, 0, false),
			components.TabSource:    yamlview.New(0, 0, false),
			components.TabFluxBuild: yamlview.New(0, 0, true),
			components.TabFluxDiff:  diffview.New(0, 0, true),
		},
		activeTab: 0,
		styles: styles{
			docStyle: lipgloss.NewStyle().Padding(0, 2, 0, 0),
			windowStyle: lipgloss.NewStyle().
				BorderForeground(theme.Colours.Blue).
				Align(lipgloss.Left).
				Border(lipgloss.RoundedBorder()).
				UnsetBorderTop(),
			inactiveTabStyle: lipgloss.NewStyle().Border(theme.TabBorder).
				BorderForeground(theme.Colours.Blue).
				Padding(0, 1),
		},
	}
	m.styles.activeTabStyle = m.styles.inactiveTabStyle.
		Border(theme.TabActiveBorder, true).
		BorderForeground(theme.Colours.Blue)
	m.styles.tabGap = m.styles.activeTabStyle.Border(theme.TabGapBorder, true)

	return &m
}

func (m *Model) NextFocus() components.FocusType {
	tab := m.tabs[m.activeTab]
	if _, ok := m.tabContent[tab].(components.Focus); ok {
		focus := m.tabContent[tab].(components.Focus).NextFocus()
		m.focus = focus != yamlview.NoFocus
		return focus
	}
	m.focus = false
	return yamlview.NoFocus
}

func (m *Model) PreviousFocus() components.FocusType {
	tab := m.tabs[m.activeTab]
	if _, ok := m.tabContent[tab].(components.Focus); ok {
		focus := m.tabContent[tab].(components.Focus).PreviousFocus()
		m.focus = focus != yamlview.NoFocus
		return focus
	}
	m.focus = false
	return yamlview.NoFocus
}

func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, tab := range m.tabContent {
		cmd := tab.Init()
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (m *Model) SetSize(w, h int) tea.Model {
	m.height = h - (2 * theme.Padding)
	m.width = w - theme.Padding
	for t, v := range m.tabContent {
		if _, ok := v.(components.Scalable); ok {
			m.tabContent[t].(components.Scalable).
				SetSize(m.width, m.height)
		}
	}
	return m
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.MouseMsg:
		cmds := make([]tea.Cmd, 0)
		for i := range m.tabContent {
			m.tabContent[i], cmd = m.tabContent[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Action != tea.MouseActionRelease {
				break
			}
			for i, tab := range m.tabs {
				if zone.Get(m.id + string(tab)).InBounds(msg) {
					m.activeTab = i
					cmd = components.TabChangedCmd(m.tabs[m.activeTab])
					cmds = append(cmds, cmd)
					break
				}
			}
		}
		cmd = tea.Batch(cmds...)
	case tea.KeyMsg:
		switch msg.String() {
		case ":":
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
			cmd = components.TabChangedCmd(m.tabs[m.activeTab])
		case ";":
			m.activeTab = max(m.activeTab-1, 0)
			cmd = components.TabChangedCmd(m.tabs[m.activeTab])
		default:
			tab := m.tabs[m.activeTab]
			m.tabContent[tab], cmd = m.tabContent[tab].Update(msg)
		}
	case splash.TickMsg:
		cmds := make([]tea.Cmd, 0)
		for k, t := range m.tabContent {
			m.tabContent[k], cmd = t.Update(msg)
			cmds = append(cmds, cmd)
		}
		cmd = tea.Batch(cmds...)
	default:
		tab := m.tabs[m.activeTab]
		m.tabContent[tab], cmd = m.tabContent[tab].Update(msg)
	}
	return m, cmd
}

func (m *Model) View() string {
	var renderedTabs []string

	for i, t := range m.tabs {
		tabTitle := string(t)
		tabTitle = zone.Mark(m.id+tabTitle, tabTitle)
		var style lipgloss.Style
		isFirst, isActive := i == 0, i == m.activeTab
		if isActive {
			style = m.styles.activeTabStyle
		} else {
			style = m.styles.inactiveTabStyle
		}
		if !m.focus {
			style = style.BorderForeground(theme.Colours.Black)
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst {
			border.BottomLeft = "│"
			if !isActive {
				border.BottomLeft = "├"
			}
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(tabTitle))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	spacer := strings.Repeat(" ", max(0, m.width-lipgloss.Width(row)-theme.Padding))
	gapStyle := m.styles.tabGap
	windowStyle := m.styles.windowStyle

	if !m.focus {
		gapStyle = gapStyle.BorderForeground(theme.Colours.Black)
		windowStyle = windowStyle.BorderForeground(theme.Colours.Black)
	}

	gap := gapStyle.Render(spacer)

	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)

	active := m.tabs[m.activeTab]
	view := viewport.New(m.width, m.height)
	view.SetContent(m.tabContent[active].View())
	doc := lipgloss.JoinVertical(lipgloss.Left,
		row,
		windowStyle.Render(view.View()))
	return m.styles.docStyle.Render(doc)
}
