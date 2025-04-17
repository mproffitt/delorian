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
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss/tree"
	"github.com/charmbracelet/log"
)

var commonNamespaces = []string{
	"flux-system", "default",
}

func (c *cluster) Add(entries []string, path string) *cluster {
	switch len(entries) {
	case 0:
		return nil
	default:
		if entries[0] == c.name {
			if len(entries) > 1 {
				entries = entries[1:]
				for _, child := range c.children {
					if child.name == entries[0] {
						return child.Add(entries, path)
					}
				}
				child := &cluster{
					name:     entries[0],
					filepath: path,
					children: make([]*cluster, 0),
				}
				log.Debug("Adding child", "cluster", entries[0], "parent", c.name, "path", path)
				c.children = append(c.children, child)
				return child
			}
		}
	}
	return nil
}

func (c *cluster) Tree() *tree.Tree {
	tree := tree.New().
		Root(c.Name())
	sort.SliceStable(c.children, func(i, j int) bool {
		return c.children[i].name < c.children[j].name
	})
	for i, v := range c.children {
		if len(v.children) > 0 {
			tree = tree.Child(c.children[i].Tree())
		} else {
			tree = tree.Child(c.children[i].Name())
		}
	}
	return tree
}

func (c *cluster) Len() int {
	l := len(c.children)
	for _, child := range c.children {
		l += child.Len()
	}
	return l
}

func (c *cluster) Matches(entry string) bool {
	dirs := strings.Split(c.filepath, string(filepath.Separator))
	return entry == dirs[len(dirs)-1]
}

func (c *cluster) Name() string {
	return c.name
}

func (c *cluster) Select(branch []string) {
	switch len(branch) {
	case 0:
		return
	default:
		switch branch[0] {
		case c.name:
			c.selected = true
			if len(branch) > 1 {
				branch = branch[1:]
			}
			for i := range c.children {
				c.children[i].Select(branch)
			}
		}
	}
}

func (c *cluster) Selected() bool {
	return c.selected
}

func (m *Model) checkClusterPath(path string) {
	// We should have already tested that this is a valid
	// location so no need to try again, just validate the
	// path and update clusters, then move on.
	path = strings.TrimRight(path, string(filepath.Separator))
	if strings.Contains(path, "/.") || strings.Contains(path, "bases/") {
		// ignore hidden paths and bases
		return
	}
	testPath := strings.TrimPrefix(path, m.root+string(filepath.Separator))
	// We accept any of
	// *clusters
	// *hub
	// as being valid cluster directory names
	//
	// This is to avoid being too opinionated about the directory structure
	// as different people have different patterns they may adhere too.
	//
	// We do have to be somewhat opinionated though ...
	re := regexp.MustCompile(`(?:[^/]*(clusters|hub))/([^/]+)`)
	matches := re.FindAllStringSubmatch(testPath, -1)
	var clusters []string
	for _, match := range matches {
		if len(match) > 2 {
			name := match[2]
			if slices.Contains(commonNamespaces, name) {
				name = match[1]
				path = strings.TrimSuffix(path, match[2])
			}
			clusters = append(clusters, name)
		}
	}

	if len(clusters) == 0 {
		return
	}
	log.Debug("matched clusters", "clusters", clusters)
	foundParent := false
	m.Lock()
	for i, c := range m.clusters {
		if c.name == clusters[0] {
			foundParent = true
			m.clusters[i].Add(clusters, path)
		}
	}
	if !foundParent {
		newCluster := cluster{
			children: make([]*cluster, 0),
			name:     clusters[0],
			filepath: path,
		}
		log.Debug("Adding cluster", "clusterName", clusters[0], "parent", nil, "filepath", path)
		m.clusters = append(m.clusters, &newCluster)
	}
	m.Unlock()
}

// Walks through the list of clusters and checks to see if any need
// to be moved to become a child of another
//
// This is achieved by checking for a file called <clustername>.yaml
// in the root of the clusters tree
func (m *Model) reparentClusters() {
	for i := range m.clusters {
		if m.clusters[i] == nil {
			continue
		}

		for j := range m.clusters {
			if j == i || m.clusters[j] == nil {
				continue
			}
			fname := filepath.Join(m.clusters[i].filepath, m.clusters[j].name) + ".yaml"
			log.Debug("checking", "fname", fname)
			if _, err := os.Stat(fname); err == nil {
				c := cluster{
					children: make([]*cluster, len(m.clusters[j].children)),
					name:     m.clusters[j].name,
					filepath: m.clusters[j].filepath,
				}
				c.children = append(c.children, m.clusters[j].children...)
				m.clusters[i].children = append(m.clusters[i].children, &c)
				m.clusters[j] = nil
			}
		}
	}

	// recreate the clusters list to ditch nil entries
	newclusters := make([]*cluster, 0)
	for _, v := range m.clusters {
		v := v
		if v != nil {
			newclusters = append(newclusters, v)
		}
	}
	sort.SliceStable(newclusters, func(i, j int) bool {
		return newclusters[i].name < newclusters[j].name
	})

	m.clusters = newclusters
}
