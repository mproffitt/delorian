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

/*
* This is the most complicated part of the application
*
* First, this must walk the entire tree and gather all Flux kustomizations and
* Kubernetes sigs kustomizations, then match Flux kustomization resources to those
* inside the sigs kustomizations
*
* - Is Flux kustomization
*   - Is this kustomization referenced as a patch
*   - Is this kustomization a likely base?
 */

import (
	"bytes"
	"cmp"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charlievieth/fastwalk"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/mproffitt/delorian/pkg/components"
	"github.com/mproffitt/delorian/pkg/kustomize"
	"golang.org/x/exp/slices"
	yaml "gopkg.in/yaml.v3"
)

const (
	kustomizationApi = "kustomize.toolkit.fluxcd.io"
	sourceApi        = "source.toolkit.fluxcd.io"
)

func (m *Model) walk() tea.Cmd {
	/*
	 * First, gather every single flux kustomization irrespective of whether
	 * this is a base or not. It will be filtered later
	 */
	rootFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fi, err := os.Stat(path)
		if err != nil || fi.IsDir() {
			m.checkClusterPath(path)
			return err
		}

		filetypes := []string{".yaml", ".yml"}
		ext := filepath.Ext(d.Name())
		if !slices.Contains(filetypes, strings.ToLower(ext)) {
			return nil
		}

		// Collect any kustomizations or sources stored in this file
		k, s := parseYamlFromFile(m.root, path)
		m.Lock()
		m.kustomizations = append(m.kustomizations, k...)
		m.sources = append(m.sources, s...)
		m.Unlock()
		return err
	}

	// Load all kustomizations and sources first from the repo
	if err := fastwalk.Walk(&m.conf, m.root, rootFn); err != nil {
		return components.ModelErrorCmd(err)
	}

	if len(m.kustomizations) == 0 {
		err := fmt.Errorf("no kustomizations found\nare you sure this is a flux repository?")
		return components.ModelFatalCmd(err)
	}

	// Now we have all kustomizations in the repo, we can start to organise them
	//
	// Ones that are used as bases will be ignored for now but those that are
	// merged from bases and patches will be kept as the final rendered value
	var cmds []tea.Cmd
	ready := true
	for i := range m.kustomizations {
		m.kustomizations[i].children = make([]*shortApi, 0)
		err := m.followFluxKustomization(i, &m.kustomizations[i])
		if err != nil {
			cmds = append(cmds, components.ModelErrorCmd(err))
			ready = false
		}
		m.setSource(i)
	}

	m.reparentClusters()

	slices.SortStableFunc(m.kustomizations, func(a, b shortApi) int {
		if len(a.children) == len(b.children) {
			return strings.Compare(a.GetName(), b.GetName())
		}
		return cmp.Compare(len(b.children), len(a.children))
	})

	cmds = append(cmds, ModelReadyCmd(ready))
	return tea.Batch(cmds...)
}

