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
	"github.com/charmbracelet/lipgloss"
	bmx "github.com/mproffitt/bmx/pkg/theme"
)

var Colours ColourStyles

type ColourStyles struct {
	Fg           lipgloss.AdaptiveColor
	Bg           lipgloss.AdaptiveColor
	SelectionBg  lipgloss.AdaptiveColor
	Cursor       lipgloss.AdaptiveColor
	BrightBlack  lipgloss.AdaptiveColor
	BrightBlue   lipgloss.AdaptiveColor
	BrightCyan   lipgloss.AdaptiveColor
	BrightGreen  lipgloss.AdaptiveColor
	BrightPurple lipgloss.AdaptiveColor
	BrightRed    lipgloss.AdaptiveColor
	BrightWhite  lipgloss.AdaptiveColor
	BrightYellow lipgloss.AdaptiveColor
	Black        lipgloss.AdaptiveColor
	Blue         lipgloss.AdaptiveColor
	Cyan         lipgloss.AdaptiveColor
	Green        lipgloss.AdaptiveColor
	Purple       lipgloss.AdaptiveColor
	Red          lipgloss.AdaptiveColor
	White        lipgloss.AdaptiveColor
	Yellow       lipgloss.AdaptiveColor
}

func init() {
	Colours = ColourStyles{
		Fg:           lipgloss.AdaptiveColor{Dark: "#a9b1d6", Light: "#343b58"}, // Editor Foreground
		Bg:           lipgloss.AdaptiveColor{Dark: "#1a1b26", Light: "#e6e7ed"}, // Editor background
		SelectionBg:  lipgloss.AdaptiveColor{Dark: "#545c7e", Light: "#707280"}, // Focus Border
		Cursor:       lipgloss.AdaptiveColor{Dark: "#c0caf5", Light: "#343b58"}, // Terminal white
		BrightBlack:  lipgloss.AdaptiveColor{Dark: "#565f89", Light: "#6c6e75"}, // Comments
		BrightBlue:   lipgloss.AdaptiveColor{Dark: "#2ac3de", Light: "#2959aa"}, // Function names
		BrightCyan:   lipgloss.AdaptiveColor{Dark: "#b4f9f8", Light: "#33635c"}, // Regex Literal strings
		BrightGreen:  lipgloss.AdaptiveColor{Dark: "#9ece6a", Light: "#385f0d"}, // Strings, ClassNames
		BrightPurple: lipgloss.AdaptiveColor{Dark: "#bb9af7", Light: "#7b43ba"}, // Terminal Magenta
		BrightRed:    lipgloss.AdaptiveColor{Dark: "#db4b4b", Light: "#942f2f"}, // Error foreground
		BrightWhite:  lipgloss.AdaptiveColor{Dark: "#cfc9c2", Light: "#634f30"}, // Semantic Highlight
		BrightYellow: lipgloss.AdaptiveColor{Dark: "#ff9e64", Light: "#965027"}, // Constants
		Black:        lipgloss.AdaptiveColor{Dark: "#414868", Light: "#343B58"}, // Terminal Black
		Blue:         lipgloss.AdaptiveColor{Dark: "#7aa2f7", Light: "#2959aa"}, // Terminal Blue
		Cyan:         lipgloss.AdaptiveColor{Dark: "#7dcfff", Light: "#0f4b6e"}, // Terminal Cyan
		Green:        lipgloss.AdaptiveColor{Dark: "#73daca", Light: "#33635c"}, // Terminal Green
		Purple:       lipgloss.AdaptiveColor{Dark: "#9d7cd8", Light: "#5a3e8e"}, // Charts Purple
		Red:          lipgloss.AdaptiveColor{Dark: "#f7768e", Light: "#8c4351"}, // Terminal Red
		White:        lipgloss.AdaptiveColor{Dark: "#c0caf5", Light: "#343b58"}, // Terminal white
		Yellow:       lipgloss.AdaptiveColor{Dark: "#e0af68", Light: "#8f5e15"}, // Terminal Yellow
	}
	bmx.Colours = bmx.ColourStyles(Colours)
}
