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

package manager

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	zone "github.com/lrstanley/bubblezone"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/components/toast"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/delorian/pkg/components"
	"github.com/mproffitt/delorian/pkg/components/tabview"
	"github.com/mproffitt/delorian/pkg/components/yamlview"
	fluxrepo "github.com/mproffitt/delorian/pkg/repo/flux"
	"github.com/mproffitt/delorian/pkg/theme"
)

type Focus int

const (
	sidebar Focus = iota
	primary
)

type Model struct {
	height int
	keymap *keyMap
	layout layout
	width  int
	focus  Focus
}

type layout struct {
	sidebar tea.Model
	primary tea.Model
	toasts  []*toast.Model
	fatal   *toast.Model
}

// The maximum number of toast messages
// we display at any given time
const MaxToasts = 10

func New() *Model {
	rootPath, _ := os.Getwd()
	m := Model{
		keymap: mapKeys(),
		layout: layout{
			sidebar: fluxrepo.New(rootPath),
			primary: tabview.New(),
			toasts:  make([]*toast.Model, 0, MaxToasts),
		},
	}
	return &m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.layout.sidebar.Init(),
		m.layout.primary.Init(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m, cmd = m.updateKeyMsg(msg)
	case fluxrepo.ModelReadyMsg:
		m.layout.sidebar, cmd = m.layout.sidebar.Update(msg)
	case components.ModelErrorMsg:
		log.Error("model", "error", msg.Error)
		// forward the error to the primary view
		m.layout.primary, _ = m.layout.primary.Update(msg)
		cmd = toast.NewToastCmd(toast.Error, msg.Error.Error())
	case components.ModelFatalMsg:
		m.layout.fatal = toast.New(toast.Error, msg.Error.Error(),
			config.ColourStyles(theme.Colours),
		).SetTickDuration(45 * time.Millisecond).SetCompletionCommand(tea.Quit)
		cmd = m.layout.fatal.Init()
	case tea.WindowSizeMsg:
		cmd = m.resize(msg)
	case toast.NewToastMsg:
		// To prevent flooding, we use a capped slice for toast messages
		// therefore we want to use the last available index to display
		// a warning if we recieve more toast messages than we have
		// capacity for
		if len(m.layout.toasts) < MaxToasts-1 {
			toast := toast.New(msg.Type, msg.Message,
				config.ColourStyles(theme.Colours)).SetTickDuration(25 * time.Millisecond)
			cmd = toast.Init()
			m.layout.toasts = append(m.layout.toasts, toast)
			break
		} else if len(m.layout.toasts) < cap(m.layout.toasts) {
			toast := toast.New(
				toast.Warning,
				"Too many messages to display\nSee log for details",
				config.ColourStyles(theme.Colours))
			cmd = toast.Init()
			m.layout.toasts = append(m.layout.toasts, toast)
		}
	case toast.FrameMsg:
		var cmds []tea.Cmd
		if m.layout.fatal != nil {
			m.layout.fatal, cmd = m.layout.fatal.Update(msg)
			cmds = append(cmds, cmd)
		}
		for i := range m.layout.toasts {
			if m.layout.toasts[i] != nil {
				m.layout.toasts[i], cmd = m.layout.toasts[i].Update(msg)
				cmds = append(cmds, cmd)
			}
		}
		// remove any completed toasts
		newToasts := make([]*toast.Model, 0, MaxToasts)
		for _, v := range m.layout.toasts {
			v := v
			if v != nil {
				newToasts = append(newToasts, v)
			}
		}
		m.layout.toasts = newToasts
		cmd = tea.Batch(cmds...)
	case tea.MouseMsg:
		switch m.focus {
		case sidebar:
			m.layout.sidebar, cmd = m.layout.sidebar.Update(msg)
		case primary:
			m.layout.primary, cmd = m.layout.primary.Update(msg)
		}

	case components.TabChangedMsg:
		// These messages need to go to both the sidebar and
		// the primary view
		var sc, pc tea.Cmd
		m.layout.sidebar, sc = m.layout.sidebar.Update(msg)
		m.layout.primary, pc = m.layout.primary.Update(msg)
		cmd = tea.Batch(sc, pc)

	default:
		// Everything else, send to the primary view
		m.layout.primary, cmd = m.layout.primary.Update(msg)
	}
	return m, cmd
}

func (m *Model) View() string {
	if m.layout.fatal != nil {
		view := m.layout.fatal.View()
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, view)
		return view
	}
	view := viewport.New(m.width-theme.Padding, m.height)
	sidebar := m.layout.sidebar.View()
	primary := m.layout.primary.View()

	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, primary)
	view.SetContent(content)
	content = view.View()
	if len(m.layout.toasts) > 0 {
		lastheight := m.height
		for _, toast := range m.layout.toasts {
			if toast != nil {
				lastheight -= toast.Height + 2
				content = overlay.PlaceOverlay(1, lastheight,
					toast.View(), content, false)
			}
		}
	}
	return zone.Scan(content)
}

func (m *Model) resize(msg tea.WindowSizeMsg) tea.Cmd {
	m.height = msg.Height
	m.width = msg.Width + theme.Padding

	var sidebarWidth, sidebarHeight, primaryWidth, primaryHeight int
	sidebarWidth = max(fluxrepo.MinListWidth, int(float64(m.width)*.15)) + theme.Padding
	sidebarHeight = m.height
	primaryWidth = (m.width - sidebarWidth) - theme.Padding
	primaryHeight = m.height

	if s, ok := m.layout.sidebar.(components.Scalable); ok {
		m.layout.sidebar = s.SetSize(sidebarWidth, sidebarHeight)
	}

	if p, ok := m.layout.primary.(components.Scalable); ok {
		m.layout.primary = p.SetSize(primaryWidth, primaryHeight)
	}
	return nil
}

func (m *Model) updateKeyMsg(msg tea.KeyMsg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, m.keymap.Quit):
		cmd = tea.Quit
	case key.Matches(msg, m.keymap.Tab):
		switch m.focus {
		case sidebar:
			m.focus = primary
			m.layout.primary.(components.Focus).NextFocus()
			m.layout.sidebar.(components.Focusable).Blur()
		case primary:
			if m.layout.primary.(components.Focus).NextFocus() == yamlview.NoFocus {
				m.focus = sidebar
				m.layout.sidebar.(components.Focusable).Focus()
			}
		}
	case key.Matches(msg, m.keymap.ShiftTab):
		switch m.focus {
		case sidebar:
			m.focus = primary
			m.layout.primary.(components.Focus).PreviousFocus()
			m.layout.sidebar.(components.Focusable).Blur()
		case primary:
			if m.layout.primary.(components.Focus).PreviousFocus() == yamlview.NoFocus {
				m.focus = sidebar
				m.layout.sidebar.(components.Focusable).Focus()
			}
		}

	default:
		switch m.focus {
		case sidebar:
			m.layout.sidebar, cmd = m.layout.sidebar.Update(msg)
		case primary:
			m.layout.primary, cmd = m.layout.primary.Update(msg)
		}
	}
	return m, cmd
}
