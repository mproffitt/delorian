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

package kustomize

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/mproffitt/delorian/pkg/yaml"
	v3 "gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const (
	Kustomization      = "kustomization"
	loadRestrictor     = types.LoadRestrictionsNone
	enableAlphaPlugins = false
)

func ExecKustomize(path string) ([]byte, error) {
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

// FilterKustomization is a convenience wrapper to filter for targetting kustomizations
func FilterKustomization(input []byte, opts ...string) ([]byte, error) {
	options := []string{
		".kind", "Kustomization",
	}
	options = append(options, opts...)
	return yaml.Filter(input, options...)
}

// FilterGitRepository is a convenience wrapper to filter for targetting GitRepository types
func FilterGitRepository(input []byte, opts ...string) ([]byte, error) {
	options := []string{
		".kind", "GitRepository",
	}
	options = append(options, opts...)
	return yaml.Filter(input, options...)
}

func GetKustomization(path string) (string, *types.Kustomization) {
	dirname := filepath.Dir(path)
	sigskustpath := filepath.Join(dirname, fmt.Sprintf("%s.%s", Kustomization, "yaml"))
	_, err := os.Stat(sigskustpath)
	if err != nil {
		sigskustpath = filepath.Join(dirname, fmt.Sprintf("%s.%s", Kustomization, "yml"))
		if _, err = os.Stat(sigskustpath); err != nil {
			return "", nil
		}
	}

	content, err := os.ReadFile(sigskustpath)
	if err != nil {
		return "", nil
	}
	var kustomization types.Kustomization
	err = v3.Unmarshal(content, &kustomization)
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
	// kustomize references this has helmV3 so lets check
	// that one for safety too
	helm, err = exec.LookPath("helmV3")
	if err == nil {
		return helm
	}
	log.Info("helm binary not found. helm will be disabled")
	return ""
}
