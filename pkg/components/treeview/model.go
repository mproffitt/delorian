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

package treeview

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/mproffitt/delorian/pkg/theme"
)

type Tree interface {
	Tree() *tree.Tree
	Matches(string) bool
	Select([]string)
}

type Model struct {
	branches []Tree
	height   int
	styles   styles
	title    string
	viewport viewport.Model
	width    int
}

type styles struct {
	enumerator lipgloss.Style
	root       lipgloss.Style
	item       lipgloss.Style
	selected   lipgloss.Style
}

func New(title string, t []Tree, w, h int) *Model {
	m := Model{
		branches: t,
		height:   h,
		styles: styles{
			enumerator: lipgloss.NewStyle().Foreground(theme.Colours.Black),
			root:       lipgloss.NewStyle().Foreground(theme.Colours.BrightBlack),
			item:       lipgloss.NewStyle().Foreground(theme.Colours.Purple),
			selected:   lipgloss.NewStyle().Foreground(theme.Colours.Fg),
		},
		title:    title,
		viewport: viewport.New(w, h),
		width:    w,
	}
	return &m

	/*
			 Need: - tree.EnumeratorStyleFunc for walking the tree
			         and highlighting selected items
			       - To know  How to index the entire tree so the
			         correct item is highlighted
		           - Left / Right = collapse / expand branch
		           - Enter logs in to current branch
	*/
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetSize(w, h int) tea.Model {
	m.width = w
	m.height = h
	m.viewport.Width = w //- (2 * theme.Padding)
	m.viewport.Height = h
	return m
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
		case "down":
		case "left":
		case "right":
		case "enter":
		}
	}
	return m, nil
}

func (m *Model) View() string {
	m.viewport.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, false, true).
		BorderForeground(theme.Colours.Black)

	tree := m.renderTree()
	m.viewport.SetContent(tree)
	return m.viewport.View()
}

func (m *Model) renderTree() string {
	if len(m.branches) == 0 {
		text := lipgloss.NewStyle().
			Foreground(theme.Colours.Cyan).Render("No " + m.title)
		text = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, text)
		return text
	}
	tree := tree.New().Root(m.title).
		Enumerator(tree.RoundedEnumerator).
		EnumeratorStyle(m.styles.enumerator).
		RootStyle(m.styles.root).
		ItemStyle(m.styles.item)

	for i := range m.branches {
		tree = tree.Child(m.branches[i].Tree())
	}

	return tree.String()
}
