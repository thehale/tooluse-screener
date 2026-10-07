// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import "fmt"

func (p Policy) CheckPath(path string) Verdict {
	denyingGlobs := p.denied.pathRules.GlobsMatching(path)
	allowingGlobs := p.allowed.pathRules.GlobsMatching(path)

	switch {
	case len(denyingGlobs) > 0:
		return Verdict{Deny, refusal("Path", denyingGlobs[0], denyingGlobs[0].Reason())}
	case len(allowingGlobs) > 0:
		return Verdict{Allow, fmt.Sprintf("Path is allowed: %s", allowingGlobs[0])}
	default:
		return unlistedPath
	}
}

var unlistedPath = Verdict{Ask, "Path is not in the shared allow list"}
