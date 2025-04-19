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
	"bytes"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	v3 "gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/types"
)

func (m *Model) followKustomization(index int, path string, fluxKust *shortApi) {
	f, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return
	}

	dec := v3.NewDecoder(bytes.NewReader(f))

	var kustomization types.Kustomization
	for dec.Decode(&kustomization) == nil {
		for _, resource := range kustomization.Resources {
			// If the resources is a yaml file, get the real path
			// to the file to allow for relative bases, then check
			// if that file is a defined flux kustomization
			np := filepath.Join(path, resource)

			// parse out relative paths, etc...
			rp, err := filepath.Abs(np)
			if err != nil {
				log.Error("error getting absolute path", "rp", rp, "error", err)
				continue
			}

			// Is this resource pointing at a directory?
			if fi, err := os.Stat(rp); err != nil || fi.IsDir() {
				if err == nil {
					m.followKustomization(index, rp, fluxKust)
					return
				}
			}

			// is this a resource we're interested in?
			//
			// Match the kustomization that exists at this path then
			// add that to the children of fluxKust
			for j, v := range m.kustomizations {
				if v.filepath == rp {
					m.kustomizations[j].parent = &m.kustomizations[index]
					m.kustomizations[index].children = append(
						m.kustomizations[index].children, &m.kustomizations[j])
				}
			}

			// Try to map the kustomisation to a source
			for s, v := range m.sources {
				if v.filepath == rp {
					m.sources[s].parent = &m.kustomizations[index]
					m.kustomizations[index].source = &m.sources[s]
				}
			}
		}
	}
}
