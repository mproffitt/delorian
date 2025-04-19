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

package yaml

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func Filter(input []byte, opts ...string) ([]byte, error) {
	if len(opts)%2 != 0 {
		return nil, fmt.Errorf("options must be pairs")
	}

	var pair string
	var options []string
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
	log.Debug("yaml filter", "filter", filter)
	prefs := yqlib.NewDefaultYamlPreferences()
	decoder := yqlib.NewYamlDecoder(prefs)
	encoder := yqlib.NewYamlEncoder(prefs)
	output, err := yqlib.NewStringEvaluator().
		Evaluate(filter, string(input), encoder, decoder)
	out := []byte(output)
	return out, err
}
