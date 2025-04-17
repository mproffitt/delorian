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
	"bufio"
	"strings"
)

// ParseFluxDiff parses the flux diff into structured data
//
// This is basically a lexer for flux diff output
func (m *Model) parseFluxDiff(input string) []DiffEntry {
	scanner := bufio.NewScanner(strings.NewReader(input))
	var (
		results        []DiffEntry
		currentEntry   *DiffEntry
		currentChange  *DiffChange
		lastChange     *ChangeSet
		lastType       LineType
		lastChangeType ChangeType
		expected       rune
	)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if len(trimmed) == 0 {
			switch lastType {
			case Change:
				if currentChange != nil {
					if lastChange != nil {
						currentChange.Changes = append(currentChange.Changes, *lastChange)
						lastChange = nil
					}

					currentEntry.Changes = append(currentEntry.Changes, *currentChange)
					currentChange = nil
				}
				lastChangeType = None
				lastType = Empty
			}
			continue
		}

		// Detect new entry
		if strings.HasPrefix(line, EntryIndicator) {
			if currentEntry != nil {
				results = append(results, *currentEntry)
			}
			title := strings.TrimPrefix(line, EntryIndicator)
			parts := strings.Split(strings.TrimSuffix(title, " drifted"), "/")
			currentEntry = &DiffEntry{
				Title:     strings.TrimSpace(title),
				Kind:      parts[0],
				Name:      parts[2],
				Namespace: parts[1],
				Changes:   []DiffChange{},
				state:     EntryOpenIndicator,
			}
			lastType = Entry
			continue
		}

		switch lastType {
		case Entry, Empty:
			lastType = Key
			currentChange = &DiffChange{
				Key: trimmed,
			}
		case Key:
			lastType = Title
			currentChange.Title = trimmed
			expected = []rune(trimmed)[0]
		case Title:
			lastType = Change
			lastChange = &ChangeSet{}
			// Last type was title so we're now into the change
			// fallthrough to parse the first line of the change
			fallthrough
		case Change:
			c := []rune(trimmed)[0]
			switch expected {
			case ChangeIndicator:
				switch c {
				case AdditionIndicator:
					lastChange.Addition = append(lastChange.Addition, trimmed)
					lastChangeType = Addition
				case DeletionIndicator:
					if lastChangeType == Addition {
						currentChange.Changes = append(currentChange.Changes, *lastChange)
						lastChange = &ChangeSet{}
					}

					lastChange.Deletion = append(lastChange.Deletion, trimmed)
					lastChangeType = Deletion
				}
			case AdditionIndicator:
				lastChange.Addition = append(lastChange.Addition, trimmed)
			case DeletionIndicator:
				lastChange.Deletion = append(lastChange.Deletion, trimmed)
			}
		}
	}

	if currentEntry != nil {
		if currentChange != nil {
			if lastChange != nil {
				currentChange.Changes = append(currentChange.Changes, *lastChange)
			}
			currentEntry.Changes = append(currentEntry.Changes, *currentChange)
		}
		results = append(results, *currentEntry)
	}

	return results
}
