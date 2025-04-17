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

package filter

import (
	"math"
	"slices"
	"sort"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	zone "github.com/lrstanley/bubblezone"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/delorian/pkg/theme"
	"github.com/muesli/reflow/truncate"
)

const MaxWidth uint = 30

type Model struct {
	form        tea.Model
	formOptions [][]huh.Option[string]
	height      int
	itemWidth   uint
	options     []string
	selected    []string
	width       int
	focused     bool
	fields      []huh.Field
	values      [][]string
	zones       map[string]string
	groups      []*huh.Group
}

func unique(options []string) (uint, []string) {
	var longest uint
	mappedOptions := make(map[string]bool)
	uniqueOptions := make([]string, 0)
	for _, item := range options {
		longest = max(uint(len(item)), longest)
		if _, ok := mappedOptions[item]; !ok {
			mappedOptions[item] = true
			uniqueOptions = append(uniqueOptions, item)
		}
	}
	return min(longest, MaxWidth), uniqueOptions
}

func New(options, selected []string) *Model {
	var longest uint
	longest, options = unique(options)
	m := Model{
		formOptions: make([][]huh.Option[string], 0),
		itemWidth:   longest,
		options:     options,
		selected:    selected,
		fields:      make([]huh.Field, 0),
		zones:       map[string]string{},
		groups:      make([]*huh.Group, 0),
	}
	return &m
}

func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) Focus() {
	m.focused = true
}

func (m *Model) SetSize(w, h int) tea.Model {
	m.width = w
	m.height = h
	return m.setFilterLayout()
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) GetHeight() int {
	return m.height
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Action != tea.MouseActionRelease {
				break
			}
			for i := range m.formOptions {
				for k, v := range m.formOptions[i] {
					m.formOptions[i][k].Selected(false)
					if zone.Get(m.zones[v.Key]).InBounds(msg) {
						log.Debug("form", "key", v.Key, "zone", zone.Get(m.zones[v.Key]))
						value := m.formOptions[i][k].Value
						if slices.Contains(m.values[i], value) {
							// remove value from list
							newvalues := make([]string, 0)
							for _, v := range m.values[i] {
								if v != value {
									newvalues = append(newvalues, v)
								}
							}
							m.values[i] = newvalues
							break
						}
						m.formOptions[i][k].Selected(true)
						m.values[i] = append(m.values[i], value)
						break
					}
				}
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			for i := range m.fields {
				m.fields[i].Blur()
			}
			cmd = m.form.(*huh.Form).PrevGroup()
		case "right":
			for i := range m.fields {
				m.fields[i].Blur()
			}
			cmd = m.form.(*huh.Form).NextGroup()
		case "up", "down":
			for i := range m.fields {
				m.fields[i].Update(msg)
			}
		default:
			m.form, cmd = m.form.Update(msg)
		}
	default:
		m.form, cmd = m.form.Update(msg)
	}

	// Re-initialise the form if it's completed
	if m.form.(*huh.Form).State == huh.StateCompleted {
		cmd = tea.Batch(cmd, m.form.Init())
	}
	return m, cmd
}

func (m *Model) Values() []string {
	values := make([]string, 0)
	for i := range m.values {
		values = append(values, m.values[i]...)
	}
	return values
}

func (m *Model) View() string {
	form := m.form.View()
	if m.form.(*huh.Form).State == huh.StateCompleted || form == "" {
		m.setFilterLayout()
	}

	view := viewport.New(m.width, m.height)
	form = lipgloss.NewStyle().PaddingLeft(0).Render(m.form.View())
	view.SetContent(form)

	borderColour := theme.Colours.Black
	titleColour := theme.Colours.Black
	if m.focused {
		borderColour = theme.Colours.Blue
		titleColour = theme.Colours.BrightYellow
	}

	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(borderColour).Render(view.View())
	title := lipgloss.NewStyle().Foreground(titleColour).Render("Filters")
	return overlay.PlaceOverlay(2, 0, title, content, false)
}

func (m *Model) setFilterLayout() tea.Model {
	cols := int(math.Floor(float64(m.width) / float64(m.itemWidth)))
	var length int
	m.formOptions = make([][]huh.Option[string], cols)
	m.values = make([][]string, cols)

	sort.SliceStable(m.options, func(i, j int) bool {
		return m.options[i] < m.options[j]
	})

	for i := range cols {
		m.values[i] = make([]string, 0)
		start := (i * len(m.options) / cols)
		end := ((i + 1) * len(m.options)) / cols
		options := make([]huh.Option[string], 0)
		for _, option := range m.options[start:end] {
			if slices.Contains(m.selected, option) {
				m.values[i] = append(m.values[i], option)
			}
			key := truncate.String(option, m.itemWidth)
			uuid := uuid.NewString()[:8]
			zone := zone.Mark(uuid, key)
			m.zones[zone] = uuid
			options = append(options, huh.NewOption(zone, option))
		}
		length = max(length, len(options))
		m.formOptions[i] = options
	}

	m.height = length + theme.Padding

	m.fields = make([]huh.Field, len(m.formOptions))
	for i, group := range m.formOptions {
		m.fields[i] = huh.NewMultiSelect[string]().
			Options(group...).
			Value(&m.values[i])
		m.groups = append(m.groups, huh.NewGroup(m.fields[i]))
	}

	m.form = huh.NewForm(m.groups...).
		WithLayout(huh.LayoutColumns(cols)).
		WithShowErrors(false).
		WithHeight(m.height).
		WithShowHelp(false).WithKeyMap(keymap()).
		WithTheme(formTheme())

	m.form.Init()
	return m
}

func keymap() *huh.KeyMap {
	h := huh.NewDefaultKeyMap()
	return h
}

func formTheme() *huh.Theme {
	t := huh.ThemeBase()
	t.Focused.Base = t.Focused.Base.Border(lipgloss.HiddenBorder(), true)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(theme.Colours.Red).SetString("✕ ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(theme.Colours.Green).SetString("✓ ")
	t.Focused.MultiSelectSelector = lipgloss.NewStyle().Foreground(theme.Colours.BrightRed).SetString("> ")
	t.Blurred = t.Focused
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	return t
}
