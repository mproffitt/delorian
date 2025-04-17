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

package flux

import (
	"path/filepath"
	"strings"
	"sync"

	"github.com/charlievieth/fastwalk"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/evertras/bubble-table/table"
	zone "github.com/lrstanley/bubblezone"
	"github.com/mproffitt/delorian/pkg/components"
	"github.com/mproffitt/delorian/pkg/components/treeview"
)

const MinListWidth = 26

type Model struct {
	sync.Mutex
	id             string
	conf           fastwalk.Config
	clusters       []*cluster
	delegates      delegates
	height         int
	kustomizations []shortApi
	lasttab        components.TabType
	list           *list.Model
	table          *table.Model
	root           string
	sources        []shortSource
	width          int
	focus          bool

	treeview tea.Model
}

type delegates struct {
	normal list.ItemDelegate
	shaded list.ItemDelegate
}

func New(root string) *Model {
	root = strings.TrimRight(root, string(filepath.Separator))
	m := Model{
		id: zone.NewPrefix(),
		conf: fastwalk.Config{
			Follow: true,
		},
		lasttab:        components.TabKustomize,
		root:           root,
		kustomizations: make([]shortApi, 0),
		sources:        make([]shortSource, 0),
	}
	m.delegates = delegates{
		normal: m.createListNormalDelegate(),
		shaded: m.createListShadedDelegate(),
	}

	return &m
}

func (m *Model) Focus() {
	m.focus = true
	m.list.SetDelegate(m.delegates.normal)
}

func (m *Model) Blur() {
	m.focus = false
	m.list.SetDelegate(m.delegates.shaded)
}

func (m *Model) Init() tea.Cmd {
	cmd := m.walk()

	var clusters []treeview.Tree
	{
		for i := range m.clusters {
			clusters = append(clusters, m.clusters[i])
			log.Debug("Adding cluster", "cluster", m.clusters[i].Name())
		}
	}
	m.treeview = treeview.New("clusters", clusters, m.width, m.height)
	return cmd
}

func (m *Model) SetSize(w, h int) tea.Model {
	m.height = h
	m.width = w
	if m.treeview != nil {
		m.treeview = m.treeview.(components.Scalable).SetSize(w+1, h)
	}
	return m
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.list.CursorUp()
			cmd = m.defaultHandler(msg)
		case tea.MouseButtonWheelDown:
			m.list.CursorDown()
			cmd = m.defaultHandler(msg)
		case tea.MouseButtonLeft:
			log.Debug("Mouse", "left", msg)
			if msg.Action != tea.MouseActionRelease {
				break
			}
			for i, listItem := range m.list.VisibleItems() {
				v, _ := listItem.(*shortApi)
				if zone.Get(v.id) == nil {
					continue
				}
				log.Debug("zone", "get", zone.Get(v.id))
				if zone.Get(v.id).InBounds(msg) {
					log.Debug("Mouse", "listitem", listItem)
					m.list.Select(i)
					cmd = m.defaultHandler(msg)
					break
				}
			}
		}
	case ModelReadyMsg:
		/*if !msg.Ready {
			break
		}*/
		m.table = nil
		m.list = m.newlist()
		m.list.SetItems(m.Items())
		api, ok := m.FindSelected()
		cmd = components.FileCmd(api, ok)
	case components.TabChangedMsg:
		m.lasttab = msg.NewTab
		api, ok := m.FindSelected()
		if ok {
			switch m.lasttab {
			case components.TabFluxBuild:
				cmd = api.(components.Flux).Build()
			case components.TabFluxDiff:
				cmd = api.(components.Flux).Diff()
			case components.TabGraph:
			default:
				cmd = components.FileCmd(api, ok)
			}
		}
	default:
		cmd = m.defaultHandler(msg)
	}
	return m, cmd
}

func (m *Model) defaultHandler(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	var list list.Model
	list, cmd = m.list.Update(msg)
	list.SetDelegate(m.delegates.normal)
	m.list = &list
	api, ok := m.FindSelected()
	var fcmd tea.Cmd
	if ok {
		switch m.lasttab {
		case components.TabFluxBuild:
			fcmd = api.(components.Flux).Build()
		case components.TabFluxDiff:
			fcmd = api.(components.Flux).Diff()
		case components.TabGraph:
		default:
			fcmd = components.FileCmd(api, ok)
		}
	}
	cmd = tea.Batch(cmd, fcmd)
	return cmd
}

func (m *Model) FindSelected() (api components.File, ok bool) {
	var path, name string
	item := m.list.SelectedItem().(*shortApi)
	path = item.GetPath()
	name = item.GetName()
	for i, v := range m.kustomizations {
		if v.GetPath() == path && v.GetName() == name {
			a := &m.kustomizations[i]
			api = a
			ok = a != nil
			break
		}
	}

	switch m.lasttab {
	case components.TabSource:
		if api != nil {
			a := api.(*shortApi).GetSource()
			ok = a != nil
			api = a
		}
	}
	return
}

func (m *Model) View() string {
	treeviewHeight := len(m.clusters) + 3
	for _, child := range m.clusters {
		treeviewHeight += child.Len()
	}

	var content string
	if m.list == nil {
		return ""
	}
	m.list.SetWidth(m.width)
	m.list.SetHeight(m.height - treeviewHeight)
	m.treeview = m.treeview.(components.Scalable).SetSize(m.width, treeviewHeight)
	tree := m.treeview.View()
	content = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - treeviewHeight).
		Render(m.list.View())
	content = lipgloss.JoinVertical(lipgloss.Left, content, tree)
	return content
}
