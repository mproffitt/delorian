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
	"github.com/charmbracelet/bubbles/list"
)

func (m *Model) newlist() *list.Model {
	list := list.New(m.Items(), m.delegates.normal, 0, 0)
	{
		list.SetShowFilter(true)
		list.SetFilteringEnabled(true)
		list.SetShowHelp(false)
		list.SetShowPagination(true)
		list.SetShowStatusBar(false)
		list.SetShowTitle(false)
	}
	return &list
}

func (m *Model) Items() []list.Item {
	items := make([]list.Item, 0)
	for _, k := range m.kustomizations {
		if k.ftype != Base {
			items = append(items, &k)
		}
	}
	return items
}
