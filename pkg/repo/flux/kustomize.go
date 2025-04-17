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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	yaml "gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	loadRestrictor     = types.LoadRestrictionsNone
	enableAlphaPlugins = false
)

func (m *Model) followKustomization(index int, path string, fluxKust *shortApi) {
	f, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return
	}

	dec := yaml.NewDecoder(bytes.NewReader(f))

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

			for s, v := range m.sources {
				if v.filepath == rp {
					m.sources[s].parent = &m.kustomizations[index]
					m.kustomizations[index].source = &m.sources[s]
				}
			}
		}
	}
}

func execKustomize(path string) ([]byte, error) {
	helm := findHelm()
	// Kustomize prints deprecation warnings to Stderr that are
	// not trapped by bubbletea and interfere with the UI.
	//
	// To overcome this, we redirect all Stderr to /dev/null as
	// these messages are not relevant for what we're doing
	o := os.Stderr
	devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	defer func() {
		_ = devNull.Close()
		os.Stderr = o
	}()
	os.Stderr = devNull
	options := krusty.Options{
		Reorder:           krusty.ReorderOptionNone,
		AddManagedbyLabel: false,
		LoadRestrictions:  loadRestrictor,

		PluginConfig: &types.PluginConfig{
			PluginRestrictions: types.PluginRestrictionsBuiltinsOnly,
			BpLoadingOptions:   types.BploUseStaticallyLinked,
			FnpLoadingOptions: types.FnPluginLoadingOptions{ // These are the defaults from the flags to kustomize
				EnableExec:    false,
				Network:       false,
				NetworkName:   "bridge",
				Mounts:        []string{},
				AsCurrentUser: false,
			},
			// Helm is enabled only if it's found in path
			HelmConfig: types.HelmConfig{
				Enabled: helm != "",
				Command: helm,
			},
		},
	}
	fsys := filesys.MakeFsOnDisk()
	k := krusty.MakeKustomizer(&options)
	m, err := k.Run(fsys, path)
	if err != nil {
		return nil, err
	}
	return m.AsYaml()
}

func filterKustomization(input []byte, opts ...string) ([]byte, error) {
	if len(opts)%2 != 0 {
		return nil, fmt.Errorf("options must be pairs")
	}
	options := []string{
		// fmt.Sprintf(`.apiVersion == "%s"`, kustomizationApi),
		`.kind == "Kustomization"`,
	}
	var pair string
	for i, v := range opts {
		if i%2 == 0 {
			if v[0] != '.' {
				v = "." + v
			}
			pair = fmt.Sprintf("%s ==", v)
		} else {
			pair = fmt.Sprintf(`%s "%s"`, pair, v)
			options = append(options, pair)
		}
	}
	filter := fmt.Sprintf(`select(%s)`, strings.Join(options, " and "))
	log.Debug("kustomization filter", "filter", filter)
	prefs := yqlib.NewDefaultYamlPreferences()
	decoder := yqlib.NewYamlDecoder(prefs)
	encoder := yqlib.NewYamlEncoder(prefs)
	output, err := yqlib.NewStringEvaluator().
		Evaluate(filter, string(input), encoder, decoder)
	out := []byte(output)
	return out, err
}

func getKustomization(path string) (string, *types.Kustomization) {
	dirname := filepath.Dir(path)
	sigskustpath := filepath.Join(dirname, fmt.Sprintf("%s.%s", kustomization, "yaml"))
	_, err := os.Stat(sigskustpath)
	if err != nil {
		sigskustpath = filepath.Join(dirname, fmt.Sprintf("%s.%s", kustomization, "yml"))
		if _, err = os.Stat(sigskustpath); err != nil {
			return "", nil
		}
	}

	content, err := os.ReadFile(sigskustpath)
	if err != nil {
		return "", nil
	}
	var kustomization types.Kustomization
	err = yaml.Unmarshal(content, &kustomization)
	if err != nil {
		return "", nil
	}

	return sigskustpath, &kustomization
}

func findHelm() string {
	helm, err := exec.LookPath("helm")
	if err == nil {
		return helm
	}
	helm, err = exec.LookPath("helmV3")
	if err == nil {
		return helm
	}
	return ""
}