// This function is for walking the kustomization path and
// detecting which kustomization, and git repository kustomizations
// should be part of
func (m *Model) followFluxKustomization(index int, fluxKust *shortApi) error {
	log.Debug("walking", "path", fluxKust.filepath)
	path := fluxKust.filepath
	if !strings.HasPrefix(path, m.root) {
		path = filepath.Join(m.root, path)
	}
	fp, kust := kustomize.GetKustomization(path)
	fluxKust.kustomize = fp
	if kust == nil || slices.Contains(kust.Resources, filepath.Base(path)) {
		fluxKust.ftype = Complete
	} else {
		for _, p := range kust.Patches {
			if p.Path == filepath.Base(path) {
				fluxKust.ftype = Patch
			}
		}
	}

	pathFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// parse directory with kustomization
		filename := d.Name()
		filename = filename[0 : len(filename)-len(filepath.Ext(filename))]
		if filename == kustomize.Kustomization {
			m.followKustomization(index, path, fluxKust)
			return nil
		}

		// parse non-kust directory
		switch {
		case d.Type().IsRegular():
			for i := range m.kustomizations {
				// Match the kustomization at this path. This then becomes a child of fluxKust
				if path == m.kustomizations[i].filepath {
					log.Debug("Matching", "path", path, "kust", *fluxKust.Spec.Path)
					(*fluxKust).children = append((*fluxKust).children, &m.kustomizations[i])
					m.kustomizations[i].parent = fluxKust

					if fluxKust.Spec.PostBuild != nil {
						m.kustomizations[i].Metadata.Name = m.ParseSubstitutions(
							m.kustomizations[i].Metadata.Name,
							fluxKust.Spec.PostBuild.Substitute)
						*m.kustomizations[i].Spec.Path = m.ParseSubstitutions(
							filepath.Join(m.root, *m.kustomizations[i].Spec.Path),
							fluxKust.Spec.PostBuild.Substitute)
					}
					return nil
				}
			}
			for s, v := range m.sources {
				if v.filepath == path {
					m.sources[s].parent = &m.kustomizations[index]
				}
			}
		}
		return nil
	}

	if fluxKust.Spec.Path == nil {
		return nil
	}

	kpath := fluxKust.GetAbsoluteSpecPath()
	return fastwalk.Walk(&m.conf, kpath, pathFn)
}

func (m *Model) setSource(index int) {
	for s := range m.sources {
		if m.kustomizations[index].Spec.Source == nil {
			return
		}
		if m.kustomizations[index].Spec.Source.Kind == m.sources[s].Kind {
			var (
				kName      = m.kustomizations[index].GetSourceName()
				kNamespace = m.kustomizations[index].GetSourceNamespace()
				sName      = m.sources[s].GetName()
				sNamespace = m.sources[s].GetNamespace()
			)

			log.Debug("checking source", "kName", kName, "kNamespace",
				kNamespace, "sName", sName, "sNamespace", sNamespace)
			if kName == sName && kNamespace == sNamespace {
				var has bool
				for _, c := range m.sources[s].children {
					// block duplication
					if c.GetName() == kName && c.GetNamespace() == kNamespace {
						has = true
					}
				}
				if !has {
					m.sources[s].children = append(m.sources[s].children, &m.kustomizations[index])
					m.kustomizations[index].source = &m.sources[s]
				}
			}
		}
	}
}

func (m *Model) ParseSubstitutions(where string, substitutions map[string]string) string {
	for k, v := range substitutions {
		replace := fmt.Sprintf("${%s}", k)
		where = strings.ReplaceAll(where, replace, v)
	}
	return where
}

func parseYamlFromFile(root, path string) (kustomizations []shortApi, sources []shortSource) {
	kustomizations = make([]shortApi, 0)
	sources = make([]shortSource, 0)
	f, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return
	}
	return parseYaml(f, root, path)
}

func parseYaml(input []byte, root, path string) (kustomizations []shortApi, sources []shortSource) {
	dec := yaml.NewDecoder(bytes.NewReader(input))

	var doc shortApi
	for dec.Decode(&doc) == nil {
		api := strings.Split(doc.ApiVersion, "/")[0]
		switch api {
		case kustomizationApi:
			if doc.Spec.Source != nil && doc.Spec.Source.Namespace == nil {
				doc.Spec.Source.Namespace = doc.Metadata.Namespace
			}
			doc.id = uuid.NewString()[:8]
			doc.root = root
			doc.filepath = strings.TrimPrefix(path, root+string(filepath.Separator))
			log.Debug("ROOT STRING", "root", root, "filepath", doc.filepath)
			// Everything starts out as a base until determined otherwise
			doc.ftype = Base
			kustomizations = append(kustomizations, doc)
		case sourceApi:
			source := shortSource{
				id:   uuid.NewString()[:8],
				Kind: doc.Kind,
				shortMeta: shortMeta{
					Name:      doc.Metadata.Name,
					Namespace: doc.Metadata.Namespace,
				},
				filepath: path,
			}
			sources = append(sources, source)
		}
	}
	return
}
