// Copyright (c) 2026 Joseph Hale
// SPDX-License-Identifier: MPL-2.0

package policy

import "github.com/thehale/tooluse-screener/internal/directories"

func (p Policy) InDir(directory string) Policy {
	expandedPath, _ := directories.WithHomeExpanded(directory)
	workingDir := directories.WorkingDir(expandedPath)
	return Policy{denied: p.denied.inDir(workingDir), allowed: p.allowed.inDir(workingDir)}
}

func (s side) inDir(workingDir directories.WorkingDir) side {
	return side{commandRules: s.commandRules.InDir(workingDir), pathRules: s.pathRules.InDir(workingDir)}
}
