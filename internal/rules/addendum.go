// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package rules

import "github.com/thehale/tooluse-screener/internal/directories"

type Addendum struct {
	Only   Scope
	Reason string
}

func (a Addendum) InDir(workingDir directories.WorkingDir) Addendum {
	a.Only = a.Only.InDir(workingDir)
	return a
}
