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

package theme

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

type Orientation int

const (
	Landscape Orientation = iota
	Portrait
)

// Allows an object to switch its layout  mechanism
// if the orientation of the window changes
type Orient interface {
	SetOrientation(Orientation) tea.Model
}

type OrientationChangedMsg struct {
	Orientation Orientation
}

func OrientationChangedCmd(o Orientation) tea.Cmd {
	return func() tea.Msg {
		return OrientationChangedMsg{
			Orientation: o,
		}
	}
}

const Padding = 2

var (
	HiddenTableBorder = table.Border{
		Top:    "",
		Left:   "",
		Right:  "",
		Bottom: "",

		TopRight:    "",
		TopLeft:     "",
		BottomRight: "",
		BottomLeft:  "",

		TopJunction:    "",
		LeftJunction:   "",
		RightJunction:  "",
		BottomJunction: "",

		InnerJunction: "",
		InnerDivider:  "",
	}

	TabBorder = lipgloss.Border{
		Top:      "─",
		Bottom:   "─",
		Left:     "│",
		Right:    "│",
		TopLeft:  "╭",
		TopRight: "╮",

		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	TabActiveBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	TabGapBorder = lipgloss.Border{
		Top:         "",
		Left:        "",
		Right:       "",
		TopLeft:     "",
		TopRight:    "",
		Bottom:      "─",
		BottomLeft:  "─",
		BottomRight: "╮",
	}
)
