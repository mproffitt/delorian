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
	"fmt"
	"slices"

	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/delorian/pkg/components"
	"github.com/mproffitt/delorian/pkg/theme"
	"github.com/muesli/reflow/wrap"
)

type LineType int

const (
	NoFocus components.FocusType = iota
	FilterFocus
	ViewportFocus
)

const (
	TypeNone LineType = iota
	Empty
	Entry
	Key
	Title
	Change
)

type ChangeType int

const (
	None ChangeType = iota
	Addition
	Deletion
)

const (
	ChangeIndicator   rune = '±'
	AdditionIndicator rune = '+'
	DeletionIndicator rune = '-'
)

const EntryIndicator = "► "

type DrawerState rune

const (
	EntryOpenIndicator   DrawerState = '⮟'
	EntryClosedIndicator DrawerState = '➤'
)

// DiffEntry represents a single drift entry
type DiffEntry struct {
	Title     string
	Kind      string
	Name      string
	Namespace string
	Changes   []DiffChange
	filter    []string
	state     DrawerState
}

func (d DiffEntry) GetKind() string {
	return d.Kind
}

func (d DiffEntry) GetName() string {
	return d.Name
}

func (d DiffEntry) GetNamespace() string {
	return d.Namespace
}

func (d DiffEntry) WithFilter(filter ...string) DiffEntry {
	d.filter = append(d.filter, filter...)
	return d
}

func (d DiffEntry) WithState(s DrawerState) DiffEntry {
	d.state = s
	return d
}

func (d DiffEntry) View(width int) string {
	d.state = EntryOpenIndicator
	changes := make([]string, 0)
	for _, change := range d.Changes {
		if !slices.Contains(d.filter, change.Key) {
			changes = append(changes, change.View(width))
		}
	}
	if len(changes) == 0 {
		d.state = EntryClosedIndicator
	}

	title := lipgloss.NewStyle().
		Foreground(theme.Colours.BrightYellow).
		Render(fmt.Sprintf("%s %s", string(d.state), d.Title))

	if d.state == EntryClosedIndicator {
		return lipgloss.NewStyle().MarginBottom(1).Render(title)
	}

	return lipgloss.NewStyle().MarginBottom(1).Render(
		lipgloss.JoinVertical(lipgloss.Left, append([]string{title}, changes...)...))
}

// DiffChange represents an individual key change
type DiffChange struct {
	Key     string
	Title   string
	Changes []ChangeSet
}

func (d DiffChange) View(width int) string {
	key := lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(theme.Colours.BrightBlue).
		Render(d.Key)
	title := lipgloss.NewStyle().
		PaddingLeft(4).
		Foreground(theme.Colours.Yellow).
		Render(d.Title)
	changes := make([]string, 0)
	for _, change := range d.Changes {
		changes = append(changes, change.View(width))
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		append([]string{key, title}, changes...)...)
}

// ChangeSet represents a changeset pair
type ChangeSet struct {
	Addition []string
	Deletion []string
}

func (c ChangeSet) View(width int) string {
	padding := 6
	width -= padding
	additionLines := make([]string, 0)
	for _, line := range c.Addition {
		if line == "" {
			continue
		}
		line = wrap.String(line, width)
		additionLines = append(additionLines, lipgloss.NewStyle().
			Foreground(theme.Colours.Green).
			PaddingLeft(padding).
			Render(line))
	}

	deletionLines := make([]string, 0)
	for _, line := range c.Deletion {
		if line == "" {
			continue
		}
		line = wrap.String(line, width)
		deletionLines = append(deletionLines, lipgloss.NewStyle().
			Foreground(theme.Colours.Red).
			PaddingLeft(padding).
			Render(line))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left, additionLines...),
		lipgloss.JoinVertical(lipgloss.Left, deletionLines...),
	)
}
