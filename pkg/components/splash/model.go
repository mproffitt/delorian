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

package splash

import (
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const fluxLogo = `
                    ▣▣▣
                 =▣▣▣▣▣▣▣≠
              ≠▣▣▣▣▣▣▣▣▣▣▣▣▣≠
           ≠▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣≠
        ≠▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣=
      =▣▣▣▣▣▣▣▣▣▣▣▣◤◿◣◥▣▣▣▣▣▣▣▣▣▣▣▣=
   =▣▣▣▣▣▣▣▣▣▣▣▣▣▣◤◿◼◼◣◥▣▣▣▣▣▣▣▣▣▣▣▣▣▣=
=▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣◤◿◼◼◼◼◣◥▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣=
=▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣◤◿◼◼◼◼◼◼◣◥▣▣▣▣▣▣▣▣▣▣▣▣▣▣▣=
  =▣▣▣▣▣▣▣▣▣▣▣▣◤◿◼◼◼◼◼◼◼◼◣◥▣▣▣▣▣▣▣▣▣▣▣▣=
    ≈▣▣▣▣▣▣▣▣▣▣▣▣▣ ◨■■■ ▣▣▣▣▣▣▣▣▣▣▣▣=
       ≠▣▣▣▣▣▣▣▣▣▣ ◨■■■ ▣▣▣▣▣▣▣▣▣▣≠
          =▣▣▣▣▣▣▣ ◨■■■ ▣▣▣▣▣▣▣≠
             =▣▣▣▣ ◨■■■ ▣▣▣▣=
               =▣▣ ◨■■■ ▣▣=
                 ≈ ◨■■■ ≈
`

func FluxLogo(colourA, colourB string, width int) string {
	var (
		darkBlueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(colourA))
		lightBlueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(colourB))
		output          strings.Builder
		style           lipgloss.Style
		currentBlock    strings.Builder
		usedPrimaryLast = false
		backgroundRunes = []rune{'=', '≠', '≈', '◤', '◥', '▣'}
	)

	maxLen := 0
	for l := range strings.SplitSeq(fluxLogo, "\n") {
		maxLen = max(maxLen, len(l))
	}

	for _, char := range fluxLogo {
		if char == ' ' || char == '\t' {
			char = '\u00A0'
		}

		usePrimary := false
		style = darkBlueStyle
		if slices.Contains(backgroundRunes, char) {
			style = lightBlueStyle
			usePrimary = true
		}

		if usedPrimaryLast && !usePrimary || !usedPrimaryLast && usePrimary {
			usedPrimaryLast = !usedPrimaryLast
			output.WriteString(style.Render(currentBlock.String()))
			currentBlock.Reset()
		}

		currentBlock.WriteRune(char)
	}
	lines := strings.Split(output.String(), "\n")
	for i, line := range lines {
		lines[i] = lipgloss.NewStyle().Width(maxLen).Align(lipgloss.Left).Render(line)
	}
	content := lipgloss.JoinVertical(lipgloss.Center, lines...)

	return content
}

type (
	TickMsg time.Time

	Model struct {
		left             progress.Model
		msg              string
		percent          float64
		visbible         bool
		colourA, colourB string
		width            int
	}
)

func New(msg string) *Model {
	m := Model{
		msg:      msg,
		visbible: true,
		colourA:  "#3d6ddd",
		colourB:  "#c3d2f4",
	}

	m.left = progress.New(
		progress.WithScaledGradient(m.colourA, m.colourB),
		progress.WithScaledEmptyGradient(m.colourB, m.colourA),
		progress.WithoutPercentage(),
		progress.WithFillCharacters('━', '━'),
	)
	m.left.Width = 45
	return &m
}

func (m *Model) Init() tea.Cmd {
	return TickCmd()
}

func (m *Model) SetVisible(v bool) {
	m.visbible = v
}

func (m *Model) Visible() bool {
	return m.visbible
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		if !m.visbible {
			return m, nil
		}
		m.percent += 0.01
		if m.percent >= 1.0 {
			m.percent = 0.
		}
		return m, TickCmd()

	default:
		return m, nil
	}
}

func (m *Model) SetWidth(w int) *Model {
	m.width = w
	return m
}

func (m *Model) View() string {
	if !m.visbible {
		return ""
	}
	left := m.left.ViewAs(m.percent)
	msg := lipgloss.NewStyle().
		Width(m.width).Align(lipgloss.Center).
		Foreground(lipgloss.Color(m.colourA)).Render(m.msg)
	logo := FluxLogo(m.colourA, m.colourB, m.width)
	lview := viewport.New(45, 20)
	lview.SetContent(logo)
	logo = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(lview.View())
	content := lipgloss.JoinVertical(lipgloss.Center, logo, msg, left)
	return content
}

func TickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
