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
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/mproffitt/delorian/pkg/components"
	"github.com/mproffitt/delorian/pkg/kustomize"
)

func (s *shortApi) Build() tea.Cmd {
	args := []string{
		"build", "kustomization", s.GetName(),
		"-n", s.GetNamespace(),
		"--path", s.GetAbsoluteSpecPath(),
		"--kustomization-file", s.GetPath(),
		"--dry-run", "--strict-substitute",
	}
	return components.FluxExecCmd(args)
}

func (s *shortApi) Diff() tea.Cmd {
	args := []string{
		"diff", "kustomization", s.GetName(),
		"-n", s.GetNamespace(),
		"--path", s.GetAbsoluteSpecPath(),
		"--kustomization-file", s.GetPath(),
		"--strict-substitute",
		"--progress-bar=false",
	}
	return components.FluxExecCmd(args)
}

func (s *shortApi) Title() string {
	return zone.Mark(s.id, s.GetName())
}

func (s *shortApi) Description() string {
	desc := fmt.Sprintf("%s (%d)", s.GetNamespace(), len(s.children))
	return desc
}

func (s *shortApi) FilterValue() string {
	return zone.Mark(s.id, s.GetName())
}

func (s *shortApi) GetContent() string {
	options := []string{
		"metadata.name",
		s.GetName(),
	}
	if s.GetNamespace() != "" {
		options = append(options, "metadata.namespace", s.GetNamespace())
	}

	if s.ftype == Complete {
		return readFile(s.GetPath(), options...)
	}

	// We should not be seeing bases in the final view
	if s.ftype == Base {
		return ""
	}
	content, err := kustomize.ExecKustomize(filepath.Dir(s.kustomize))
	if err != nil {
		return err.Error()
	}
	content, err = kustomize.FilterKustomization(content, options...)
	if err != nil {
		return err.Error()
	}
	return string(content)
}

func (s *shortApi) GetName() string {
	return strings.TrimSpace(s.Metadata.Name)
}

func (s *shortApi) GetNamespace() string {
	if s.Metadata.Namespace == nil {
		return ""
	}
	return strings.TrimSpace(*s.Metadata.Namespace)
}

func (s *shortApi) GetPath() string {
	path, _ := filepath.Abs(filepath.Join(s.root, s.filepath))
	return path
}

func (s *shortApi) GetAbsoluteSpecPath() string {
	path := ""
	if s.Spec.Path != nil {
		path, _ = filepath.Abs(filepath.Join(s.root, *s.Spec.Path))
	}
	return path
}

func (s *shortApi) GetSource() *shortSource {
	return s.source
}

func (s *shortApi) GetSourceName() string {
	return s.Spec.Source.Name
}

func (s *shortApi) GetSourceNamespace() string {
	if s.Spec.Source.Namespace == nil {
		return s.GetNamespace()
	}
	return *s.Spec.Source.Namespace
}
