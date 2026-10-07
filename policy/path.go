// Copyright (c) Joseph Hale, 2026
// SPDX-License-Identifier: MPL-2.0

package policy

import "fmt"

func (p Policy) CheckPath(path string) (Verdict, bool) {
	refused := p.denied.pathRules.Matching(path)
	permitted := p.allowed.pathRules.Matching(path)

	switch {
	case len(refused) > 0:
		return Verdict{Deny, refusal("Path", refused[0], refused[0].Reason())}, true
	case len(permitted) > 0:
		return Verdict{Allow, fmt.Sprintf("Path is allowed: %s", permitted[0])}, true
	default:
		return Verdict{}, false
	}
}
