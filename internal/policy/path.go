// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"fmt"

	"github.com/thehale/tooluse-screener/internal/rules"
)

func (p Policy) CheckPath(path string) (Verdict, bool) {
	refused := globsMatching(p.Denied.PathRules, path)
	permitted := globsMatching(p.Allowed.PathRules, path)

	switch {
	case len(refused) > 0:
		return Verdict{Deny, refusal("Path", refused[0], refused[0].Reason())}, true
	case len(permitted) > 0:
		return Verdict{Allow, fmt.Sprintf("Path is allowed: %s", permitted[0])}, true
	default:
		return Verdict{}, false
	}
}

func globsMatching(globs []rules.Glob, path string) []rules.Glob {
	var matched []rules.Glob
	for _, glob := range globs {
		if glob.Matches(path) {
			matched = append(matched, glob)
		}
	}
	return matched
}
