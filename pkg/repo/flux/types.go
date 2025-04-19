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
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/delorian/pkg/kustomize"
)

type FluxFileType uint

const (
	Base FluxFileType = iota
	Patch
	Complete
)

// cluster is for building a tree of how clusters fit together in the repo
type cluster struct {
	name     string
	filepath string
	children []*cluster
	selected bool
}

// shortApi is a generic for capturing just enough
// information out of a yaml doc to reresent a
// kustomization or git repository resource
type shortApi struct {
	ApiVersion string    `yaml:"apiVersion"`
	Kind       string    `yaml:"kind"`
	Metadata   shortMeta `yaml:"metadata"`
	Spec       shortSpec `yaml:"spec"`

	id        string
	children  []*shortApi
	filepath  string
	ftype     FluxFileType
	kustomize string
	parent    *shortApi
	source    *shortSource
	root      string
}

// shortMeta contains only the relevant information
// from metadata to distinctly identify a kustomization
type shortMeta struct {
	Name      string  `yaml:"name"`
	Namespace *string `yaml:"namespace,omitempty"`
}

// shortSpec is used by the kustomization type to ensure
// enough information is gathered to allow identification
// of flux kustomizations without requiring the full
// object to be loaded
type shortSpec struct {
	Path      *string      `yaml:"path,omitempty"`
	Source    *shortSource `yaml:"sourceRef,omitempty"`
	PostBuild *postBuild   `yaml:"postBuild,omitempty"`
}

// postBuild contains relevant substitutions.
//
// Note: with this, we ignore ConfigMap and Secret
// substitutions as they require accessing the cluster
// and that would seriously impact loading performance
type postBuild struct {
	Substitute map[string]string `yaml:"substitute,omitempty"`
}

// shortSource is just enough information to distinctly
// identify a gitrepository resource type
type shortSource struct {
	shortMeta `yaml:",inline"`
	Kind      string `yaml:"kind"`

	children []*shortApi
	filepath string
	id       string
	parent   *shortApi
}

// GetName gets the name of the source
func (s *shortSource) GetName() string {
	return s.Name
}

// GetNamespace gets the namespace for the source
// if namespace is nil, this returns the empty string
func (s *shortSource) GetNamespace() string {
	if s.Namespace == nil {
		return ""
	}
	return *s.Namespace
}

// GetContent for source this only reads the
// details from the file
func (s *shortSource) GetContent() string {
	return readFile(s.filepath)
}

// GetPath gets the filepath for the source
func (s *shortSource) GetPath() string {
	return s.filepath
}

// ModelReadyMsg is sent when the model is loaded
type ModelReadyMsg struct {
	Ready bool
}

// ModelReadyCmd is returned by the loading process when
// the model traversal is complete
func ModelReadyCmd(ready bool) tea.Cmd {
	return func() tea.Msg {
		return ModelReadyMsg{Ready: ready}
	}
}

func readFile(filename string, filterOpts ...string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err.Error()
	}
	if len(filterOpts) == 0 {
		return string(content)
	}
	nc, err := kustomize.FilterKustomization(content, filterOpts...)
	if err != nil {
		return err.Error()
	}
	return string(nc)
}
